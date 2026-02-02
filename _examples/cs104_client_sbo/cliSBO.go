package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/marrasen/go-iecp5/asdu"
	"github.com/marrasen/go-iecp5/clog"
	"github.com/marrasen/go-iecp5/cs104"
)

// sboClient demonstrates the IEC 60870-5-104 Select-Before-Operate (SBO) sequence
// using single commands (C_SC_NA_1). The same pattern applies to double commands
// (C_DC_NA_1) and set-point commands (C_SE_*):
//   1) Send SELECT (Qualifier.InSelect = true, Cause = Activation)
//   2) Wait for positive confirmation from the outstation
//   3) Send EXECUTE (Qualifier.InSelect = false, Cause = Activation)
//   4) Optionally wait for termination/negative confirmations
//
// Notes about SBO in this library:
// - The select/execute semantics are encoded in the qualifier (QOC/QOS) InSelect flag.
// - The CauseOfTransmission should typically be Activation for both select and execute.
// - This example listens to incoming ASDUs in the client handler and matches
//   confirmations for the command type and IOA to drive the flow.
// - For brevity, we showcase a single IOA selection/operation; real systems handle
//   multiple concurrent operations and richer error handling/timeouts.

type sboClient struct {
	// simple synchronizers for this demo
	selectAckCh  chan struct{}
	executeAckCh chan struct{}
}

func main() {
	// Configure remote endpoint (default: 127.0.0.1:2404). You can override via env.
	remote := os.Getenv("IEC104_REMOTE")
	if remote == "" {
		remote = "127.0.0.1:2404"
	}

	opt := cs104.NewOption()
	if err := opt.SetRemoteServer(remote); err != nil {
		panic(err)
	}

	cliHandler := &sboClient{
		selectAckCh:  make(chan struct{}, 1),
		executeAckCh: make(chan struct{}, 1),
	}
	client := cs104.NewClient(cliHandler, opt)
	client.SetLogLevel(clog.LevelError)

	client.SetConnStateHandler(func(c asdu.Connect, s cs104.ConnState) {
		switch s {
		case cs104.ConnStateNew:
			fmt.Println("Connected, sending StartDT_ACT...")
			c.(*cs104.Client).SendStartDt()
		case cs104.ConnStateActive:
			fmt.Println("Link activated. Demonstrating SBO (Select then Operate)...")

			// Define the common address (CA) and information object address (IOA)
			ca := asdu.CommonAddr(1)
			ioa := asdu.InfoObjAddr(1)

			// 1) SEND SELECT: build a single command with InSelect=true.
			//    Here we command Value=true (e.g., close circuit breaker) with a short pulse.
			selectQoc := asdu.QualifierOfCommand{Qual: asdu.QOCShortPulseDuration, InSelect: true}
			coa := asdu.CauseOfTransmission{Cause: asdu.Activation}
			fmt.Printf("Sending SELECT for IOA=%d CA=%d...\n", ioa, ca)
			if err := asdu.SingleCmd(c, asdu.C_SC_NA_1, coa, ca, asdu.SingleCommandInfo{
				Ioa:   ioa,
				Value: true,
				Qoc:   selectQoc,
			}); err != nil {
				fmt.Println("Failed to send SELECT:", err)
				return
			}

			// Wait for select confirmation (basic demo timeout)
			selectCtx, cancelSelect := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelSelect()
			select {
			case <-cliHandler.selectAckCh:
				fmt.Println("SELECT confirmed by outstation")
			case <-selectCtx.Done():
				fmt.Println("SELECT timed out waiting for confirmation")
				return
			}

			// 2) SEND EXECUTE: same command but InSelect=false.
			execQoc := asdu.QualifierOfCommand{Qual: asdu.QOCShortPulseDuration, InSelect: false}
			fmt.Printf("Sending EXECUTE for IOA=%d CA=%d...\n", ioa, ca)
			if err := asdu.SingleCmd(c, asdu.C_SC_NA_1, coa, ca, asdu.SingleCommandInfo{
				Ioa:   ioa,
				Value: true,
				Qoc:   execQoc,
			}); err != nil {
				fmt.Println("Failed to send EXECUTE:", err)
				return
			}

			// Wait for execute confirmation (basic demo timeout)
			execCtx, cancelExec := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelExec()
			select {
			case <-cliHandler.executeAckCh:
				fmt.Println("EXECUTE confirmed by outstation")
			case <-execCtx.Done():
				fmt.Println("EXECUTE timed out waiting for confirmation")
				return
			}
		}
	})

	// Run until Ctrl+C
	notifyCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()
	if err := client.Start(notifyCtx); err != nil {
		fmt.Println("Connection error:", err)
	} else {
		fmt.Println("Connection closed")
	}
}

// The following handler methods receive ASDUs back from the outstation.
// For SELECT/EXECUTE confirmations, outstations typically respond with the
// same type (e.g., C_SC_NA_1) and causes like Activation confirmation/termination.
// We parse the ASDU and signal our waiting goroutines.

func (s *sboClient) Handle(c asdu.Connect, msg asdu.Message) {
	switch m := msg.(type) {
	case *asdu.SingleCommandMsg:
		cause := m.Header().Identifier.Coa.Cause
		cmd := m.Cmd
		if cmd.Qoc.InSelect && cause == asdu.ActivationCon {
			select {
			case s.selectAckCh <- struct{}{}:
			default:
			}
			fmt.Printf("SELECT confirmation received: IOA=%d Value=%v\n", cmd.Ioa, cmd.Value)
		}
		if !cmd.Qoc.InSelect && cause == asdu.ActivationCon {
			select {
			case s.executeAckCh <- struct{}{}:
			default:
			}
			fmt.Printf("EXECUTE confirmation received: IOA=%d Value=%v\n", cmd.Ioa, cmd.Value)
		}
	}
}

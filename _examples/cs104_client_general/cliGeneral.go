package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/marrasen/go-iecp5/asdu"
	"github.com/marrasen/go-iecp5/clog"
	"github.com/marrasen/go-iecp5/cs104"
)

type myClient struct{}

func main() {
	var err error

	option := cs104.NewOption()
	if err = option.AddRemoteServer("172.22.27.81:2404"); err != nil {
		panic(err)
	}

	mycli := &myClient{}

	client := cs104.NewClient(mycli, option)

	client.SetLogLevel(clog.LevelError)

	client.SetConnStateHandler(func(c asdu.Connect, s cs104.ConnState) {
		switch s {
		case cs104.ConnStateNew:
			fmt.Println("Connected")
			c.(*cs104.Client).SendStartDt() // Send startDt activation command
		case cs104.ConnStateActive:
			fmt.Println("Activated, sending interrogation command")
			err := c.(*cs104.Client).InterrogationCmd(asdu.CauseOfTransmission{
				IsTest:     false,
				IsNegative: false,
				Cause:      asdu.Activation,
			}, asdu.CommonAddr(1), asdu.QOIStation)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
		}
	})

	notifyContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	err = client.Start(notifyContext)
	if err != nil {
		fmt.Printf("Failed to connect. error:%v\n", err)
	} else {
		fmt.Println("Connection closed")
	}

}

func (myClient) Handle(c asdu.Connect, msg asdu.Message) error {
	switch m := msg.(type) {
	case asdu.InterrogationCmdMsg:
		fmt.Printf("InterrogationCmd: %+v\n", m)
	case asdu.CounterInterrogationCmdMsg:
		fmt.Printf("CounterInterrogationCmd: %+v\n", m)
	case asdu.ReadCmdMsg:
		fmt.Printf("ReadCmd: %+v\n", m)
	case asdu.TestCmdMsg:
		fmt.Printf("TestCmd: %+v\n", m)
	case asdu.ClockSyncCmdMsg:
		fmt.Printf("ClockSyncCmd: %+v\n", m)
	case asdu.ResetProcessCmdMsg:
		fmt.Printf("ResetProcessCmd: %+v\n", m)
	case asdu.DelayAcquireCmdMsg:
		fmt.Printf("DelayAcquireCmd: %+v\n", m)
	default:
		fmt.Printf("ASDU: %+v\n", msg)
	}
	return nil
}

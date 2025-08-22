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

	client.SetOnConnectHandler(func(c *cs104.Client) {
		fmt.Println("Connected")
		c.SendStartDt() // Send startDt activation command
	})
	client.SetOnActivatedHandler(func(c *cs104.Client) {
		fmt.Println("Activated, sending interrogation command")
		err := c.InterrogationCmd(asdu.CauseOfTransmission{
			IsTest:     false,
			IsNegative: false,
			Cause:      asdu.Activation,
		}, asdu.CommonAddr(1), asdu.QOIStation)
		if err != nil {
			fmt.Println("Error:", err)
			return
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

func (myClient) InterrogationHandler(c asdu.Connect, a *asdu.ASDU) error {
	fmt.Printf("InterrogationHandler: %v\n", a)
	return nil
}

func (myClient) CounterInterrogationHandler(c asdu.Connect, a *asdu.ASDU) error {
	fmt.Printf("CounterInterrogationHandler: %v\n", a)
	return nil
}

func (myClient) ReadHandler(c asdu.Connect, a *asdu.ASDU) error {
	fmt.Printf("ReadHandler: %v\n", a)
	return nil
}

func (myClient) TestCommandHandler(c asdu.Connect, a *asdu.ASDU) error {
	fmt.Printf("TestCommandHandler: %v\n", a)
	return nil
}

func (myClient) ClockSyncHandler(c asdu.Connect, a *asdu.ASDU) error {
	fmt.Printf("ClockSyncHandler: %v\n", a)
	return nil
}

func (myClient) ResetProcessHandler(c asdu.Connect, a *asdu.ASDU) error {
	fmt.Printf("ResetProcessHandler: %v\n", a)
	return nil
}

func (myClient) DelayAcquisitionHandler(c asdu.Connect, a *asdu.ASDU) error {
	fmt.Printf("DelayAcquisitionHandler: %v\n", a)
	return nil
}

func (myClient) ASDUHandler(c asdu.Connect, a *asdu.ASDU) error {
	fmt.Printf("ASDUHandler: %v\n", a)
	return nil
}

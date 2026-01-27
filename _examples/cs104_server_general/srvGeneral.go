package main

import (
	"log"

	"github.com/marrasen/go-iecp5/asdu"
	"github.com/marrasen/go-iecp5/cs104"
)

func main() {
	srv := cs104.NewServer(&mysrv{})
	srv.SetOnConnectionHandler(func(c asdu.Connect) {
		log.Println("on connect")
	})
	srv.SetConnectionLostHandler(func(c asdu.Connect) {
		log.Println("connect lost")
	})
	srv.LogMode(true)
	// go func() {
	// 	time.Sleep(time.Second * 20)
	// 	log.Println("try ooooooo", err)
	// 	err := srv.Close()
	// 	log.Println("ooooooo", err)
	// }()
	srv.ListenAndServer(":2404")
}

type mysrv struct{}

func (sf *mysrv) Handle(c asdu.Connect, msg asdu.Message) error {
	switch m := msg.(type) {
	case asdu.InterrogationCmdMsg:
		log.Println("qoi", m.QOI)
		if mirror := m.Header().ASDU(); mirror != nil {
			_ = mirror.SendReplyMirror(c, asdu.ActivationCon)
		}
		_ = asdu.Single(c, false, asdu.CauseOfTransmission{Cause: asdu.InterrogatedByStation}, asdu.GlobalCommonAddr,
			asdu.SinglePointInfo{})
		// go func() {
		// 	for {
		// 		err := asdu.Single(c, false, asdu.CauseOfTransmission{Cause: asdu.Spontaneous}, asdu.GlobalCommonAddr,
		// 			asdu.SinglePointInfo{})
		// 		if err != nil {
		// 			log.Println("falied", err)
		// 		} else {
		// 			log.Println("success", err)
		// 		}
		//
		// 		time.Sleep(time.Second * 1)
		// 	}
		// }()
		if mirror := m.Header().ASDU(); mirror != nil {
			_ = mirror.SendReplyMirror(c, asdu.ActivationTerm)
		}
	}
	return nil
}

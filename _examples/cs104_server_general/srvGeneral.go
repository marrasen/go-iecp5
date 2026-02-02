package main

import (
	"log"

	"github.com/marrasen/go-iecp5/asdu"
	"github.com/marrasen/go-iecp5/cs104"
)

func main() {
	srv := cs104.NewServer(&mysrv{})
	srv.ConnState = func(c asdu.Connect, s cs104.ConnState) {
		log.Printf("conn state: %s", s)
	}
	// go func() {
	// 	time.Sleep(time.Second * 20)
	// 	log.Println("try ooooooo", err)
	// 	err := srv.Close()
	// 	log.Println("ooooooo", err)
	// }()
	_ = srv.ListenAndServe(":2404")
}

type mysrv struct{}

func (sf *mysrv) Handle(c asdu.Connect, msg asdu.Message) {
	switch m := msg.(type) {
	case *asdu.InterrogationCmdMsg:
		log.Println("qoi", m.QOI)
		if mirror := m.Header().ASDU(); mirror != nil {
			if err := mirror.SendReplyMirror(c, asdu.ActivationCon); err != nil {
				log.Printf("failed to send reply mirror: %v", err)
			}
		}
		if err := asdu.Single(c, false, asdu.CauseOfTransmission{Cause: asdu.InterrogatedByStation}, asdu.GlobalCommonAddr,
			asdu.SinglePointInfo{}); err != nil {
			log.Printf("failed to send single point: %v", err)
		}
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
			if err := mirror.SendReplyMirror(c, asdu.ActivationTerm); err != nil {
				log.Printf("failed to send reply mirror: %v", err)
			}
		}
	}
}

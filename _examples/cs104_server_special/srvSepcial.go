package main

import (
	"context"
	"log"
	"net/http"

	_ "net/http/pprof"

	"github.com/marrasen/go-iecp5/asdu"
	"github.com/marrasen/go-iecp5/cs104"
)

func main() {
	option := cs104.NewOption()
	err := option.SetRemoteServer("127.0.0.1:2404")
	if err != nil {
		panic(err)
	}

	srv := cs104.NewServerSpecial(&mysrv{}, option)

	srv.SetConnStateHandler(func(c asdu.Connect, s cs104.ConnState) {
		if s == cs104.ConnStateNew {
			_, _ = c.UnderlyingConn().Write([]byte{0x68, 0x0e, 0x00, 0x00, 0x00, 0x00, 0x46, 0x01, 0x04, 0x00, 0xa0, 0xaf, 0xbd, 0xd8, 0x0a, 0xf4})
		}
		log.Printf("conn state: %s", s)
	})
	if err = srv.Start(context.Background()); err != nil {
		panic(err)
	}

	if err := http.ListenAndServe(":6060", nil); err != nil {
		panic(err)
	}
}

type mysrv struct{}

func (sf *mysrv) Handle(c asdu.Connect, msg asdu.Message) error {
	switch m := msg.(type) {
	case asdu.InterrogationCmdMsg:
		log.Println("qoi", m.QOI)
		// mirror := m.Header().ASDU()
		// _ = mirror.SendReplyMirror(c, asdu.ActivationCon)
	}
	return nil
}

package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/marrasen/go-iecp5/asdu"
	"github.com/marrasen/go-iecp5/cs104"
)

type proxy struct {
	mu       sync.RWMutex
	outbound map[asdu.CommonAddr]*cs104.Client
	inbound  map[asdu.Connect]struct{}
	logger   *log.Logger
}

func newProxy(logger *log.Logger) *proxy {
	return &proxy{
		outbound: make(map[asdu.CommonAddr]*cs104.Client),
		inbound:  make(map[asdu.Connect]struct{}),
		logger:   logger,
	}
}

func (p *proxy) addInbound(c asdu.Connect) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.inbound[c] = struct{}{}
}

func (p *proxy) dropInbound(c asdu.Connect) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.inbound, c)
}

func (p *proxy) getAllInbound() []asdu.Connect {
	p.mu.RLock()
	defer p.mu.RUnlock()
	conns := make([]asdu.Connect, 0, len(p.inbound))
	for c := range p.inbound {
		conns = append(conns, c)
	}
	return conns
}

func (p *proxy) getOutbound(ca asdu.CommonAddr) *cs104.Client {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.outbound[ca]
}

type inboundHandler struct {
	proxy *proxy
}

func (h inboundHandler) Handle(c asdu.Connect, msg asdu.Message) {
	header := msg.Header()
	ca := header.Identifier.CommonAddr
	if ca == asdu.InvalidCommonAddr {
		if mirror := header.ASDU(); mirror != nil {
			if err := mirror.SendReplyMirror(c, asdu.UnknownCA); err != nil {
				h.proxy.logger.Printf("failed to send reply mirror: %v", err)
			}
		}
		return
	}

	if ca == asdu.GlobalCommonAddr {
		err := h.broadcast(header)
		if err != nil {
			h.proxy.logger.Printf("failed to broadcast: %v", err)
		}
		return
	}

	out := h.proxy.getOutbound(ca)
	if out == nil {
		if mirror := header.ASDU(); mirror != nil {
			if err := mirror.SendReplyMirror(c, asdu.UnknownCA); err != nil {
				h.proxy.logger.Printf("failed to send reply mirror: %v", err)
			}
		}
		return
	}
	outMsg := header.ASDU()
	if outMsg == nil {
		return
	}
	outMsg.Identifier.CommonAddr = ca
	if err := out.Send(outMsg); err != nil {
		h.proxy.logger.Printf("failed to send to outbound: %v", err)
	}
}

func (h inboundHandler) broadcast(header asdu.Header) error {
	outMsg := header.ASDU()
	if outMsg == nil {
		return errors.New("failed to build outbound asdu")
	}
	h.proxy.mu.RLock()
	outbounds := make(map[asdu.CommonAddr]*cs104.Client, len(h.proxy.outbound))
	for ca, out := range h.proxy.outbound {
		outbounds[ca] = out
	}
	h.proxy.mu.RUnlock()

	var firstErr error
	targetCount := 0
	for ca, out := range outbounds {
		cloned := outMsg.Clone()
		cloned.Identifier.CommonAddr = ca
		if err := out.Send(cloned); err != nil {
			if firstErr == nil {
				firstErr = err
			} else {
				targetCount++
			}
		}
	}
	h.proxy.logger.Printf("Broadcasted to %d target(s)", targetCount)
	return firstErr
}

type outboundHandler struct {
	logger *log.Logger
	proxy  *proxy
	ca     asdu.CommonAddr
}

func (h outboundHandler) Handle(_ asdu.Connect, msg asdu.Message) {
	h.logger.Printf("Received msg on ca %d. Message: %s", h.ca, msg.Header().ASDU().String())
	inbounds := h.proxy.getAllInbound()
	if len(inbounds) == 0 {
		h.logger.Print("No inbound connections available")
		return
	}
	inMsg := msg.Header().ASDU()
	if inMsg == nil {
		h.logger.Printf("Received msg on ca %d: failed to build inbound asdu, dropping message", h.ca)
		return
	}
	inMsg.Identifier.CommonAddr = h.ca
	for _, in := range inbounds {
		if err := in.Send(inMsg); err != nil {
			h.logger.Printf("Failed to send to inbound: %v", err)
		}
	}
	h.logger.Printf("Sent msg on ca %d to %d inbound connection(s)", h.ca, len(inbounds))
}

func main() {
	listenAddr := flag.String("listen", ":2404", "listen address for incoming IEC104 connections")
	remoteList := flag.String("remote", "", "comma-separated upstream servers (host:port)")
	flag.Parse()

	if *remoteList == "" {
		log.Fatal("missing -remote list")
	}

	logger := log.New(os.Stdout, "proxy: ", log.LstdFlags)
	p := newProxy(logger)

	remotes := strings.Split(*remoteList, ",")
	for i, raw := range remotes {
		remote := strings.TrimSpace(raw)
		if remote == "" {
			continue
		}
		ca := asdu.CommonAddr(i + 1)
		opt := cs104.NewOption()
		if err := opt.SetRemoteServer(remote); err != nil {
			log.Fatalf("invalid remote %q: %v", remote, err)
		}
		handler := outboundHandler{proxy: p, ca: ca, logger: logger}
		client := cs104.NewClient(handler, opt)
		client.SetConnStateHandler(func(c asdu.Connect, s cs104.ConnState) {
			switch s {
			case cs104.ConnStateNew:
				logger.Printf("outbound %s connected, sending StartDT_ACT...", remote)
				c.(*cs104.Client).SendStartDt()
			case cs104.ConnStateClosed:
				logger.Printf("outbound %s disconnected", remote)
			case cs104.ConnStateActive:
				logger.Printf("outbound %s connected", remote)
			case cs104.ConnStateIdle:
				logger.Printf("outbound %s idle", remote)
			}
		})
		p.outbound[ca] = client
		logger.Printf("mapped outbound %s -> CA=%d", remote, ca)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for ca, client := range p.outbound {
		go func(ca asdu.CommonAddr, cli *cs104.Client) {
			if err := cli.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
				logger.Printf("outbound CA=%d stopped: %v", ca, err)
			}
		}(ca, client)
	}

	server := cs104.NewServer(inboundHandler{proxy: p})
	server.ConnState = func(c asdu.Connect, s cs104.ConnState) {
		remoteAddr := c.UnderlyingConn().RemoteAddr().String()
		switch s {
		case cs104.ConnStateNew:
			logger.Printf("New inbound connection: %s", remoteAddr)
		case cs104.ConnStateActive:
			logger.Printf("Inbound connection active: %s", remoteAddr)
			p.addInbound(c)
		case cs104.ConnStateClosed:
			logger.Printf("Inbound connection closed: %s", remoteAddr)
			p.dropInbound(c)
		case cs104.ConnStateIdle:
			logger.Printf("Inbound connection idle: %s", remoteAddr)
		}
	}

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(*listenAddr); err != nil && !errors.Is(err, cs104.ErrServerClosed) {
		logger.Fatalf("listen failed: %v", err)
	}
}

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
	mu         sync.RWMutex
	upstream   map[asdu.CommonAddr]*cs104.Client
	downstream map[asdu.CommonAddr]asdu.Connect
	casByConn  map[asdu.Connect]map[asdu.CommonAddr]struct{}
	logger     *log.Logger
}

func newProxy(logger *log.Logger) *proxy {
	return &proxy{
		upstream:   make(map[asdu.CommonAddr]*cs104.Client),
		downstream: make(map[asdu.CommonAddr]asdu.Connect),
		casByConn:  make(map[asdu.Connect]map[asdu.CommonAddr]struct{}),
		logger:     logger,
	}
}

func (p *proxy) setDownstream(c asdu.Connect, ca asdu.CommonAddr) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.downstream[ca] = c
	if _, ok := p.casByConn[c]; !ok {
		p.casByConn[c] = make(map[asdu.CommonAddr]struct{})
	}
	p.casByConn[c][ca] = struct{}{}
}

func (p *proxy) dropDownstreamConn(c asdu.Connect) {
	p.mu.Lock()
	defer p.mu.Unlock()
	cas := p.casByConn[c]
	for ca := range cas {
		if cur, ok := p.downstream[ca]; ok && cur == c {
			delete(p.downstream, ca)
		}
	}
	delete(p.casByConn, c)
}

func (p *proxy) getDownstream(ca asdu.CommonAddr) asdu.Connect {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.downstream[ca]
}

func (p *proxy) getUpstream(ca asdu.CommonAddr) *cs104.Client {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.upstream[ca]
}

type inboundHandler struct {
	proxy *proxy
}

func (h inboundHandler) Handle(c asdu.Connect, msg asdu.Message) error {
	header := msg.Header()
	ca := header.Identifier.CommonAddr
	if ca == asdu.InvalidCommonAddr {
		if mirror := header.ASDU(); mirror != nil {
			return mirror.SendReplyMirror(c, asdu.UnknownCA)
		}
		return errors.New("invalid common address")
	}

	if ca == asdu.GlobalCommonAddr {
		return h.broadcast(c, header)
	}

	up := h.proxy.getUpstream(ca)
	if up == nil {
		if mirror := header.ASDU(); mirror != nil {
			return mirror.SendReplyMirror(c, asdu.UnknownCA)
		}
		return errors.New("unknown common address")
	}
	h.proxy.setDownstream(c, ca)
	out := header.ASDU()
	if out == nil {
		return errors.New("failed to build outbound asdu")
	}
	out.Identifier.CommonAddr = ca
	return up.Send(out)
}

func (h inboundHandler) broadcast(c asdu.Connect, header asdu.Header) error {
	out := header.ASDU()
	if out == nil {
		return errors.New("failed to build outbound asdu")
	}
	h.proxy.mu.RLock()
	upstreams := make(map[asdu.CommonAddr]*cs104.Client, len(h.proxy.upstream))
	for ca, up := range h.proxy.upstream {
		upstreams[ca] = up
	}
	h.proxy.mu.RUnlock()

	var firstErr error
	for ca, up := range upstreams {
		h.proxy.setDownstream(c, ca)
		cloned := out.Clone()
		cloned.Identifier.CommonAddr = ca
		if err := up.Send(cloned); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type upstreamHandler struct {
	proxy *proxy
	ca    asdu.CommonAddr
}

func (h upstreamHandler) Handle(c asdu.Connect, msg asdu.Message) error {
	down := h.proxy.getDownstream(h.ca)
	if down == nil {
		return nil
	}
	out := msg.Header().ASDU()
	if out == nil {
		return errors.New("failed to build outbound asdu")
	}
	out.Identifier.CommonAddr = h.ca
	return down.Send(out)
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
		if err := opt.AddRemoteServer(remote); err != nil {
			log.Fatalf("invalid remote %q: %v", remote, err)
		}
		handler := upstreamHandler{proxy: p, ca: ca}
		client := cs104.NewClient(handler, opt)
		client.SetConnStateHandler(func(c asdu.Connect, s cs104.ConnState) {
			if s == cs104.ConnStateNew {
				c.(*cs104.Client).SendStartDt()
			}
		})
		p.upstream[ca] = client
		logger.Printf("mapped upstream %s -> CA=%d", remote, ca)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for ca, client := range p.upstream {
		go func(ca asdu.CommonAddr, cli *cs104.Client) {
			if err := cli.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
				logger.Printf("upstream CA=%d stopped: %v", ca, err)
			}
		}(ca, client)
	}

	server := cs104.NewServer(inboundHandler{proxy: p})
	server.SetConnStateHandler(func(c asdu.Connect, s cs104.ConnState) {
		if s == cs104.ConnStateClosed {
			p.dropDownstreamConn(c)
		}
	})

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(*listenAddr); err != nil && !errors.Is(err, cs104.ErrServerClosed) {
		logger.Fatalf("listen failed: %v", err)
	}
}

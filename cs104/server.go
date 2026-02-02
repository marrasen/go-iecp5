// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package cs104

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/marrasen/go-iecp5/asdu"
	"github.com/marrasen/go-iecp5/clog"
)

// timeoutResolution is seconds according to companion standard 104,
// subclass 6.9, caption "Definition of time outs". However, then
// of a second make this system much more responsive i.c.w. S-frames.
const timeoutResolution = 100 * time.Millisecond

// Server the common server
type Server struct {
	config    Config
	params    asdu.Params
	handler   asdu.Handler
	ConnState func(asdu.Connect, ConnState)
	TLSConfig *tls.Config
	mux       sync.Mutex
	sessions  map[*SrvSession]struct{}
	listen    net.Listener
	clog.Clog
	wg      sync.WaitGroup
	closing uint32
}

// NewServer new a server, default config and default asdu.ParamsWide params
func NewServer(handler asdu.Handler) *Server {
	return &Server{
		config:   DefaultConfig(),
		params:   *asdu.ParamsWide,
		handler:  handler,
		sessions: make(map[*SrvSession]struct{}),
		Clog:     clog.NewLogger("cs104 server => "),
	}
}

// SetConfig set config if config is valid it will use DefaultConfig()
func (sf *Server) SetConfig(cfg Config) *Server {
	if err := cfg.Valid(); err != nil {
		sf.config = DefaultConfig()
	} else {
		sf.config = cfg
	}
	return sf
}

// SetParams set asdu params if params is valid it will use asdu.ParamsWide
func (sf *Server) SetParams(p *asdu.Params) *Server {
	if err := p.Valid(); err != nil {
		sf.params = *asdu.ParamsWide
	} else {
		sf.params = *p
	}
	return sf
}

// ListenAndServe runs the server until stopped or it fails.
func (sf *Server) ListenAndServe(addr string) error {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		sf.Error("server run failed, %v", err)
		return err
	}
	sf.mux.Lock()
	sf.listen = listen
	sf.mux.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		_ = sf.Close()
		sf.Debug("server stop")
	}()
	sf.Debug("server run")
	for {
		conn, err := listen.Accept()
		if err != nil {
			if atomic.LoadUint32(&sf.closing) != 0 {
				return ErrServerClosed
			}
			sf.Error("server run failed, %v", err)
			return err
		}

		sf.wg.Add(1)
		go func() {
			sess := &SrvSession{
				config:   &sf.config,
				params:   &sf.params,
				handler:  sf.handler,
				conn:     conn,
				rcvASDU:  make(chan []byte, sf.config.RecvUnAckLimitW<<4),
				sendASDU: make(chan []byte, sf.config.SendUnAckLimitK<<4),
				rcvRaw:   make(chan []byte, sf.config.RecvUnAckLimitW<<5),
				sendRaw:  make(chan []byte, sf.config.SendUnAckLimitK<<5), // may not block!

				connState: sf.ConnState,
				Clog:      sf.Clog,
			}
			sf.mux.Lock()
			sf.sessions[sess] = struct{}{}
			sf.mux.Unlock()
			sess.run(ctx)
			sf.mux.Lock()
			delete(sf.sessions, sess)
			sf.mux.Unlock()
			sf.wg.Done()
		}()
	}
}

// Close close the server
func (sf *Server) Close() error {
	atomic.StoreUint32(&sf.closing, 1)
	var err error

	sf.mux.Lock()
	if sf.listen != nil {
		err = sf.listen.Close()
		sf.listen = nil
	}
	sessions := make([]*SrvSession, 0, len(sf.sessions))
	for s := range sf.sessions {
		sessions = append(sessions, s)
	}
	sf.mux.Unlock()

	for _, s := range sessions {
		_ = s.Close()
	}
	return err
}

// Shutdown gracefully stops the server and waits for sessions to close.
func (sf *Server) Shutdown(ctx context.Context) error {
	if err := sf.Close(); err != nil {
		return err
	}
	done := make(chan struct{})
	go func() {
		sf.wg.Wait()
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// Send imp interface Connect
func (sf *Server) Send(a *asdu.ASDU) error {
	sf.mux.Lock()
	for k := range sf.sessions {
		_ = k.Send(a.Clone())
	}
	sf.mux.Unlock()
	return nil
}

// Params imp interface Connect
func (sf *Server) Params() *asdu.Params { return &sf.params }

// SetInfoObjTimeZone set info object time zone
func (sf *Server) SetInfoObjTimeZone(zone *time.Location) {
	sf.params.InfoObjTimeZone = zone
}

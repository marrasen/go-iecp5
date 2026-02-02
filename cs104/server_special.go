// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package cs104

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/marrasen/go-iecp5/asdu"
	"github.com/marrasen/go-iecp5/clog"
)

// ServerSpecial server special interface
type ServerSpecial interface {
	asdu.Connect

	IsConnected() bool
	IsClosed() bool
	Start(ctx context.Context) error
	Close() error
	SetConnStateHandler(f func(c asdu.Connect, s ConnState))

	SetLogLevel(level clog.Level)
	SetLogProvider(p clog.LogProvider)
}

type serverSpec struct {
	SrvSession
	option      ClientOption
	closeCancel context.CancelFunc
}

// NewServerSpecial new special server
func NewServerSpecial(handler asdu.Handler, o *ClientOption) ServerSpecial {
	return &serverSpec{
		SrvSession: SrvSession{
			config:  &o.config,
			params:  &o.params,
			handler: handler,

			rcvASDU:  make(chan []byte, 1024),
			sendASDU: make(chan []byte, 1024),
			rcvRaw:   make(chan []byte, 1024),
			sendRaw:  make(chan []byte, 1024), // may not block!

			Clog: clog.NewLogger("cs104 serverSpec => "),
		},
		option: *o,
	}
}

// SetConnStateHandler sets the connection lifecycle handler.
func (sf *serverSpec) SetConnStateHandler(f func(c asdu.Connect, s ConnState)) {
	sf.connState = f
}

// Start begins the server's operation and establishes a connection to the remote server if specified.
func (sf *serverSpec) Start(ctx context.Context) error {
	sf.rwMux.Lock()
	if !atomic.CompareAndSwapUint32(&sf.status, initial, disconnected) {
		sf.rwMux.Unlock()
		return errors.New("server already started")
	}
	ctx, sf.closeCancel = context.WithCancel(context.Background())
	sf.rwMux.Unlock()
	defer sf.setConnectStatus(initial)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	sf.Debug("connecting server %+v", sf.option.server)
	conn, err := openConnection(ctx, sf.option.server, sf.option.TLSConfig, sf.config.ConnectTimeout0, nil)
	if err != nil {
		sf.Error("connect failed, %v", err)
		return err
	}
	sf.Debug("connect success")
	sf.conn = conn
	err = sf.run(ctx)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		sf.Debug("disconnected, %v", err)
	} else {
		sf.Error("run failed, %v", err)
	}
	sf.Debug("disconnected server %+v", sf.option.server)
	return err
}

func (sf *serverSpec) IsClosed() bool {
	return sf.connectStatus() == initial
}

func (sf *serverSpec) Close() error {
	sf.rwMux.Lock()
	if sf.closeCancel != nil {
		sf.closeCancel()
	}
	sf.rwMux.Unlock()
	return nil
}

// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/url"
	"time"
)

// DefaultReconnectInterval defined default value
const DefaultReconnectInterval = 1 * time.Minute

type seqPending struct {
	seq      uint16
	sendTime time.Time
}

func openConnection(ctx context.Context, uri *url.URL, tlsc *tls.Config, timeout time.Duration, dialCtx func(ctx context.Context, network, address string) (net.Conn, error)) (net.Conn, error) {
	if uri == nil {
		return nil, errors.New("nil uri")
	}
	addr := uri.Host
	if addr == "" {
		return nil, errors.New("empty host")
	}
	// default dialer
	if dialCtx == nil {
		d := &net.Dialer{Timeout: timeout}
		dialCtx = d.DialContext
	}
	switch uri.Scheme {
	case "tcp":
		return dialCtx(ctx, "tcp", addr)
	case "ssl", "tls", "tcps":
		// Use provided dialer to establish the underlying TCP connection, then wrap with TLS
		rawConn, err := dialCtx(ctx, "tcp", addr)
		if err != nil {
			return nil, err
		}
		if tlsc == nil {
			tlsc = &tls.Config{}
		}
		// Set handshake timeout via deadline on the raw connection
		_ = rawConn.SetDeadline(time.Now().Add(timeout))
		tlsConn := tls.Client(rawConn, tlsc)
		if err := tlsConn.Handshake(); err != nil {
			_ = rawConn.Close()
			return nil, err
		}
		// Clear deadline after successful handshake
		_ = rawConn.SetDeadline(time.Time{})
		return tlsConn, nil
	}
	return nil, errors.New("unknown protocol")
}

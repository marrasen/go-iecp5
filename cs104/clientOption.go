// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/marrasen/go-iecp5/asdu"
)

// ClientOption client configuration
type ClientOption struct {
	config            Config
	params            asdu.Params
	server            *url.URL      // Connected server endpoint
	autoReconnect     bool          // Enable auto reconnect
	reconnectInterval time.Duration // Reconnection interval
	TLSConfig         *tls.Config   // TLS configuration
	// DialContext allows providing a custom dialer (e.g., SSH jump). If nil, net.Dialer is used.
	DialContext func(ctx context.Context, network, address string) (net.Conn, error)
}

// NewOption with default config and default asdu.ParamsWide params
func NewOption() *ClientOption {
	return &ClientOption{
		DefaultConfig(),
		*asdu.ParamsWide,
		nil,
		true,
		DefaultReconnectInterval,
		nil,
		nil,
	}
}

// SetConfig set config if config is valid it will use DefaultConfig()
func (sf *ClientOption) SetConfig(cfg Config) *ClientOption {
	if err := cfg.Valid(); err != nil {
		sf.config = DefaultConfig()
	} else {
		sf.config = cfg
	}
	return sf
}

// SetParams set asdu params if params is valid it will use asdu.ParamsWide
func (sf *ClientOption) SetParams(p *asdu.Params) *ClientOption {
	if err := p.Valid(); err != nil {
		sf.params = *asdu.ParamsWide
	} else {
		sf.params = *p
	}
	return sf
}

// SetReconnectInterval set tcp  reconnect the host interval when connect failed after try
func (sf *ClientOption) SetReconnectInterval(t time.Duration) *ClientOption {
	if t > 0 {
		sf.reconnectInterval = t
	}
	return sf
}

// SetAutoReconnect enable auto reconnect
func (sf *ClientOption) SetAutoReconnect(b bool) *ClientOption {
	sf.autoReconnect = b
	return sf
}

// SetTLSConfig set tls config
func (sf *ClientOption) SetTLSConfig(t *tls.Config) *ClientOption {
	sf.TLSConfig = t
	return sf
}

// SetDialContext sets a custom dialer function used to establish TCP connections (e.g., SSH jump).
func (sf *ClientOption) SetDialContext(dial func(ctx context.Context, network, address string) (net.Conn, error)) *ClientOption {
	sf.DialContext = dial
	return sf
}

// AddRemoteServer adds a broker URI to the list of brokers to be used.
// The format should be scheme://host:port
// Default values for hostname are "127.0.0.1", for schema is "tcp://".
// An example broker URI would look like: tcp://foobar.com:1204
func (sf *ClientOption) AddRemoteServer(server string) error {
	if len(server) > 0 && server[0] == ':' {
		server = "127.0.0.1" + server
	}
	if !strings.Contains(server, "://") {
		server = "tcp://" + server
	}
	remoteURL, err := url.Parse(server)
	if err != nil {
		return err
	}
	sf.server = remoteURL
	return nil
}

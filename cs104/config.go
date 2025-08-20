// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package cs104

import (
	"errors"
	"time"
)

const (
	// Port is the IANA registered port number for unsecure connection.
	Port = 2404

	// PortSecure is the IANA registered port number for secure connection.
	PortSecure = 19998
)

// defines an IEC 60870-5-104 configuration range
const (
	// "t₀" range [1, 255]s, default 30s
	ConnectTimeout0Min = 1 * time.Second
	ConnectTimeout0Max = 255 * time.Second

	// "t₁" range [1, 255]s, default 15s. See IEC 60870-5-104, figure 18.
	SendUnAckTimeout1Min = 1 * time.Second
	SendUnAckTimeout1Max = 255 * time.Second

	// "t₂" range [1, 255]s, default 10s. See IEC 60870-5-104, figure 10.
	RecvUnAckTimeout2Min = 1 * time.Second
	RecvUnAckTimeout2Max = 255 * time.Second

	// "t₃" range [1 second, 48 hours], default 20s. See IEC 60870-5-104, subclass 5.2.
	IdleTimeout3Min = 1 * time.Second
	IdleTimeout3Max = 48 * time.Hour

	// "k" range [1, 32767], default 12. See IEC 60870-5-104, subclass 5.5.
	SendUnAckLimitKMin = 1
	SendUnAckLimitKMax = 32767

	// "w" range [1, 32767], default 8. See IEC 60870-5-104, subclass 5.5.
	RecvUnAckLimitWMin = 1
	RecvUnAckLimitWMax = 32767
)

// Config defines an IEC 60870-5-104 configuration.
// The default is applied for each unspecified value.
type Config struct {
	// Maximum timeout for establishing a TCP connection
	// "t₀" range [1, 255]s, default 30s.
	ConnectTimeout0 time.Duration

	// Maximum number of unacknowledged I-frames that can be sent; once reached, transmission stops
	// "k" range [1, 32767], default 12.
	// See IEC 60870-5-104, subclass 5.5.
	SendUnAckLimitK uint16

	// Maximum timeout waiting for confirmation of received frames; if exceeded, the connection is closed immediately
	// "t₁" range [1, 255]s, default 15s.
	// See IEC 60870-5-104, figure 18.
	SendUnAckTimeout1 time.Duration

	// The receiver must acknowledge at latest after receiving w I-frames (ASDUs). w must not exceed 2/3 of k (2/3 SendUnAckLimitK)
	// "w" range [1, 32767], default 8.
	// See IEC 60870-5-104, subclass 5.5.
	RecvUnAckLimitW uint16

	// Maximum time before sending a receive acknowledgement; in practice this stack replies within 1 second
	// "t₂" range [1, 255]s, default 10s
	// See IEC 60870-5-104, figure 10.
	RecvUnAckTimeout2 time.Duration

	// Idle time that triggers the "TESTFR" keep-alive
	// "t₃" range [1 second, 48 hours], default 20s
	// See IEC 60870-5-104, subclass 5.2.
	IdleTimeout3 time.Duration
}

// Valid applies the default (defined by IEC) for each unspecified value.
func (sf *Config) Valid() error {
	if sf == nil {
		return errors.New("invalid pointer")
	}

	if sf.ConnectTimeout0 == 0 {
		sf.ConnectTimeout0 = 30 * time.Second
	} else if sf.ConnectTimeout0 < ConnectTimeout0Min || sf.ConnectTimeout0 > ConnectTimeout0Max {
		return errors.New(`ConnectTimeout0 "t₀" not in [1, 255]s`)
	}

	if sf.SendUnAckLimitK == 0 {
		sf.SendUnAckLimitK = 12
	} else if sf.SendUnAckLimitK < SendUnAckLimitKMin || sf.SendUnAckLimitK > SendUnAckLimitKMax {
		return errors.New(`SendUnAckLimitK "k" not in [1, 32767]`)
	}

	if sf.SendUnAckTimeout1 == 0 {
		sf.SendUnAckTimeout1 = 15 * time.Second
	} else if sf.SendUnAckTimeout1 < SendUnAckTimeout1Min || sf.SendUnAckTimeout1 > SendUnAckTimeout1Max {
		return errors.New(`SendUnAckTimeout1 "t₁" not in [1, 255]s`)
	}

	if sf.RecvUnAckLimitW == 0 {
		sf.RecvUnAckLimitW = 8
	} else if sf.RecvUnAckLimitW < RecvUnAckLimitWMin || sf.RecvUnAckLimitW > RecvUnAckLimitWMax {
		return errors.New(`RecvUnAckLimitW "w" not in [1, 32767]`)
	}

	if sf.RecvUnAckTimeout2 == 0 {
		sf.RecvUnAckTimeout2 = 10 * time.Second
	} else if sf.RecvUnAckTimeout2 < RecvUnAckTimeout2Min || sf.RecvUnAckTimeout2 > RecvUnAckTimeout2Max {
		return errors.New(`RecvUnAckTimeout2 "t₂" not in [1, 255]s`)
	}

	if sf.IdleTimeout3 == 0 {
		sf.IdleTimeout3 = 20 * time.Second
	} else if sf.IdleTimeout3 < IdleTimeout3Min || sf.IdleTimeout3 > IdleTimeout3Max {
		return errors.New(`IdleTimeout3 "t₃" not in [1 second, 48 hours]`)
	}

	return nil
}

// DefaultConfig default config
func DefaultConfig() Config {
	return Config{
		30 * time.Second,
		12,
		15 * time.Second,
		8,
		10 * time.Second,
		20 * time.Second,
	}
}

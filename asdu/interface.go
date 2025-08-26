// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package asdu

import (
	"net"
)

// Connect interface
type Connect interface {
	Params() *Params
	Send(a *ASDU) error
	UnderlyingConn() net.Conn
}

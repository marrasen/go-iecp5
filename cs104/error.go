// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package cs104

import (
	"errors"
)

// error defined
var (
	ErrUseClosedConnection = errors.New("use of closed connection")
	ErrBufferFulled        = errors.New("buffer is full")
	ErrNotActive           = errors.New("server is not active")
)

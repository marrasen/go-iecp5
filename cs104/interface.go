// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package cs104

import "github.com/marrasen/go-iecp5/asdu"

// Handler processes parsed ASDUs using type assertions.
type Handler interface {
	Handle(asdu.Connect, asdu.Message) error
}

# go-iecp5

Note about this repository
- This repository is a fork of https://github.com/thinkgos/go-iecp5.
- The original upstream repository has been archived by its author and is no longer maintained.
- The original README contained the following Chinese notice: "已归档, 不再维护, 放弃License. 有需要的可以自由分发". Translation: "Archived, no longer maintained, license abandoned. If needed, you can freely redistribute."

go-iecp5 library for IEC 60870-5 based protocols in pure Go.
The current implementation contains code for IEC 60870-5-104 (protocol over TCP/IP) specifications.



[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/thinkgos/go-iecp5?tab=doc)
[![Tests](https://github.com/thinkgos/go-iecp5/actions/workflows/ci.yml/badge.svg)](https://github.com/thinkgos/go-iecp5/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/thinkgos/go-iecp5/branch/master/graph/badge.svg)](https://codecov.io/gh/thinkgos/go-iecp5)
[![Go Report Card](https://goreportcard.com/badge/github.com/thinkgos/go-iecp5)](https://goreportcard.com/report/github.com/thinkgos/go-iecp5)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Tag](https://img.shields.io/github/v/tag/thinkgos/go-iecp5)](https://github.com/thinkgos/go-iecp5/tags)
[![Sourcegraph](https://sourcegraph.com/github.com/thinkgos/go-iecp5/-/badge.svg)](https://sourcegraph.com/github.com/thinkgos/go-iecp5?badge)


asdu package: [![GoDoc](https://godoc.org/github.com/thinkgos/go-iecp5/asdu?status.svg)](https://godoc.org/github.com/thinkgos/go-iecp5/asdu)  
clog package: [![GoDoc](https://godoc.org/github.com/thinkgos/go-iecp5/clog?status.svg)](https://godoc.org/github.com/thinkgos/go-iecp5/clog)  
cs104 package: [![GoDoc](https://godoc.org/github.com/thinkgos/go-iecp5/cs104?status.svg)](https://godoc.org/github.com/thinkgos/go-iecp5/cs104)  

## License

This fork adopts the MIT License.

- Background: the original upstream repository (thinkgos/go-iecp5) was archived with the notice: "已归档, 不再维护, 放弃License. 有需要的可以自由分发" ("Archived, no longer maintained, license abandoned. If needed, you can freely redistribute.").
- Change: this fork changes the licensing of the code present here to MIT to simplify reuse and clarify terms.
- Implementation details:
  - All source files include SPDX-License-Identifier: MIT headers.
  - The root LICENSE file contains the full MIT License text.

See the LICENSE file in the repository root for the full text and terms.

## Feature:

- client/server for CS 104 TCP/IP communication
- support for much application layer (except file object) message types,

## Handler API (cs104)

All inbound ASDUs are parsed once and delivered to a single handler. Use type assertions
to access the specific message payloads.

```go
type Handler interface {
	Handle(asdu.Connect, asdu.Message) error
}

type myHandler struct{}

func (myHandler) Handle(c asdu.Connect, msg asdu.Message) error {
	switch m := msg.(type) {
	case asdu.InterrogationCmdMsg:
		if mirror := m.Header().ASDU(); mirror != nil {
			_ = mirror.SendReplyMirror(c, asdu.ActivationCon)
		}
	default:
		// handle other message types
	}
	return nil
}
```

## Connection lifecycle (cs104)

Use a ConnState callback for connection lifecycle events and `ListenAndServe`/`Shutdown` for server
lifecycle control.

```go
srv := cs104.NewServer(&myHandler{})
srv.SetConnStateHandler(func(c asdu.Connect, s cs104.ConnState) {
	log.Printf("conn state: %s", s)
})

go func() {
	if err := srv.ListenAndServe(":2404"); err != nil && !errors.Is(err, cs104.ErrServerClosed) {
		log.Printf("listen failed: %v", err)
	}
}()

// Later...
_ = srv.Shutdown(context.Background())
```

Clients use the same ConnState mechanism:

```go
cli := cs104.NewClient(&myHandler{}, option)
cli.SetConnStateHandler(func(c asdu.Connect, s cs104.ConnState) {
	if s == cs104.ConnStateNew {
		c.(*cs104.Client).SendStartDt()
	}
})
```

# Reference
lib60870 C library [lib60870](https://github.com/mz-automation/lib60870)  
lib60870 C library docs [lib60870 doc](https://support.mz-automation.de/doc/lib60870/latest/group__CS104__MASTER.html)

## Donation

If this package helps you a lot, you can support the original author:

**Alipay**

![alipay](https://github.com/thinkgos/thinkgos/blob/master/asserts/alipay.jpg)

**WeChat Pay**

![wxpay](https://github.com/thinkgos/thinkgos/blob/master/asserts/wxpay.jpg)

# go-iecp5

IEC 60870-5-104 client and server in pure Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/marrasen/go-iecp5.svg)](https://pkg.go.dev/github.com/marrasen/go-iecp5)
[![Tests](https://github.com/marrasen/go-iecp5/actions/workflows/go.yml/badge.svg)](https://github.com/marrasen/go-iecp5/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Tag](https://img.shields.io/github/v/tag/marrasen/go-iecp5)](https://github.com/marrasen/go-iecp5/tags)

IEC 60870-5-104 (often called IEC 104) is the TCP/IP transport used by SCADA
systems to talk to substations and other outstations. This library implements
the transport layer (APCI) and the application layer (ASDU) for that protocol.
You can use it to build a controlling station (client), a controlled station
(server), or something in between such as a proxy.

## Features

- Client and server for IEC 60870-5-104 over TCP.
- Optional TLS on both client and server.
- Custom dialer on the client, for example to tunnel through SSH.
- All common ASDU types in both directions: monitoring, commands, parameters
  and system commands. File transfer is not implemented.
- Inbound ASDUs are parsed once into typed Go structs. Your handler gets a
  typed message and picks it apart with a type switch.
- Outbound ASDUs are built with plain functions, one per type.
- Every ASDU and parsed message has a readable `String()` form.
- ASDUs serialise to JSON.
- Static metadata for every type ID: name, direction, time tag format and
  information object size.
- Connection lifecycle callbacks, context-based cancellation and graceful
  server shutdown.
- Pluggable, levelled logging.

## Install

```
go get github.com/marrasen/go-iecp5
```

Requires Go 1.25 or newer.

## Packages

| Package | What it holds |
| --- | --- |
| `cs104` | The IEC 104 client, server and connection handling. |
| `asdu` | ASDU types, parsing, building and sending. |
| `clog` | The small logging interface used by the other packages. |
| `cs101` | Placeholder for IEC 60870-5-101. Only FT1.2 frame constants exist. |

## The handler

Both client and server deliver every inbound ASDU to one handler:

```go
type Handler interface {
	Handle(asdu.Connect, asdu.Message)
}
```

The `asdu.Connect` is the connection the message arrived on. Use it to send
replies. The `asdu.Message` is already parsed. Assert on the concrete type to
get at the payload:

```go
type handler struct{}

func (handler) Handle(c asdu.Connect, msg asdu.Message) {
	switch m := msg.(type) {
	case *asdu.SinglePointMsg:
		for _, item := range m.Items {
			log.Printf("IOA %d = %v (%s)", item.Ioa, item.Value, item.Qds)
		}
	case *asdu.MeasuredValueFloatMsg:
		for _, item := range m.Items {
			log.Printf("IOA %d = %f", item.Ioa, item.Value)
		}
	case *asdu.UnknownMsg:
		log.Printf("unsupported type %s", m.TypeID())
	default:
		log.Printf("%s", msg)
	}
}
```

Message types are named after the ASDU family, not the individual type ID.
`SinglePointMsg` covers `M_SP_NA_1`, `M_SP_TA_1` and `M_SP_TB_1`. Check
`msg.TypeID()` or the `Time` field on each item when the difference matters.

Every message carries a header with the identifier and the raw payload.
`m.Header().ASDU()` rebuilds the original ASDU. That is how you send a
mirrored reply:

```go
case *asdu.InterrogationCmdMsg:
	mirror := m.Header().ASDU()
	_ = mirror.SendReplyMirror(c, asdu.ActivationCon)
	// send the interrogated data here
	_ = mirror.SendReplyMirror(c, asdu.ActivationTerm)
```

## Client

```go
opt := cs104.NewOption()
if err := opt.SetRemoteServer("10.0.0.5:2404"); err != nil {
	log.Fatal(err)
}

cli := cs104.NewClient(handler{}, opt)
cli.SetLogLevel(clog.LevelWarn)

cli.SetConnStateHandler(func(c asdu.Connect, s cs104.ConnState) {
	switch s {
	case cs104.ConnStateNew:
		// TCP connection is up. Ask the server to start data transfer.
		c.(*cs104.Client).SendStartDt()
	case cs104.ConnStateActive:
		// Data transfer is active. Ask for everything.
		coa := asdu.CauseOfTransmission{Cause: asdu.Activation}
		_ = c.(*cs104.Client).InterrogationCmd(coa, asdu.CommonAddr(1), asdu.QOIStation)
	}
})

ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
defer stop()

// Start blocks until the connection drops or the context is cancelled.
if err := cli.Start(ctx); err != nil {
	log.Println(err)
}
```

`Start` makes one connection attempt and returns when that connection ends.
It does not reconnect. Wrap it in a loop if you want retries.

The client has convenience methods for the system commands:
`InterrogationCmd`, `CounterInterrogationCmd`, `ReadCmd`,
`ClockSynchronizationCmd`, `ResetProcessCmd`, `DelayAcquireCommand` and
`TestCommand`. Everything else is sent with the functions in the `asdu`
package, described below.

### TLS and custom dialers

Use a `tls://` address and give the option a TLS config:

```go
_ = opt.SetRemoteServer("tls://10.0.0.5:19998")
opt.SetTLSConfig(&tls.Config{ /* ... */ })
```

For an SSH tunnel or any other transport, set a dialer. The library calls it
to get the raw TCP connection and wraps it in TLS if the scheme asks for it:

```go
opt.SetDialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
	return sshClient.Dial(network, addr)
})
```

## Server

```go
srv := cs104.NewServer(handler{})
srv.SetLogLevel(clog.LevelWarn)
srv.ConnState = func(c asdu.Connect, s cs104.ConnState) {
	log.Printf("conn %s: %s", c.UnderlyingConn().RemoteAddr(), s)
}

go func() {
	err := srv.ListenAndServe(":2404")
	if err != nil && !errors.Is(err, cs104.ErrServerClosed) {
		log.Fatal(err)
	}
}()

// Later, on shutdown:
_ = srv.Shutdown(context.Background())
```

`Serve` accepts a `net.Listener` if you want to control the socket yourself.
`Shutdown` stops accepting, closes every session and waits for them to
finish. `Close` does the same without waiting.

The server validates system commands before they reach your handler. An
interrogation with the wrong cause of transmission, common address or
information object address gets a negative mirror reply and is dropped. Test
commands are answered by the server itself. Everything else goes to your
handler, which is responsible for the confirmation and termination replies.

Sending from the server side uses the connection passed to the handler, or
`srv.Send`, which broadcasts to every active session.

### Server TLS

```go
srv.SetTLSConfig(&tls.Config{
	Certificates: []tls.Certificate{cert},
	ClientAuth:   tls.RequireAndVerifyClientCert,
	ClientCAs:    pool,
})
```

The handshake runs before a session is admitted. A client that fails it never
reaches the handler. Inside the handler, the connection can be asserted to
`*cs104.SrvSession` and `PeerCertificates()` returns the client chain.

### Reverse-connecting server

`cs104.NewServerSpecial` builds a server that dials out to a fixed peer
instead of listening. It takes the same option as the client. This covers
setups where the outstation must open the connection.

## Sending ASDUs

The `asdu` package has one send function per ASDU family. Each takes the
connection, the cause of transmission, the common address and the payload.

Monitoring direction, for a server:

```go
coa := asdu.CauseOfTransmission{Cause: asdu.Spontaneous}
ca := asdu.CommonAddr(1)

_ = asdu.Single(c, false, coa, ca, asdu.SinglePointInfo{Ioa: 100, Value: true, Qds: asdu.QDSGood})
_ = asdu.MeasuredValueFloatCP56Time2a(c, coa, ca, asdu.MeasuredValueFloatInfo{
	Ioa:   200,
	Value: 230.5,
	Qds:   asdu.QDSGood,
	Time:  time.Now(),
})
```

Control direction, for a client. The type ID picks the variant with or without
a time tag:

```go
coa := asdu.CauseOfTransmission{Cause: asdu.Activation}

_ = asdu.SingleCmd(c, asdu.C_SC_NA_1, coa, ca, asdu.SingleCommandInfo{
	Ioa:   100,
	Value: true,
	Qoc:   asdu.QualifierOfCommand{Qual: asdu.QOCShortPulseDuration},
})
_ = asdu.SetpointCmdFloat(c, asdu.C_SE_NC_1, coa, ca, asdu.SetpointCommandFloatInfo{
	Ioa:   300,
	Value: 42.0,
})
```

Select-before-operate is expressed with the `InSelect` flag on the qualifier.
Send once with it set, wait for the confirmation, then send again with it
cleared. See `_examples/cs104_client_sbo` for a complete run-through.

You can also build an ASDU from a parsed message with `asdu.EncodeMessage`,
or work on the raw ASDU with `asdu.NewASDU` and `Send`.

## Type metadata

`TypeID.Info()` returns static facts about any type ID:

```go
info := asdu.M_ME_TF_1.Info()
// info.Name           "M_ME_TF_1"
// info.Description    "measured value, short floating point number with time tag CP56Time2a"
// info.Direction      Monitor
// info.TimeTagFormat  CP56Time2a
// info.InfoObjectSize 12
```

## Logging

Client and server embed a logger. It is off by default.

```go
cli.SetLogLevel(clog.LevelDebug)
cli.SetLogProvider(myLogger) // anything that implements clog.LogProvider
```

Levels are `LevelOff`, `LevelCritical`, `LevelError`, `LevelWarn` and
`LevelDebug`. Debug prints every frame.

## Examples

The `_examples` folder has runnable programs:

- `cs104_client_general`: connect, activate and interrogate.
- `cs104_client_sbo`: select-before-operate command sequence.
- `cs104_server_general`: answer interrogations with data.
- `cs104_server_special`: reverse-connecting server.
- `cs104_proxy`: route between one inbound side and several outbound
  servers by common address.

## References

- [IEC 60870-5-104 on Wikipedia](https://en.wikipedia.org/wiki/IEC_60870-5#IEC_60870-5-104)
- [lib60870](https://github.com/mz-automation/lib60870), a C implementation
  that is useful for cross-checking behaviour.

## License

MIT. See [LICENSE](LICENSE).

## Origin

This library started as a fork of
[thinkgos/go-iecp5](https://github.com/thinkgos/go-iecp5). The upstream
project has been archived by its author and is no longer maintained. Its
archive notice released the code without a licence, and this fork relicensed
it under MIT. Since then the API has been rewritten and the two are no longer
compatible. Code written against the original will not build against this
library without changes.

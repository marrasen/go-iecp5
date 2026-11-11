package cs104

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/marrasen/go-iecp5/asdu"
)

// badASDU has a type id no table knows, so it cannot be decoded.
var badASDU = []byte{0x7f, 0x01, 0x03, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00}

// iFrame wraps an ASDU in an I-frame with send and receive numbers 0.
func iFrame(a []byte) []byte {
	return append([]byte{startFrame, byte(4 + len(a)), 0, 0, 0, 0}, a...)
}

func waitRx(t *testing.T, ch <-chan RxError) RxError {
	t.Helper()
	select {
	case e := <-ch:
		return e
	case <-time.After(5 * time.Second):
		t.Fatal("no receive error reported")
		return RxError{}
	}
}

func TestClientReportsDroppedFrames(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	// A raw server: confirm STARTDT, then send an undecodable ASDU and an unknown U-frame.
	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 6)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return
		}
		conn.Write([]byte{startFrame, 4, uStartDtConfirm | 0x03, 0, 0, 0})
		conn.Write(iFrame(badASDU))
		conn.Write([]byte{startFrame, 4, 0xff, 0, 0, 0})
		time.Sleep(2 * time.Second)
	}()

	got := make(chan RxError, 4)
	opt := NewOption()
	if err := opt.SetRemoteServer(l.Addr().String()); err != nil {
		t.Fatal(err)
	}
	opt.SetRxErrorHandler(func(_ asdu.Connect, e RxError) { got <- e })
	c := NewClient(&captureHandler{}, opt)
	c.SetConnStateHandler(func(_ asdu.Connect, s ConnState) {
		if s == ConnStateNew {
			c.SendStartDt()
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.Start(ctx)

	// The ASDU and the U-frame are handled on different goroutines, so either may come first.
	seen := map[RxErrorKind]RxError{}
	for range 2 {
		e := waitRx(t, got)
		seen[e.Kind] = e
	}
	if e := seen[RxDecode]; string(e.Raw) != string(badASDU) || e.Err == nil || e.ASDU != nil {
		t.Fatalf("unexpected decode report %+v", e)
	}
	if e := seen[RxUFrame]; len(e.Raw) != 6 || e.Err == nil {
		t.Fatalf("unexpected U-frame report %+v", e)
	}
}

func TestServerReportsIFrameBeforeStartDt(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	got := make(chan RxError, 4)
	srv := NewServer(&captureHandler{})
	srv.SetRxErrorHandler(func(_ asdu.Connect, e RxError) { got <- e })
	go srv.Serve(l)
	defer srv.Close()

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.Write(iFrame(badASDU))

	if e := waitRx(t, got); e.Kind != RxInactive || e.Raw[0] != startFrame {
		t.Fatalf("unexpected report %+v", e)
	}
}

func TestServerReportsDroppedFrames(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	got := make(chan RxError, 4)
	srv := NewServer(&captureHandler{})
	srv.SetRxErrorHandler(func(_ asdu.Connect, e RxError) { got <- e })
	go srv.Serve(l)
	defer srv.Close()

	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.Write([]byte{startFrame, 4, uStartDtActive | 0x03, 0, 0, 0})
	buf := make([]byte, 6)
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatal(err)
	}
	conn.Write(iFrame(badASDU))

	e := waitRx(t, got)
	if e.Kind != RxDecode || string(e.Raw) != string(badASDU) || e.Err == nil {
		t.Fatalf("unexpected report %+v", e)
	}
}

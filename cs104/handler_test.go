package cs104

import (
	"testing"

	"github.com/marrasen/go-iecp5/asdu"
	"github.com/marrasen/go-iecp5/clog"
)

type captureHandler struct {
	msgs []asdu.Message
}

func (h *captureHandler) Handle(c asdu.Connect, msg asdu.Message) error {
	h.msgs = append(h.msgs, msg)
	return nil
}

func TestClientHandlerDispatch(t *testing.T) {
	opt := NewOption()
	opt.SetParams(asdu.ParamsNarrow)

	h := &captureHandler{}
	c := NewClient(h, opt)

	raw := []byte{
		byte(asdu.M_SP_NA_1),
		0x01, // VSQ number=1
		byte(asdu.Spontaneous),
		0x01, // common addr
		0x01, // IOA
		0x01, // value on
	}
	a := asdu.NewEmptyASDU(asdu.ParamsNarrow)
	if err := a.UnmarshalBinary(raw); err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}

	if err := c.clientHandler(a); err != nil {
		t.Fatalf("clientHandler failed: %v", err)
	}
	if len(h.msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(h.msgs))
	}
	if _, ok := h.msgs[0].(*asdu.SinglePointMsg); !ok {
		t.Fatalf("unexpected message type: %T", h.msgs[0])
	}
}

func TestServerHandlerDispatch(t *testing.T) {
	h := &captureHandler{}
	sess := &SrvSession{
		params:   asdu.ParamsNarrow,
		handler:  h,
		sendASDU: make(chan []byte, 1),
		Clog:     clog.NewLogger("test"),
	}

	raw := []byte{
		byte(asdu.C_IC_NA_1),
		0x01, // VSQ number=1
		byte(asdu.Activation),
		0x01, // common addr
		0x00, // IOA
		byte(asdu.QOIStation),
	}
	a := asdu.NewEmptyASDU(asdu.ParamsNarrow)
	if err := a.UnmarshalBinary(raw); err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}

	if err := sess.serverHandler(a); err != nil {
		t.Fatalf("serverHandler failed: %v", err)
	}
	if len(h.msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(h.msgs))
	}
	if _, ok := h.msgs[0].(*asdu.InterrogationCmdMsg); !ok {
		t.Fatalf("unexpected message type: %T", h.msgs[0])
	}
}

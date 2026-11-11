package cs104

import "github.com/marrasen/go-iecp5/asdu"

// RxErrorKind says how handling a received frame failed.
type RxErrorKind string

const (
	// RxDecode is an ASDU whose header or length could not be decoded, for example an unknown type.
	RxDecode RxErrorKind = "decode"
	// RxParse is an ASDU that decoded but whose information objects could not be parsed.
	RxParse RxErrorKind = "parse"
	// RxUFrame is a U-frame with a function the library does not know.
	RxUFrame RxErrorKind = "uframe"
	// RxInactive is an I-frame that arrived before STARTDT was confirmed.
	RxInactive RxErrorKind = "inactive"
	// RxSequence is an I-frame with a wrong sequence number; the connection is closed after it.
	RxSequence RxErrorKind = "sequence"
)

// RxError is a received frame that was dropped instead of reaching the handler.
type RxError struct {
	Kind RxErrorKind
	// Raw is the ASDU as received for RxDecode and RxParse, and the whole APDU otherwise. It is a copy the handler
	// may keep.
	Raw []byte
	// ASDU is the decoded ASDU for RxParse, and nil otherwise.
	ASDU *asdu.ASDU
	Err  error
}

// RxErrorHandler is called from the receive goroutine for every dropped frame; it must not block.
type RxErrorHandler func(c asdu.Connect, e RxError)

// reportRx calls h, if set, with a copy of raw.
func reportRx(h RxErrorHandler, c asdu.Connect, kind RxErrorKind, raw []byte, a *asdu.ASDU, err error) {
	if h == nil {
		return
	}
	h(c, RxError{Kind: kind, Raw: append([]byte(nil), raw...), ASDU: a, Err: err})
}

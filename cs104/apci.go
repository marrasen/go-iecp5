// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package cs104

import (
	"fmt"

	"github.com/marrasen/go-iecp5/asdu"
)

const startFrame byte = 0x68 // start character

// APDU form Max size 255
//
//	|              APCI                   |       ASDU         |
//	| start | APDU length | control field |       ASDU         |
//	                 |          APDU field size(253)           |
//
// bytes|    1  |    1   |        4           |                    |
const (
	APCICtlFiledSize = 4 // control filed(4)

	APDUSizeMax      = 255                                 // start(1) + length(1) + control field(4) + ASDU
	APDUFieldSizeMax = APCICtlFiledSize + asdu.ASDUSizeMax // control field(4) + ASDU
)

// U-frame control field functions
const (
	uStartDtActive  byte = 4 << iota // Start activation 0x04
	uStartDtConfirm                  // Start confirmation 0x08
	uStopDtActive                    // Stop activation 0x10
	uStopDtConfirm                   // Stop confirmation 0x20
	uTestFrActive                    // Test activation 0x40
	uTestFrConfirm                   // Test confirmation 0x80
)

// I-frame: contains APCI and ASDU. Information frame used for numbered information transfer
type iAPCI struct {
	sendSN, rcvSN uint16
}

func (sf iAPCI) String() string {
	return fmt.Sprintf("I[sendNO: %d, recvNO: %d]", sf.sendSN, sf.rcvSN)
}

// S-frame: contains only APCI. Used primarily to acknowledge correct frame transmission (supervisory)
type sAPCI struct {
	rcvSN uint16
}

func (sf sAPCI) String() string {
	return fmt.Sprintf("S[recvNO: %d]", sf.rcvSN)
}

// U-frame: contains only APCI. Unnumbered control information
type uAPCI struct {
	function byte // bit8 测试确认
}

func (sf uAPCI) String() string {
	var s string
	switch sf.function {
	case uStartDtActive:
		s = "StartDtActive"
	case uStartDtConfirm:
		s = "StartDtConfirm"
	case uStopDtActive:
		s = "StopDtActive"
	case uStopDtConfirm:
		s = "StopDtConfirm"
	case uTestFrActive:
		s = "TestFrActive"
	case uTestFrConfirm:
		s = "TestFrConfirm"
	default:
		s = "Unknown"
	}
	return fmt.Sprintf("U[function: %s]", s)
}

// newIFrame creates an I-frame and returns the APDU
func newIFrame(sendSN, RcvSN uint16, asdus []byte) ([]byte, error) {
	if len(asdus) > asdu.ASDUSizeMax {
		return nil, fmt.Errorf("ASDU filed large than max %d", asdu.ASDUSizeMax)
	}

	b := make([]byte, len(asdus)+6)

	b[0] = startFrame
	b[1] = byte(len(asdus) + 4)
	b[2] = byte(sendSN << 1)
	b[3] = byte(sendSN >> 7)
	b[4] = byte(RcvSN << 1)
	b[5] = byte(RcvSN >> 7)
	copy(b[6:], asdus)

	return b, nil
}

// newSFrame creates an S-frame and returns the APDU
func newSFrame(RcvSN uint16) []byte {
	return []byte{startFrame, 4, 0x01, 0x00, byte(RcvSN << 1), byte(RcvSN >> 7)}
}

// newUFrame creates a U-frame and returns the APDU
func newUFrame(which byte) []byte {
	return []byte{startFrame, 4, which | 0x03, 0x00, 0x00, 0x00}
}

// APCI application protocol control information
type APCI struct {
	start                  byte
	apduFiledLen           byte // length of control + ASDU
	ctr1, ctr2, ctr3, ctr4 byte
}

// return frame type , APCI, remain data
func parse(apdu []byte) (interface{}, []byte) {
	apci := APCI{apdu[0], apdu[1], apdu[2], apdu[3], apdu[4], apdu[5]}
	if apci.ctr1&0x01 == 0 {
		return iAPCI{
			sendSN: uint16(apci.ctr1)>>1 + uint16(apci.ctr2)<<7,
			rcvSN:  uint16(apci.ctr3)>>1 + uint16(apci.ctr4)<<7,
		}, apdu[6:]
	}
	if apci.ctr1&0x03 == 0x01 {
		return sAPCI{
			rcvSN: uint16(apci.ctr3)>>1 + uint16(apci.ctr4)<<7,
		}, apdu[6:]
	}
	// apci.ctrl&0x03 == 0x03
	return uAPCI{
		function: apci.ctr1 & 0xfc,
	}, apdu[6:]
}

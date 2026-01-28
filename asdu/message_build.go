// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package asdu

func newMessageHeader(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, isSequence bool, count int) Header {
	h := Header{
		Params: c.Params(),
		Identifier: Identifier{
			Type:       typeID,
			Variable:   VariableStruct{IsSequence: isSequence},
			Coa:        coa,
			CommonAddr: ca,
		},
	}
	if count > 0 && count < 128 {
		h.Identifier.Variable.Number = byte(count)
	}
	return h
}

func sendEncoded(c Connect, msg Message) error {
	a, err := EncodeMessage(msg)
	if err != nil {
		return err
	}
	return c.Send(a)
}

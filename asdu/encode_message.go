// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package asdu

import "errors"

var errEncodeUnsupported = errors.New("unsupported message type")

// EncodeMessage builds an ASDU from a parsed message.
func EncodeMessage(msg Message) (*ASDU, error) {
	if msg == nil {
		return nil, ErrParam
	}
	h := msg.Header()
	if h.Params == nil {
		return nil, ErrParam
	}

	switch m := msg.(type) {
	case UnknownMsg:
		if len(h.RawInfoObj) == 0 {
			return nil, ErrTypeIDNotMatch
		}
		return h.ASDU(), nil
	case SinglePointMsg:
		return encodeSinglePoint(h, m)
	case DoublePointMsg:
		return encodeDoublePoint(h, m)
	case StepPositionMsg:
		return encodeStepPosition(h, m)
	case BitString32Msg:
		return encodeBitString32(h, m)
	case MeasuredValueNormalMsg:
		return encodeMeasuredValueNormal(h, m)
	case MeasuredValueScaledMsg:
		return encodeMeasuredValueScaled(h, m)
	case MeasuredValueFloatMsg:
		return encodeMeasuredValueFloat(h, m)
	case IntegratedTotalsMsg:
		return encodeIntegratedTotals(h, m)
	case EventOfProtectionMsg:
		return encodeEventOfProtection(h, m)
	case PackedStartEventsMsg:
		return encodePackedStartEvents(h, m)
	case PackedOutputCircuitMsg:
		return encodePackedOutputCircuit(h, m)
	case PackedSinglePointWithSCDMsg:
		return encodePackedSinglePointWithSCD(h, m)
	case EndOfInitMsg:
		return encodeEndOfInit(h, m)
	case SingleCommandMsg:
		return encodeSingleCommand(h, m)
	case DoubleCommandMsg:
		return encodeDoubleCommand(h, m)
	case StepCommandMsg:
		return encodeStepCommand(h, m)
	case SetpointNormalMsg:
		return encodeSetpointNormal(h, m)
	case SetpointScaledMsg:
		return encodeSetpointScaled(h, m)
	case SetpointFloatMsg:
		return encodeSetpointFloat(h, m)
	case BitsString32CmdMsg:
		return encodeBitsString32Cmd(h, m)
	case ParameterNormalMsg:
		return encodeParameterNormal(h, m)
	case ParameterScaledMsg:
		return encodeParameterScaled(h, m)
	case ParameterFloatMsg:
		return encodeParameterFloat(h, m)
	case ParameterActivationMsg:
		return encodeParameterActivation(h, m)
	case InterrogationCmdMsg:
		return encodeInterrogationCmd(h, m)
	case CounterInterrogationCmdMsg:
		return encodeCounterInterrogationCmd(h, m)
	case ReadCmdMsg:
		return encodeReadCmd(h, m)
	case ClockSyncCmdMsg:
		return encodeClockSyncCmd(h, m)
	case TestCmdMsg:
		return encodeTestCmd(h, m)
	case ResetProcessCmdMsg:
		return encodeResetProcessCmd(h, m)
	case DelayAcquireCmdMsg:
		return encodeDelayAcquireCmd(h, m)
	case TestCmdCP56Msg:
		return encodeTestCmdCP56(h, m)
	default:
		return nil, errEncodeUnsupported
	}
}

func newASDUFromHeader(h Header) *ASDU {
	a := NewASDU(h.Params, h.Identifier)
	a.Identifier.Type = h.Identifier.Type
	return a
}

func setVariable(a *ASDU, count int, isSequence bool) error {
	a.Variable.IsSequence = isSequence
	return a.SetVariableNumber(count)
}

func encodeSinglePoint(h Header, m SinglePointMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if len(m.Items) == 0 {
		return nil, ErrNotAnyObjInfo
	}
	if err := setVariable(a, len(m.Items), h.Identifier.Variable.IsSequence); err != nil {
		return nil, err
	}
	once := false
	for _, it := range m.Items {
		if !h.Identifier.Variable.IsSequence || !once {
			once = true
			if err := a.appendInfoObjAddr(it.Ioa); err != nil {
				return nil, err
			}
		}
		val := byte(0)
		if it.Value {
			val = 0x01
		}
		a.appendBytes(val | byte(it.Qds&0xf0))
		switch m.TypeID() {
		case M_SP_TA_1:
			a.appendCP24Time2a(it.Time, a.InfoObjTimeZone)
		case M_SP_TB_1:
			a.appendCP56Time2a(it.Time, a.InfoObjTimeZone)
		}
	}
	return a, nil
}

func encodeDoublePoint(h Header, m DoublePointMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if len(m.Items) == 0 {
		return nil, ErrNotAnyObjInfo
	}
	if err := setVariable(a, len(m.Items), h.Identifier.Variable.IsSequence); err != nil {
		return nil, err
	}
	once := false
	for _, it := range m.Items {
		if !h.Identifier.Variable.IsSequence || !once {
			once = true
			if err := a.appendInfoObjAddr(it.Ioa); err != nil {
				return nil, err
			}
		}
		a.appendBytes(byte(it.Value&0x03) | byte(it.Qds&0xf0))
		switch m.TypeID() {
		case M_DP_TA_1:
			a.appendCP24Time2a(it.Time, a.InfoObjTimeZone)
		case M_DP_TB_1:
			a.appendCP56Time2a(it.Time, a.InfoObjTimeZone)
		}
	}
	return a, nil
}

func encodeStepPosition(h Header, m StepPositionMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if len(m.Items) == 0 {
		return nil, ErrNotAnyObjInfo
	}
	if err := setVariable(a, len(m.Items), h.Identifier.Variable.IsSequence); err != nil {
		return nil, err
	}
	once := false
	for _, it := range m.Items {
		if !h.Identifier.Variable.IsSequence || !once {
			once = true
			if err := a.appendInfoObjAddr(it.Ioa); err != nil {
				return nil, err
			}
		}
		a.appendBytes(it.Value.Value(), byte(it.Qds))
		switch m.TypeID() {
		case M_ST_TA_1:
			a.appendCP24Time2a(it.Time, a.InfoObjTimeZone)
		case M_ST_TB_1, M_SP_TB_1:
			a.appendCP56Time2a(it.Time, a.InfoObjTimeZone)
		}
	}
	return a, nil
}

func encodeBitString32(h Header, m BitString32Msg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if len(m.Items) == 0 {
		return nil, ErrNotAnyObjInfo
	}
	if err := setVariable(a, len(m.Items), h.Identifier.Variable.IsSequence); err != nil {
		return nil, err
	}
	once := false
	for _, it := range m.Items {
		if !h.Identifier.Variable.IsSequence || !once {
			once = true
			if err := a.appendInfoObjAddr(it.Ioa); err != nil {
				return nil, err
			}
		}
		a.appendBitsString32(it.Value).appendBytes(byte(it.Qds))
		switch m.TypeID() {
		case M_BO_TA_1:
			a.appendCP24Time2a(it.Time, a.InfoObjTimeZone)
		case M_BO_TB_1:
			a.appendCP56Time2a(it.Time, a.InfoObjTimeZone)
		}
	}
	return a, nil
}

func encodeMeasuredValueNormal(h Header, m MeasuredValueNormalMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if len(m.Items) == 0 {
		return nil, ErrNotAnyObjInfo
	}
	if err := setVariable(a, len(m.Items), h.Identifier.Variable.IsSequence); err != nil {
		return nil, err
	}
	once := false
	for _, it := range m.Items {
		if !h.Identifier.Variable.IsSequence || !once {
			once = true
			if err := a.appendInfoObjAddr(it.Ioa); err != nil {
				return nil, err
			}
		}
		a.appendNormalize(it.Value)
		switch m.TypeID() {
		case M_ME_NA_1:
			a.appendBytes(byte(it.Qds))
		case M_ME_TA_1:
			a.appendBytes(byte(it.Qds)).appendCP24Time2a(it.Time, a.InfoObjTimeZone)
		case M_ME_TD_1:
			a.appendBytes(byte(it.Qds)).appendCP56Time2a(it.Time, a.InfoObjTimeZone)
		case M_ME_ND_1:
		}
	}
	return a, nil
}

func encodeMeasuredValueScaled(h Header, m MeasuredValueScaledMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if len(m.Items) == 0 {
		return nil, ErrNotAnyObjInfo
	}
	if err := setVariable(a, len(m.Items), h.Identifier.Variable.IsSequence); err != nil {
		return nil, err
	}
	once := false
	for _, it := range m.Items {
		if !h.Identifier.Variable.IsSequence || !once {
			once = true
			if err := a.appendInfoObjAddr(it.Ioa); err != nil {
				return nil, err
			}
		}
		a.appendScaled(it.Value).appendBytes(byte(it.Qds))
		switch m.TypeID() {
		case M_ME_TB_1:
			a.appendCP24Time2a(it.Time, a.InfoObjTimeZone)
		case M_ME_TE_1:
			a.appendCP56Time2a(it.Time, a.InfoObjTimeZone)
		}
	}
	return a, nil
}

func encodeMeasuredValueFloat(h Header, m MeasuredValueFloatMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if len(m.Items) == 0 {
		return nil, ErrNotAnyObjInfo
	}
	if err := setVariable(a, len(m.Items), h.Identifier.Variable.IsSequence); err != nil {
		return nil, err
	}
	once := false
	for _, it := range m.Items {
		if !h.Identifier.Variable.IsSequence || !once {
			once = true
			if err := a.appendInfoObjAddr(it.Ioa); err != nil {
				return nil, err
			}
		}
		a.appendFloat32(it.Value).appendBytes(byte(it.Qds & 0xf1))
		switch m.TypeID() {
		case M_ME_TC_1:
			a.appendCP24Time2a(it.Time, a.InfoObjTimeZone)
		case M_ME_TF_1:
			a.appendCP56Time2a(it.Time, a.InfoObjTimeZone)
		}
	}
	return a, nil
}

func encodeIntegratedTotals(h Header, m IntegratedTotalsMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if len(m.Items) == 0 {
		return nil, ErrNotAnyObjInfo
	}
	if err := setVariable(a, len(m.Items), h.Identifier.Variable.IsSequence); err != nil {
		return nil, err
	}
	once := false
	for _, it := range m.Items {
		if !h.Identifier.Variable.IsSequence || !once {
			once = true
			if err := a.appendInfoObjAddr(it.Ioa); err != nil {
				return nil, err
			}
		}
		a.appendBinaryCounterReading(it.Value)
		switch m.TypeID() {
		case M_IT_TA_1:
			a.appendCP24Time2a(it.Time, a.InfoObjTimeZone)
		case M_IT_TB_1:
			a.appendCP56Time2a(it.Time, a.InfoObjTimeZone)
		}
	}
	return a, nil
}

func encodeEventOfProtection(h Header, m EventOfProtectionMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if len(m.Items) == 0 {
		return nil, ErrNotAnyObjInfo
	}
	if err := setVariable(a, len(m.Items), h.Identifier.Variable.IsSequence); err != nil {
		return nil, err
	}
	once := false
	for _, it := range m.Items {
		if !h.Identifier.Variable.IsSequence || !once {
			once = true
			if err := a.appendInfoObjAddr(it.Ioa); err != nil {
				return nil, err
			}
		}
		a.appendBytes(byte(it.Event&0x03) | byte(it.Qdp&0xf8))
		a.appendCP16Time2a(it.Msec)
		switch m.TypeID() {
		case M_EP_TA_1:
			a.appendCP24Time2a(it.Time, a.InfoObjTimeZone)
		case M_EP_TD_1:
			a.appendCP56Time2a(it.Time, a.InfoObjTimeZone)
		}
	}
	return a, nil
}

func encodePackedStartEvents(h Header, m PackedStartEventsMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Item.Ioa); err != nil {
		return nil, err
	}
	a.appendBytes(byte(m.Item.Event), byte(m.Item.Qdp)&0xf1)
	a.appendCP16Time2a(m.Item.Msec)
	switch m.TypeID() {
	case M_EP_TB_1:
		a.appendCP24Time2a(m.Item.Time, a.InfoObjTimeZone)
	case M_EP_TE_1:
		a.appendCP56Time2a(m.Item.Time, a.InfoObjTimeZone)
	}
	return a, nil
}

func encodePackedOutputCircuit(h Header, m PackedOutputCircuitMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Item.Ioa); err != nil {
		return nil, err
	}
	a.appendBytes(byte(m.Item.Oci), byte(m.Item.Qdp)&0xf1)
	a.appendCP16Time2a(m.Item.Msec)
	switch m.TypeID() {
	case M_EP_TC_1:
		a.appendCP24Time2a(m.Item.Time, a.InfoObjTimeZone)
	case M_EP_TF_1:
		a.appendCP56Time2a(m.Item.Time, a.InfoObjTimeZone)
	}
	return a, nil
}

func encodePackedSinglePointWithSCD(h Header, m PackedSinglePointWithSCDMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if len(m.Items) == 0 {
		return nil, ErrNotAnyObjInfo
	}
	if err := setVariable(a, len(m.Items), h.Identifier.Variable.IsSequence); err != nil {
		return nil, err
	}
	once := false
	for _, it := range m.Items {
		if !h.Identifier.Variable.IsSequence || !once {
			once = true
			if err := a.appendInfoObjAddr(it.Ioa); err != nil {
				return nil, err
			}
		}
		a.appendStatusAndStatusChangeDetection(it.Scd)
		a.appendBytes(byte(it.Qds))
	}
	return a, nil
}

func encodeEndOfInit(h Header, m EndOfInitMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.IOA); err != nil {
		return nil, err
	}
	a.appendBytes(m.COI.Value())
	return a, nil
}

func encodeSingleCommand(h Header, m SingleCommandMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Cmd.Ioa); err != nil {
		return nil, err
	}
	val := byte(0)
	if m.Cmd.Value {
		val = 0x01
	}
	a.appendBytes(m.Cmd.Qoc.Value() | val)
	if m.TypeID() == C_SC_TA_1 {
		a.appendCP56Time2a(m.Cmd.Time, a.InfoObjTimeZone)
	}
	return a, nil
}

func encodeDoubleCommand(h Header, m DoubleCommandMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Cmd.Ioa); err != nil {
		return nil, err
	}
	a.appendBytes(m.Cmd.Qoc.Value() | byte(m.Cmd.Value&0x03))
	if m.TypeID() == C_DC_TA_1 {
		a.appendCP56Time2a(m.Cmd.Time, a.InfoObjTimeZone)
	}
	return a, nil
}

func encodeStepCommand(h Header, m StepCommandMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Cmd.Ioa); err != nil {
		return nil, err
	}
	a.appendBytes(m.Cmd.Qoc.Value() | byte(m.Cmd.Value&0x03))
	if m.TypeID() == C_RC_TA_1 {
		a.appendCP56Time2a(m.Cmd.Time, a.InfoObjTimeZone)
	}
	return a, nil
}

func encodeSetpointNormal(h Header, m SetpointNormalMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Cmd.Ioa); err != nil {
		return nil, err
	}
	a.appendNormalize(m.Cmd.Value).appendBytes(m.Cmd.Qos.Value())
	if m.TypeID() == C_SE_TA_1 {
		a.appendCP56Time2a(m.Cmd.Time, a.InfoObjTimeZone)
	}
	return a, nil
}

func encodeSetpointScaled(h Header, m SetpointScaledMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Cmd.Ioa); err != nil {
		return nil, err
	}
	a.appendScaled(m.Cmd.Value).appendBytes(m.Cmd.Qos.Value())
	if m.TypeID() == C_SE_TB_1 {
		a.appendCP56Time2a(m.Cmd.Time, a.InfoObjTimeZone)
	}
	return a, nil
}

func encodeSetpointFloat(h Header, m SetpointFloatMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Cmd.Ioa); err != nil {
		return nil, err
	}
	a.appendFloat32(m.Cmd.Value).appendBytes(m.Cmd.Qos.Value())
	if m.TypeID() == C_SE_TC_1 {
		a.appendCP56Time2a(m.Cmd.Time, a.InfoObjTimeZone)
	}
	return a, nil
}

func encodeBitsString32Cmd(h Header, m BitsString32CmdMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Cmd.Ioa); err != nil {
		return nil, err
	}
	a.appendBitsString32(m.Cmd.Value)
	if m.TypeID() == C_BO_TA_1 {
		a.appendCP56Time2a(m.Cmd.Time, a.InfoObjTimeZone)
	}
	return a, nil
}

func encodeParameterNormal(h Header, m ParameterNormalMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Param.Ioa); err != nil {
		return nil, err
	}
	a.appendNormalize(m.Param.Value).appendBytes(m.Param.Qpm.Value())
	return a, nil
}

func encodeParameterScaled(h Header, m ParameterScaledMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Param.Ioa); err != nil {
		return nil, err
	}
	a.appendScaled(m.Param.Value).appendBytes(m.Param.Qpm.Value())
	return a, nil
}

func encodeParameterFloat(h Header, m ParameterFloatMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Param.Ioa); err != nil {
		return nil, err
	}
	a.appendFloat32(m.Param.Value).appendBytes(m.Param.Qpm.Value())
	return a, nil
}

func encodeParameterActivation(h Header, m ParameterActivationMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.Param.Ioa); err != nil {
		return nil, err
	}
	a.appendBytes(byte(m.Param.Qpa))
	return a, nil
}

func encodeInterrogationCmd(h Header, m InterrogationCmdMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.IOA); err != nil {
		return nil, err
	}
	a.appendBytes(byte(m.QOI))
	return a, nil
}

func encodeCounterInterrogationCmd(h Header, m CounterInterrogationCmdMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.IOA); err != nil {
		return nil, err
	}
	a.appendBytes(m.QCC.Value())
	return a, nil
}

func encodeReadCmd(h Header, m ReadCmdMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.IOA); err != nil {
		return nil, err
	}
	return a, nil
}

func encodeClockSyncCmd(h Header, m ClockSyncCmdMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.IOA); err != nil {
		return nil, err
	}
	a.appendCP56Time2a(m.Time, a.InfoObjTimeZone)
	return a, nil
}

func encodeTestCmd(h Header, m TestCmdMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.IOA); err != nil {
		return nil, err
	}
	val := uint16(0)
	if m.Test {
		val = FBPTestWord
	}
	a.appendUint16(val)
	return a, nil
}

func encodeResetProcessCmd(h Header, m ResetProcessCmdMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.IOA); err != nil {
		return nil, err
	}
	a.appendBytes(byte(m.QRP))
	return a, nil
}

func encodeDelayAcquireCmd(h Header, m DelayAcquireCmdMsg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.IOA); err != nil {
		return nil, err
	}
	a.appendCP16Time2a(m.Msec)
	return a, nil
}

func encodeTestCmdCP56(h Header, m TestCmdCP56Msg) (*ASDU, error) {
	a := newASDUFromHeader(h)
	a.Identifier.Type = m.TypeID()
	if err := setVariable(a, 1, false); err != nil {
		return nil, err
	}
	if err := a.appendInfoObjAddr(m.IOA); err != nil {
		return nil, err
	}
	val := uint16(0)
	if m.Test {
		val = FBPTestWord
	}
	a.appendUint16(val)
	a.appendCP56Time2a(m.Time, a.InfoObjTimeZone)
	return a, nil
}

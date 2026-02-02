// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package asdu

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"time"
)

// Header carries ASDU identification and raw payload.
type Header struct {
	Params     *Params
	Identifier Identifier
	RawInfoObj []byte
}

// ASDU recreates an ASDU that mirrors the original header and payload.
func (h Header) ASDU() *ASDU {
	if h.Params == nil {
		return nil
	}
	a := NewASDU(h.Params, h.Identifier)
	a.infoObj = append(a.infoObj, h.RawInfoObj...)
	return a
}

// Message is a parsed ASDU payload that supports type assertions.
type Message interface {
	Header() Header
	TypeID() TypeID
	String() string
}

// UnknownMsg is returned for unsupported or unknown TypeIDs.
type UnknownMsg struct {
	H Header
}

// Header returns the ASDU header.
func (m *UnknownMsg) Header() Header { return m.H }

// TypeID returns the ASDU TypeID.
func (m *UnknownMsg) TypeID() TypeID { return m.H.Identifier.Type }

// Monitoring direction messages.
type SinglePointMsg struct {
	H     Header
	Items []SinglePointInfo
}

func (m *SinglePointMsg) Header() Header { return m.H }
func (m *SinglePointMsg) TypeID() TypeID { return m.H.Identifier.Type }

type DoublePointMsg struct {
	H     Header
	Items []DoublePointInfo
}

func (m *DoublePointMsg) Header() Header { return m.H }
func (m *DoublePointMsg) TypeID() TypeID { return m.H.Identifier.Type }

type StepPositionMsg struct {
	H     Header
	Items []StepPositionInfo
}

func (m *StepPositionMsg) Header() Header { return m.H }
func (m *StepPositionMsg) TypeID() TypeID { return m.H.Identifier.Type }

type BitString32Msg struct {
	H     Header
	Items []BitString32Info
}

func (m *BitString32Msg) Header() Header { return m.H }
func (m *BitString32Msg) TypeID() TypeID { return m.H.Identifier.Type }

type MeasuredValueNormalMsg struct {
	H     Header
	Items []MeasuredValueNormalInfo
}

func (m *MeasuredValueNormalMsg) Header() Header { return m.H }
func (m *MeasuredValueNormalMsg) TypeID() TypeID { return m.H.Identifier.Type }

type MeasuredValueScaledMsg struct {
	H     Header
	Items []MeasuredValueScaledInfo
}

func (m *MeasuredValueScaledMsg) Header() Header { return m.H }
func (m *MeasuredValueScaledMsg) TypeID() TypeID { return m.H.Identifier.Type }

type MeasuredValueFloatMsg struct {
	H     Header
	Items []MeasuredValueFloatInfo
}

func (m *MeasuredValueFloatMsg) Header() Header { return m.H }
func (m *MeasuredValueFloatMsg) TypeID() TypeID { return m.H.Identifier.Type }

type IntegratedTotalsMsg struct {
	H     Header
	Items []BinaryCounterReadingInfo
}

func (m *IntegratedTotalsMsg) Header() Header { return m.H }
func (m *IntegratedTotalsMsg) TypeID() TypeID { return m.H.Identifier.Type }

type EventOfProtectionMsg struct {
	H     Header
	Items []EventOfProtectionEquipmentInfo
}

func (m *EventOfProtectionMsg) Header() Header { return m.H }
func (m *EventOfProtectionMsg) TypeID() TypeID { return m.H.Identifier.Type }

type PackedStartEventsMsg struct {
	H    Header
	Item PackedStartEventsOfProtectionEquipmentInfo
}

func (m *PackedStartEventsMsg) Header() Header { return m.H }
func (m *PackedStartEventsMsg) TypeID() TypeID { return m.H.Identifier.Type }

type PackedOutputCircuitMsg struct {
	H    Header
	Item PackedOutputCircuitInfoInfo
}

func (m *PackedOutputCircuitMsg) Header() Header { return m.H }
func (m *PackedOutputCircuitMsg) TypeID() TypeID { return m.H.Identifier.Type }

type PackedSinglePointWithSCDMsg struct {
	H     Header
	Items []PackedSinglePointWithSCDInfo
}

func (m *PackedSinglePointWithSCDMsg) Header() Header { return m.H }
func (m *PackedSinglePointWithSCDMsg) TypeID() TypeID { return m.H.Identifier.Type }

type EndOfInitMsg struct {
	H   Header
	IOA InfoObjAddr
	COI CauseOfInitial
}

func (m *EndOfInitMsg) Header() Header { return m.H }
func (m *EndOfInitMsg) TypeID() TypeID { return m.H.Identifier.Type }

// Control direction messages.
type SingleCommandMsg struct {
	H   Header
	Cmd SingleCommandInfo
}

func (m *SingleCommandMsg) Header() Header { return m.H }
func (m *SingleCommandMsg) TypeID() TypeID { return m.H.Identifier.Type }

type DoubleCommandMsg struct {
	H   Header
	Cmd DoubleCommandInfo
}

func (m *DoubleCommandMsg) Header() Header { return m.H }
func (m *DoubleCommandMsg) TypeID() TypeID { return m.H.Identifier.Type }

type StepCommandMsg struct {
	H   Header
	Cmd StepCommandInfo
}

func (m *StepCommandMsg) Header() Header { return m.H }
func (m *StepCommandMsg) TypeID() TypeID { return m.H.Identifier.Type }

type SetpointNormalMsg struct {
	H   Header
	Cmd SetpointCommandNormalInfo
}

func (m *SetpointNormalMsg) Header() Header { return m.H }
func (m *SetpointNormalMsg) TypeID() TypeID { return m.H.Identifier.Type }

type SetpointScaledMsg struct {
	H   Header
	Cmd SetpointCommandScaledInfo
}

func (m *SetpointScaledMsg) Header() Header { return m.H }
func (m *SetpointScaledMsg) TypeID() TypeID { return m.H.Identifier.Type }

type SetpointFloatMsg struct {
	H   Header
	Cmd SetpointCommandFloatInfo
}

func (m *SetpointFloatMsg) Header() Header { return m.H }
func (m *SetpointFloatMsg) TypeID() TypeID { return m.H.Identifier.Type }

type BitsString32CmdMsg struct {
	H   Header
	Cmd BitsString32CommandInfo
}

func (m *BitsString32CmdMsg) Header() Header { return m.H }
func (m *BitsString32CmdMsg) TypeID() TypeID { return m.H.Identifier.Type }

// Parameter messages.
type ParameterNormalMsg struct {
	H     Header
	Param ParameterNormalInfo
}

func (m *ParameterNormalMsg) Header() Header { return m.H }
func (m *ParameterNormalMsg) TypeID() TypeID { return m.H.Identifier.Type }

type ParameterScaledMsg struct {
	H     Header
	Param ParameterScaledInfo
}

func (m *ParameterScaledMsg) Header() Header { return m.H }
func (m *ParameterScaledMsg) TypeID() TypeID { return m.H.Identifier.Type }

type ParameterFloatMsg struct {
	H     Header
	Param ParameterFloatInfo
}

func (m *ParameterFloatMsg) Header() Header { return m.H }
func (m *ParameterFloatMsg) TypeID() TypeID { return m.H.Identifier.Type }

type ParameterActivationMsg struct {
	H     Header
	Param ParameterActivationInfo
}

func (m *ParameterActivationMsg) Header() Header { return m.H }
func (m *ParameterActivationMsg) TypeID() TypeID { return m.H.Identifier.Type }

// System command messages.
type InterrogationCmdMsg struct {
	H   Header
	IOA InfoObjAddr
	QOI QualifierOfInterrogation
}

func (m *InterrogationCmdMsg) Header() Header { return m.H }
func (m *InterrogationCmdMsg) TypeID() TypeID { return m.H.Identifier.Type }

type CounterInterrogationCmdMsg struct {
	H   Header
	IOA InfoObjAddr
	QCC QualifierCountCall
}

func (m *CounterInterrogationCmdMsg) Header() Header { return m.H }
func (m *CounterInterrogationCmdMsg) TypeID() TypeID { return m.H.Identifier.Type }

type ReadCmdMsg struct {
	H   Header
	IOA InfoObjAddr
}

func (m *ReadCmdMsg) Header() Header { return m.H }
func (m *ReadCmdMsg) TypeID() TypeID { return m.H.Identifier.Type }

type ClockSyncCmdMsg struct {
	H    Header
	IOA  InfoObjAddr
	Time time.Time
}

func (m *ClockSyncCmdMsg) Header() Header { return m.H }
func (m *ClockSyncCmdMsg) TypeID() TypeID { return m.H.Identifier.Type }

type TestCmdMsg struct {
	H    Header
	IOA  InfoObjAddr
	Test bool
}

func (m *TestCmdMsg) Header() Header { return m.H }
func (m *TestCmdMsg) TypeID() TypeID { return m.H.Identifier.Type }

type ResetProcessCmdMsg struct {
	H   Header
	IOA InfoObjAddr
	QRP QualifierOfResetProcessCmd
}

func (m *ResetProcessCmdMsg) Header() Header { return m.H }
func (m *ResetProcessCmdMsg) TypeID() TypeID { return m.H.Identifier.Type }

type DelayAcquireCmdMsg struct {
	H    Header
	IOA  InfoObjAddr
	Msec uint16
}

func (m *DelayAcquireCmdMsg) Header() Header { return m.H }
func (m *DelayAcquireCmdMsg) TypeID() TypeID { return m.H.Identifier.Type }

type TestCmdCP56Msg struct {
	H    Header
	IOA  InfoObjAddr
	Test bool
	Time time.Time
}

func (m *TestCmdCP56Msg) Header() Header { return m.H }
func (m *TestCmdCP56Msg) TypeID() TypeID { return m.H.Identifier.Type }

type decodeCursor struct {
	params *Params
	data   []byte
	off    int
}

func (d *decodeCursor) remaining() int {
	return len(d.data) - d.off
}

func (d *decodeCursor) read(n int) ([]byte, error) {
	if d.remaining() < n {
		return nil, io.EOF
	}
	b := d.data[d.off : d.off+n]
	d.off += n
	return b, nil
}

func (d *decodeCursor) readByte() (byte, error) {
	b, err := d.read(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (d *decodeCursor) readUint16() (uint16, error) {
	b, err := d.read(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b), nil
}

func (d *decodeCursor) readInfoObjAddr() (InfoObjAddr, error) {
	switch d.params.InfoObjAddrSize {
	case 1:
		b, err := d.read(1)
		if err != nil {
			return 0, err
		}
		return InfoObjAddr(b[0]), nil
	case 2:
		b, err := d.read(2)
		if err != nil {
			return 0, err
		}
		return InfoObjAddr(b[0]) | (InfoObjAddr(b[1]) << 8), nil
	case 3:
		b, err := d.read(3)
		if err != nil {
			return 0, err
		}
		return InfoObjAddr(b[0]) | (InfoObjAddr(b[1]) << 8) | (InfoObjAddr(b[2]) << 16), nil
	default:
		return 0, ErrParam
	}
}

func (d *decodeCursor) readNormalize() (Normalize, error) {
	v, err := d.readUint16()
	if err != nil {
		return 0, err
	}
	return Normalize(v), nil
}

func (d *decodeCursor) readScaled() (int16, error) {
	v, err := d.readUint16()
	if err != nil {
		return 0, err
	}
	return int16(v), nil
}

func (d *decodeCursor) readFloat32() (float32, error) {
	b, err := d.read(4)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(b)), nil
}

func (d *decodeCursor) readBinaryCounterReading() (BinaryCounterReading, error) {
	b, err := d.read(5)
	if err != nil {
		return BinaryCounterReading{}, err
	}
	v := int32(binary.LittleEndian.Uint32(b[:4]))
	flags := b[4]
	return BinaryCounterReading{
		CounterReading: v,
		SeqNumber:      flags & 0x1f,
		HasCarry:       flags&0x20 == 0x20,
		IsAdjusted:     flags&0x40 == 0x40,
		IsInvalid:      flags&0x80 == 0x80,
	}, nil
}

func (d *decodeCursor) readBitsString32() (uint32, error) {
	b, err := d.read(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (d *decodeCursor) readCP24Time2a() (time.Time, error) {
	b, err := d.read(3)
	if err != nil {
		return time.Time{}, err
	}
	return ParseCP24Time2a(b, d.params.InfoObjTimeZone), nil
}

func (d *decodeCursor) readCP56Time2a() (time.Time, error) {
	b, err := d.read(7)
	if err != nil {
		return time.Time{}, err
	}
	return ParseCP56Time2a(b, d.params.InfoObjTimeZone), nil
}

func (d *decodeCursor) readCP16Time2a() (uint16, error) {
	b, err := d.read(2)
	if err != nil {
		return 0, err
	}
	return ParseCP16Time2a(b), nil
}

func (d *decodeCursor) readStatusAndStatusChangeDetection() (StatusAndStatusChangeDetection, error) {
	b, err := d.read(4)
	if err != nil {
		return 0, err
	}
	return StatusAndStatusChangeDetection(binary.LittleEndian.Uint32(b)), nil
}

// ParseASDU decodes an ASDU into a typed message without mutating the ASDU buffer.
func ParseASDU(a *ASDU) (Message, error) {
	if a == nil || a.Params == nil {
		return nil, ErrParam
	}
	header := Header{
		Params:     a.Params,
		Identifier: a.Identifier,
		RawInfoObj: a.infoObj,
	}

	cur := decodeCursor{
		params: a.Params,
		data:   a.infoObj,
	}

	switch a.Type {
	case M_SP_NA_1, M_SP_TA_1, M_SP_TB_1:
		items := make([]SinglePointInfo, 0, a.Variable.Number)
		var ioa InfoObjAddr
		for i, once := 0, false; i < int(a.Variable.Number); i++ {
			if !a.Variable.IsSequence || !once {
				once = true
				var err error
				ioa, err = cur.readInfoObjAddr()
				if err != nil {
					return nil, err
				}
			} else {
				ioa++
			}
			value, err := cur.readByte()
			if err != nil {
				return nil, err
			}
			var t time.Time
			switch a.Type {
			case M_SP_NA_1:
			case M_SP_TA_1:
				t, err = cur.readCP24Time2a()
			case M_SP_TB_1:
				t, err = cur.readCP56Time2a()
			default:
				return nil, ErrTypeIDNotMatch
			}
			if err != nil {
				return nil, err
			}
			items = append(items, SinglePointInfo{
				Ioa:   ioa,
				Value: value&0x01 == 0x01,
				Qds:   QualityDescriptor(value & 0xf0),
				Time:  t,
			})
		}
		return &SinglePointMsg{H: header, Items: items}, nil

	case M_DP_NA_1, M_DP_TA_1, M_DP_TB_1:
		items := make([]DoublePointInfo, 0, a.Variable.Number)
		var ioa InfoObjAddr
		for i, once := 0, false; i < int(a.Variable.Number); i++ {
			if !a.Variable.IsSequence || !once {
				once = true
				var err error
				ioa, err = cur.readInfoObjAddr()
				if err != nil {
					return nil, err
				}
			} else {
				ioa++
			}
			value, err := cur.readByte()
			if err != nil {
				return nil, err
			}
			var t time.Time
			switch a.Type {
			case M_DP_NA_1:
			case M_DP_TA_1:
				t, err = cur.readCP24Time2a()
			case M_DP_TB_1:
				t, err = cur.readCP56Time2a()
			default:
				return nil, ErrTypeIDNotMatch
			}
			if err != nil {
				return nil, err
			}
			items = append(items, DoublePointInfo{
				Ioa:   ioa,
				Value: DoublePoint(value & 0x03),
				Qds:   QualityDescriptor(value & 0xf0),
				Time:  t,
			})
		}
		return &DoublePointMsg{H: header, Items: items}, nil

	case M_ST_NA_1, M_ST_TA_1, M_ST_TB_1:
		items := make([]StepPositionInfo, 0, a.Variable.Number)
		var ioa InfoObjAddr
		for i, once := 0, false; i < int(a.Variable.Number); i++ {
			if !a.Variable.IsSequence || !once {
				once = true
				var err error
				ioa, err = cur.readInfoObjAddr()
				if err != nil {
					return nil, err
				}
			} else {
				ioa++
			}
			raw, err := cur.readByte()
			if err != nil {
				return nil, err
			}
			qdsRaw, err := cur.readByte()
			if err != nil {
				return nil, err
			}
			var t time.Time
			switch a.Type {
			case M_ST_NA_1:
			case M_ST_TA_1:
				t, err = cur.readCP24Time2a()
			case M_ST_TB_1:
				t, err = cur.readCP56Time2a()
			default:
				return nil, ErrTypeIDNotMatch
			}
			if err != nil {
				return nil, err
			}
			items = append(items, StepPositionInfo{
				Ioa:   ioa,
				Value: ParseStepPosition(raw),
				Qds:   QualityDescriptor(qdsRaw),
				Time:  t,
			})
		}
		return &StepPositionMsg{H: header, Items: items}, nil

	case M_BO_NA_1, M_BO_TA_1, M_BO_TB_1:
		items := make([]BitString32Info, 0, a.Variable.Number)
		var ioa InfoObjAddr
		for i, once := 0, false; i < int(a.Variable.Number); i++ {
			if !a.Variable.IsSequence || !once {
				once = true
				var err error
				ioa, err = cur.readInfoObjAddr()
				if err != nil {
					return nil, err
				}
			} else {
				ioa++
			}
			val, err := cur.readBitsString32()
			if err != nil {
				return nil, err
			}
			qdsRaw, err := cur.readByte()
			if err != nil {
				return nil, err
			}
			var t time.Time
			switch a.Type {
			case M_BO_NA_1:
			case M_BO_TA_1:
				t, err = cur.readCP24Time2a()
			case M_BO_TB_1:
				t, err = cur.readCP56Time2a()
			default:
				return nil, ErrTypeIDNotMatch
			}
			if err != nil {
				return nil, err
			}
			items = append(items, BitString32Info{
				Ioa:   ioa,
				Value: val,
				Qds:   QualityDescriptor(qdsRaw),
				Time:  t,
			})
		}
		return &BitString32Msg{H: header, Items: items}, nil

	case M_ME_NA_1, M_ME_TA_1, M_ME_TD_1, M_ME_ND_1:
		items := make([]MeasuredValueNormalInfo, 0, a.Variable.Number)
		var ioa InfoObjAddr
		for i, once := 0, false; i < int(a.Variable.Number); i++ {
			if !a.Variable.IsSequence || !once {
				once = true
				var err error
				ioa, err = cur.readInfoObjAddr()
				if err != nil {
					return nil, err
				}
			} else {
				ioa++
			}
			val, err := cur.readNormalize()
			if err != nil {
				return nil, err
			}
			var t time.Time
			var qds QualityDescriptor
			switch a.Type {
			case M_ME_NA_1:
				b, err := cur.readByte()
				if err != nil {
					return nil, err
				}
				qds = QualityDescriptor(b)
			case M_ME_TA_1:
				b, err := cur.readByte()
				if err != nil {
					return nil, err
				}
				qds = QualityDescriptor(b)
				t, err = cur.readCP24Time2a()
				if err != nil {
					return nil, err
				}
			case M_ME_TD_1:
				b, err := cur.readByte()
				if err != nil {
					return nil, err
				}
				qds = QualityDescriptor(b)
				t, err = cur.readCP56Time2a()
				if err != nil {
					return nil, err
				}
			case M_ME_ND_1:
			default:
				return nil, ErrTypeIDNotMatch
			}
			items = append(items, MeasuredValueNormalInfo{
				Ioa:   ioa,
				Value: val,
				Qds:   qds,
				Time:  t,
			})
		}
		return &MeasuredValueNormalMsg{H: header, Items: items}, nil

	case M_ME_NB_1, M_ME_TB_1, M_ME_TE_1:
		items := make([]MeasuredValueScaledInfo, 0, a.Variable.Number)
		var ioa InfoObjAddr
		for i, once := 0, false; i < int(a.Variable.Number); i++ {
			if !a.Variable.IsSequence || !once {
				once = true
				var err error
				ioa, err = cur.readInfoObjAddr()
				if err != nil {
					return nil, err
				}
			} else {
				ioa++
			}
			val, err := cur.readScaled()
			if err != nil {
				return nil, err
			}
			qdsRaw, err := cur.readByte()
			if err != nil {
				return nil, err
			}
			var t time.Time
			switch a.Type {
			case M_ME_NB_1:
			case M_ME_TB_1:
				t, err = cur.readCP24Time2a()
			case M_ME_TE_1:
				t, err = cur.readCP56Time2a()
			default:
				return nil, ErrTypeIDNotMatch
			}
			if err != nil {
				return nil, err
			}
			items = append(items, MeasuredValueScaledInfo{
				Ioa:   ioa,
				Value: val,
				Qds:   QualityDescriptor(qdsRaw),
				Time:  t,
			})
		}
		return &MeasuredValueScaledMsg{H: header, Items: items}, nil

	case M_ME_NC_1, M_ME_TC_1, M_ME_TF_1:
		items := make([]MeasuredValueFloatInfo, 0, a.Variable.Number)
		var ioa InfoObjAddr
		for i, once := 0, false; i < int(a.Variable.Number); i++ {
			if !a.Variable.IsSequence || !once {
				once = true
				var err error
				ioa, err = cur.readInfoObjAddr()
				if err != nil {
					return nil, err
				}
			} else {
				ioa++
			}
			val, err := cur.readFloat32()
			if err != nil {
				return nil, err
			}
			qua, err := cur.readByte()
			if err != nil {
				return nil, err
			}
			var t time.Time
			switch a.Type {
			case M_ME_NC_1:
			case M_ME_TC_1:
				t, err = cur.readCP24Time2a()
			case M_ME_TF_1:
				t, err = cur.readCP56Time2a()
			default:
				return nil, ErrTypeIDNotMatch
			}
			if err != nil {
				return nil, err
			}
			items = append(items, MeasuredValueFloatInfo{
				Ioa:   ioa,
				Value: val,
				Qds:   QualityDescriptor(qua & 0xf1),
				Time:  t,
			})
		}
		return &MeasuredValueFloatMsg{H: header, Items: items}, nil

	case M_IT_NA_1, M_IT_TA_1, M_IT_TB_1:
		items := make([]BinaryCounterReadingInfo, 0, a.Variable.Number)
		var ioa InfoObjAddr
		for i, once := 0, false; i < int(a.Variable.Number); i++ {
			if !a.Variable.IsSequence || !once {
				once = true
				var err error
				ioa, err = cur.readInfoObjAddr()
				if err != nil {
					return nil, err
				}
			} else {
				ioa++
			}
			val, err := cur.readBinaryCounterReading()
			if err != nil {
				return nil, err
			}
			var t time.Time
			switch a.Type {
			case M_IT_NA_1:
			case M_IT_TA_1:
				t, err = cur.readCP24Time2a()
			case M_IT_TB_1:
				t, err = cur.readCP56Time2a()
			default:
				return nil, ErrTypeIDNotMatch
			}
			if err != nil {
				return nil, err
			}
			items = append(items, BinaryCounterReadingInfo{
				Ioa:   ioa,
				Value: val,
				Time:  t,
			})
		}
		return &IntegratedTotalsMsg{H: header, Items: items}, nil

	case M_EP_TA_1, M_EP_TD_1:
		items := make([]EventOfProtectionEquipmentInfo, 0, a.Variable.Number)
		var ioa InfoObjAddr
		for i, once := 0, false; i < int(a.Variable.Number); i++ {
			if !a.Variable.IsSequence || !once {
				once = true
				var err error
				ioa, err = cur.readInfoObjAddr()
				if err != nil {
					return nil, err
				}
			} else {
				ioa++
			}
			value, err := cur.readByte()
			if err != nil {
				return nil, err
			}
			msec, err := cur.readCP16Time2a()
			if err != nil {
				return nil, err
			}
			var t time.Time
			switch a.Type {
			case M_EP_TA_1:
				t, err = cur.readCP24Time2a()
			case M_EP_TD_1:
				t, err = cur.readCP56Time2a()
			default:
				return nil, ErrTypeIDNotMatch
			}
			if err != nil {
				return nil, err
			}
			items = append(items, EventOfProtectionEquipmentInfo{
				Ioa:   ioa,
				Event: SingleEvent(value & 0x03),
				Qdp:   QualityDescriptorProtection(value & 0xf1),
				Msec:  msec,
				Time:  t,
			})
		}
		return &EventOfProtectionMsg{H: header, Items: items}, nil

	case M_EP_TB_1, M_EP_TE_1:
		if a.Variable.IsSequence || a.Variable.Number != 1 {
			return nil, errors.New("unexpected variable structure for packed start events")
		}
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		event, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		qdpRaw, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		msec, err := cur.readCP16Time2a()
		if err != nil {
			return nil, err
		}
		var t time.Time
		switch a.Type {
		case M_EP_TB_1:
			t, err = cur.readCP24Time2a()
		case M_EP_TE_1:
			t, err = cur.readCP56Time2a()
		default:
			return nil, ErrTypeIDNotMatch
		}
		if err != nil {
			return nil, err
		}
		item := PackedStartEventsOfProtectionEquipmentInfo{
			Ioa:   ioa,
			Event: StartEvent(event),
			Qdp:   QualityDescriptorProtection(qdpRaw & 0xf1),
			Msec:  msec,
			Time:  t,
		}
		return &PackedStartEventsMsg{H: header, Item: item}, nil

	case M_EP_TC_1, M_EP_TF_1:
		if a.Variable.IsSequence || a.Variable.Number != 1 {
			return nil, errors.New("unexpected variable structure for packed output circuit info")
		}
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		oci, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		qdpRaw, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		msec, err := cur.readCP16Time2a()
		if err != nil {
			return nil, err
		}
		var t time.Time
		switch a.Type {
		case M_EP_TC_1:
			t, err = cur.readCP24Time2a()
		case M_EP_TF_1:
			t, err = cur.readCP56Time2a()
		default:
			return nil, ErrTypeIDNotMatch
		}
		if err != nil {
			return nil, err
		}
		item := PackedOutputCircuitInfoInfo{
			Ioa:  ioa,
			Oci:  OutputCircuitInfo(oci),
			Qdp:  QualityDescriptorProtection(qdpRaw & 0xf1),
			Msec: msec,
			Time: t,
		}
		return &PackedOutputCircuitMsg{H: header, Item: item}, nil

	case M_PS_NA_1:
		items := make([]PackedSinglePointWithSCDInfo, 0, a.Variable.Number)
		var ioa InfoObjAddr
		for i, once := 0, false; i < int(a.Variable.Number); i++ {
			if !a.Variable.IsSequence || !once {
				once = true
				var err error
				ioa, err = cur.readInfoObjAddr()
				if err != nil {
					return nil, err
				}
			} else {
				ioa++
			}
			scd, err := cur.readStatusAndStatusChangeDetection()
			if err != nil {
				return nil, err
			}
			qdsRaw, err := cur.readByte()
			if err != nil {
				return nil, err
			}
			items = append(items, PackedSinglePointWithSCDInfo{
				Ioa: ioa,
				Scd: scd,
				Qds: QualityDescriptor(qdsRaw),
			})
		}
		return &PackedSinglePointWithSCDMsg{H: header, Items: items}, nil

	case M_EI_NA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		b, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		return &EndOfInitMsg{H: header, IOA: ioa, COI: ParseCauseOfInitial(b)}, nil

	case C_SC_NA_1, C_SC_TA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		val, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		cmd := SingleCommandInfo{
			Ioa:   ioa,
			Value: val&0x01 == 0x01,
			Qoc:   ParseQualifierOfCommand(val & 0xfe),
		}
		if a.Type == C_SC_TA_1 {
			cmd.Time, err = cur.readCP56Time2a()
			if err != nil {
				return nil, err
			}
		}
		return &SingleCommandMsg{H: header, Cmd: cmd}, nil

	case C_DC_NA_1, C_DC_TA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		val, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		cmd := DoubleCommandInfo{
			Ioa:   ioa,
			Value: DoubleCommand(val & 0x03),
			Qoc:   ParseQualifierOfCommand(val & 0xfc),
		}
		if a.Type == C_DC_TA_1 {
			cmd.Time, err = cur.readCP56Time2a()
			if err != nil {
				return nil, err
			}
		}
		return &DoubleCommandMsg{H: header, Cmd: cmd}, nil

	case C_RC_NA_1, C_RC_TA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		val, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		cmd := StepCommandInfo{
			Ioa:   ioa,
			Value: StepCommand(val & 0x03),
			Qoc:   ParseQualifierOfCommand(val & 0xfc),
		}
		if a.Type == C_RC_TA_1 {
			cmd.Time, err = cur.readCP56Time2a()
			if err != nil {
				return nil, err
			}
		}
		return &StepCommandMsg{H: header, Cmd: cmd}, nil

	case C_SE_NA_1, C_SE_TA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		val, err := cur.readNormalize()
		if err != nil {
			return nil, err
		}
		qosRaw, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		cmd := SetpointCommandNormalInfo{
			Ioa:   ioa,
			Value: val,
			Qos:   ParseQualifierOfSetpointCmd(qosRaw),
		}
		if a.Type == C_SE_TA_1 {
			cmd.Time, err = cur.readCP56Time2a()
			if err != nil {
				return nil, err
			}
		}
		return &SetpointNormalMsg{H: header, Cmd: cmd}, nil

	case C_SE_NB_1, C_SE_TB_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		val, err := cur.readScaled()
		if err != nil {
			return nil, err
		}
		qosRaw, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		cmd := SetpointCommandScaledInfo{
			Ioa:   ioa,
			Value: val,
			Qos:   ParseQualifierOfSetpointCmd(qosRaw),
		}
		if a.Type == C_SE_TB_1 {
			cmd.Time, err = cur.readCP56Time2a()
			if err != nil {
				return nil, err
			}
		}
		return &SetpointScaledMsg{H: header, Cmd: cmd}, nil

	case C_SE_NC_1, C_SE_TC_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		val, err := cur.readFloat32()
		if err != nil {
			return nil, err
		}
		qosRaw, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		cmd := SetpointCommandFloatInfo{
			Ioa:   ioa,
			Value: val,
			Qos:   ParseQualifierOfSetpointCmd(qosRaw),
		}
		if a.Type == C_SE_TC_1 {
			cmd.Time, err = cur.readCP56Time2a()
			if err != nil {
				return nil, err
			}
		}
		return &SetpointFloatMsg{H: header, Cmd: cmd}, nil

	case C_BO_NA_1, C_BO_TA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		val, err := cur.readBitsString32()
		if err != nil {
			return nil, err
		}
		cmd := BitsString32CommandInfo{
			Ioa:   ioa,
			Value: val,
		}
		if a.Type == C_BO_TA_1 {
			cmd.Time, err = cur.readCP56Time2a()
			if err != nil {
				return nil, err
			}
		}
		return &BitsString32CmdMsg{H: header, Cmd: cmd}, nil

	case C_IC_NA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		b, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		return &InterrogationCmdMsg{H: header, IOA: ioa, QOI: QualifierOfInterrogation(b)}, nil

	case C_CI_NA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		b, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		return &CounterInterrogationCmdMsg{H: header, IOA: ioa, QCC: ParseQualifierCountCall(b)}, nil

	case C_RD_NA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		return &ReadCmdMsg{H: header, IOA: ioa}, nil

	case C_CS_NA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		t, err := cur.readCP56Time2a()
		if err != nil {
			return nil, err
		}
		return &ClockSyncCmdMsg{H: header, IOA: ioa, Time: t}, nil

	case C_TS_NA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		v, err := cur.readUint16()
		if err != nil {
			return nil, err
		}
		return &TestCmdMsg{H: header, IOA: ioa, Test: v == FBPTestWord}, nil

	case C_RP_NA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		b, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		return &ResetProcessCmdMsg{H: header, IOA: ioa, QRP: QualifierOfResetProcessCmd(b)}, nil

	case C_CD_NA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		msec, err := cur.readUint16()
		if err != nil {
			return nil, err
		}
		return &DelayAcquireCmdMsg{H: header, IOA: ioa, Msec: msec}, nil

	case C_TS_TA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		v, err := cur.readUint16()
		if err != nil {
			return nil, err
		}
		t, err := cur.readCP56Time2a()
		if err != nil {
			return nil, err
		}
		return &TestCmdCP56Msg{H: header, IOA: ioa, Test: v == FBPTestWord, Time: t}, nil

	case P_ME_NA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		val, err := cur.readNormalize()
		if err != nil {
			return nil, err
		}
		qpmRaw, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		return &ParameterNormalMsg{H: header, Param: ParameterNormalInfo{Ioa: ioa, Value: val, Qpm: ParseQualifierOfParamMV(qpmRaw)}}, nil

	case P_ME_NB_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		val, err := cur.readScaled()
		if err != nil {
			return nil, err
		}
		qpmRaw, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		return &ParameterScaledMsg{H: header, Param: ParameterScaledInfo{Ioa: ioa, Value: val, Qpm: ParseQualifierOfParamMV(qpmRaw)}}, nil

	case P_ME_NC_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		val, err := cur.readFloat32()
		if err != nil {
			return nil, err
		}
		qpmRaw, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		return &ParameterFloatMsg{H: header, Param: ParameterFloatInfo{Ioa: ioa, Value: val, Qpm: ParseQualifierOfParamMV(qpmRaw)}}, nil

	case P_AC_NA_1:
		ioa, err := cur.readInfoObjAddr()
		if err != nil {
			return nil, err
		}
		qpaRaw, err := cur.readByte()
		if err != nil {
			return nil, err
		}
		return &ParameterActivationMsg{H: header, Param: ParameterActivationInfo{Ioa: ioa, Qpa: QualifierOfParameterAct(qpaRaw)}}, nil
	}

	return &UnknownMsg{H: header}, nil
}

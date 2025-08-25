// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

// Package asdu provides the OSI presentation layer.
package asdu

import (
	"fmt"
	"io"
	"math/bits"
	"strings"
	"time"
)

// ASDUSizeMax asdu max size
const (
	ASDUSizeMax = 249
)

// ASDU format
//       | data unit identification | information object <1..n> |
//
//       | <------------  data unit identification ------------>|
//       | typeID | variable struct | cause  |  common address  |
// bytes |    1   |      1          | [1,2]  |      [1,2]       |
//       | <------------  information object ------------------>|
//       | object address | element set  |  object time scale   |
// bytes |     [1,2,3]    |              |                      |

var (
	// ParamsNarrow is the smallest configuration.
	ParamsNarrow = &Params{CauseSize: 1, CommonAddrSize: 1, InfoObjAddrSize: 1, InfoObjTimeZone: time.UTC}
	// ParamsWide is the largest configuration.
	ParamsWide = &Params{CauseSize: 2, CommonAddrSize: 2, InfoObjAddrSize: 3, InfoObjTimeZone: time.UTC}
)

// Params 定义了ASDU相关特定参数
// See companion standard 101, subclass 7.1.
type Params struct {
	// cause of transmission, 传输原因字节数
	// The standard requires "b" in [1, 2].
	// Value 2 includes/activates the originator address.
	CauseSize int
	// Originator Address [1, 255] or 0 for the default.
	// The applicability is controlled by Params.CauseSize.
	OrigAddress OriginAddr
	// size of ASDU common address， ASDU 公共地址字节数
	// 应用服务数据单元公共地址的八位位组数目,公共地址是站地址
	// The standard requires "a" in [1, 2].
	CommonAddrSize int

	// size of ASDU information object address. 信息对象地址字节数
	// The standard requires "c" in [1, 3].
	InfoObjAddrSize int

	// InfoObjTimeZone controls the time tag interpretation.
	// The standard fails to mention this one.
	InfoObjTimeZone *time.Location
}

// Valid returns the validation result of params.
func (sf Params) Valid() error {
	if (sf.CauseSize < 1 || sf.CauseSize > 2) ||
		(sf.CommonAddrSize < 1 || sf.CommonAddrSize > 2) ||
		(sf.InfoObjAddrSize < 1 || sf.InfoObjAddrSize > 3) ||
		(sf.InfoObjTimeZone == nil) {
		return ErrParam
	}
	return nil
}

// ValidCommonAddr returns the validation result of a station common address.
func (sf Params) ValidCommonAddr(addr CommonAddr) error {
	if addr == InvalidCommonAddr {
		return ErrCommonAddrZero
	}
	if bits.Len(uint(addr)) > sf.CommonAddrSize*8 {
		return ErrCommonAddrFit
	}
	return nil
}

// IdentifierSize return the application service data unit identifies size
func (sf Params) IdentifierSize() int {
	return 2 + int(sf.CauseSize) + int(sf.CommonAddrSize)
}

// Identifier the application service data unit identifies.
type Identifier struct {
	// type identification, information content
	Type TypeID
	// Variable is variable structure qualifier
	Variable VariableStruct
	// cause of transmission submission category
	Coa CauseOfTransmission
	// Originator Address [1, 255] or 0 for the default.
	// The applicability is controlled by Params.CauseSize.
	OrigAddr OriginAddr
	// CommonAddr is a station address. Zero is not used.
	// The width is controlled by Params.CommonAddrSize.
	// See companion standard 101, subclass 7.2.4.
	CommonAddr CommonAddr // station address
}

// String returns the information of data unit identifier, e.g.: "TypeID Cause OrigAddr@CommonAddr"
func (id Identifier) String() string {
	if id.OrigAddr == 0 {
		return fmt.Sprintf("TID<%s> COT<%s> @%d", id.Type, id.Coa, id.CommonAddr)
	}
	return fmt.Sprintf("TID<%s> COT<%s> %d@%d ", id.Type, id.Coa, id.OrigAddr, id.CommonAddr)
}

// ASDU (Application Service Data Unit) is an application message.
type ASDU struct {
	*Params
	Identifier
	infoObj   []byte            // information object serial
	bootstrap [ASDUSizeMax]byte // prevents Info malloc
}

// NewEmptyASDU new empty asdu with special params
func NewEmptyASDU(p *Params) *ASDU {
	a := &ASDU{Params: p}
	lenDUI := a.IdentifierSize()
	a.infoObj = a.bootstrap[lenDUI:lenDUI]
	return a
}

// NewASDU new asdu with special params and identifier
func NewASDU(p *Params, identifier Identifier) *ASDU {
	a := NewEmptyASDU(p)
	a.Identifier = identifier
	return a
}

// Clone deep clone asdu
func (sf *ASDU) Clone() *ASDU {
	r := NewASDU(sf.Params, sf.Identifier)
	r.infoObj = append(r.infoObj, sf.infoObj...)
	return r
}

// SetVariableNumber See companion standard 101, subclass 7.2.2.
func (sf *ASDU) SetVariableNumber(n int) error {
	if n >= 128 {
		return ErrInfoObjIndexFit
	}
	sf.Variable.Number = byte(n)
	return nil
}

// Respond returns a new "responding" ASDU which addresses "initiating" u.
//func (u *ASDU) Respond(t TypeID, c Cause) *ASDU {
//	return NewASDU(u.Params, Identifier{
//		CommonAddr: u.CommonAddr,
//		OrigAddr:   u.OrigAddr,
//		Type:       t,
//		Cause:      c | u.Cause&TestFlag,
//	})
//}

// Reply returns a new "responding" ASDU which addresses "initiating" addr with a copy of Info.
func (sf *ASDU) Reply(c Cause, addr CommonAddr) *ASDU {
	sf.CommonAddr = addr
	r := NewASDU(sf.Params, sf.Identifier)
	r.Coa.Cause = c
	r.infoObj = append(r.infoObj, sf.infoObj...)
	return r
}

// SendReplyMirror send a reply of the mirror request but cause different
func (sf *ASDU) SendReplyMirror(c Connect, cause Cause) error {
	r := NewASDU(sf.Params, sf.Identifier)
	r.Coa.Cause = cause
	r.infoObj = append(r.infoObj, sf.infoObj...)
	return c.Send(r)
}

// String returns a human-readable description of the ASDU without dumping raw byte arrays.
func (sf *ASDU) String() string {
	if sf == nil {
		return "<nil>"
	}
	var b strings.Builder
	// Header: Type, VSQ, Cause, Addresses
	b.WriteString(sf.Identifier.String())
	b.WriteByte(' ')
	b.WriteString("VSQ<" + sf.Variable.String() + ">")
	_, _ = fmt.Fprintf(&b, " IOA-Width=%d", sf.InfoObjAddrSize)

	// If there's no information object payload, return header
	if len(sf.infoObj) == 0 {
		return b.String()
	}

	// Work on a non-destructive copy of the infoObj by saving and restoring slice
	saved := sf.infoObj
	defer func() { sf.infoObj = saved }()

	switch sf.Type {
	// Monitored information (common)
	case M_SP_NA_1, M_SP_TA_1, M_SP_TB_1:
		infos := sf.GetSinglePoint()
		_, _ = fmt.Fprintf(&b, " items=%d", len(infos))
		for i, it := range infos {
			if i == 0 {
				b.WriteString(" [")
			} else {
				b.WriteString(", ")
			}
			_, _ = fmt.Fprintf(&b, "%d=%t", it.Ioa, it.Value)
			if it.Qds != QDSGood {
				_, _ = fmt.Fprintf(&b, " QDS=0x%02x", byte(it.Qds))
			}
			if !it.Time.IsZero() {
				_, _ = fmt.Fprintf(&b, " @%s", it.Time.Format(time.RFC3339Nano))
			}
		}
		if len(infos) > 0 {
			b.WriteByte(']')
		}
	case M_DP_NA_1, M_DP_TA_1, M_DP_TB_1:
		infos := sf.GetDoublePoint()
		_, _ = fmt.Fprintf(&b, " items=%d", len(infos))
		for i, it := range infos {
			if i == 0 {
				b.WriteString(" [")
			} else {
				b.WriteString(", ")
			}
			_, _ = fmt.Fprintf(&b, "%d=%d", it.Ioa, it.Value)
			if it.Qds != QDSGood {
				_, _ = fmt.Fprintf(&b, " QDS=0x%02x", byte(it.Qds))
			}
			if !it.Time.IsZero() {
				_, _ = fmt.Fprintf(&b, " @%s", it.Time.Format(time.RFC3339Nano))
			}
		}
		if len(infos) > 0 {
			b.WriteByte(']')
		}
	case M_ST_NA_1, M_ST_TA_1, M_ST_TB_1:
		infos := sf.GetStepPosition()
		_, _ = fmt.Fprintf(&b, " items=%d", len(infos))
		for i, it := range infos {
			if i == 0 {
				b.WriteString(" [")
			} else {
				b.WriteString(", ")
			}
			_, _ = fmt.Fprintf(&b, "%d=val(%d)", it.Ioa, it.Value.Val)
			if it.Value.HasTransient {
				b.WriteString(" transient")
			}
			if it.Qds != QDSGood {
				_, _ = fmt.Fprintf(&b, " QDS=0x%02x", byte(it.Qds))
			}
			if !it.Time.IsZero() {
				_, _ = fmt.Fprintf(&b, " @%s", it.Time.Format(time.RFC3339Nano))
			}
		}
		if len(infos) > 0 {
			b.WriteByte(']')
		}
	case M_BO_NA_1, M_BO_TA_1, M_BO_TB_1:
		infos := sf.GetBitString32()
		_, _ = fmt.Fprintf(&b, " items=%d", len(infos))
		for i, it := range infos {
			if i == 0 {
				b.WriteString(" [")
			} else {
				b.WriteString(", ")
			}
			_, _ = fmt.Fprintf(&b, "%d=0x%08x", it.Ioa, it.Value)
			if it.Qds != QDSGood {
				_, _ = fmt.Fprintf(&b, " QDS=0x%02x", byte(it.Qds))
			}
			if !it.Time.IsZero() {
				_, _ = fmt.Fprintf(&b, " @%s", it.Time.Format(time.RFC3339Nano))
			}
		}
		if len(infos) > 0 {
			b.WriteByte(']')
		}
	case M_ME_NA_1, M_ME_TA_1, M_ME_TD_1, M_ME_ND_1:
		infos := sf.GetMeasuredValueNormal()
		_, _ = fmt.Fprintf(&b, " items=%d", len(infos))
		for i, it := range infos {
			if i == 0 {
				b.WriteString(" [")
			} else {
				b.WriteString(", ")
			}
			_, _ = fmt.Fprintf(&b, "%d=%.6f", it.Ioa, it.Value.Float64())
			if it.Qds != QDSGood {
				_, _ = fmt.Fprintf(&b, " QDS=0x%02x", byte(it.Qds))
			}
			if !it.Time.IsZero() {
				_, _ = fmt.Fprintf(&b, " @%s", it.Time.Format(time.RFC3339Nano))
			}
		}
		if len(infos) > 0 {
			b.WriteByte(']')
		}
	case M_ME_NB_1, M_ME_TB_1, M_ME_TE_1:
		infos := sf.GetMeasuredValueScaled()
		_, _ = fmt.Fprintf(&b, " items=%d", len(infos))
		for i, it := range infos {
			if i == 0 {
				b.WriteString(" [")
			} else {
				b.WriteString(", ")
			}
			_, _ = fmt.Fprintf(&b, "%d=%d", it.Ioa, it.Value)
			if it.Qds != QDSGood {
				_, _ = fmt.Fprintf(&b, " QDS=0x%02x", byte(it.Qds))
			}
			if !it.Time.IsZero() {
				_, _ = fmt.Fprintf(&b, " @%s", it.Time.Format(time.RFC3339Nano))
			}
		}
		if len(infos) > 0 {
			b.WriteByte(']')
		}
	case M_ME_NC_1, M_ME_TC_1, M_ME_TF_1:
		infos := sf.GetMeasuredValueFloat()
		_, _ = fmt.Fprintf(&b, " items=%d", len(infos))
		for i, it := range infos {
			if i == 0 {
				b.WriteString(" [")
			} else {
				b.WriteString(", ")
			}
			_, _ = fmt.Fprintf(&b, "%d=%g", it.Ioa, it.Value)
			if it.Qds != QDSGood {
				_, _ = fmt.Fprintf(&b, " QDS=0x%02x", byte(it.Qds))
			}
			if !it.Time.IsZero() {
				_, _ = fmt.Fprintf(&b, " @%s", it.Time.Format(time.RFC3339Nano))
			}
		}
		if len(infos) > 0 {
			b.WriteByte(']')
		}
	case M_IT_NA_1, M_IT_TA_1, M_IT_TB_1:
		infos := sf.GetIntegratedTotals()
		_, _ = fmt.Fprintf(&b, " items=%d", len(infos))
		for i, it := range infos {
			if i == 0 {
				b.WriteString(" [")
			} else {
				b.WriteString(", ")
			}
			v := it.Value
			_, _ = fmt.Fprintf(&b, "%d=count(%d) seq=%d", it.Ioa, v.CounterReading, v.SeqNumber)
			if v.HasCarry {
				b.WriteString(" carry")
			}
			if v.IsAdjusted {
				b.WriteString(" adjusted")
			}
			if v.IsInvalid {
				b.WriteString(" invalid")
			}
			if !it.Time.IsZero() {
				_, _ = fmt.Fprintf(&b, " @%s", it.Time.Format(time.RFC3339Nano))
			}
		}
		if len(infos) > 0 {
			b.WriteByte(']')
		}
	case M_EP_TA_1, M_EP_TD_1:
		infos := sf.GetEventOfProtectionEquipment()
		_, _ = fmt.Fprintf(&b, " items=%d", len(infos))
		for i, it := range infos {
			if i == 0 {
				b.WriteString(" [")
			} else {
				b.WriteString(", ")
			}
			_, _ = fmt.Fprintf(&b, "%d=event(%d) QDP=0x%02x msec=%d", it.Ioa, it.Event, byte(it.Qdp), it.Msec)
			if !it.Time.IsZero() {
				_, _ = fmt.Fprintf(&b, " @%s", it.Time.Format(time.RFC3339Nano))
			}
		}
		if len(infos) > 0 {
			b.WriteByte(']')
		}
	case M_EP_TB_1, M_EP_TE_1:
		it := sf.GetPackedStartEventsOfProtectionEquipment()
		_, _ = fmt.Fprintf(&b, " IOA=%d start=0x%02x QDP=0x%02x msec=%d", it.Ioa, byte(it.Event), byte(it.Qdp), it.Msec)
		if !it.Time.IsZero() {
			_, _ = fmt.Fprintf(&b, " @%s", it.Time.Format(time.RFC3339Nano))
		}
	case M_EP_TC_1, M_EP_TF_1:
		it := sf.GetPackedOutputCircuitInfo()
		_, _ = fmt.Fprintf(&b, " IOA=%d oci=0x%02x QDP=0x%02x msec=%d", it.Ioa, byte(it.Oci), byte(it.Qdp), it.Msec)
		if !it.Time.IsZero() {
			_, _ = fmt.Fprintf(&b, " @%s", it.Time.Format(time.RFC3339Nano))
		}
	case M_PS_NA_1:
		infos := sf.GetPackedSinglePointWithSCD()
		_, _ = fmt.Fprintf(&b, " items=%d", len(infos))
		for i, it := range infos {
			if i == 0 {
				b.WriteString(" [")
			} else {
				b.WriteString(", ")
			}
			_, _ = fmt.Fprintf(&b, "%d=SCD(0x%08x) QDS=0x%02x", it.Ioa, uint32(it.Scd), byte(it.Qds))
		}
		if len(infos) > 0 {
			b.WriteByte(']')
		}

	// System and control directions
	case M_EI_NA_1:
		ioa, coi := sf.GetEndOfInitialization()
		_, _ = fmt.Fprintf(&b, " IOA=%d cause=%d localChange=%t", ioa, coi.Cause, coi.IsLocalChange)
	case C_SC_NA_1, C_SC_TA_1:
		cmd := sf.GetSingleCmd()
		_, _ = fmt.Fprintf(&b, " IOA=%d val=%t QOC=0x%02x", cmd.Ioa, cmd.Value, cmd.Qoc.Value())
		if !cmd.Time.IsZero() {
			_, _ = fmt.Fprintf(&b, " @%s", cmd.Time.Format(time.RFC3339Nano))
		}
	case C_DC_NA_1, C_DC_TA_1:
		cmd := sf.GetDoubleCmd()
		_, _ = fmt.Fprintf(&b, " IOA=%d val=%d QOC=0x%02x", cmd.Ioa, cmd.Value, cmd.Qoc.Value())
		if !cmd.Time.IsZero() {
			_, _ = fmt.Fprintf(&b, " @%s", cmd.Time.Format(time.RFC3339Nano))
		}
	case C_RC_NA_1, C_RC_TA_1:
		cmd := sf.GetStepCmd()
		_, _ = fmt.Fprintf(&b, " IOA=%d val=%d QOC=0x%02x", cmd.Ioa, cmd.Value, cmd.Qoc.Value())
		if !cmd.Time.IsZero() {
			_, _ = fmt.Fprintf(&b, " @%s", cmd.Time.Format(time.RFC3339Nano))
		}
	case C_SE_NA_1, C_SE_TA_1:
		cmd := sf.GetSetpointNormalCmd()
		_, _ = fmt.Fprintf(&b, " IOA=%d val=%.6f QOS=0x%02x", cmd.Ioa, cmd.Value.Float64(), byte(cmd.Qos.Value()))
		if !cmd.Time.IsZero() {
			_, _ = fmt.Fprintf(&b, " @%s", cmd.Time.Format(time.RFC3339Nano))
		}
	case C_SE_NB_1, C_SE_TB_1:
		cmd := sf.GetSetpointCmdScaled()
		_, _ = fmt.Fprintf(&b, " IOA=%d val=%d QOS=0x%02x", cmd.Ioa, cmd.Value, byte(cmd.Qos.Value()))
		if !cmd.Time.IsZero() {
			_, _ = fmt.Fprintf(&b, " @%s", cmd.Time.Format(time.RFC3339Nano))
		}
	case C_SE_NC_1, C_SE_TC_1:
		cmd := sf.GetSetpointFloatCmd()
		_, _ = fmt.Fprintf(&b, " IOA=%d val=%g QOS=0x%02x", cmd.Ioa, cmd.Value, byte(cmd.Qos.Value()))
		if !cmd.Time.IsZero() {
			_, _ = fmt.Fprintf(&b, " @%s", cmd.Time.Format(time.RFC3339Nano))
		}
	case C_BO_NA_1, C_BO_TA_1:
		cmd := sf.GetBitsString32Cmd()
		_, _ = fmt.Fprintf(&b, " IOA=%d bits=0x%08x", cmd.Ioa, cmd.Value)
		if !cmd.Time.IsZero() {
			_, _ = fmt.Fprintf(&b, " @%s", cmd.Time.Format(time.RFC3339Nano))
		}

	// Parameters
	case P_ME_NA_1:
		p := sf.GetParameterNormal()
		_, _ = fmt.Fprintf(&b, " IOA=%d val=%.6f QPM=0x%02x", p.Ioa, p.Value.Float64(), byte(p.Qpm.Value()))
	case P_ME_NB_1:
		p := sf.GetParameterScaled()
		_, _ = fmt.Fprintf(&b, " IOA=%d val=%d QPM=0x%02x", p.Ioa, p.Value, byte(p.Qpm.Value()))
	case P_ME_NC_1:
		p := sf.GetParameterFloat()
		_, _ = fmt.Fprintf(&b, " IOA=%d val=%g QPM=0x%02x", p.Ioa, p.Value, byte(p.Qpm.Value()))
	case P_AC_NA_1:
		p := sf.GetParameterActivation()
		_, _ = fmt.Fprintf(&b, " IOA=%d QPA=%d", p.Ioa, p.Qpa)

	// System command: Interrogation Command
	case C_IC_NA_1:
		ioa, qoi := sf.GetInterrogationCmd()
		_, _ = fmt.Fprintf(&b, " IOA=%d QOI=%d", ioa, byte(qoi))

	default:
		// Unknown or not yet formatted types: provide concise summary without dumping raw bytes
		n := int(sf.Variable.Number)
		if n == 0 {
			n = 1
		}
		_, _ = fmt.Fprintf(&b, " items=%d payload=%dB", n, len(sf.infoObj))
	}

	return b.String()
}

// MarshalBinary honors the encoding.BinaryMarshaler interface.
func (sf *ASDU) MarshalBinary() (data []byte, err error) {
	switch {
	case sf.Coa.Cause == Unused:
		return nil, ErrCauseZero
	case !(sf.CauseSize == 1 || sf.CauseSize == 2):
		return nil, ErrParam
	case sf.CauseSize == 1 && sf.OrigAddr != 0:
		return nil, ErrOriginAddrFit
	case sf.CommonAddr == InvalidCommonAddr:
		return nil, ErrCommonAddrZero
	case !(sf.CommonAddrSize == 1 || sf.CommonAddrSize == 2):
		return nil, ErrParam
	case sf.CommonAddrSize == 1 && sf.CommonAddr != GlobalCommonAddr && sf.CommonAddr >= 255:
		return nil, ErrParam
	}

	raw := sf.bootstrap[:(sf.IdentifierSize() + len(sf.infoObj))]
	raw[0] = byte(sf.Type)
	raw[1] = sf.Variable.Value()
	raw[2] = sf.Coa.Value()
	offset := 3
	if sf.CauseSize == 2 {
		raw[offset] = byte(sf.OrigAddr)
		offset++
	}
	if sf.CommonAddrSize == 1 {
		if sf.CommonAddr == GlobalCommonAddr {
			raw[offset] = 255
		} else {
			raw[offset] = byte(sf.CommonAddr)
		}
	} else { // 2
		raw[offset] = byte(sf.CommonAddr)
		offset++
		raw[offset] = byte(sf.CommonAddr >> 8)
	}
	return raw, nil
}

// UnmarshalBinary honors the encoding.BinaryUnmarshaler interface.
// ASDUParams must be set in advance. All other fields are initialized.
func (sf *ASDU) UnmarshalBinary(rawAsdu []byte) error {
	if !(sf.CauseSize == 1 || sf.CauseSize == 2) ||
		!(sf.CommonAddrSize == 1 || sf.CommonAddrSize == 2) {
		return ErrParam
	}

	// rawAsdu unit identifier size check
	lenDUI := sf.IdentifierSize()
	if lenDUI > len(rawAsdu) {
		return io.EOF
	}

	// parse rawAsdu unit identifier
	sf.Type = TypeID(rawAsdu[0])
	sf.Variable = ParseVariableStruct(rawAsdu[1])
	sf.Coa = ParseCauseOfTransmission(rawAsdu[2])
	if sf.CauseSize == 1 {
		sf.OrigAddr = 0
	} else {
		sf.OrigAddr = OriginAddr(rawAsdu[3])
	}
	if sf.CommonAddrSize == 1 {
		sf.CommonAddr = CommonAddr(rawAsdu[lenDUI-1])
		if sf.CommonAddr == 255 { // map 8-bit variant to 16-bit equivalent
			sf.CommonAddr = GlobalCommonAddr
		}
	} else { // 2
		sf.CommonAddr = CommonAddr(rawAsdu[lenDUI-2]) | CommonAddr(rawAsdu[lenDUI-1])<<8
	}
	// information object
	sf.infoObj = append(sf.bootstrap[lenDUI:lenDUI], rawAsdu[lenDUI:]...)
	return sf.fixInfoObjSize()
}

// fixInfoObjSize fix information object size
func (sf *ASDU) fixInfoObjSize() error {
	// fixed element size
	objSize, err := GetInfoObjSize(sf.Type)
	if err != nil {
		return err
	}

	var size int
	// read the variable structure qualifier
	if sf.Variable.IsSequence {
		size = sf.InfoObjAddrSize + int(sf.Variable.Number)*objSize
	} else {
		size = int(sf.Variable.Number) * (sf.InfoObjAddrSize + objSize)
	}

	switch {
	case size == 0:
		return ErrInfoObjIndexFit
	case size > len(sf.infoObj):
		return io.EOF
	case size < len(sf.infoObj): // not explicitly prohibited
		sf.infoObj = sf.infoObj[:size]
	}

	return nil
}

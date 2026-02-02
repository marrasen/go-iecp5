// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

// Package asdu provides the OSI presentation layer.
package asdu

import (
	"encoding/json"
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

// Params defines ASDU-specific parameters
// See companion standard 101, subclass 7.1.
type Params struct {
	// cause of transmission (byte size)
	// The standard requires "b" in [1, 2].
	// Value 2 includes/activates the originator address.
	CauseSize int
	// Originator Address [1, 255] or 0 for the default.
	// The applicability is controlled by Params.CauseSize.
	OrigAddress OriginAddr
	// size of ASDU common address (in bytes)
	// Number of octets of the ASDU common address; the common address is the station address
	// The standard requires "a" in [1, 2].
	CommonAddrSize int

	// size of ASDU information object address (in bytes)
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
	msg, err := ParseASDU(sf)
	if err != nil {
		n := int(sf.Variable.Number)
		if n == 0 {
			n = 1
		}
		_, _ = fmt.Fprintf(&b, " items=%d payload=%dB", n, len(sf.infoObj))
		return b.String()
	}

	b.WriteByte(' ')
	b.WriteString(msg.String())

	return b.String()
}

// MarshalJSON encodes ASDU into a JSON object with a dynamic "value" field.
/*
TypeScript types for the JSON produced by ASDU.MarshalJSON()

// Base header shared by all ASDUs
interface ASDUBase {
  type: TypeString;        // discriminant (TypeID string)
  variable: string;        // e.g. "sq,3" or "5"
  cause: string;           // e.g. "Spontaneous[,neg][,test]"
  origAddr: number;        // 0..255
  commonAddr: number;      // 1..65535 (65535=broadcast)
}

type TypeString =
  | "M_SP_NA_1" | "M_SP_TA_1" | "M_SP_TB_1"
  | "M_DP_NA_1" | "M_DP_TA_1" | "M_DP_TB_1"
  | "M_ST_NA_1" | "M_ST_TA_1" | "M_ST_TB_1"
  | "M_BO_NA_1" | "M_BO_TA_1" | "M_BO_TB_1"
  | "M_ME_NA_1" | "M_ME_TA_1" | "M_ME_TD_1" | "M_ME_ND_1"
  | "M_ME_NB_1" | "M_ME_TB_1" | "M_ME_TE_1"
  | "M_ME_NC_1" | "M_ME_TC_1" | "M_ME_TF_1"
  | "M_IT_NA_1" | "M_IT_TA_1" | "M_IT_TB_1"
  | "M_EP_TA_1" | "M_EP_TD_1"
  | "M_EP_TB_1" | "M_EP_TE_1"
  | "M_EP_TC_1" | "M_EP_TF_1"
  | "M_PS_NA_1"
  | "M_EI_NA_1"
  | "C_SC_NA_1" | "C_SC_TA_1"
  | "C_DC_NA_1" | "C_DC_TA_1"
  | "C_RC_NA_1" | "C_RC_TA_1"
  | "C_SE_NA_1" | "C_SE_TA_1"
  | "C_SE_NB_1" | "C_SE_TB_1"
  | "C_SE_NC_1" | "C_SE_TC_1"
  | "C_BO_NA_1" | "C_BO_TA_1"
  | "C_IC_NA_1"
  | string; // fallback for private/unknown types

// Helpers
interface Timed { time?: string } // RFC3339Nano; may be absent if not applicable

// Monitored info (arrays)
type M_SP = ASDUBase & {
  type: "M_SP_NA_1" | "M_SP_TA_1" | "M_SP_TB_1";
  value: Array<{ ioa: number; value: boolean; qds: number; } & Timed>;
};

type M_DP = ASDUBase & {
  type: "M_DP_NA_1" | "M_DP_TA_1" | "M_DP_TB_1";
  value: Array<{ ioa: number; value: number; qds: number; } & Timed>; // value in [0..3]
};

type M_ST = ASDUBase & {
  type: "M_ST_NA_1" | "M_ST_TA_1" | "M_ST_TB_1";
  value: Array<{ ioa: number; value: { val: number; transient: boolean }; qds: number; } & Timed>;
};

type M_BO = ASDUBase & {
  type: "M_BO_NA_1" | "M_BO_TA_1" | "M_BO_TB_1";
  value: Array<{ ioa: number; value: number; qds: number; } & Timed>;
};

type M_ME_Normal_WithQ = ASDUBase & {
  type: "M_ME_NA_1" | "M_ME_TA_1" | "M_ME_TD_1";
  value: Array<{ ioa: number; value: number; qds: number; } & Timed>;
};

type M_ME_Normal_NoQ = ASDUBase & {
  type: "M_ME_ND_1";
  value: Array<{ ioa: number; value: number; } & Timed>; // no qds
};

type M_ME_Scaled = ASDUBase & {
  type: "M_ME_NB_1" | "M_ME_TB_1" | "M_ME_TE_1";
  value: Array<{ ioa: number; value: number; qds: number; } & Timed>;
};

type M_ME_Float = ASDUBase & {
  type: "M_ME_NC_1" | "M_ME_TC_1" | "M_ME_TF_1";
  value: Array<{ ioa: number; value: number; qds: number; } & Timed>;
};

type M_IT = ASDUBase & {
  type: "M_IT_NA_1" | "M_IT_TA_1" | "M_IT_TB_1";
  value: Array<{ ioa: number; value: { count: number; seq: number; carry: boolean; adjusted: boolean; invalid: boolean }; } & Timed>;
};

type M_EP_List = ASDUBase & {
  type: "M_EP_TA_1" | "M_EP_TD_1";
  value: Array<{ ioa: number; event: number; qdp: number; msec: number; } & Timed>;
};

type M_EP_Start = ASDUBase & {
  type: "M_EP_TB_1" | "M_EP_TE_1";
  value: { ioa: number; event: number; qdp: number; msec: number; } & Timed;
};

type M_EP_Oci = ASDUBase & {
  type: "M_EP_TC_1" | "M_EP_TF_1";
  value: { ioa: number; oci: number; qdp: number; msec: number; } & Timed;
};

type M_PS = ASDUBase & {
  type: "M_PS_NA_1";
  value: Array<{ ioa: number; scd: number; qds: number }>;
};

type M_EI = ASDUBase & {
  type: "M_EI_NA_1";
  value: { ioa: number; cause: number; localChange: boolean };
};

// Control direction (single object)
type C_SC = ASDUBase & {
  type: "C_SC_NA_1" | "C_SC_TA_1";
  value: { ioa: number; value: boolean; qoc: number } & Timed;
};

type C_DC = ASDUBase & {
  type: "C_DC_NA_1" | "C_DC_TA_1";
  value: { ioa: number; value: number; qoc: number } & Timed;
};

type C_RC = ASDUBase & {
  type: "C_RC_NA_1" | "C_RC_TA_1";
  value: { ioa: number; value: number; qoc: number } & Timed;
};

type C_SE_Normal = ASDUBase & {
  type: "C_SE_NA_1" | "C_SE_TA_1";
  value: { ioa: number; value: number; qos: number } & Timed;
};

type C_SE_Scaled = ASDUBase & {
  type: "C_SE_NB_1" | "C_SE_TB_1";
  value: { ioa: number; value: number; qos: number } & Timed;
};

type C_SE_Float = ASDUBase & {
  type: "C_SE_NC_1" | "C_SE_TC_1";
  value: { ioa: number; value: number; qos: number } & Timed;
};

type C_BO = ASDUBase & {
  type: "C_BO_NA_1" | "C_BO_TA_1";
  value: { ioa: number; value: number } & Timed;
};

type C_IC = ASDUBase & {
  type: "C_IC_NA_1";
  value: { ioa: number; qoi: number };
};

// Fallback for unknown/private types
interface ASDUUnknown extends ASDUBase {
  value:
    | { items: number; payload: number } // default fallback used by marshaller for unknown types
    | any; // in case of future extensions
}

// Discriminated union for all known shapes
export type ASDUJson =
  | M_SP
  | M_DP
  | M_ST
  | M_BO
  | M_ME_Normal_WithQ
  | M_ME_Normal_NoQ
  | M_ME_Scaled
  | M_ME_Float
  | M_IT
  | M_EP_List
  | M_EP_Start
  | M_EP_Oci
  | M_PS
  | M_EI
  | C_SC
  | C_DC
  | C_RC
  | C_SE_Normal
  | C_SE_Scaled
  | C_SE_Float
  | C_BO
  | C_IC
  | ASDUUnknown;
*/
func (sf *ASDU) MarshalJSON() ([]byte, error) {
	if sf == nil {
		return []byte("null"), nil
	}
	// Helper for time formatting
	ts := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format(time.RFC3339Nano)
	}
	var value interface{}
	msg, err := ParseASDU(sf)
	if err != nil {
		value = map[string]interface{}{"items": int(sf.Variable.Number), "payload": len(sf.infoObj)}
	} else {
		switch m := msg.(type) {
		case *SinglePointMsg:
			arr := []map[string]interface{}{}
			for _, it := range m.Items {
				item := map[string]interface{}{"ioa": uint(it.Ioa), "value": it.Value, "qds": byte(it.Qds)}
				if !it.Time.IsZero() {
					item["time"] = ts(it.Time)
				}
				arr = append(arr, item)
			}
			value = arr
		case *DoublePointMsg:
			arr := []map[string]interface{}{}
			for _, it := range m.Items {
				item := map[string]interface{}{"ioa": uint(it.Ioa), "value": byte(it.Value), "qds": byte(it.Qds)}
				if !it.Time.IsZero() {
					item["time"] = ts(it.Time)
				}
				arr = append(arr, item)
			}
			value = arr
		case *StepPositionMsg:
			arr := []map[string]interface{}{}
			for _, it := range m.Items {
				item := map[string]interface{}{
					"ioa":   uint(it.Ioa),
					"value": map[string]interface{}{"val": it.Value.Val, "transient": it.Value.HasTransient},
					"qds":   byte(it.Qds),
				}
				if !it.Time.IsZero() {
					item["time"] = ts(it.Time)
				}
				arr = append(arr, item)
			}
			value = arr
		case *BitString32Msg:
			arr := []map[string]interface{}{}
			for _, it := range m.Items {
				item := map[string]interface{}{"ioa": uint(it.Ioa), "value": it.Value, "qds": byte(it.Qds)}
				if !it.Time.IsZero() {
					item["time"] = ts(it.Time)
				}
				arr = append(arr, item)
			}
			value = arr
		case *MeasuredValueNormalMsg:
			arr := []map[string]interface{}{}
			for _, it := range m.Items {
				item := map[string]interface{}{"ioa": uint(it.Ioa), "value": it.Value.Float64()}
				if m.TypeID() != M_ME_ND_1 {
					item["qds"] = byte(it.Qds)
				}
				if !it.Time.IsZero() {
					item["time"] = ts(it.Time)
				}
				arr = append(arr, item)
			}
			value = arr
		case *MeasuredValueScaledMsg:
			arr := []map[string]interface{}{}
			for _, it := range m.Items {
				item := map[string]interface{}{"ioa": uint(it.Ioa), "value": it.Value, "qds": byte(it.Qds)}
				if !it.Time.IsZero() {
					item["time"] = ts(it.Time)
				}
				arr = append(arr, item)
			}
			value = arr
		case *MeasuredValueFloatMsg:
			arr := []map[string]interface{}{}
			for _, it := range m.Items {
				item := map[string]interface{}{"ioa": uint(it.Ioa), "value": it.Value, "qds": byte(it.Qds)}
				if !it.Time.IsZero() {
					item["time"] = ts(it.Time)
				}
				arr = append(arr, item)
			}
			value = arr
		case *IntegratedTotalsMsg:
			arr := []map[string]interface{}{}
			for _, it := range m.Items {
				v := it.Value
				item := map[string]interface{}{
					"ioa": uint(it.Ioa),
					"value": map[string]interface{}{
						"count":    v.CounterReading,
						"seq":      v.SeqNumber,
						"carry":    v.HasCarry,
						"adjusted": v.IsAdjusted,
						"invalid":  v.IsInvalid,
					},
				}
				if !it.Time.IsZero() {
					item["time"] = ts(it.Time)
				}
				arr = append(arr, item)
			}
			value = arr
		case *EventOfProtectionMsg:
			arr := []map[string]interface{}{}
			for _, it := range m.Items {
				item := map[string]interface{}{"ioa": uint(it.Ioa), "event": byte(it.Event), "qdp": byte(it.Qdp), "msec": it.Msec}
				if !it.Time.IsZero() {
					item["time"] = ts(it.Time)
				}
				arr = append(arr, item)
			}
			value = arr
		case *PackedStartEventsMsg:
			it := m.Item
			value = map[string]interface{}{"ioa": uint(it.Ioa), "event": byte(it.Event), "qdp": byte(it.Qdp), "msec": it.Msec, "time": ts(it.Time)}
		case *PackedOutputCircuitMsg:
			it := m.Item
			value = map[string]interface{}{"ioa": uint(it.Ioa), "oci": byte(it.Oci), "qdp": byte(it.Qdp), "msec": it.Msec, "time": ts(it.Time)}
		case *PackedSinglePointWithSCDMsg:
			arr := []map[string]interface{}{}
			for _, it := range m.Items {
				item := map[string]interface{}{"ioa": uint(it.Ioa), "scd": uint32(it.Scd), "qds": byte(it.Qds)}
				arr = append(arr, item)
			}
			value = arr
		case *EndOfInitMsg:
			value = map[string]interface{}{"ioa": uint(m.IOA), "cause": byte(m.COI.Cause), "localChange": m.COI.IsLocalChange}
		case *SingleCommandMsg:
			cmd := m.Cmd
			value = map[string]interface{}{"ioa": uint(cmd.Ioa), "value": cmd.Value, "qoc": cmd.Qoc.Value(), "time": ts(cmd.Time)}
		case *DoubleCommandMsg:
			cmd := m.Cmd
			value = map[string]interface{}{"ioa": uint(cmd.Ioa), "value": byte(cmd.Value), "qoc": cmd.Qoc.Value(), "time": ts(cmd.Time)}
		case *StepCommandMsg:
			cmd := m.Cmd
			value = map[string]interface{}{"ioa": uint(cmd.Ioa), "value": byte(cmd.Value), "qoc": cmd.Qoc.Value(), "time": ts(cmd.Time)}
		case *SetpointNormalMsg:
			cmd := m.Cmd
			value = map[string]interface{}{"ioa": uint(cmd.Ioa), "value": cmd.Value.Float64(), "qos": cmd.Qos.Value(), "time": ts(cmd.Time)}
		case *SetpointScaledMsg:
			cmd := m.Cmd
			value = map[string]interface{}{"ioa": uint(cmd.Ioa), "value": cmd.Value, "qos": cmd.Qos.Value(), "time": ts(cmd.Time)}
		case *SetpointFloatMsg:
			cmd := m.Cmd
			value = map[string]interface{}{"ioa": uint(cmd.Ioa), "value": cmd.Value, "qos": cmd.Qos.Value(), "time": ts(cmd.Time)}
		case *BitsString32CmdMsg:
			cmd := m.Cmd
			value = map[string]interface{}{"ioa": uint(cmd.Ioa), "value": cmd.Value, "time": ts(cmd.Time)}
		case *InterrogationCmdMsg:
			value = map[string]interface{}{"ioa": uint(m.IOA), "qoi": byte(m.QOI)}
		default:
			value = map[string]interface{}{"items": int(sf.Variable.Number), "payload": len(sf.infoObj)}
		}
	}

	out := map[string]interface{}{
		"type":       sf.Type,
		"variable":   sf.Variable,
		"cause":      sf.Coa,
		"origAddr":   sf.OrigAddr,
		"commonAddr": sf.CommonAddr,
		"value":      value,
	}
	return json.Marshal(out)
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

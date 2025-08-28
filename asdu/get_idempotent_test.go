package asdu

import (
	"reflect"
	"testing"
)

// helper to unmarshal with wide params
func mustUnmarshal(t *testing.T, raw []byte) *ASDU {
	t.Helper()
	a := NewEmptyASDU(ParamsWide)
	if err := a.UnmarshalBinary(raw); err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}
	return a
}

// helper to unmarshal with custom params
func mustUnmarshalWithParams(t *testing.T, p *Params, raw []byte) *ASDU {
	t.Helper()
	a := NewEmptyASDU(p)
	if err := a.UnmarshalBinary(raw); err != nil {
		t.Fatalf("UnmarshalBinary failed: %v", err)
	}
	return a
}

func cloneBytes(b []byte) []byte { c := make([]byte, len(b)); copy(c, b); return c }

func marshal(t *testing.T, a *ASDU) []byte {
	t.Helper()
	b, err := a.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}
	return cloneBytes(b)
}

// build minimal raw for a given header and payload
func buildRaw(params *Params, id Identifier, payload []byte) []byte {
	a := NewASDU(params, id)
	a.infoObj = append(a.infoObj, payload...)
	b, _ := a.MarshalBinary()
	return cloneBytes(b)
}

func TestGetSinglePoint_Idempotent(t *testing.T) {
	// one element, no time, IOA size 3
	id := Identifier{Type: M_SP_NA_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Spontaneous}, CommonAddr: 1}
	ioa := InfoObjAddr(0x010203)
	payload := []byte{byte(ioa), byte(ioa >> 8), byte(ioa >> 16), 0x01 /*on + qds=0*/}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)

	before := marshal(t, a)
	v1 := a.GetSinglePoint()
	v2 := a.GetSinglePoint()
	if !reflect.DeepEqual(v1, v2) {
		t.Fatalf("values differ on second call: %#v vs %#v", v1, v2)
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by GetSinglePoint; before %x after %x", before, after)
	}
}

func TestGetMeasuredValueScaled_Idempotent(t *testing.T) {
	id := Identifier{Type: M_ME_NB_1, Variable: VariableStruct{IsSequence: false, Number: 2}, Coa: CauseOfTransmission{Cause: Periodic}, CommonAddr: 2}
	// two objects, ioa=5 and 6, scaled values 100,-1, QDS good
	payload := []byte{5, 0, 0, 100, 0, 0, 6, 0, 0, 255, 255, 0}
	raw := buildRaw(ParamsNarrow, id, payload)
	a := NewEmptyASDU(ParamsNarrow)
	if err := a.UnmarshalBinary(raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	before := marshal(t, a)
	v1 := a.GetMeasuredValueScaled()
	v2 := a.GetMeasuredValueScaled()
	if !reflect.DeepEqual(v1, v2) {
		t.Fatalf("values differ on second call")
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by GetMeasuredValueScaled")
	}
}

func TestGetBitString32_Idempotent(t *testing.T) {
	id := Identifier{Type: M_BO_NA_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Background}, CommonAddr: 3}
	ioa := InfoObjAddr(7)
	payload := []byte{byte(ioa), 0, 0, 0x78, 0x56, 0x34, 0x12, 0x10}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)
	before := marshal(t, a)
	v1 := a.GetBitString32()
	v2 := a.GetBitString32()
	if !reflect.DeepEqual(v1, v2) {
		t.Fatalf("values differ on second call")
	}
	if v1[0].Value != 0x12345678 {
		t.Fatalf("unexpected value: %x", v1[0].Value)
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by GetBitString32")
	}
}

func TestGetIntegratedTotals_Idempotent(t *testing.T) {
	id := Identifier{Type: M_IT_NA_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Spontaneous}, CommonAddr: 9}
	ioa := InfoObjAddr(1)
	// BinaryCounterReading: 4 bytes counter + status byte
	payload := []byte{byte(ioa), 0, 0, 0x11, 0x22, 0x33, 0x44, 0x5f}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)
	before := marshal(t, a)
	v1 := a.GetIntegratedTotals()
	v2 := a.GetIntegratedTotals()
	if !reflect.DeepEqual(v1, v2) {
		t.Fatalf("values differ on second call")
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by GetIntegratedTotals")
	}
}

func TestGetEventOfProtectionEquipment_Idempotent(t *testing.T) {
	id := Identifier{Type: M_EP_TA_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Spontaneous}, CommonAddr: 1}
	ioa := InfoObjAddr(10)
	// value byte, CP16 (2 bytes), CP24 (3 bytes)
	payload := []byte{byte(ioa), 0, 0, 0x03, 0x34, 0x12, 0x01, 0x02, 0x03}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)
	before := marshal(t, a)
	v1 := a.GetEventOfProtectionEquipment()
	v2 := a.GetEventOfProtectionEquipment()
	if !reflect.DeepEqual(v1, v2) {
		t.Fatalf("values differ on second call")
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by GetEventOfProtectionEquipment")
	}
}

func TestGetPackedStartEvents_Idempotent(t *testing.T) {
	id := Identifier{Type: M_EP_TB_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Spontaneous}, CommonAddr: 1}
	ioa := InfoObjAddr(3)
	payload := []byte{byte(ioa), 0, 0, 0xAA, 0x11 /*QDP*/, 0x00, 0x78, 0x56 /*CP16*/, 0x01, 0x02, 0x03 /*CP24*/}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)
	before := marshal(t, a)
	v1 := a.GetPackedStartEventsOfProtectionEquipment()
	v2 := a.GetPackedStartEventsOfProtectionEquipment()
	if !reflect.DeepEqual(v1, v2) {
		t.Fatalf("values differ on second call")
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by GetPackedStartEventsOfProtectionEquipment")
	}
}

func TestGetPackedOutputCircuitInfo_Idempotent(t *testing.T) {
	id := Identifier{Type: M_EP_TC_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Spontaneous}, CommonAddr: 1}
	ioa := InfoObjAddr(4)
	payload := []byte{byte(ioa), 0, 0, 0x0F, 0x01 /*QDP*/, 0x00, 0x34, 0x12 /*CP16*/, 0x01, 0x02, 0x03 /*CP24*/}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)
	before := marshal(t, a)
	v1 := a.GetPackedOutputCircuitInfo()
	v2 := a.GetPackedOutputCircuitInfo()
	if !reflect.DeepEqual(v1, v2) {
		t.Fatalf("values differ on second call")
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by GetPackedOutputCircuitInfo")
	}
}

func TestSystemGetters_Idempotent(t *testing.T) {
	// C_IC_NA_1 interrogation
	id := Identifier{Type: C_IC_NA_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Activation}, CommonAddr: 1}
	ioa := InfoObjAddrIrrelevant
	payload := []byte{byte(ioa), byte(QOIStation)}
	raw := buildRaw(ParamsNarrow, id, payload)
	a := mustUnmarshalWithParams(t, ParamsNarrow, raw)
	before := marshal(t, a)
	_, _ = a.GetInterrogationCmd()
	_, _ = a.GetInterrogationCmd()
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by GetInterrogationCmd")
	}

	// C_TS_TA_1 test with CP56
	id = Identifier{Type: C_TS_TA_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Activation}, CommonAddr: 1}
	payload = []byte{0 /*IOA*/, 0xAA, 0x55 /*test word*/, 0, 0, 0, 0, 0, 0, 0}
	raw = buildRaw(ParamsNarrow, id, payload)
	a = mustUnmarshalWithParams(t, ParamsNarrow, raw)
	before = marshal(t, a)
	_, _, _ = a.GetTestCommandCP56Time2a()
	_, _, _ = a.GetTestCommandCP56Time2a()
	after = marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by GetTestCommandCP56Time2a")
	}
}

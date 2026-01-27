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

func TestParseASDU_IdempotentSinglePoint(t *testing.T) {
	id := Identifier{Type: M_SP_NA_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Spontaneous}, CommonAddr: 1}
	ioa := InfoObjAddr(0x010203)
	payload := []byte{byte(ioa), byte(ioa >> 8), byte(ioa >> 16), 0x01}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)

	before := marshal(t, a)
	msg1, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	msg2, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	m1 := msg1.(SinglePointMsg)
	m2 := msg2.(SinglePointMsg)
	if !reflect.DeepEqual(m1, m2) {
		t.Fatalf("values differ on second call: %#v vs %#v", m1, m2)
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by ParseASDU; before %x after %x", before, after)
	}
}

func TestParseASDU_IdempotentMeasuredValueScaled(t *testing.T) {
	id := Identifier{Type: M_ME_NB_1, Variable: VariableStruct{IsSequence: false, Number: 2}, Coa: CauseOfTransmission{Cause: Periodic}, CommonAddr: 2}
	payload := []byte{5, 0, 0, 100, 0, 0, 6, 0, 0, 255, 255, 0}
	raw := buildRaw(ParamsNarrow, id, payload)
	a := mustUnmarshalWithParams(t, ParamsNarrow, raw)
	before := marshal(t, a)

	msg1, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	msg2, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	m1 := msg1.(MeasuredValueScaledMsg)
	m2 := msg2.(MeasuredValueScaledMsg)
	if !reflect.DeepEqual(m1, m2) {
		t.Fatalf("values differ on second call")
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by ParseASDU")
	}
}

func TestParseASDU_IdempotentBitString32(t *testing.T) {
	id := Identifier{Type: M_BO_NA_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Background}, CommonAddr: 3}
	ioa := InfoObjAddr(7)
	payload := []byte{byte(ioa), 0, 0, 0x78, 0x56, 0x34, 0x12, 0x10}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)
	before := marshal(t, a)

	msg1, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	msg2, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	m1 := msg1.(BitString32Msg)
	m2 := msg2.(BitString32Msg)
	if !reflect.DeepEqual(m1, m2) {
		t.Fatalf("values differ on second call")
	}
	if m1.Items[0].Value != 0x12345678 {
		t.Fatalf("unexpected value: %x", m1.Items[0].Value)
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by ParseASDU")
	}
}

func TestParseASDU_IdempotentIntegratedTotals(t *testing.T) {
	id := Identifier{Type: M_IT_NA_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Spontaneous}, CommonAddr: 9}
	ioa := InfoObjAddr(1)
	payload := []byte{byte(ioa), 0, 0, 0x11, 0x22, 0x33, 0x44, 0x5f}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)
	before := marshal(t, a)

	msg1, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	msg2, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	m1 := msg1.(IntegratedTotalsMsg)
	m2 := msg2.(IntegratedTotalsMsg)
	if !reflect.DeepEqual(m1, m2) {
		t.Fatalf("values differ on second call")
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by ParseASDU")
	}
}

func TestParseASDU_IdempotentEventOfProtection(t *testing.T) {
	id := Identifier{Type: M_EP_TA_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Spontaneous}, CommonAddr: 1}
	ioa := InfoObjAddr(10)
	payload := []byte{byte(ioa), 0, 0, 0x03, 0x34, 0x12, 0x01, 0x02, 0x03}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)
	before := marshal(t, a)

	msg1, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	msg2, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	m1 := msg1.(EventOfProtectionMsg)
	m2 := msg2.(EventOfProtectionMsg)
	if !reflect.DeepEqual(m1, m2) {
		t.Fatalf("values differ on second call")
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by ParseASDU")
	}
}

func TestParseASDU_IdempotentPackedStartEvents(t *testing.T) {
	id := Identifier{Type: M_EP_TB_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Spontaneous}, CommonAddr: 1}
	ioa := InfoObjAddr(3)
	payload := []byte{byte(ioa), 0, 0, 0xAA, 0x11, 0x00, 0x78, 0x56, 0x01, 0x02, 0x03}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)
	before := marshal(t, a)

	msg1, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	msg2, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	m1 := msg1.(PackedStartEventsMsg)
	m2 := msg2.(PackedStartEventsMsg)
	if !reflect.DeepEqual(m1, m2) {
		t.Fatalf("values differ on second call")
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by ParseASDU")
	}
}

func TestParseASDU_IdempotentPackedOutputCircuit(t *testing.T) {
	id := Identifier{Type: M_EP_TC_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Spontaneous}, CommonAddr: 1}
	ioa := InfoObjAddr(4)
	payload := []byte{byte(ioa), 0, 0, 0x0F, 0x01, 0x00, 0x34, 0x12, 0x01, 0x02, 0x03}
	raw := buildRaw(ParamsWide, id, payload)
	a := mustUnmarshal(t, raw)
	before := marshal(t, a)

	msg1, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	msg2, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	m1 := msg1.(PackedOutputCircuitMsg)
	m2 := msg2.(PackedOutputCircuitMsg)
	if !reflect.DeepEqual(m1, m2) {
		t.Fatalf("values differ on second call")
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by ParseASDU")
	}
}

func TestParseASDU_IdempotentSystem(t *testing.T) {
	id := Identifier{Type: C_IC_NA_1, Variable: VariableStruct{IsSequence: false, Number: 1}, Coa: CauseOfTransmission{Cause: Activation}, CommonAddr: 1}
	payload := []byte{0x00, byte(QOIStation)}
	raw := buildRaw(ParamsNarrow, id, payload)
	a := mustUnmarshalWithParams(t, ParamsNarrow, raw)
	before := marshal(t, a)

	msg1, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	msg2, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	m1 := msg1.(InterrogationCmdMsg)
	m2 := msg2.(InterrogationCmdMsg)
	if !reflect.DeepEqual(m1, m2) {
		t.Fatalf("values differ on second call")
	}
	after := marshal(t, a)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("ASDU mutated by ParseASDU")
	}
}

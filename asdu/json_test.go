package asdu

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTypeID_JSON(t *testing.T) {
	// marshal named constant
	b, err := json.Marshal(M_SP_NA_1)
	if err != nil {
		t.Fatalf("marshal TypeID: %v", err)
	}
	if string(b) != `"M_SP_NA_1"` {
		t.Fatalf("unexpected json: %s", string(b))
	}
	// unmarshal from name
	var tid TypeID
	if err := json.Unmarshal([]byte(`"M_SP_NA_1"`), &tid); err != nil {
		t.Fatalf("unmarshal TypeID by name: %v", err)
	}
	if tid != M_SP_NA_1 {
		t.Fatalf("want %v got %v", M_SP_NA_1, tid)
	}
	// unmarshal from numeric string
	if err := json.Unmarshal([]byte(`"1"`), &tid); err != nil {
		t.Fatalf("unmarshal TypeID by numeric string: %v", err)
	}
	if tid != M_SP_NA_1 {
		t.Fatalf("want %v got %v", M_SP_NA_1, tid)
	}
	// unmarshal from number
	if err := json.Unmarshal([]byte(`1`), &tid); err != nil {
		t.Fatalf("unmarshal TypeID number: %v", err)
	}
	if tid != M_SP_NA_1 {
		t.Fatalf("want %v got %v", M_SP_NA_1, tid)
	}
}

func TestVariableStruct_JSON(t *testing.T) {
	vs := VariableStruct{Number: 5}
	b, err := json.Marshal(vs)
	if err != nil {
		t.Fatalf("marshal VariableStruct: %v", err)
	}
	if string(b) != `"5"` {
		t.Fatalf("unexpected json: %s", string(b))
	}
	var got VariableStruct
	if err := json.Unmarshal([]byte(`"sq,7"`), &got); err != nil {
		t.Fatalf("unmarshal VariableStruct: %v", err)
	}
	if !(got.IsSequence && got.Number == 7) {
		t.Fatalf("want seq,7 got %+v", got)
	}
	if err := json.Unmarshal([]byte(`9`), &got); err != nil {
		t.Fatalf("unmarshal VariableStruct number: %v", err)
	}
	if got.IsSequence || got.Number != 9 {
		t.Fatalf("want 9 got %+v", got)
	}
}

func TestCauseOfTransmission_JSON(t *testing.T) {
	c := CauseOfTransmission{Cause: Spontaneous, IsNegative: true, IsTest: true}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal Cause: %v", err)
	}
	if string(b) != `"Spontaneous,neg,test"` {
		t.Fatalf("unexpected json: %s", string(b))
	}
	var got CauseOfTransmission
	if err := json.Unmarshal([]byte(`"Spontaneous,neg,test"`), &got); err != nil {
		t.Fatalf("unmarshal Cause: %v", err)
	}
	if !(got.Cause == Spontaneous && got.IsNegative && got.IsTest) {
		t.Fatalf("unexpected value: %+v", got)
	}
	// numeric fallback
	if err := json.Unmarshal([]byte(`3`), &got); err != nil {
		t.Fatalf("unmarshal numeric: %v", err)
	}
	if got.Cause != Spontaneous {
		t.Fatalf("want Spontaneous got %v", got.Cause)
	}
}

// helper to create a base ASDU
func newBaseASDU(p *Params, t TypeID, n byte) *ASDU {
	a := NewASDU(p, Identifier{
		Type:       t,
		Variable:   VariableStruct{Number: n},
		Coa:        CauseOfTransmission{Cause: Spontaneous},
		OrigAddr:   0,
		CommonAddr: 0x80,
	})
	return a
}

func TestASDU_MarshalJSON_MeasuredValueNormalNoQ(t *testing.T) {
	a := newBaseASDU(ParamsNarrow, M_ME_ND_1, 1)
	// payload: IOA + Normalize (no qds, no time)
	if err := a.AppendInfoObjAddr(10); err != nil {
		t.Fatalf("addr: %v", err)
	}
	a.AppendNormalize(Normalize(16384)) // 0.5

	b, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal asdu: %v", err)
	}
	// decode into generic map for assertions
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["type"] != "M_ME_ND_1" {
		t.Fatalf("type: %v", m["type"])
	}
	if m["variable"] != "1" {
		t.Fatalf("variable: %v", m["variable"])
	}
	if m["cause"] != "Spontaneous" {
		t.Fatalf("cause: %v", m["cause"])
	}
	if m["origAddr"].(float64) != 0 {
		t.Fatalf("origAddr: %v", m["origAddr"])
	}
	if m["commonAddr"].(float64) != 128 {
		t.Fatalf("commonAddr: %v", m["commonAddr"])
	}
	// value array
	vals, ok := m["value"].([]interface{})
	if !ok || len(vals) != 1 {
		t.Fatalf("value len: %T %v", m["value"], m["value"])
	}
	item := vals[0].(map[string]interface{})
	if item["ioa"].(float64) != 10 {
		t.Fatalf("ioa: %v", item["ioa"])
	}
	if _, hasQ := item["qds"]; hasQ {
		t.Fatalf("unexpected qds present")
	}
	if _, hasT := item["time"]; hasT {
		t.Fatalf("unexpected time present")
	}
	// value should be roughly 0.5
	if v := item["value"].(float64); v < 0.499 || v > 0.501 {
		t.Fatalf("value: %v", v)
	}
}

func TestASDU_MarshalJSON_SinglePointWithTime(t *testing.T) {
	a := newBaseASDU(ParamsNarrow, M_SP_TB_1, 1)
	// payload: IOA + val|qds + CP56Time2a
	if err := a.AppendInfoObjAddr(1); err != nil {
		t.Fatalf("addr: %v", err)
	}
	// value true with QDSBlocked
	a.AppendBytes(0x01 | byte(QDSBlocked))
	// timestamp
	utc := time.UTC
	ts := time.Date(2025, 8, 25, 12, 34, 56, 789*1e6, utc)
	a.AppendCP56Time2a(ts, utc)

	b, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal asdu: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["type"] != "M_SP_TB_1" {
		t.Fatalf("type: %v", m["type"])
	}
	vals := m["value"].([]interface{})
	item := vals[0].(map[string]interface{})
	if item["ioa"].(float64) != 1 {
		t.Fatalf("ioa: %v", item["ioa"])
	}
	if item["value"].(bool) != true {
		t.Fatalf("value: %v", item["value"])
	}
	if int(item["qds"].(float64)) != int(QDSBlocked) {
		t.Fatalf("qds: %v", item["qds"])
	}
	if _, ok := item["time"].(string); !ok {
		t.Fatalf("time missing or not string: %v", item["time"])
	}
}

func TestASDU_MarshalJSON_ControlSingleCommand(t *testing.T) {
	a := newBaseASDU(ParamsNarrow, C_SC_NA_1, 1)
	if err := a.AppendInfoObjAddr(5); err != nil {
		t.Fatalf("addr: %v", err)
	}
	qoc := QualifierOfCommand{Qual: QOCShortPulseDuration, InSelect: true}
	val := qoc.Value() | 0x01 // command true
	a.AppendBytes(val)

	b, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal asdu: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["type"] != "C_SC_NA_1" {
		t.Fatalf("type: %v", m["type"])
	}
	obj := m["value"].(map[string]interface{})
	if obj["ioa"].(float64) != 5 {
		t.Fatalf("ioa: %v", obj["ioa"])
	}
	if obj["value"].(bool) != true {
		t.Fatalf("value: %v", obj["value"])
	}
	if int(obj["qoc"].(float64)) != int(qoc.Value()) {
		t.Fatalf("qoc: %v", obj["qoc"])
	}
}

func TestASDU_MarshalJSON_UnknownTypeFallback(t *testing.T) {
	a := newBaseASDU(ParamsNarrow, TypeID(200), 2)
	// arbitrary payload bytes (not necessarily valid IOA etc.)
	a.AppendBytes(0xAA, 0xBB, 0xCC)
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal asdu: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["type"] != "200" {
		t.Fatalf("type: %v", m["type"])
	}
	val := m["value"].(map[string]interface{})
	if int(val["items"].(float64)) != 2 {
		t.Fatalf("items: %v", val["items"])
	}
	if int(val["payload"].(float64)) != 3 {
		t.Fatalf("payload: %v", val["payload"])
	}
}

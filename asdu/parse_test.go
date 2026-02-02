package asdu

import (
	"math"
	"testing"
	"time"
)

func ioaBytes(ioa InfoObjAddr) []byte {
	return []byte{byte(ioa), byte(ioa >> 8), byte(ioa >> 16)}
}

func newASDUForParse(tid TypeID, vs VariableStruct, payload []byte) *ASDU {
	a := NewASDU(ParamsWide, Identifier{Type: tid, Variable: vs})
	a.infoObj = append(a.infoObj, payload...)
	return a
}

func mustParse(t *testing.T, a *ASDU) Message {
	t.Helper()
	msg, err := ParseASDU(a)
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	return msg
}

func TestParseASDU_MonitoringFamilies(t *testing.T) {
	tm := time.Date(2025, 8, 25, 12, 34, 56, 0, time.UTC)

	t.Run("SinglePoint", func(t *testing.T) {
		payload := append(append(ioaBytes(1), 0x11), append(ioaBytes(2), 0x10)...)
		a := newASDUForParse(M_SP_NA_1, VariableStruct{Number: 2}, payload)
		msg := mustParse(t, a).(*SinglePointMsg)
		if len(msg.Items) != 2 || msg.Items[0].Ioa != 1 || !msg.Items[0].Value {
			t.Fatalf("unexpected single point: %+v", msg.Items)
		}
	})

	t.Run("DoublePoint", func(t *testing.T) {
		payload := append(append(ioaBytes(1), 0x12), append(ioaBytes(2), 0x11)...)
		a := newASDUForParse(M_DP_NA_1, VariableStruct{Number: 2}, payload)
		msg := mustParse(t, a).(*DoublePointMsg)
		if len(msg.Items) != 2 || msg.Items[0].Value != DPIDeterminedOn {
			t.Fatalf("unexpected double point: %+v", msg.Items)
		}
	})

	t.Run("StepPosition", func(t *testing.T) {
		payload := append(append(ioaBytes(1), 0x01, 0x10), append(ioaBytes(2), 0x02, 0x10)...)
		a := newASDUForParse(M_ST_NA_1, VariableStruct{Number: 2}, payload)
		msg := mustParse(t, a).(*StepPositionMsg)
		if msg.Items[0].Value.Val != 1 || msg.Items[1].Value.Val != 2 {
			t.Fatalf("unexpected step position: %+v", msg.Items)
		}
	})

	t.Run("BitString32", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x78, 0x56, 0x34, 0x12, 0x10)
		a := newASDUForParse(M_BO_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*BitString32Msg)
		if msg.Items[0].Value != 0x12345678 {
			t.Fatalf("unexpected bitstring: %+v", msg.Items)
		}
	})

	t.Run("MeasuredNormal", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x01, 0x00)
		a := newASDUForParse(M_ME_ND_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*MeasuredValueNormalMsg)
		if msg.Items[0].Value != 1 {
			t.Fatalf("unexpected measured normal: %+v", msg.Items)
		}
	})

	t.Run("MeasuredScaled", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x64, 0x00, 0x10)
		a := newASDUForParse(M_ME_NB_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*MeasuredValueScaledMsg)
		if msg.Items[0].Value != 100 {
			t.Fatalf("unexpected measured scaled: %+v", msg.Items)
		}
	})

	t.Run("MeasuredFloat", func(t *testing.T) {
		bits := math.Float32bits(100)
		payload := append(ioaBytes(1), byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24), 0x10)
		a := newASDUForParse(M_ME_NC_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*MeasuredValueFloatMsg)
		if msg.Items[0].Value != 100 {
			t.Fatalf("unexpected measured float: %+v", msg.Items)
		}
	})

	t.Run("IntegratedTotals", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x01, 0x02, 0x03, 0x04, 0x85)
		a := newASDUForParse(M_IT_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*IntegratedTotalsMsg)
		if msg.Items[0].Value.CounterReading != 0x04030201 {
			t.Fatalf("unexpected integrated totals: %+v", msg.Items)
		}
	})

	t.Run("EventOfProtection", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x03, 0x10, 0x00)
		payload = append(payload, CP24Time2a(tm, time.UTC)...)
		a := newASDUForParse(M_EP_TA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*EventOfProtectionMsg)
		if msg.Items[0].Event != SEIndeterminate {
			t.Fatalf("unexpected event: %+v", msg.Items)
		}
	})

	t.Run("PackedStartEvents", func(t *testing.T) {
		payload := append(ioaBytes(1), 0xAA, 0x11, 0x10, 0x00)
		payload = append(payload, CP24Time2a(tm, time.UTC)...)
		a := newASDUForParse(M_EP_TB_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*PackedStartEventsMsg)
		if msg.Item.Event != StartEvent(0xAA) {
			t.Fatalf("unexpected start events: %+v", msg.Item)
		}
	})

	t.Run("PackedOutputCircuit", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x0F, 0x11, 0x10, 0x00)
		payload = append(payload, CP24Time2a(tm, time.UTC)...)
		a := newASDUForParse(M_EP_TC_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*PackedOutputCircuitMsg)
		if msg.Item.Oci != OutputCircuitInfo(0x0F) {
			t.Fatalf("unexpected output circuit: %+v", msg.Item)
		}
	})

	t.Run("PackedSinglePointWithSCD", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x04, 0x03, 0x02, 0x01, 0x10)
		a := newASDUForParse(M_PS_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*PackedSinglePointWithSCDMsg)
		if msg.Items[0].Scd != 0x01020304 {
			t.Fatalf("unexpected scd: %+v", msg.Items)
		}
	})

	t.Run("EndOfInitialization", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x01)
		a := newASDUForParse(M_EI_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*EndOfInitMsg)
		if msg.COI.Cause != COILocalHandReset {
			t.Fatalf("unexpected coi: %+v", msg)
		}
	})
}

func TestParseASDU_ControlFamilies(t *testing.T) {
	bits := math.Float32bits(100)

	t.Run("SingleCommand", func(t *testing.T) {
		qoc := QualifierOfCommand{Qual: QOCShortPulseDuration, InSelect: true}
		payload := append(ioaBytes(1), qoc.Value()|0x01)
		a := newASDUForParse(C_SC_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*SingleCommandMsg)
		if !msg.Cmd.Value || !msg.Cmd.Qoc.InSelect {
			t.Fatalf("unexpected single command: %+v", msg.Cmd)
		}
	})

	t.Run("DoubleCommand", func(t *testing.T) {
		qoc := QualifierOfCommand{Qual: QOCShortPulseDuration, InSelect: false}
		payload := append(ioaBytes(1), qoc.Value()|byte(DCOOn))
		a := newASDUForParse(C_DC_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*DoubleCommandMsg)
		if msg.Cmd.Value != DCOOn {
			t.Fatalf("unexpected double command: %+v", msg.Cmd)
		}
	})

	t.Run("StepCommand", func(t *testing.T) {
		qoc := QualifierOfCommand{Qual: QOCShortPulseDuration, InSelect: false}
		payload := append(ioaBytes(1), qoc.Value()|byte(SCOStepUP))
		a := newASDUForParse(C_RC_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*StepCommandMsg)
		if msg.Cmd.Value != SCOStepUP {
			t.Fatalf("unexpected step command: %+v", msg.Cmd)
		}
	})

	t.Run("SetpointNormal", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x34, 0x12, 0x01)
		a := newASDUForParse(C_SE_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*SetpointNormalMsg)
		if msg.Cmd.Value != 0x1234 {
			t.Fatalf("unexpected setpoint normal: %+v", msg.Cmd)
		}
	})

	t.Run("SetpointScaled", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x34, 0x12, 0x01)
		a := newASDUForParse(C_SE_NB_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*SetpointScaledMsg)
		if msg.Cmd.Value != 0x1234 {
			t.Fatalf("unexpected setpoint scaled: %+v", msg.Cmd)
		}
	})

	t.Run("SetpointFloat", func(t *testing.T) {
		payload := append(ioaBytes(1), byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24), 0x01)
		a := newASDUForParse(C_SE_NC_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*SetpointFloatMsg)
		if msg.Cmd.Value != 100 {
			t.Fatalf("unexpected setpoint float: %+v", msg.Cmd)
		}
	})

	t.Run("BitsString32Cmd", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x78, 0x56, 0x34, 0x12)
		a := newASDUForParse(C_BO_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*BitsString32CmdMsg)
		if msg.Cmd.Value != 0x12345678 {
			t.Fatalf("unexpected bits command: %+v", msg.Cmd)
		}
	})
}

func TestParseASDU_ParameterFamilies(t *testing.T) {
	bits := math.Float32bits(100)

	t.Run("ParameterNormal", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x34, 0x12, 0x01)
		a := newASDUForParse(P_ME_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*ParameterNormalMsg)
		if msg.Param.Value != 0x1234 {
			t.Fatalf("unexpected parameter normal: %+v", msg.Param)
		}
	})

	t.Run("ParameterScaled", func(t *testing.T) {
		payload := append(ioaBytes(1), 0x34, 0x12, 0x01)
		a := newASDUForParse(P_ME_NB_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*ParameterScaledMsg)
		if msg.Param.Value != 0x1234 {
			t.Fatalf("unexpected parameter scaled: %+v", msg.Param)
		}
	})

	t.Run("ParameterFloat", func(t *testing.T) {
		payload := append(ioaBytes(1), byte(bits), byte(bits>>8), byte(bits>>16), byte(bits>>24), 0x01)
		a := newASDUForParse(P_ME_NC_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*ParameterFloatMsg)
		if msg.Param.Value != 100 {
			t.Fatalf("unexpected parameter float: %+v", msg.Param)
		}
	})

	t.Run("ParameterActivation", func(t *testing.T) {
		payload := append(ioaBytes(1), byte(QPADeActObjectParameter))
		a := newASDUForParse(P_AC_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*ParameterActivationMsg)
		if msg.Param.Qpa != QPADeActObjectParameter {
			t.Fatalf("unexpected parameter activation: %+v", msg.Param)
		}
	})
}

func TestParseASDU_SystemFamilies(t *testing.T) {
	tm := time.Date(2025, 8, 25, 12, 34, 56, 0, time.UTC)

	t.Run("InterrogationCmd", func(t *testing.T) {
		payload := append(ioaBytes(0), byte(QOIStation))
		a := newASDUForParse(C_IC_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*InterrogationCmdMsg)
		if msg.QOI != QOIStation {
			t.Fatalf("unexpected qoi: %+v", msg)
		}
	})

	t.Run("CounterInterrogationCmd", func(t *testing.T) {
		qcc := QualifierCountCall{Request: QCCGroup1, Freeze: QCCFrzRead}
		payload := append(ioaBytes(0), qcc.Value())
		a := newASDUForParse(C_CI_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*CounterInterrogationCmdMsg)
		if msg.QCC.Request != QCCGroup1 {
			t.Fatalf("unexpected qcc: %+v", msg)
		}
	})

	t.Run("ReadCmd", func(t *testing.T) {
		payload := ioaBytes(0x010203)
		a := newASDUForParse(C_RD_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*ReadCmdMsg)
		if msg.IOA != 0x010203 {
			t.Fatalf("unexpected ioa: %+v", msg)
		}
	})

	t.Run("ClockSyncCmd", func(t *testing.T) {
		payload := append(ioaBytes(0), CP56Time2a(tm, time.UTC)...)
		a := newASDUForParse(C_CS_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*ClockSyncCmdMsg)
		if msg.Time.Year() != tm.Year() {
			t.Fatalf("unexpected time: %+v", msg)
		}
	})

	t.Run("TestCmd", func(t *testing.T) {
		payload := append(ioaBytes(0), 0xAA, 0x55)
		a := newASDUForParse(C_TS_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*TestCmdMsg)
		if !msg.Test {
			t.Fatalf("unexpected test flag: %+v", msg)
		}
	})

	t.Run("ResetProcessCmd", func(t *testing.T) {
		payload := append(ioaBytes(0), byte(QPRGeneralRest))
		a := newASDUForParse(C_RP_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*ResetProcessCmdMsg)
		if msg.QRP != QPRGeneralRest {
			t.Fatalf("unexpected qrp: %+v", msg)
		}
	})

	t.Run("DelayAcquireCmd", func(t *testing.T) {
		payload := append(ioaBytes(0), 0x10, 0x27)
		a := newASDUForParse(C_CD_NA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*DelayAcquireCmdMsg)
		if msg.Msec != 10000 {
			t.Fatalf("unexpected msec: %+v", msg)
		}
	})

	t.Run("TestCmdCP56", func(t *testing.T) {
		payload := append(ioaBytes(0), 0xAA, 0x55)
		payload = append(payload, CP56Time2a(tm, time.UTC)...)
		a := newASDUForParse(C_TS_TA_1, VariableStruct{Number: 1}, payload)
		msg := mustParse(t, a).(*TestCmdCP56Msg)
		if !msg.Test || msg.Time.Year() != tm.Year() {
			t.Fatalf("unexpected test cmd cp56: %+v", msg)
		}
	})
}

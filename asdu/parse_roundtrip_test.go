package asdu

import (
	"net"
	"reflect"
	"testing"
)

type captureConn struct {
	params *Params
	last   *ASDU
}

func (c *captureConn) Params() *Params          { return c.params }
func (c *captureConn) UnderlyingConn() net.Conn { return nil }
func (c *captureConn) Send(a *ASDU) error       { c.last = a.Clone(); return nil }

func (c *captureConn) mustRaw(t *testing.T) []byte {
	t.Helper()
	if c.last == nil {
		t.Fatal("no ASDU captured")
	}
	raw, err := c.last.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}
	return append([]byte(nil), raw...)
}

func mustEncodeBinary(t *testing.T, msg Message) []byte {
	t.Helper()
	a, err := EncodeMessage(msg)
	if err != nil {
		t.Fatalf("EncodeMessage failed: %v", err)
	}
	raw, err := a.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary failed: %v", err)
	}
	return raw
}

func roundTripFromHelper(t *testing.T, build func(*captureConn) error) {
	t.Helper()
	conn := &captureConn{params: ParamsWide}
	if err := build(conn); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	raw := conn.mustRaw(t)
	msg, err := ParseASDU(mustUnmarshal(t, raw))
	if err != nil {
		t.Fatalf("ParseASDU failed: %v", err)
	}
	round := mustEncodeBinary(t, msg)
	if !reflect.DeepEqual(raw, round) {
		t.Fatalf("round-trip mismatch: %x vs %x", raw, round)
	}
}

func TestParseASDU_RoundTripSinglePoint(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return Single(c, true, coa, 1,
			SinglePointInfo{Ioa: 100, Value: true, Qds: QDSGood},
			SinglePointInfo{Ioa: 101, Value: false, Qds: QDSGood})
	})
}

func TestParseASDU_RoundTripDoublePoint(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return Double(c, false, coa, 2, DoublePointInfo{Ioa: 5, Value: DPIDeterminedOn, Qds: QDSGood})
	})
}

func TestParseASDU_RoundTripMeasuredValueScaled(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return MeasuredValueScaled(c, true, coa, 3,
			MeasuredValueScaledInfo{Ioa: 10, Value: 123, Qds: QDSGood},
			MeasuredValueScaledInfo{Ioa: 11, Value: -12, Qds: QDSGood})
	})
}

func TestParseASDU_RoundTripInterrogationCmd(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Activation}
		return InterrogationCmd(c, coa, 1, QOIStation)
	})
}

func TestParseASDU_RoundTripStepPosition(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return Step(c, true, coa, 4,
			StepPositionInfo{Ioa: 1, Value: StepPosition{Val: 2}, Qds: QDSGood},
			StepPositionInfo{Ioa: 2, Value: StepPosition{Val: 3}, Qds: QDSGood})
	})
}

func TestParseASDU_RoundTripBitString32(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return BitString32CP24Time2a(c, coa, 5, BitString32Info{Ioa: 9, Value: 0x12345678, Qds: QDSGood, Time: tm0})
	})
}

func TestParseASDU_RoundTripMeasuredValueNormal(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return MeasuredValueNormal(c, true, coa, 6,
			MeasuredValueNormalInfo{Ioa: 1, Value: Normalize(123), Qds: QDSGood},
			MeasuredValueNormalInfo{Ioa: 2, Value: Normalize(456), Qds: QDSGood})
	})
}

func TestParseASDU_RoundTripMeasuredValueFloat(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return MeasuredValueFloatCP24Time2a(c, coa, 7, MeasuredValueFloatInfo{Ioa: 1, Value: 1.5, Qds: QDSGood, Time: tm0})
	})
}

func TestParseASDU_RoundTripIntegratedTotals(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return IntegratedTotals(c, true, coa, 8, BinaryCounterReadingInfo{Ioa: 1, Value: BinaryCounterReading{CounterReading: 42}})
	})
}

func TestParseASDU_RoundTripEventOfProtection(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return EventOfProtectionEquipmentCP24Time2a(c, coa, 9, EventOfProtectionEquipmentInfo{
			Ioa:   1,
			Event: SingleEvent(2),
			Qdp:   QDPGood,
			Msec:  123,
			Time:  tm0,
		})
	})
}

func TestParseASDU_RoundTripPackedStartEvents(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return PackedStartEventsOfProtectionEquipmentCP24Time2a(c, coa, 10, PackedStartEventsOfProtectionEquipmentInfo{
			Ioa:   1,
			Event: StartEvent(1),
			Qdp:   QDPGood,
			Msec:  200,
			Time:  tm0,
		})
	})
}

func TestParseASDU_RoundTripPackedOutputCircuit(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return PackedOutputCircuitInfoCP24Time2a(c, coa, 11, PackedOutputCircuitInfoInfo{
			Ioa:  2,
			Oci:  OutputCircuitInfo(3),
			Qdp:  QDPGood,
			Msec: 500,
			Time: tm0,
		})
	})
}

func TestParseASDU_RoundTripPackedSinglePointWithSCD(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Spontaneous}
		return PackedSinglePointWithSCD(c, true, coa, 12, PackedSinglePointWithSCDInfo{
			Ioa: 1,
			Scd: StatusAndStatusChangeDetection(0x01020304),
			Qds: QDSGood,
		})
	})
}

func TestParseASDU_RoundTripEndOfInit(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Initialized}
		return EndOfInitialization(c, coa, 13, 1, CauseOfInitial{Cause: COILocalPowerOn})
	})
}

func TestParseASDU_RoundTripSingleCmd(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Activation}
		return SingleCmd(c, C_SC_NA_1, coa, 14, SingleCommandInfo{
			Ioa:   0x010203,
			Value: true,
			Qoc:   QualifierOfCommand{Qual: QOCShortPulseDuration},
		})
	})
}

func TestParseASDU_RoundTripSetpointScaled(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Activation}
		return SetpointCmdScaled(c, C_SE_NB_1, coa, 15, SetpointCommandScaledInfo{
			Ioa:   1,
			Value: 42,
			Qos:   QualifierOfSetpointCmd{Qual: 1},
		})
	})
}

func TestParseASDU_RoundTripParameterNormal(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Activation}
		return ParameterNormal(c, coa, 16, ParameterNormalInfo{
			Ioa:   1,
			Value: Normalize(1234),
			Qpm:   QualifierOfParameterMV{Category: QPMThreshold},
		})
	})
}

func TestParseASDU_RoundTripResetProcess(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Activation}
		return ResetProcessCmd(c, coa, 17, QPRGeneralRest)
	})
}

func TestParseASDU_RoundTripDelayAcquire(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Activation}
		return DelayAcquireCommand(c, coa, 18, 250)
	})
}

func TestParseASDU_RoundTripTestCmdCP56(t *testing.T) {
	roundTripFromHelper(t, func(c *captureConn) error {
		coa := CauseOfTransmission{Cause: Activation}
		return TestCommandCP56Time2a(c, coa, 19, tm0)
	})
}

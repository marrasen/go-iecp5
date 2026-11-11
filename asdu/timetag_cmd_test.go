package asdu

import (
	"testing"
	"time"
)

// TestTimeTaggedCommandsRoundTrip checks that every time-tagged command type decodes after it is encoded.
func TestTimeTaggedCommandsRoundTrip(t *testing.T) {
	at := time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	c := &captureConn{params: ParamsWide}
	coa := CauseOfTransmission{Cause: Activation}
	sends := map[TypeID]func() error{
		C_SC_TA_1: func() error {
			return SingleCmd(c, C_SC_TA_1, coa, 1, SingleCommandInfo{Ioa: 100, Value: true, Time: at})
		},
		C_DC_TA_1: func() error {
			return DoubleCmd(c, C_DC_TA_1, coa, 1, DoubleCommandInfo{Ioa: 100, Value: DCOOn, Time: at})
		},
		C_RC_TA_1: func() error {
			return StepCmd(c, C_RC_TA_1, coa, 1, StepCommandInfo{Ioa: 100, Value: SCOStepUP, Time: at})
		},
		C_SE_TA_1: func() error {
			return SetpointCmdNormal(c, C_SE_TA_1, coa, 1, SetpointCommandNormalInfo{Ioa: 100, Value: 1000, Time: at})
		},
		C_SE_TB_1: func() error {
			return SetpointCmdScaled(c, C_SE_TB_1, coa, 1, SetpointCommandScaledInfo{Ioa: 100, Value: 1000, Time: at})
		},
		C_SE_TC_1: func() error {
			return SetpointCmdFloat(c, C_SE_TC_1, coa, 1, SetpointCommandFloatInfo{Ioa: 100, Value: 1.5, Time: at})
		},
		C_BO_TA_1: func() error {
			return BitsString32Cmd(c, C_BO_TA_1, coa, 1, BitsString32CommandInfo{Ioa: 100, Value: 0xdeadbeef, Time: at})
		},
	}
	for typ, send := range sends {
		c.last = nil
		if err := send(); err != nil {
			t.Fatalf("%v: send: %v", typ, err)
		}
		raw, err := c.last.MarshalBinary()
		if err != nil {
			t.Fatalf("%v: marshal: %v", typ, err)
		}
		got := NewEmptyASDU(ParamsWide)
		if err := got.UnmarshalBinary(raw); err != nil {
			t.Fatalf("%v: unmarshal: %v", typ, err)
		}
		if _, err := ParseASDU(got); err != nil {
			t.Fatalf("%v: parse: %v", typ, err)
		}
	}
}

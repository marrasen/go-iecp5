// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package asdu

import (
	"fmt"
	"strings"
	"time"
)

// String returns a human-readable description of UnknownMsg.
func (m *UnknownMsg) String() string {
	n := int(m.H.Identifier.Variable.Number)
	if n == 0 {
		n = 1
	}
	return fmt.Sprintf("items=%d payload=%dB", n, len(m.H.RawInfoObj))
}

// String returns a human-readable description of SinglePointMsg.
func (m *SinglePointMsg) String() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "items=%d", len(m.Items))
	for i, it := range m.Items {
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
	if len(m.Items) > 0 {
		b.WriteByte(']')
	}
	return b.String()
}

// String returns a human-readable description of DoublePointMsg.
func (m *DoublePointMsg) String() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "items=%d", len(m.Items))
	for i, it := range m.Items {
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
	if len(m.Items) > 0 {
		b.WriteByte(']')
	}
	return b.String()
}

// String returns a human-readable description of StepPositionMsg.
func (m *StepPositionMsg) String() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "items=%d", len(m.Items))
	for i, it := range m.Items {
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
	if len(m.Items) > 0 {
		b.WriteByte(']')
	}
	return b.String()
}

// String returns a human-readable description of BitString32Msg.
func (m *BitString32Msg) String() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "items=%d", len(m.Items))
	for i, it := range m.Items {
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
	if len(m.Items) > 0 {
		b.WriteByte(']')
	}
	return b.String()
}

// String returns a human-readable description of MeasuredValueNormalMsg.
func (m *MeasuredValueNormalMsg) String() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "items=%d", len(m.Items))
	for i, it := range m.Items {
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
	if len(m.Items) > 0 {
		b.WriteByte(']')
	}
	return b.String()
}

// String returns a human-readable description of MeasuredValueScaledMsg.
func (m *MeasuredValueScaledMsg) String() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "items=%d", len(m.Items))
	for i, it := range m.Items {
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
	if len(m.Items) > 0 {
		b.WriteByte(']')
	}
	return b.String()
}

// String returns a human-readable description of MeasuredValueFloatMsg.
func (m *MeasuredValueFloatMsg) String() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "items=%d", len(m.Items))
	for i, it := range m.Items {
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
	if len(m.Items) > 0 {
		b.WriteByte(']')
	}
	return b.String()
}

// String returns a human-readable description of IntegratedTotalsMsg.
func (m *IntegratedTotalsMsg) String() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "items=%d", len(m.Items))
	for i, it := range m.Items {
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
	if len(m.Items) > 0 {
		b.WriteByte(']')
	}
	return b.String()
}

// String returns a human-readable description of EventOfProtectionMsg.
func (m *EventOfProtectionMsg) String() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "items=%d", len(m.Items))
	for i, it := range m.Items {
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
	if len(m.Items) > 0 {
		b.WriteByte(']')
	}
	return b.String()
}

// String returns a human-readable description of PackedStartEventsMsg.
func (m *PackedStartEventsMsg) String() string {
	it := m.Item
	s := fmt.Sprintf("IOA=%d start=0x%02x QDP=0x%02x msec=%d", it.Ioa, byte(it.Event), byte(it.Qdp), it.Msec)
	if !it.Time.IsZero() {
		s += fmt.Sprintf(" @%s", it.Time.Format(time.RFC3339Nano))
	}
	return s
}

// String returns a human-readable description of PackedOutputCircuitMsg.
func (m *PackedOutputCircuitMsg) String() string {
	it := m.Item
	s := fmt.Sprintf("IOA=%d oci=0x%02x QDP=0x%02x msec=%d", it.Ioa, byte(it.Oci), byte(it.Qdp), it.Msec)
	if !it.Time.IsZero() {
		s += fmt.Sprintf(" @%s", it.Time.Format(time.RFC3339Nano))
	}
	return s
}

// String returns a human-readable description of PackedSinglePointWithSCDMsg.
func (m *PackedSinglePointWithSCDMsg) String() string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "items=%d", len(m.Items))
	for i, it := range m.Items {
		if i == 0 {
			b.WriteString(" [")
		} else {
			b.WriteString(", ")
		}
		_, _ = fmt.Fprintf(&b, "%d=SCD(0x%08x) QDS=0x%02x", it.Ioa, uint32(it.Scd), byte(it.Qds))
	}
	if len(m.Items) > 0 {
		b.WriteByte(']')
	}
	return b.String()
}

// String returns a human-readable description of EndOfInitMsg.
func (m *EndOfInitMsg) String() string {
	return fmt.Sprintf("IOA=%d cause=%d localChange=%t", m.IOA, m.COI.Cause, m.COI.IsLocalChange)
}

// String returns a human-readable description of SingleCommandMsg.
func (m *SingleCommandMsg) String() string {
	cmd := m.Cmd
	s := fmt.Sprintf("IOA=%d val=%t QOC=%s", cmd.Ioa, cmd.Value, cmd.Qoc)
	if !cmd.Time.IsZero() {
		s += fmt.Sprintf(" @%s", cmd.Time.Format(time.RFC3339Nano))
	}
	return s
}

// String returns a human-readable description of DoubleCommandMsg.
func (m *DoubleCommandMsg) String() string {
	cmd := m.Cmd
	s := fmt.Sprintf("IOA=%d val=%d QOC=%s", cmd.Ioa, cmd.Value, cmd.Qoc)
	if !cmd.Time.IsZero() {
		s += fmt.Sprintf(" @%s", cmd.Time.Format(time.RFC3339Nano))
	}
	return s
}

// String returns a human-readable description of StepCommandMsg.
func (m *StepCommandMsg) String() string {
	cmd := m.Cmd
	s := fmt.Sprintf("IOA=%d val=%d QOC=%s", cmd.Ioa, cmd.Value, cmd.Qoc)
	if !cmd.Time.IsZero() {
		s += fmt.Sprintf(" @%s", cmd.Time.Format(time.RFC3339Nano))
	}
	return s
}

// String returns a human-readable description of SetpointNormalMsg.
func (m *SetpointNormalMsg) String() string {
	cmd := m.Cmd
	s := fmt.Sprintf("IOA=%d val=%.6f QOS=%s", cmd.Ioa, cmd.Value.Float64(), cmd.Qos)
	if !cmd.Time.IsZero() {
		s += fmt.Sprintf(" @%s", cmd.Time.Format(time.RFC3339Nano))
	}
	return s
}

// String returns a human-readable description of SetpointScaledMsg.
func (m *SetpointScaledMsg) String() string {
	cmd := m.Cmd
	s := fmt.Sprintf("IOA=%d val=%d QOS=%s", cmd.Ioa, cmd.Value, cmd.Qos)
	if !cmd.Time.IsZero() {
		s += fmt.Sprintf(" @%s", cmd.Time.Format(time.RFC3339Nano))
	}
	return s
}

// String returns a human-readable description of SetpointFloatMsg.
func (m *SetpointFloatMsg) String() string {
	cmd := m.Cmd
	s := fmt.Sprintf("IOA=%d val=%g QOS=%s", cmd.Ioa, cmd.Value, cmd.Qos)
	if !cmd.Time.IsZero() {
		s += fmt.Sprintf(" @%s", cmd.Time.Format(time.RFC3339Nano))
	}
	return s
}

// String returns a human-readable description of BitsString32CmdMsg.
func (m *BitsString32CmdMsg) String() string {
	cmd := m.Cmd
	s := fmt.Sprintf("IOA=%d bits=0x%08x", cmd.Ioa, cmd.Value)
	if !cmd.Time.IsZero() {
		s += fmt.Sprintf(" @%s", cmd.Time.Format(time.RFC3339Nano))
	}
	return s
}

// String returns a human-readable description of ParameterNormalMsg.
func (m *ParameterNormalMsg) String() string {
	p := m.Param
	return fmt.Sprintf("IOA=%d val=%.6f QPM=0x%02x", p.Ioa, p.Value.Float64(), byte(p.Qpm.Value()))
}

// String returns a human-readable description of ParameterScaledMsg.
func (m *ParameterScaledMsg) String() string {
	p := m.Param
	return fmt.Sprintf("IOA=%d val=%d QPM=0x%02x", p.Ioa, p.Value, byte(p.Qpm.Value()))
}

// String returns a human-readable description of ParameterFloatMsg.
func (m *ParameterFloatMsg) String() string {
	p := m.Param
	return fmt.Sprintf("IOA=%d val=%g QPM=0x%02x", p.Ioa, p.Value, byte(p.Qpm.Value()))
}

// String returns a human-readable description of ParameterActivationMsg.
func (m *ParameterActivationMsg) String() string {
	p := m.Param
	return fmt.Sprintf("IOA=%d QPA=%d", p.Ioa, p.Qpa)
}

// String returns a human-readable description of InterrogationCmdMsg.
func (m *InterrogationCmdMsg) String() string {
	return fmt.Sprintf("IOA=%d QOI=%d", m.IOA, byte(m.QOI))
}

// String returns a human-readable description of CounterInterrogationCmdMsg.
func (m *CounterInterrogationCmdMsg) String() string {
	return fmt.Sprintf("IOA=%d QCC=%d", m.IOA, m.QCC.Value())
}

// String returns a human-readable description of ReadCmdMsg.
func (m *ReadCmdMsg) String() string {
	return fmt.Sprintf("IOA=%d", m.IOA)
}

// String returns a human-readable description of ClockSyncCmdMsg.
func (m *ClockSyncCmdMsg) String() string {
	return fmt.Sprintf("IOA=%d @%s", m.IOA, m.Time.Format(time.RFC3339Nano))
}

// String returns a human-readable description of TestCmdMsg.
func (m *TestCmdMsg) String() string {
	return fmt.Sprintf("IOA=%d test=%t", m.IOA, m.Test)
}

// String returns a human-readable description of ResetProcessCmdMsg.
func (m *ResetProcessCmdMsg) String() string {
	return fmt.Sprintf("IOA=%d QRP=%d", m.IOA, byte(m.QRP))
}

// String returns a human-readable description of DelayAcquireCmdMsg.
func (m *DelayAcquireCmdMsg) String() string {
	return fmt.Sprintf("IOA=%d msec=%d", m.IOA, m.Msec)
}

// String returns a human-readable description of TestCmdCP56Msg.
func (m *TestCmdCP56Msg) String() string {
	return fmt.Sprintf("IOA=%d test=%t @%s", m.IOA, m.Test, m.Time.Format(time.RFC3339Nano))
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package asdu

import (
	"time"
)

// Application Service Data Units for control-direction process information

// SingleCommandInfo single command information object
type SingleCommandInfo struct {
	Ioa   InfoObjAddr
	Value bool
	Qoc   QualifierOfCommand
	Time  time.Time
}

// SingleCmd sends a type identification [C_SC_NA_1] or [C_SC_TA_1]. Single command; single information object (SQ = 0)
// [C_SC_NA_1] See companion standard 101, subsection 7.3.2.1
// [C_SC_TA_1] See companion standard 101
// Cause of transmission (coa) used for
// Control direction:
// <6> := activation
// <8> := deactivation
// Monitoring direction:
// <7> := activation confirmation
// <9> := deactivation confirmation
// <10> := activation termination
// <44> := unknown type identification
// <45> := unknown cause of transmission
// <46> := unknown ASDU common address
// <47> := unknown information object address
func SingleCmd(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, cmd SingleCommandInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})

	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}
	value := cmd.Qoc.Value()
	if cmd.Value {
		value |= 0x01
	}
	u.AppendBytes(value)
	switch typeID {
	case C_SC_NA_1:
	case C_SC_TA_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}
	return c.Send(u)
}

// DoubleCommandInfo double command information object
type DoubleCommandInfo struct {
	Ioa   InfoObjAddr
	Value DoubleCommand
	Qoc   QualifierOfCommand
	Time  time.Time
}

// DoubleCmd sends a type identification [C_DC_NA_1] or [C_DC_TA_1]. Double command; single information object (SQ = 0)
// [C_DC_NA_1] See companion standard 101, subsection 7.3.2.2
// [C_DC_TA_1] See companion standard 101
// Cause of transmission (coa) used for
// Control direction:
// <6> := activation
// <8> := deactivation
// Monitoring direction:
// <7> := activation confirmation
// <9> := deactivation confirmation
// <10> := activation termination
// <44> := unknown type identification
// <45> := unknown cause of transmission
// <46> := unknown ASDU common address
// <47> := unknown information object address
func DoubleCmd(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr,
	cmd DoubleCommandInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}
	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})

	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}

	u.AppendBytes(cmd.Qoc.Value() | byte(cmd.Value&0x03))
	switch typeID {
	case C_DC_NA_1:
	case C_DC_TA_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}
	return c.Send(u)
}

// StepCommandInfo step command information object
type StepCommandInfo struct {
	Ioa   InfoObjAddr
	Value StepCommand
	Qoc   QualifierOfCommand
	Time  time.Time
}

// StepCmd sends a type [C_RC_NA_1] or [C_RC_TA_1]. Step command; single information object (SQ = 0)
// [C_RC_NA_1] See companion standard 101, subsection 7.3.2.3
// [C_RC_TA_1] See companion standard 101
// Cause of transmission (coa) used for
// Control direction:
// <6> := activation
// <8> := deactivation
// Monitoring direction:
// <7> := activation confirmation
// <9> := deactivation confirmation
// <10> := activation termination
// <44> := unknown type identification
// <45> := unknown cause of transmission
// <46> := unknown ASDU common address
// <47> := unknown information object address
func StepCmd(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, cmd StepCommandInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}
	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})

	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}

	u.AppendBytes(cmd.Qoc.Value() | byte(cmd.Value&0x03))
	switch typeID {
	case C_RC_NA_1:
	case C_RC_TA_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}
	return c.Send(u)
}

// SetpointCommandNormalInfo setpoint command, normalized value information object
type SetpointCommandNormalInfo struct {
	Ioa   InfoObjAddr
	Value Normalize
	Qos   QualifierOfSetpointCmd
	Time  time.Time
}

// SetpointCmdNormal sends a type [C_SE_NA_1] or [C_SE_TA_1]. Setpoint command, normalized value; single information object (SQ = 0)
// [C_SE_NA_1] See companion standard 101, subsection 7.3.2.4
// [C_SE_TA_1] See companion standard 101
// Cause of transmission (coa) used for
// Control direction:
// <6> := activation
// <8> := deactivation
// Monitoring direction:
// <7> := activation confirmation
// <9> := deactivation confirmation
// <10> := activation termination
// <44> := unknown type identification
// <45> := unknown cause of transmission
// <46> := unknown ASDU common address
// <47> := unknown information object address
func SetpointCmdNormal(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, cmd SetpointCommandNormalInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}
	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})

	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}
	u.AppendNormalize(cmd.Value).AppendBytes(cmd.Qos.Value())
	switch typeID {
	case C_SE_NA_1:
	case C_SE_TA_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}
	return c.Send(u)
}

// SetpointCommandScaledInfo setpoint command, scaled value information object
type SetpointCommandScaledInfo struct {
	Ioa   InfoObjAddr
	Value int16
	Qos   QualifierOfSetpointCmd
	Time  time.Time
}

// SetpointCmdScaled sends a type [C_SE_NB_1] or [C_SE_TB_1]. Setpoint command, scaled value; single information object (SQ = 0)
// [C_SE_NB_1] See companion standard 101, subsection 7.3.2.5
// [C_SE_TB_1] See companion standard 101
// Cause of transmission (coa) used for
// Control direction:
// <6> := activation
// <8> := deactivation
// Monitoring direction:
// <7> := activation confirmation
// <9> := deactivation confirmation
// <10> := activation termination
// <44> := unknown type identification
// <45> := unknown cause of transmission
// <46> := unknown ASDU common address
// <47> := unknown information object address
func SetpointCmdScaled(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, cmd SetpointCommandScaledInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}
	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})

	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}
	u.AppendScaled(cmd.Value).AppendBytes(cmd.Qos.Value())
	switch typeID {
	case C_SE_NB_1:
	case C_SE_TB_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}
	return c.Send(u)
}

// SetpointCommandFloatInfo setpoint command, short floating-point value information object
type SetpointCommandFloatInfo struct {
	Ioa   InfoObjAddr
	Value float32
	Qos   QualifierOfSetpointCmd
	Time  time.Time
}

// SetpointCmdFloat sends a type [C_SE_NC_1] or [C_SE_TC_1]. Setpoint command, short floating-point value; single information object (SQ = 0)
// [C_SE_NC_1] See companion standard 101, subsection 7.3.2.6
// [C_SE_TC_1] See companion standard 101
// Cause of transmission (coa) used for
// Control direction:
// <6> := activation
// <8> := deactivation
// Monitoring direction:
// <7> := activation confirmation
// <9> := deactivation confirmation
// <10> := activation termination
// <44> := unknown type identification
// <45> := unknown cause of transmission
// <46> := unknown ASDU common address
// <47> := unknown information object address
func SetpointCmdFloat(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, cmd SetpointCommandFloatInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}
	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}

	u.AppendFloat32(cmd.Value).AppendBytes(cmd.Qos.Value())

	switch typeID {
	case C_SE_NC_1:
	case C_SE_TC_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}

	return c.Send(u)
}

// BitsString32CommandInfo bitstring (32-bit) command information object
type BitsString32CommandInfo struct {
	Ioa   InfoObjAddr
	Value uint32
	Time  time.Time
}

// BitsString32Cmd sends a type [C_BO_NA_1] or [C_BO_TA_1]. Bitstring (32-bit) command; single information object (SQ = 0)
// [C_BO_NA_1] See companion standard 101, subsection 7.3.2.7
// [C_BO_TA_1] See companion standard 101
// Cause of transmission (coa) used for
// Control direction:
// <6> := activation
// <8> := deactivation
// Monitoring direction:
// <7> := activation confirmation
// <9> := deactivation confirmation
// <10> := activation termination
// <44> := unknown type identification
// <45> := unknown cause of transmission
// <46> := unknown ASDU common address
// <47> := unknown information object address
func BitsString32Cmd(c Connect, typeID TypeID, coa CauseOfTransmission, commonAddr CommonAddr,
	cmd BitsString32CommandInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}
	u := NewASDU(c.Params(), Identifier{
		typeID,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		commonAddr,
	})
	if err := u.AppendInfoObjAddr(cmd.Ioa); err != nil {
		return err
	}

	u.AppendBitsString32(cmd.Value)

	switch typeID {
	case C_BO_NA_1:
	case C_BO_TA_1:
		u.AppendBytes(CP56Time2a(cmd.Time, u.InfoObjTimeZone)...)
	default:
		return ErrTypeIDNotMatch
	}

	return c.Send(u)
}

// GetSingleCmd [C_SC_NA_1] or [C_SC_TA_1] get single command information object
func (sf *ASDU) GetSingleCmd() SingleCommandInfo {
	var s SingleCommandInfo

	s.Ioa = sf.DecodeInfoObjAddr()
	value := sf.DecodeByte()
	s.Value = value&0x01 == 0x01
	s.Qoc = ParseQualifierOfCommand(value & 0xfe)

	switch sf.Type {
	case C_SC_NA_1:
	case C_SC_TA_1:
		s.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return s
}

// GetDoubleCmd [C_DC_NA_1] or [C_DC_TA_1] get double command information object
func (sf *ASDU) GetDoubleCmd() DoubleCommandInfo {
	var cmd DoubleCommandInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	value := sf.DecodeByte()
	cmd.Value = DoubleCommand(value & 0x03)
	cmd.Qoc = ParseQualifierOfCommand(value & 0xfc)

	switch sf.Type {
	case C_DC_NA_1:
	case C_DC_TA_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

// GetStepCmd [C_RC_NA_1] or [C_RC_TA_1] get step command information object
func (sf *ASDU) GetStepCmd() StepCommandInfo {
	var cmd StepCommandInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	value := sf.DecodeByte()
	cmd.Value = StepCommand(value & 0x03)
	cmd.Qoc = ParseQualifierOfCommand(value & 0xfc)

	switch sf.Type {
	case C_RC_NA_1:
	case C_RC_TA_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

// GetSetpointNormalCmd [C_SE_NA_1] or [C_SE_TA_1] get setpoint command, normalized value information object
func (sf *ASDU) GetSetpointNormalCmd() SetpointCommandNormalInfo {
	var cmd SetpointCommandNormalInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	cmd.Value = sf.DecodeNormalize()
	cmd.Qos = ParseQualifierOfSetpointCmd(sf.DecodeByte())

	switch sf.Type {
	case C_SE_NA_1:
	case C_SE_TA_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

// GetSetpointCmdScaled [C_SE_NB_1] or [C_SE_TB_1] get setpoint command, scaled value information object
func (sf *ASDU) GetSetpointCmdScaled() SetpointCommandScaledInfo {
	var cmd SetpointCommandScaledInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	cmd.Value = sf.DecodeScaled()
	cmd.Qos = ParseQualifierOfSetpointCmd(sf.DecodeByte())

	switch sf.Type {
	case C_SE_NB_1:
	case C_SE_TB_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

// GetSetpointFloatCmd [C_SE_NC_1] or [C_SE_TC_1] get setpoint command, short floating-point value information object
func (sf *ASDU) GetSetpointFloatCmd() SetpointCommandFloatInfo {
	var cmd SetpointCommandFloatInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	cmd.Value = sf.DecodeFloat32()
	cmd.Qos = ParseQualifierOfSetpointCmd(sf.DecodeByte())

	switch sf.Type {
	case C_SE_NC_1:
	case C_SE_TC_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

// GetBitsString32Cmd [C_BO_NA_1] or [C_BO_TA_1] get bitstring (32-bit) command information object
func (sf *ASDU) GetBitsString32Cmd() BitsString32CommandInfo {
	var cmd BitsString32CommandInfo

	cmd.Ioa = sf.DecodeInfoObjAddr()
	cmd.Value = sf.DecodeBitsString32()
	switch sf.Type {
	case C_BO_NA_1:
	case C_BO_TA_1:
		cmd.Time = sf.DecodeCP56Time2a()
	default:
		panic(ErrTypeIDNotMatch)
	}

	return cmd
}

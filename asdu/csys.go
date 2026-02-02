// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package asdu

import (
	"time"
)

// InterrogationCmd send a new interrogation command [C_IC_NA_1]. Single information object (SQ = 0)
// [C_IC_NA_1] See companion standard 101, subclass 7.3.4.1
// Cause of transmission (coa) used for:
// Control direction:
// <6> := Activation
// <8> := Deactivation
// Monitor direction:
// <7> := Activation confirmation
// <9> := Deactivation confirmation
// <10> := Activation termination
// <44> := Unknown type identification
// <45> := Unknown cause of transmission
// <46> := Unknown common address of ASDU
// <47> := Unknown information object address
func InterrogationCmd(c Connect, coa CauseOfTransmission, ca CommonAddr, qoi QualifierOfInterrogation) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}
	msg := &InterrogationCmdMsg{
		H:   newMessageHeader(c, C_IC_NA_1, coa, ca, false, 1),
		IOA: InfoObjAddrIrrelevant,
		QOI: qoi,
	}
	return sendEncoded(c, msg)
}

// CounterInterrogationCmd send Counter Interrogation command [C_CI_NA_1]，计数量召唤命令，只有单个信息对象(SQ = 0)
// [C_CI_NA_1] See companion standard 101, subclass 7.3.4.2
// Cause of transmission (coa) used for:
// Control direction:
// <6> := Activation
// Monitor direction:
// <7> := Activation confirmation
// <10> := Activation termination
// <44> := Unknown type identification
// <45> := Unknown cause of transmission
// <46> := Unknown common address of ASDU
// <47> := Unknown information object address
func CounterInterrogationCmd(c Connect, coa CauseOfTransmission, ca CommonAddr, qcc QualifierCountCall) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}
	coa.Cause = Activation
	msg := &CounterInterrogationCmdMsg{
		H:   newMessageHeader(c, C_CI_NA_1, coa, ca, false, 1),
		IOA: InfoObjAddrIrrelevant,
		QCC: qcc,
	}
	return sendEncoded(c, msg)
}

// ReadCmd send read command [C_RD_NA_1], Read command, single information object (SQ = 0)
// [C_RD_NA_1] See companion standard 101, subclass 7.3.4.3
// Cause of transmission (coa) used for:
// Control direction:
// <5> := Request
// Monitor direction:
// <44> := Unknown type identification
// <45> := Unknown cause of transmission
// <46> := Unknown common address of ASDU
// <47> := Unknown information object address
func ReadCmd(c Connect, coa CauseOfTransmission, ca CommonAddr, ioa InfoObjAddr) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}
	coa.Cause = Request
	msg := &ReadCmdMsg{
		H:   newMessageHeader(c, C_RD_NA_1, coa, ca, false, 1),
		IOA: ioa,
	}
	return sendEncoded(c, msg)
}

// ClockSynchronizationCmd send clock sync command [C_CS_NA_1], Clock synchronization command, single information object (SQ = 0)
// [C_CS_NA_1] See companion standard 101, subclass 7.3.4.4
// Cause of transmission (coa) used for:
// Control direction:
// <6> := Activation
// Monitor direction:
// <7> := Activation confirmation
// <10> := Activation termination
// <44> := Unknown type identification
// <45> := Unknown cause of transmission
// <46> := Unknown common address of ASDU
// <47> := Unknown information object address
func ClockSynchronizationCmd(c Connect, coa CauseOfTransmission, ca CommonAddr, t time.Time) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}
	coa.Cause = Activation
	msg := &ClockSyncCmdMsg{
		H:    newMessageHeader(c, C_CS_NA_1, coa, ca, false, 1),
		IOA:  InfoObjAddrIrrelevant,
		Time: t,
	}
	return sendEncoded(c, msg)
}

// TestCommand send test command [C_TS_NA_1], Test command, single information object (SQ = 0)
// [C_TS_NA_1] See companion standard 101, subclass 7.3.4.5
// Cause of transmission (coa) used for:
// Control direction:
// <6> := Activation
// Monitor direction:
// <7> := Activation confirmation
// <44> := Unknown type identification
// <45> := Unknown cause of transmission
// <46> := Unknown common address of ASDU
// <47> := Unknown information object address
func TestCommand(c Connect, coa CauseOfTransmission, ca CommonAddr) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}
	coa.Cause = Activation
	msg := &TestCmdMsg{
		H:    newMessageHeader(c, C_TS_NA_1, coa, ca, false, 1),
		IOA:  InfoObjAddrIrrelevant,
		Test: true,
	}
	return sendEncoded(c, msg)
}

// ResetProcessCmd send reset process command [C_RP_NA_1], Reset process command, single information object (SQ = 0)
// [C_RP_NA_1] See companion standard 101, subclass 7.3.4.6
// Cause of transmission (coa) used for:
// Control direction:
// <6> := Activation
// Monitor direction:
// <7> := Activation confirmation
// <44> := Unknown type identification
// <45> := Unknown cause of transmission
// <46> := Unknown common address of ASDU
// <47> := Unknown information object address
func ResetProcessCmd(c Connect, coa CauseOfTransmission, ca CommonAddr, qrp QualifierOfResetProcessCmd) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}
	coa.Cause = Activation
	msg := &ResetProcessCmdMsg{
		H:   newMessageHeader(c, C_RP_NA_1, coa, ca, false, 1),
		IOA: InfoObjAddrIrrelevant,
		QRP: qrp,
	}
	return sendEncoded(c, msg)
}

// DelayAcquireCommand send delay acquire command [C_CD_NA_1], Delay acquisition command, single information object (SQ = 0)
// [C_CD_NA_1] See companion standard 101, subclass 7.3.4.7
// Cause of transmission (coa) used for:
// Control direction:
// <3> := Spontaneous
// <6> := Activation
// Monitor direction:
// <7> := Activation confirmation
// <44> := Unknown type identification
// <45> := Unknown cause of transmission
// <46> := Unknown common address of ASDU
// <47> := Unknown information object address
func DelayAcquireCommand(c Connect, coa CauseOfTransmission, ca CommonAddr, msec uint16) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Activation) {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}
	msg := &DelayAcquireCmdMsg{
		H:    newMessageHeader(c, C_CD_NA_1, coa, ca, false, 1),
		IOA:  InfoObjAddrIrrelevant,
		Msec: msec,
	}
	return sendEncoded(c, msg)
}

// TestCommandCP56Time2a send test command [C_TS_TA_1], Test command, single information object (SQ = 0)
// Cause of transmission (coa) used for:
// Control direction:
// <6> := Activation
// Monitor direction:
// <7> := Activation confirmation
// <44> := Unknown type identification
// <45> := Unknown cause of transmission
// <46> := Unknown common address of ASDU
// <47> := Unknown information object address
func TestCommandCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, t time.Time) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}
	msg := &TestCmdCP56Msg{
		H:    newMessageHeader(c, C_TS_TA_1, coa, ca, false, 1),
		IOA:  InfoObjAddrIrrelevant,
		Test: true,
		Time: t,
	}
	return sendEncoded(c, msg)
}

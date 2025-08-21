// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

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

	u := NewASDU(c.Params(), Identifier{
		C_IC_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}
	u.AppendBytes(byte(qoi))
	return c.Send(u)
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
	u := NewASDU(c.Params(), Identifier{
		C_CI_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}
	u.AppendBytes(qcc.Value())
	return c.Send(u)
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
	u := NewASDU(c.Params(), Identifier{
		C_RD_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(ioa); err != nil {
		return err
	}
	return c.Send(u)
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
	u := NewASDU(c.Params(), Identifier{
		C_CS_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}
	u.AppendBytes(CP56Time2a(t, u.InfoObjTimeZone)...)
	return c.Send(u)
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
	u := NewASDU(c.Params(), Identifier{
		C_TS_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}
	u.AppendBytes(byte(FBPTestWord&0xff), byte(FBPTestWord>>8))
	return c.Send(u)
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
	u := NewASDU(c.Params(), Identifier{
		C_RP_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}
	u.AppendBytes(byte(qrp))
	return c.Send(u)
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

	u := NewASDU(c.Params(), Identifier{
		C_CD_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}
	u.AppendCP16Time2a(msec)
	return c.Send(u)
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
	u := NewASDU(c.Params(), Identifier{
		C_TS_TA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(InfoObjAddrIrrelevant); err != nil {
		return err
	}
	u.AppendUint16(FBPTestWord)
	u.AppendCP56Time2a(t, u.InfoObjTimeZone)
	return c.Send(u)
}

// GetInterrogationCmd [C_IC_NA_1] Get general interrogation information body (information object address, qualifier of interrogation)
func (sf *ASDU) GetInterrogationCmd() (InfoObjAddr, QualifierOfInterrogation) {
	return sf.DecodeInfoObjAddr(), QualifierOfInterrogation(sf.infoObj[0])
}

// GetCounterInterrogationCmd [C_CI_NA_1] Get counter interrogation information body (information object address, qualifier of counter call)
func (sf *ASDU) GetCounterInterrogationCmd() (InfoObjAddr, QualifierCountCall) {
	return sf.DecodeInfoObjAddr(), ParseQualifierCountCall(sf.infoObj[0])
}

// GetReadCmd [C_RD_NA_1] Get read command information address
func (sf *ASDU) GetReadCmd() InfoObjAddr {
	return sf.DecodeInfoObjAddr()
}

// GetClockSynchronizationCmd [C_CS_NA_1] Get clock synchronization command information body (information object address, time)
func (sf *ASDU) GetClockSynchronizationCmd() (InfoObjAddr, time.Time) {
	return sf.DecodeInfoObjAddr(), sf.DecodeCP56Time2a()
}

// GetTestCommand [C_TS_NA_1] Get test command information body (information object address, is test word)
func (sf *ASDU) GetTestCommand() (InfoObjAddr, bool) {
	return sf.DecodeInfoObjAddr(), sf.DecodeUint16() == FBPTestWord
}

// GetResetProcessCmd [C_RP_NA_1] Get reset process command information body (information object address, qualifier of reset process command)
func (sf *ASDU) GetResetProcessCmd() (InfoObjAddr, QualifierOfResetProcessCmd) {
	return sf.DecodeInfoObjAddr(), QualifierOfResetProcessCmd(sf.infoObj[0])
}

// GetDelayAcquireCommand [C_CD_NA_1] Get delay acquire command information body (information object address, delay milliseconds)
func (sf *ASDU) GetDelayAcquireCommand() (InfoObjAddr, uint16) {
	return sf.DecodeInfoObjAddr(), sf.DecodeUint16()
}

// GetTestCommandCP56Time2a [C_TS_TA_1] Get test command information body (information object address, is test word, time)
func (sf *ASDU) GetTestCommandCP56Time2a() (InfoObjAddr, bool, time.Time) {
	return sf.DecodeInfoObjAddr(), sf.DecodeUint16() == FBPTestWord, sf.DecodeCP56Time2a()
}

// Copyright 2020 thinkgos (thinkgo@aliyun.com).  All rights reserved.
// Use of this source code is governed by a version 3 of the GNU General
// Public License, license that can be found in the LICENSE file.

package asdu

// Application Service Data Units for control-direction parameters

// ParameterNormalInfo measurement parameter, normalized value information object
type ParameterNormalInfo struct {
	Ioa   InfoObjAddr
	Value Normalize
	Qpm   QualifierOfParameterMV
}

// ParameterNormal measurement parameter, normalized value; single information object (SQ = 0)
// [P_ME_NA_1], See companion standard 101, subsection 7.3.5.1
// Cause of transmission (coa) used for
// Control direction:
// <6> := activation
// Monitoring direction:
// <7> := activation confirmation
// <20> := response to station interrogation
// <21> := response to group 1 interrogation
// <22> := response to group 2 interrogation
// ...
// <36> := response to group 16 interrogation
// <44> := unknown type identification
// <45> := unknown cause of transmission
// <46> := unknown ASDU common address
// <47> := unknown information object address
func ParameterNormal(c Connect, coa CauseOfTransmission, ca CommonAddr, p ParameterNormalInfo) error {
	if coa.Cause != Activation {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		P_ME_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.AppendNormalize(p.Value)
	u.AppendBytes(p.Qpm.Value())
	return c.Send(u)
}

// ParameterScaledInfo measurement parameter, scaled value information object
type ParameterScaledInfo struct {
	Ioa   InfoObjAddr
	Value int16
	Qpm   QualifierOfParameterMV
}

// ParameterScaled measurement parameter, scaled value; single information object (SQ = 0)
// [P_ME_NB_1], See companion standard 101, subsection 7.3.5.2
// Cause of transmission (coa) used for
// Control direction:
// <6> := activation
// Monitoring direction:
// <7> := activation confirmation
// <20> := response to station interrogation
// <21> := response to group 1 interrogation
// <22> := response to group 2 interrogation
// ...
// <36> := response to group 16 interrogation
// <44> := unknown type identification
// <45> := unknown cause of transmission
// <46> := unknown ASDU common address
// <47> := unknown information object address
func ParameterScaled(c Connect, coa CauseOfTransmission, ca CommonAddr, p ParameterScaledInfo) error {
	if coa.Cause != Activation {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		P_ME_NB_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.AppendScaled(p.Value).AppendBytes(p.Qpm.Value())
	return c.Send(u)
}

// ParameterFloatInfo measurement parameter, short floating-point value information object
type ParameterFloatInfo struct {
	Ioa   InfoObjAddr
	Value float32
	Qpm   QualifierOfParameterMV
}

// ParameterFloat measurement parameter, short floating-point value; single information object (SQ = 0)
// [P_ME_NC_1], See companion standard 101, subsection 7.3.5.3
// Cause of transmission (coa) used for
// Control direction:
// <6> := activation
// Monitoring direction:
// <7> := activation confirmation
// <20> := response to station interrogation
// <21> := response to group 1 interrogation
// <22> := response to group 2 interrogation
// ...
// <36> := response to group 16 interrogation
// <44> := unknown type identification
// <45> := unknown cause of transmission
// <46> := unknown ASDU common address
// <47> := unknown information object address
func ParameterFloat(c Connect, coa CauseOfTransmission, ca CommonAddr, p ParameterFloatInfo) error {
	if coa.Cause != Activation {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		P_ME_NC_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.AppendFloat32(p.Value).AppendBytes(p.Qpm.Value())
	return c.Send(u)
}

// ParameterActivationInfo parameter activation information object
type ParameterActivationInfo struct {
	Ioa InfoObjAddr
	Qpa QualifierOfParameterAct
}

// ParameterActivation parameter activation; single information object (SQ = 0)
// [P_AC_NA_1], See companion standard 101, subsection 7.3.5.4
// Cause of transmission (coa) used for
// Control direction:
// <6> := activation
// <8> := deactivation
// Monitoring direction:
// <7> := activation confirmation
// <9> := deactivation confirmation
// <44> := unknown type identification
// <45> := unknown cause of transmission
// <46> := unknown ASDU common address
// <47> := unknown information object address
func ParameterActivation(c Connect, coa CauseOfTransmission, ca CommonAddr, p ParameterActivationInfo) error {
	if !(coa.Cause == Activation || coa.Cause == Deactivation) {
		return ErrCmdCause
	}
	if err := c.Params().Valid(); err != nil {
		return err
	}

	u := NewASDU(c.Params(), Identifier{
		P_AC_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})
	if err := u.AppendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.AppendBytes(byte(p.Qpa))
	return c.Send(u)
}

// GetParameterNormal [P_ME_NA_1] get measurement parameter, normalized value information object
func (sf *ASDU) GetParameterNormal() ParameterNormalInfo {
	return ParameterNormalInfo{
		sf.DecodeInfoObjAddr(),
		sf.DecodeNormalize(),
		ParseQualifierOfParamMV(sf.infoObj[0]),
	}
}

// GetParameterScaled [P_ME_NB_1] get measurement parameter, scaled value information object
func (sf *ASDU) GetParameterScaled() ParameterScaledInfo {
	return ParameterScaledInfo{
		sf.DecodeInfoObjAddr(),
		sf.DecodeScaled(),
		ParseQualifierOfParamMV(sf.infoObj[0]),
	}
}

// GetParameterFloat [P_ME_NC_1] get measurement parameter, short floating-point value information object
func (sf *ASDU) GetParameterFloat() ParameterFloatInfo {
	return ParameterFloatInfo{
		sf.DecodeInfoObjAddr(),
		sf.DecodeFloat32(),
		ParseQualifierOfParamMV(sf.infoObj[0]),
	}
}

// GetParameterActivation [P_AC_NA_1] get parameter activation information object
func (sf *ASDU) GetParameterActivation() ParameterActivationInfo {
	return ParameterActivationInfo{
		sf.DecodeInfoObjAddr(),
		QualifierOfParameterAct(sf.infoObj[0]),
	}
}

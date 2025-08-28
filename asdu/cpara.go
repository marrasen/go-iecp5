// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

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
	if err := u.appendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.appendNormalize(p.Value)
	u.appendBytes(p.Qpm.Value())
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
	if err := u.appendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.appendScaled(p.Value).appendBytes(p.Qpm.Value())
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
	if err := u.appendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.appendFloat32(p.Value).appendBytes(p.Qpm.Value())
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
	if err := u.appendInfoObjAddr(p.Ioa); err != nil {
		return err
	}
	u.appendBytes(byte(p.Qpa))
	return c.Send(u)
}

// GetParameterNormal [P_ME_NA_1] get measurement parameter, normalized value information object
func (sf *ASDU) GetParameterNormal() ParameterNormalInfo {
	saved := sf.infoObj
	defer func() { sf.infoObj = saved }()
	return ParameterNormalInfo{
		sf.decodeInfoObjAddr(),
		sf.decodeNormalize(),
		ParseQualifierOfParamMV(sf.infoObj[0]),
	}
}

// GetParameterScaled [P_ME_NB_1] get measurement parameter, scaled value information object
func (sf *ASDU) GetParameterScaled() ParameterScaledInfo {
	saved := sf.infoObj
	defer func() { sf.infoObj = saved }()
	return ParameterScaledInfo{
		sf.decodeInfoObjAddr(),
		sf.decodeScaled(),
		ParseQualifierOfParamMV(sf.infoObj[0]),
	}
}

// GetParameterFloat [P_ME_NC_1] get measurement parameter, short floating-point value information object
func (sf *ASDU) GetParameterFloat() ParameterFloatInfo {
	saved := sf.infoObj
	defer func() { sf.infoObj = saved }()
	return ParameterFloatInfo{
		sf.decodeInfoObjAddr(),
		sf.decodeFloat32(),
		ParseQualifierOfParamMV(sf.infoObj[0]),
	}
}

// GetParameterActivation [P_AC_NA_1] get parameter activation information object
func (sf *ASDU) GetParameterActivation() ParameterActivationInfo {
	saved := sf.infoObj
	defer func() { sf.infoObj = saved }()
	return ParameterActivationInfo{
		sf.decodeInfoObjAddr(),
		QualifierOfParameterAct(sf.infoObj[0]),
	}
}

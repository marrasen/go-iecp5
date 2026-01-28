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
	msg := ParameterNormalMsg{
		H:     newMessageHeader(c, P_ME_NA_1, coa, ca, false, 1),
		Param: p,
	}
	return sendEncoded(c, msg)
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
	msg := ParameterScaledMsg{
		H:     newMessageHeader(c, P_ME_NB_1, coa, ca, false, 1),
		Param: p,
	}
	return sendEncoded(c, msg)
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
	msg := ParameterFloatMsg{
		H:     newMessageHeader(c, P_ME_NC_1, coa, ca, false, 1),
		Param: p,
	}
	return sendEncoded(c, msg)
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
	msg := ParameterActivationMsg{
		H:     newMessageHeader(c, P_AC_NA_1, coa, ca, false, 1),
		Param: p,
	}
	return sendEncoded(c, msg)
}

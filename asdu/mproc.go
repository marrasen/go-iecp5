// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package asdu

import (
	"time"
)

// Application Service Data Units (ASDUs) for process information in the monitoring direction

// checkValid check common parameter of request is valid
func checkValid(c Connect, typeID TypeID, isSequence bool, infosLen int) error {
	if infosLen == 0 {
		return ErrNotAnyObjInfo
	}
	objSize, err := GetInfoObjSize(typeID)
	if err != nil {
		return err
	}
	param := c.Params()
	if err := param.Valid(); err != nil {
		return err
	}

	var asduLen int
	if isSequence {
		asduLen = param.IdentifierSize() + infosLen*objSize + param.InfoObjAddrSize
	} else {
		asduLen = param.IdentifierSize() + infosLen*(objSize+param.InfoObjAddrSize)
	}

	if asduLen > ASDUSizeMax {
		return ErrLengthOutOfRange
	}
	return nil
}

// SinglePointInfo the measured value attributes.
type SinglePointInfo struct {
	Ioa InfoObjAddr
	// value of single point
	Value bool
	// Quality descriptor asdu.OK means no remarks.
	Qds QualityDescriptor
	// the type does not include timing will ignore
	Time time.Time
}

// single sends a type identification [M_SP_NA_1], [M_SP_TA_1] or [M_SP_TB_1]. Single-point information
// [M_SP_NA_1] See companion standard 101,subclass 7.3.1.1
// [M_SP_TA_1] See companion standard 101,subclass 7.3.1.2
// [M_SP_TB_1] See companion standard 101,subclass 7.3.1.22
func single(c Connect, typeID TypeID, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...SinglePointInfo) error {
	if err := checkValid(c, typeID, isSequence, len(infos)); err != nil {
		return err
	}
	switch typeID {
	case M_SP_NA_1, M_SP_TA_1, M_SP_TB_1:
	default:
		return ErrTypeIDNotMatch
	}
	msg := SinglePointMsg{
		H:     newMessageHeader(c, typeID, coa, ca, isSequence, len(infos)),
		Items: infos,
	}
	return sendEncoded(c, msg)
}

// Single sends a type identification [M_SP_NA_1]. Single-point information without timestamp
// [M_SP_NA_1] See companion standard 101, subclass 7.3.1.1
// Cause of transmission (coa) used for monitoring direction:
// <2> := Background scan
// <3> := Spontaneous
// <5> := Requested
// <11> := Return information caused by remote command
// <12> := Return information caused by local command
// <20> := Response to station interrogation
// <21> := Response to group 1 interrogation
// ...
// <36> := Response to group 16 interrogation
func Single(c Connect, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...SinglePointInfo) error {
	if !(coa.Cause == Background || coa.Cause == Spontaneous || coa.Cause == Request ||
		coa.Cause == ReturnInfoRemote || coa.Cause == ReturnInfoLocal ||
		(coa.Cause >= InterrogatedByStation && coa.Cause <= InterrogatedByGroup16)) {
		return ErrCmdCause
	}
	return single(c, M_SP_NA_1, isSequence, coa, ca, infos...)
}

// SingleCP24Time2a sends a type identification [M_SP_TA_1]. Single-point information with CP24Time2a timestamp, only (SQ = 0) single information elements
// [M_SP_TA_1] See companion standard 101, subclass 7.3.1.2
// Cause of transmission (coa) used for monitoring direction:
// <3> := Spontaneous
// <5> := Requested
// <11> := Return information caused by remote command
// <12> := Return information caused by local command
func SingleCP24Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...SinglePointInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request ||
		coa.Cause == ReturnInfoRemote || coa.Cause == ReturnInfoLocal) {
		return ErrCmdCause
	}
	return single(c, M_SP_TA_1, false, coa, ca, infos...)
}

// SingleCP56Time2a sends a type identification [M_SP_TB_1]. Single-point information with CP56Time2a timestamp, only (SQ = 0) single information elements
// [M_SP_TB_1] See companion standard 101, subclass 7.3.1.22
// Cause of transmission (coa) used for monitoring direction:
// <3> := Spontaneous
// <5> := Requested
// <11> := Return information caused by remote command
// <12> := Return information caused by local command
func SingleCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...SinglePointInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request ||
		coa.Cause == ReturnInfoRemote || coa.Cause == ReturnInfoLocal) {
		return ErrCmdCause
	}
	return single(c, M_SP_TB_1, false, coa, ca, infos...)
}

// DoublePointInfo the measured value attributes.
type DoublePointInfo struct {
	Ioa   InfoObjAddr
	Value DoublePoint
	// Quality descriptor asdu.QDSGood means no remarks.
	Qds QualityDescriptor
	// the type does not include timing will ignore
	Time time.Time
}

// double sends a type identification [M_DP_NA_1], [M_DP_TA_1] or [M_DP_TB_1]. Double-point information
// [M_DP_NA_1] See companion standard 101,subclass 7.3.1.3
// [M_DP_TA_1] See companion standard 101,subclass 7.3.1.4
// [M_DP_TB_1] See companion standard 101,subclass 7.3.1.23
func double(c Connect, typeID TypeID, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...DoublePointInfo) error {
	if err := checkValid(c, typeID, isSequence, len(infos)); err != nil {
		return err
	}
	switch typeID {
	case M_DP_NA_1, M_DP_TA_1, M_DP_TB_1:
	default:
		return ErrTypeIDNotMatch
	}
	msg := DoublePointMsg{
		H:     newMessageHeader(c, typeID, coa, ca, isSequence, len(infos)),
		Items: infos,
	}
	return sendEncoded(c, msg)
}

// Double sends a type identification [M_DP_NA_1]. Double-point information
// [M_DP_NA_1] See companion standard 101, subclass 7.3.1.3
// Cause of transmission (coa) used for monitoring direction:
// <2> := Background scan
// <3> := Spontaneous
// <5> := Requested
// <11> := Return information caused by remote command
// <12> := Return information caused by local command
// <20> := Response to station interrogation
// <21> := Response to group 1 interrogation
// ...
// <36> := Response to group 16 interrogation
func Double(c Connect, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...DoublePointInfo) error {
	if !(coa.Cause == Background || coa.Cause == Spontaneous || coa.Cause == Request ||
		coa.Cause == ReturnInfoRemote || coa.Cause == ReturnInfoLocal ||
		(coa.Cause >= InterrogatedByStation && coa.Cause <= InterrogatedByGroup16)) {
		return ErrCmdCause
	}
	return double(c, M_DP_NA_1, isSequence, coa, ca, infos...)
}

// DoubleCP24Time2a sends a type identification [M_DP_TA_1]. Double-point information with CP24Time2a timestamp, only (SQ = 0) single information elements
// [M_DP_TA_1] See companion standard 101, subclass 7.3.1.4
// Cause of transmission (coa) used for monitoring direction:
// <3> := Spontaneous
// <5> := Requested
// <11> := Return information caused by remote command
// <12> := Return information caused by local command
func DoubleCP24Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...DoublePointInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request ||
		coa.Cause == ReturnInfoRemote || coa.Cause == ReturnInfoLocal) {
		return ErrCmdCause
	}
	return double(c, M_DP_TA_1, false, coa, ca, infos...)
}

// DoubleCP56Time2a sends a type identification [M_DP_TB_1]. Double-point information with CP56Time2a timestamp, only (SQ = 0) single information elements
// [M_DP_TB_1] See companion standard 101, subclass 7.3.1.23
// Cause of transmission (coa) used for monitoring direction:
// <3> := Spontaneous
// <5> := Requested
// <11> := Return information caused by remote command
// <12> := Return information caused by local command
func DoubleCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...DoublePointInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request ||
		coa.Cause == ReturnInfoRemote || coa.Cause == ReturnInfoLocal) {
		return ErrCmdCause
	}
	return double(c, M_DP_TB_1, false, coa, ca, infos...)
}

// StepPositionInfo the measured value attributes.
type StepPositionInfo struct {
	Ioa   InfoObjAddr
	Value StepPosition
	// Quality descriptor asdu.GOOD means no remarks.
	Qds QualityDescriptor
	// the type does not include timing will ignore
	Time time.Time
}

// step sends a type identification [M_ST_NA_1], [M_ST_TA_1] or [M_ST_TB_1]. Step position information
// [M_ST_NA_1] See companion standard 101, subclass 7.3.1.5
// [M_ST_TA_1] See companion standard 101, subclass 7.3.1.6
// [M_ST_TB_1] See companion standard 101, subclass 7.3.1.24
func step(c Connect, typeID TypeID, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...StepPositionInfo) error {
	if err := checkValid(c, typeID, isSequence, len(infos)); err != nil {
		return err
	}
	switch typeID {
	case M_ST_NA_1, M_ST_TA_1, M_ST_TB_1, M_SP_TB_1:
	default:
		return ErrTypeIDNotMatch
	}
	msg := StepPositionMsg{
		H:     newMessageHeader(c, typeID, coa, ca, isSequence, len(infos)),
		Items: infos,
	}
	return sendEncoded(c, msg)
}

// Step sends a type identification [M_ST_NA_1]. Step position information
// [M_ST_NA_1] See companion standard 101, subclass 7.3.1.5
// Cause of transmission (coa) used for monitoring direction:
// <2> := Background scan
// <3> := Spontaneous
// <5> := Requested
// <11> := Return information caused by remote command
// <12> := Return information caused by local command
// <20> := Response to station interrogation
// <21> := Response to group 1 interrogation
// ...
// <36> := Response to group 16 interrogation
func Step(c Connect, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...StepPositionInfo) error {
	if !(coa.Cause == Background || coa.Cause == Spontaneous || coa.Cause == Request ||
		coa.Cause == ReturnInfoRemote || coa.Cause == ReturnInfoLocal ||
		(coa.Cause >= InterrogatedByStation && coa.Cause <= InterrogatedByGroup16)) {
		return ErrCmdCause
	}
	return step(c, M_ST_NA_1, isSequence, coa, ca, infos...)
}

// StepCP24Time2a sends a type identification [M_ST_TA_1]. Step position information with CP24Time2a timestamp, only (SQ = 0) single information elements
// [M_ST_TA_1] See companion standard 101, subclass 7.3.1.6
// Cause of transmission (coa) used for monitoring direction:
// <3> := Spontaneous
// <5> := Requested
// <11> := Return information caused by remote command
// <12> := Return information caused by local command
func StepCP24Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...StepPositionInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request ||
		coa.Cause == ReturnInfoRemote || coa.Cause == ReturnInfoLocal) {
		return ErrCmdCause
	}
	return step(c, M_ST_TA_1, false, coa, ca, infos...)
}

// StepCP56Time2a sends a type identification [M_ST_TB_1]. Step position information with CP56Time2a timestamp, only (SQ = 0) single information elements
// [M_ST_TB_1] See companion standard 101, subclass 7.3.1.24
// Cause of transmission (coa) used for monitoring direction:
// <3> := Spontaneous
// <5> := Requested
// <11> := Return information caused by remote command
// <12> := Return information caused by local command
func StepCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...StepPositionInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request ||
		coa.Cause == ReturnInfoRemote || coa.Cause == ReturnInfoLocal) {
		return ErrCmdCause
	}
	return step(c, M_SP_TB_1, false, coa, ca, infos...)
}

// BitString32Info the measured value attributes.
type BitString32Info struct {
	Ioa   InfoObjAddr
	Value uint32
	// Quality descriptor asdu.GOOD means no remarks.
	Qds QualityDescriptor
	// the type does not include timing will ignore
	Time time.Time
}

// bitString32 sends a type identification [M_BO_NA_1], [M_BO_TA_1] or [M_BO_TB_1]. Bitstring (32 bits)
// [M_ST_NA_1] See companion standard 101, subclass 7.3.1.7
// [M_ST_TA_1] See companion standard 101, subclass 7.3.1.8
// [M_ST_TB_1] See companion standard 101, subclass 7.3.1.25
func bitString32(c Connect, typeID TypeID, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...BitString32Info) error {
	if err := checkValid(c, typeID, isSequence, len(infos)); err != nil {
		return err
	}
	switch typeID {
	case M_BO_NA_1, M_BO_TA_1, M_BO_TB_1:
	default:
		return ErrTypeIDNotMatch
	}
	msg := BitString32Msg{
		H:     newMessageHeader(c, typeID, coa, ca, isSequence, len(infos)),
		Items: infos,
	}
	return sendEncoded(c, msg)
}

// BitString32 sends a type identification [M_BO_NA_1]. Bitstring (32 bits)
// [M_ST_NA_1] See companion standard 101, subclass 7.3.1.7
// Cause of transmission (coa) used for monitoring direction:
// <2> := Background scan
// <3> := Spontaneous
// <5> := Requested
// <20> := Response to station interrogation
// <21> := Response to group 1 interrogation
// ...
// <36> := Response to group 16 interrogation
func BitString32(c Connect, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...BitString32Info) error {
	if !(coa.Cause == Background || coa.Cause == Spontaneous || coa.Cause == Request ||
		(coa.Cause >= InterrogatedByStation && coa.Cause <= InterrogatedByGroup16)) {
		return ErrCmdCause
	}
	return bitString32(c, M_BO_NA_1, isSequence, coa, ca, infos...)
}

// BitString32CP24Time2a sends a type identification [M_BO_TA_1]. Bitstring (32 bits) with CP24Time2a timestamp, only (SQ = 0) single information elements
// [M_ST_TA_1] See companion standard 101, subclass 7.3.1.8
// Cause of transmission (coa) used for monitoring direction:
// <3> := Spontaneous
// <5> := Requested
func BitString32CP24Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...BitString32Info) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request) {
		return ErrCmdCause
	}
	return bitString32(c, M_BO_TA_1, false, coa, ca, infos...)
}

// BitString32CP56Time2a sends a type identification [M_BO_TB_1]. Bitstring (32 bits) with CP56Time2a timestamp, only (SQ = 0) single information elements
// [M_ST_TB_1] See companion standard 101, subclass 7.3.1.25
// Cause of transmission (coa) used for monitoring direction:
// <3> := Spontaneous
// <5> := Requested
func BitString32CP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...BitString32Info) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request) {
		return ErrCmdCause
	}
	return bitString32(c, M_BO_TB_1, false, coa, ca, infos...)
}

// MeasuredValueNormalInfo the measured value attributes.
type MeasuredValueNormalInfo struct {
	Ioa   InfoObjAddr
	Value Normalize
	// Quality descriptor asdu.GOOD means no remarks.
	Qds QualityDescriptor
	// the type does not include timing will ignore
	Time time.Time
}

// measuredValueNormal sends a type identification [M_ME_NA_1], [M_ME_TA_1], [M_ME_TD_1] or [M_ME_ND_1]. Measured value, normalized value
// [M_ME_NA_1] See companion standard 101, subclass 7.3.1.9
// [M_ME_TA_1] See companion standard 101, subclass 7.3.1.10
// [M_ME_TD_1] See companion standard 101, subclass 7.3.1.26
// [M_ME_ND_1] See companion standard 101, subclass 7.3.1.21. The quality descriptor must default to asdu.GOOD
func measuredValueNormal(c Connect, typeID TypeID, isSequence bool, coa CauseOfTransmission, ca CommonAddr, attrs ...MeasuredValueNormalInfo) error {
	if err := checkValid(c, typeID, isSequence, len(attrs)); err != nil {
		return err
	}
	switch typeID {
	case M_ME_NA_1, M_ME_TA_1, M_ME_TD_1, M_ME_ND_1:
	default:
		return ErrTypeIDNotMatch
	}
	msg := MeasuredValueNormalMsg{
		H:     newMessageHeader(c, typeID, coa, ca, isSequence, len(attrs)),
		Items: attrs,
	}
	return sendEncoded(c, msg)
}

// MeasuredValueNormal sends a type identification [M_ME_NA_1]. Measured value, normalized value
// [M_ME_NA_1] See companion standard 101, subclass 7.3.1.9
// Cause of transmission (coa) used for monitoring direction:
// <1> := Periodic/cyclic
// <2> := Background scan
// <3> := Spontaneous
// <5> := Requested
// <20> := Response to station interrogation
// <21> := Response to group 1 interrogation
// ...
// <36> := Response to group 16 interrogation
func MeasuredValueNormal(c Connect, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...MeasuredValueNormalInfo) error {
	if !(coa.Cause == Periodic || coa.Cause == Background ||
		coa.Cause == Spontaneous || coa.Cause == Request ||
		(coa.Cause >= InterrogatedByStation && coa.Cause <= InterrogatedByGroup16)) {
		return ErrCmdCause
	}
	return measuredValueNormal(c, M_ME_NA_1, isSequence, coa, ca, infos...)
}

// MeasuredValueNormalCP24Time2a sends a type identification [M_ME_TA_1]. Measured value, normalized value with CP24Time2a timestamp, only (SQ = 0) single information elements
// [M_ME_TA_1] See companion standard 101, subclass 7.3.1.10
// Cause of transmission (coa) used for monitoring direction:
// <3> := Spontaneous
// <5> := Requested
func MeasuredValueNormalCP24Time2a(c Connect, coa CauseOfTransmission,
	ca CommonAddr, infos ...MeasuredValueNormalInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request) {
		return ErrCmdCause
	}
	return measuredValueNormal(c, M_ME_TA_1, false, coa, ca, infos...)
}

// MeasuredValueNormalCP56Time2a sends a type identification [M_ME_TD_1]. Measured value, normalized value with CP56Time2a timestamp, only (SQ = 0) single information elements
// [M_ME_TD_1] See companion standard 101, subclass 7.3.1.26
// Cause of transmission (coa) used for monitoring direction:
// <3> := Spontaneous
// <5> := Requested
func MeasuredValueNormalCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...MeasuredValueNormalInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request) {
		return ErrCmdCause
	}
	return measuredValueNormal(c, M_ME_TD_1, false, coa, ca, infos...)
}

// MeasuredValueNormalNoQuality sends a type identification [M_ME_ND_1]. Measured value, normalized value without quality
// [M_ME_ND_1] See companion standard 101, subclass 7.3.1.21
// The quality descriptor must default to asdu.GOOD
// Cause of transmission (coa) used for monitoring direction:
// <1> := Periodic/cyclic
// <2> := Background scan
// <3> := Spontaneous
// <5> := Requested
// <20> := Response to station interrogation
// <21> := Response to group 1 interrogation
// ...
// <36> := Response to group 16 interrogation
func MeasuredValueNormalNoQuality(c Connect, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...MeasuredValueNormalInfo) error {
	if !(coa.Cause == Periodic || coa.Cause == Background ||
		coa.Cause == Spontaneous || coa.Cause == Request ||
		(coa.Cause >= InterrogatedByStation && coa.Cause <= InterrogatedByGroup16)) {
		return ErrCmdCause
	}
	return measuredValueNormal(c, M_ME_ND_1, isSequence, coa, ca, infos...)
}

// MeasuredValueScaledInfo the measured value attributes.
type MeasuredValueScaledInfo struct {
	Ioa   InfoObjAddr
	Value int16
	// Quality descriptor asdu.GOOD means no remarks.
	Qds QualityDescriptor
	// the type does not include timing will ignore
	Time time.Time
}

// measuredValueScaled sends a type identification [M_ME_NB_1], [M_ME_TB_1] or [M_ME_TE_1]. Measured value, scaled value
// [M_ME_NB_1] See companion standard 101, subclass 7.3.1.11
// [M_ME_TB_1] See companion standard 101, subclass 7.3.1.12
// [M_ME_TE_1] See companion standard 101, subclass 7.3.1.27
func measuredValueScaled(c Connect, typeID TypeID, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...MeasuredValueScaledInfo) error {
	if err := checkValid(c, typeID, isSequence, len(infos)); err != nil {
		return err
	}
	switch typeID {
	case M_ME_NB_1, M_ME_TB_1, M_ME_TE_1:
	default:
		return ErrTypeIDNotMatch
	}
	msg := MeasuredValueScaledMsg{
		H:     newMessageHeader(c, typeID, coa, ca, isSequence, len(infos)),
		Items: infos,
	}
	return sendEncoded(c, msg)
}

// MeasuredValueScaled sends a type identification [M_ME_NB_1]. Measured value, scaled value
// [M_ME_NB_1] See companion standard 101, subclass 7.3.1.11
// Cause of transmission (coa) used for
// Monitoring direction:
// <1> := Periodic/cyclic
// <2> := Background scan
// <3> := Spontaneous
// <5> := Requested
// <20> := Response to station interrogation
// <21> := Response to group 1 interrogation
// ...
// <36> := Response to group 16 interrogation
func MeasuredValueScaled(c Connect, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...MeasuredValueScaledInfo) error {
	if !(coa.Cause == Periodic || coa.Cause == Background ||
		coa.Cause == Spontaneous || coa.Cause == Request ||
		(coa.Cause >= InterrogatedByStation && coa.Cause <= InterrogatedByGroup16)) {
		return ErrCmdCause
	}
	return measuredValueScaled(c, M_ME_NB_1, isSequence, coa, ca, infos...)
}

// MeasuredValueScaledCP24Time2a sends a type identification [M_ME_TB_1]. Measured value, scaled value with CP24Time2a timestamp, only (SQ = 0) single information elements
// [M_ME_TB_1] See companion standard 101, subclass 7.3.1.12
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
// <5> := Requested
func MeasuredValueScaledCP24Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...MeasuredValueScaledInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request) {
		return ErrCmdCause
	}
	return measuredValueScaled(c, M_ME_TB_1, false, coa, ca, infos...)
}

// MeasuredValueScaledCP56Time2a sends a type identification [M_ME_TE_1]. Measured value, scaled value with CP56Time2a timestamp, only (SQ = 0) single information elements
// [M_ME_TE_1] See companion standard 101, subclass 7.3.1.27
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
// <5> := Requested
func MeasuredValueScaledCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...MeasuredValueScaledInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request) {
		return ErrCmdCause
	}
	return measuredValueScaled(c, M_ME_TE_1, false, coa, ca, infos...)
}

// MeasuredValueFloatInfo the measured value attributes.
type MeasuredValueFloatInfo struct {
	Ioa   InfoObjAddr
	Value float32
	// Quality descriptor asdu.GOOD means no remarks.
	Qds QualityDescriptor
	// the type does not include timing will ignore
	Time time.Time
}

// measuredValueFloat sends a type identification [M_ME_NC_1], [M_ME_TC_1] or [M_ME_TF_1]. Measured value, short floating point
// [M_ME_NC_1] See companion standard 101, subclass 7.3.1.13
// [M_ME_TC_1] See companion standard 101, subclass 7.3.1.14
// [M_ME_TF_1] See companion standard 101, subclass 7.3.1.28
func measuredValueFloat(c Connect, typeID TypeID, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...MeasuredValueFloatInfo) error {
	if err := checkValid(c, typeID, isSequence, len(infos)); err != nil {
		return err
	}
	switch typeID {
	case M_ME_NC_1, M_ME_TC_1, M_ME_TF_1:
	default:
		return ErrTypeIDNotMatch
	}
	msg := MeasuredValueFloatMsg{
		H:     newMessageHeader(c, typeID, coa, ca, isSequence, len(infos)),
		Items: infos,
	}
	return sendEncoded(c, msg)
}

// MeasuredValueFloat sends a type identification [M_ME_TF_1]. Measured value, short floating point
// [M_ME_NC_1] See companion standard 101, subclass 7.3.1.13
// Cause of transmission (coa) used for
// Monitoring direction:
// <1> := Periodic/cyclic
// <2> := Background scan
// <3> := Spontaneous
// <5> := Requested
// <20> := Response to station interrogation
// <21> := Response to group 1 interrogation
// ...
// <36> := Response to group 16 interrogation
func MeasuredValueFloat(c Connect, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...MeasuredValueFloatInfo) error {
	if !(coa.Cause == Periodic || coa.Cause == Background ||
		coa.Cause == Spontaneous || coa.Cause == Request ||
		(coa.Cause >= InterrogatedByStation && coa.Cause <= InterrogatedByGroup16)) {
		return ErrCmdCause
	}
	return measuredValueFloat(c, M_ME_NC_1, isSequence, coa, ca, infos...)
}

// MeasuredValueFloatCP24Time2a sends a type identification [M_ME_TC_1]. Measured value, short floating point with CP24Time2a timestamp, only (SQ = 0) single information elements
// [M_ME_TC_1] See companion standard 101, subclass 7.3.1.14
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
// <5> := Requested
func MeasuredValueFloatCP24Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...MeasuredValueFloatInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request) {
		return ErrCmdCause
	}
	return measuredValueFloat(c, M_ME_TC_1, false, coa, ca, infos...)
}

// MeasuredValueFloatCP56Time2a sends a type identification [M_ME_TF_1]. Measured value, short floating point with CP56Time2a timestamp, only (SQ = 0) single information elements
// [M_ME_TF_1] See companion standard 101, subclass 7.3.1.28
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
// <5> := Requested
func MeasuredValueFloatCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...MeasuredValueFloatInfo) error {
	if !(coa.Cause == Spontaneous || coa.Cause == Request) {
		return ErrCmdCause
	}
	return measuredValueFloat(c, M_ME_TF_1, false, coa, ca, infos...)
}

// BinaryCounterReadingInfo the counter reading attributes. Binary counter reading
type BinaryCounterReadingInfo struct {
	Ioa   InfoObjAddr
	Value BinaryCounterReading
	// the type does not include timing will ignore
	Time time.Time
}

// integratedTotals sends a type identification [M_IT_NA_1], [M_IT_TA_1] or [M_IT_TB_1]. Integrated totals
// [M_IT_NA_1] See companion standard 101, subclass 7.3.1.15
// [M_IT_TA_1] See companion standard 101, subclass 7.3.1.16
// [M_IT_TB_1] See companion standard 101, subclass 7.3.1.29
func integratedTotals(c Connect, typeID TypeID, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...BinaryCounterReadingInfo) error {
	if err := checkValid(c, typeID, isSequence, len(infos)); err != nil {
		return err
	}
	switch typeID {
	case M_IT_NA_1, M_IT_TA_1, M_IT_TB_1:
	default:
		return ErrTypeIDNotMatch
	}
	msg := IntegratedTotalsMsg{
		H:     newMessageHeader(c, typeID, coa, ca, isSequence, len(infos)),
		Items: infos,
	}
	return sendEncoded(c, msg)
}

// IntegratedTotals sends a type identification [M_IT_NA_1]. Integrated totals
// [M_IT_NA_1] See companion standard 101, subclass 7.3.1.15
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
// <37> := Response to general counter interrogation
// <38> := Response to group 1 counter interrogation
// <39> := Response to group 2 counter interrogation
// <40> := Response to group 3 counter interrogation
// <41> := Response to group 4 counter interrogation
func IntegratedTotals(c Connect, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...BinaryCounterReadingInfo) error {
	if !(coa.Cause == Spontaneous || (coa.Cause >= RequestByGeneralCounter && coa.Cause <= RequestByGroup4Counter)) {
		return ErrCmdCause
	}
	return integratedTotals(c, M_IT_NA_1, isSequence, coa, ca, infos...)
}

// IntegratedTotalsCP24Time2a sends a type identification [M_IT_TA_1]. Integrated totals with CP24Time2a timestamp, only (SQ = 0) single information elements
// [M_IT_TA_1] See companion standard 101, subclass 7.3.1.16
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
// <37> := Response to general counter interrogation
// <38> := Response to group 1 counter interrogation
// <39> := Response to group 2 counter interrogation
// <40> := Response to group 3 counter interrogation
// <41> := Response to group 4 counter interrogation
func IntegratedTotalsCP24Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...BinaryCounterReadingInfo) error {
	if !(coa.Cause == Spontaneous || (coa.Cause >= RequestByGeneralCounter && coa.Cause <= RequestByGroup4Counter)) {
		return ErrCmdCause
	}
	return integratedTotals(c, M_IT_TA_1, false, coa, ca, infos...)
}

// IntegratedTotalsCP56Time2a sends a type identification [M_IT_TB_1]. Integrated totals with CP56Time2a timestamp, only (SQ = 0) single information elements
// [M_IT_TB_1] See companion standard 101, subclass 7.3.1.29
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
// <37> := Response to general counter interrogation
// <38> := Response to group 1 counter interrogation
// <39> := Response to group 2 counter interrogation
// <40> := Response to group 3 counter interrogation
// <41> := Response to group 4 counter interrogation
func IntegratedTotalsCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...BinaryCounterReadingInfo) error {
	if !(coa.Cause == Spontaneous || (coa.Cause >= RequestByGeneralCounter && coa.Cause <= RequestByGroup4Counter)) {
		return ErrCmdCause
	}
	return integratedTotals(c, M_IT_TB_1, false, coa, ca, infos...)
}

// EventOfProtectionEquipmentInfo the protection equipment event attributes.
type EventOfProtectionEquipmentInfo struct {
	Ioa   InfoObjAddr
	Event SingleEvent
	Qdp   QualityDescriptorProtection
	Msec  uint16
	// the type does not include timing will ignore
	Time time.Time
}

// eventOfProtectionEquipment sends a type identification [M_EP_TA_1], [M_EP_TD_1]. Event of protection equipment (relay protection device event)
// [M_EP_TA_1] See companion standard 101, subclass 7.3.1.17
// [M_EP_TD_1] See companion standard 101, subclass 7.3.1.30
func eventOfProtectionEquipment(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, infos ...EventOfProtectionEquipmentInfo) error {
	if coa.Cause != Spontaneous {
		return ErrCmdCause
	}
	if err := checkValid(c, typeID, false, len(infos)); err != nil {
		return err
	}
	switch typeID {
	case M_EP_TA_1, M_EP_TD_1:
	default:
		return ErrTypeIDNotMatch
	}
	msg := EventOfProtectionMsg{
		H:     newMessageHeader(c, typeID, coa, ca, false, len(infos)),
		Items: infos,
	}
	return sendEncoded(c, msg)
}

// EventOfProtectionEquipmentCP24Time2a sends a type identification [M_EP_TA_1]. Event of protection equipment with CP24Time2a timestamp
// [M_EP_TA_1] See companion standard 101, subclass 7.3.1.17
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
func EventOfProtectionEquipmentCP24Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...EventOfProtectionEquipmentInfo) error {
	return eventOfProtectionEquipment(c, M_EP_TA_1, coa, ca, infos...)
}

// EventOfProtectionEquipmentCP56Time2a sends a type identification [M_EP_TD_1]. Event of protection equipment with CP56Time2a timestamp
// [M_EP_TD_1] See companion standard 101, subclass 7.3.1.30
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
func EventOfProtectionEquipmentCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, infos ...EventOfProtectionEquipmentInfo) error {
	return eventOfProtectionEquipment(c, M_EP_TD_1, coa, ca, infos...)
}

// PackedStartEventsOfProtectionEquipmentInfo Packed start events of protection equipment (group start events)
type PackedStartEventsOfProtectionEquipmentInfo struct {
	Ioa   InfoObjAddr
	Event StartEvent
	Qdp   QualityDescriptorProtection
	Msec  uint16
	// the type does not include timing will ignore
	Time time.Time
}

// packedStartEventsOfProtectionEquipment sends a type identification [M_EP_TB_1], [M_EP_TE_1]. Packed start events of protection equipment
// [M_EP_TB_1] See companion standard 101, subclass 7.3.1.18
// [M_EP_TE_1] See companion standard 101, subclass 7.3.1.31
func packedStartEventsOfProtectionEquipment(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, info PackedStartEventsOfProtectionEquipmentInfo) error {
	if coa.Cause != Spontaneous {
		return ErrCmdCause
	}
	if err := checkValid(c, typeID, false, 1); err != nil {
		return err
	}
	switch typeID {
	case M_EP_TB_1, M_EP_TE_1:
	default:
		return ErrTypeIDNotMatch
	}
	msg := PackedStartEventsMsg{
		H:    newMessageHeader(c, typeID, coa, ca, false, 1),
		Item: info,
	}
	return sendEncoded(c, msg)
}

// PackedStartEventsOfProtectionEquipmentCP24Time2a sends a type identification [M_EP_TB_1]. Packed start events of protection equipment with CP24Time2a timestamp
// [M_EP_TB_1] See companion standard 101, subclass 7.3.1.18
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
func PackedStartEventsOfProtectionEquipmentCP24Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, info PackedStartEventsOfProtectionEquipmentInfo) error {
	return packedStartEventsOfProtectionEquipment(c, M_EP_TB_1, coa, ca, info)
}

// PackedStartEventsOfProtectionEquipmentCP56Time2a sends a type identification [M_EP_TE_1]. Packed start events of protection equipment with CP56Time2a timestamp
// [M_EP_TE_1] See companion standard 101, subclass 7.3.1.31
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
func PackedStartEventsOfProtectionEquipmentCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, info PackedStartEventsOfProtectionEquipmentInfo) error {
	return packedStartEventsOfProtectionEquipment(c, M_EP_TE_1, coa, ca, info)
}

// PackedOutputCircuitInfoInfo Packed output circuit information of protection equipment (grouped)
type PackedOutputCircuitInfoInfo struct {
	Ioa  InfoObjAddr
	Oci  OutputCircuitInfo
	Qdp  QualityDescriptorProtection
	Msec uint16
	// the type does not include timing will ignore
	Time time.Time
}

// packedOutputCircuitInfo sends a type identification [M_EP_TC_1], [M_EP_TF_1]. Packed output circuit information of protection equipment (grouped)
// [M_EP_TC_1] See companion standard 101, subclass 7.3.1.19
// [M_EP_TF_1] See companion standard 101, subclass 7.3.1.32
func packedOutputCircuitInfo(c Connect, typeID TypeID, coa CauseOfTransmission, ca CommonAddr, info PackedOutputCircuitInfoInfo) error {
	if coa.Cause != Spontaneous {
		return ErrCmdCause
	}
	if err := checkValid(c, typeID, false, 1); err != nil {
		return err
	}
	switch typeID {
	case M_EP_TC_1, M_EP_TF_1:
	default:
		return ErrTypeIDNotMatch
	}
	msg := PackedOutputCircuitMsg{
		H:    newMessageHeader(c, typeID, coa, ca, false, 1),
		Item: info,
	}
	return sendEncoded(c, msg)
}

// PackedOutputCircuitInfoCP24Time2a sends a type identification [M_EP_TC_1]. Packed output circuit information of protection equipment with CP24Time2a timestamp (grouped)
// [M_EP_TC_1] See companion standard 101, subclass 7.3.1.19
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
func PackedOutputCircuitInfoCP24Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, info PackedOutputCircuitInfoInfo) error {
	return packedOutputCircuitInfo(c, M_EP_TC_1, coa, ca, info)
}

// PackedOutputCircuitInfoCP56Time2a sends a type identification [M_EP_TF_1]. Packed output circuit information of protection equipment with CP56Time2a timestamp (grouped)
// [M_EP_TF_1] See companion standard 101, subclass 7.3.1.32
// Cause of transmission (coa) used for
// Monitoring direction:
// <3> := Spontaneous
func PackedOutputCircuitInfoCP56Time2a(c Connect, coa CauseOfTransmission, ca CommonAddr, info PackedOutputCircuitInfoInfo) error {
	return packedOutputCircuitInfo(c, M_EP_TF_1, coa, ca, info)
}

// PackedSinglePointWithSCDInfo Grouped single-point information with change detection
type PackedSinglePointWithSCDInfo struct {
	Ioa InfoObjAddr
	Scd StatusAndStatusChangeDetection
	Qds QualityDescriptor
}

// PackedSinglePointWithSCD sends a type identification [M_PS_NA_1]. Grouped single-point information with change detection
// [M_PS_NA_1] See companion standard 101, subclass 7.3.1.20
// Cause of transmission (coa) used for
// Monitoring direction:
// <2> := Background scan
// <3> := Spontaneous
// <5> := Requested
// <11> := Return information caused by a remote command
// <12> := Return information caused by a local command
// <20> := Response to station interrogation
// <21> := Response to group 1 interrogation
// to
// <36> := Response to group 16 interrogation
func PackedSinglePointWithSCD(c Connect, isSequence bool, coa CauseOfTransmission, ca CommonAddr, infos ...PackedSinglePointWithSCDInfo) error {
	if !(coa.Cause == Background || coa.Cause == Spontaneous || coa.Cause == Request ||
		coa.Cause == ReturnInfoRemote || coa.Cause == ReturnInfoLocal ||
		(coa.Cause >= InterrogatedByStation && coa.Cause <= InterrogatedByGroup16)) {
		return ErrCmdCause
	}
	if err := checkValid(c, M_PS_NA_1, isSequence, len(infos)); err != nil {
		return err
	}
	msg := PackedSinglePointWithSCDMsg{
		H:     newMessageHeader(c, M_PS_NA_1, coa, ca, isSequence, len(infos)),
		Items: infos,
	}
	return sendEncoded(c, msg)
}

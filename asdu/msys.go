// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package asdu

// Application service data unit for system information in the monitoring direction

// EndOfInitialization sends type identification [M_EI_NA_1]; end of initialization; only a single information object (SQ = 0)
// [M_EI_NA_1] See companion standard 101,subclass 7.3.3.1
// Cause of transmission (coa) used
// in the monitoring direction:
// <4> := initialized
func EndOfInitialization(c Connect, coa CauseOfTransmission, ca CommonAddr, ioa InfoObjAddr, coi CauseOfInitial) error {
	if err := c.Params().Valid(); err != nil {
		return err
	}

	coa.Cause = Initialized
	u := NewASDU(c.Params(), Identifier{
		M_EI_NA_1,
		VariableStruct{IsSequence: false, Number: 1},
		coa,
		0,
		ca,
	})

	if err := u.appendInfoObjAddr(ioa); err != nil {
		return err
	}
	u.appendBytes(coi.Value())
	return c.Send(u)
}

// GetEndOfInitialization [M_EI_NA_1] Retrieve end of initialization (idempotent)
func (sf *ASDU) GetEndOfInitialization() (InfoObjAddr, CauseOfInitial) { // idempotent
	saved := sf.infoObj
	defer func() { sf.infoObj = saved }()
	return sf.decodeInfoObjAddr(), ParseCauseOfInitial(sf.infoObj[0])
}

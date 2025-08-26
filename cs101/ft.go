// SPDX-License-Identifier: MIT
// Copyright (c) 2025 go-iecp5 contributors.

package cs101

// Use FT1.2 frame format
const (
	startVarFrame byte = 0x68 // start character of variable-length frame
	startFixFrame byte = 0x10 // start character of fixed-length frame
	endFrame      byte = 0x16
)

// Control field definitions
const (

	// Master-to-slave specific
	FCV = 1 << 4 // frame count valid bit
	FCB = 1 << 5 // frame count bit
	// Slave-to-master specific
	DFC     = 1 << 4 // Data Flow Control bit
	ACD_RES = 1 << 5 // Access Demand bit (unbalanced ACD, balanced reserved)
	// Primary Message bit:
	// PRM = 0, message transmitted from controlled station to primary station;
	// PRM = 1, message transmitted from primary station to controlled station
	RPM     = 1 << 6
	RES_DIR = 1 << 7 // Unbalanced reserved, balanced for direction

	// Control field function codes for messages transmitted from primary station to controlled station (PRM = 1)
	FccResetRemoteLink                 = iota // Reset remote link
	FccResetUserProcess                       // Reset user process
	FccBalanceTestLink                        // Link test function
	FccUserDataWithConfirmed                  // User data, confirmation required
	FccUserDataWithUnconfirmed                // User data, no confirmation required
	_                                         // Reserved
	_                                         // Defined by manufacturer and user agreement
	_                                         // Defined by manufacturer and user agreement
	FccUnbalanceWithRequestBitResponse        // Response with request bit
	FccLinkStatus                             // Request link status
	FccUnbalanceLevel1UserData                // Request level 1 user data
	FccUnbalanceLevel2UserData                // Request level 2 user data
	// 12-13: Reserved
	// 14-15: Defined by manufacturer and user agreement

	// Control field function codes for messages transmitted from controlled station to primary station (PRM = 0)
	FcsConfirmed                 = iota // Confirm: Positive acknowledgment
	FcsNConfirmed                       // Negative acknowledgment: Message not received, link busy
	_                                   // Reserved
	_                                   // Reserved
	_                                   // Reserved
	_                                   // Reserved
	_                                   // Defined by manufacturer and user agreement
	_                                   // Defined by manufacturer and user agreement
	FcsUnbalanceResponse                // User data
	FcsUnbalanceNegativeResponse        // Negative acknowledgment: No data requested
	_                                   // Reserved
	FcsStatus                           // Link status or access demand
	// 12: Reserved
	// 13: Defined by manufacturer and user agreement
	// 14: Link service not functioning
	// 15: Link service not completed
)

// Ft12 ...
type Ft12 struct {
	start        byte
	apduFiledLen byte
	ctrl         byte
	address      uint16
	checksum     byte
	end          byte
}

package asdu

import (
	"testing"
	"time"
)

func TestInterrogationCmd(t *testing.T) {
	type args struct {
		c   Connect
		coa CauseOfTransmission
		ca  CommonAddr
		qoi QualifierOfInterrogation
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"cause not Activation and Deactivation",
			args{
				newConn(nil, t),
				CauseOfTransmission{Cause: Unused},
				0x1234,
				QOIGroup1},
			true,
		},
		{
			"C_IC_NA_1",
			args{
				newConn([]byte{byte(C_IC_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 21}, t),
				CauseOfTransmission{Cause: Activation},
				0x1234,
				QOIGroup1},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InterrogationCmd(tt.args.c, tt.args.coa, tt.args.ca, tt.args.qoi); (err != nil) != tt.wantErr {
				t.Errorf("InterrogationCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCounterInterrogationCmd(t *testing.T) {
	type args struct {
		c   Connect
		coa CauseOfTransmission
		ca  CommonAddr
		qcc QualifierCountCall
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_CI_NA_1",
			args{
				newConn([]byte{byte(C_CI_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 0x01}, t),
				CauseOfTransmission{Cause: Activation},
				0x1234,
				QualifierCountCall{QCCGroup1, QCCFrzRead}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CounterInterrogationCmd(tt.args.c, tt.args.coa, tt.args.ca, tt.args.qcc); (err != nil) != tt.wantErr {
				t.Errorf("CounterInterrogationCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReadCmd(t *testing.T) {
	type args struct {
		c   Connect
		coa CauseOfTransmission
		ca  CommonAddr
		ioa InfoObjAddr
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_RD_NA_1",
			args{
				newConn([]byte{byte(C_RD_NA_1), 0x01, 0x05, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56}, t),
				CauseOfTransmission{Cause: Request},
				0x1234,
				0x567890},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ReadCmd(tt.args.c, tt.args.coa, tt.args.ca, tt.args.ioa); (err != nil) != tt.wantErr {
				t.Errorf("ReadCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClockSynchronizationCmd(t *testing.T) {
	type args struct {
		c   Connect
		coa CauseOfTransmission
		ca  CommonAddr
		t   time.Time
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_CS_NA_1",
			args{
				newConn(append([]byte{byte(C_CS_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00}, tm0CP56Time2aBytes...), t),
				CauseOfTransmission{Cause: Activation},
				0x1234,
				tm0},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ClockSynchronizationCmd(tt.args.c, tt.args.coa, tt.args.ca, tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("ClockSynchronizationCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTestCommand(t *testing.T) {
	type args struct {
		c   Connect
		coa CauseOfTransmission
		ca  CommonAddr
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_TS_NA_1",
			args{
				newConn([]byte{byte(C_TS_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 0xaa, 0x55}, t),
				CauseOfTransmission{Cause: Activation},
				0x1234},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TestCommand(tt.args.c, tt.args.coa, tt.args.ca); (err != nil) != tt.wantErr {
				t.Errorf("TestCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResetProcessCmd(t *testing.T) {
	type args struct {
		c   Connect
		coa CauseOfTransmission
		ca  CommonAddr
		qrp QualifierOfResetProcessCmd
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_RP_NA_1",
			args{
				newConn([]byte{byte(C_RP_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 0x01}, t),
				CauseOfTransmission{Cause: Activation},
				0x1234,
				QPRGeneralRest},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ResetProcessCmd(tt.args.c, tt.args.coa, tt.args.ca, tt.args.qrp); (err != nil) != tt.wantErr {
				t.Errorf("ResetProcessCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDelayAcquireCommand(t *testing.T) {
	type args struct {
		c    Connect
		coa  CauseOfTransmission
		ca   CommonAddr
		msec uint16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"cause not act and spont",
			args{
				newConn(nil, t),
				CauseOfTransmission{Cause: Unused},
				0x1234,
				10000},
			true,
		},
		{
			"C_CD_NA_1",
			args{
				newConn([]byte{byte(C_CD_NA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 0x10, 0x27}, t),
				CauseOfTransmission{Cause: Activation},
				0x1234,
				10000},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DelayAcquireCommand(tt.args.c, tt.args.coa, tt.args.ca, tt.args.msec); (err != nil) != tt.wantErr {
				t.Errorf("DelayAcquireCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTestCommandCP56Time2a(t *testing.T) {
	type args struct {
		c   Connect
		coa CauseOfTransmission
		ca  CommonAddr
		t   time.Time
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"C_TS_TA_1",
			args{
				newConn(append([]byte{byte(C_TS_TA_1), 0x01, 0x06, 0x00, 0x34, 0x12,
					0x00, 0x00, 0x00, 0xaa, 0x55}, tm0CP56Time2aBytes...), t),
				CauseOfTransmission{Cause: Activation},
				0x1234,
				tm0},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TestCommandCP56Time2a(tt.args.c, tt.args.coa, tt.args.ca, tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("TestCommandCP56Time2a() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

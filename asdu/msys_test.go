package asdu

import "testing"

func TestEndOfInitialization(t *testing.T) {
	type args struct {
		c   Connect
		coa CauseOfTransmission
		ca  CommonAddr
		ioa InfoObjAddr
		coi CauseOfInitial
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"M_EI_NA_1",
			args{
				newConn([]byte{byte(M_EI_NA_1), 0x01, 0x04, 0x00, 0x34, 0x12,
					0x90, 0x78, 0x56, 0x01}, t),
				CauseOfTransmission{Cause: Initialized},
				0x1234,
				0x567890,
				CauseOfInitial{COILocalHandReset, false}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EndOfInitialization(tt.args.c, tt.args.coa, tt.args.ca, tt.args.ioa, tt.args.coi); (err != nil) != tt.wantErr {
				t.Errorf("EndOfInitialization() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

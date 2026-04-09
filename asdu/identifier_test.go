package asdu

import (
	"reflect"
	"strconv"
	"testing"
)

func TestGetInfoObjSize(t *testing.T) {
	type args struct {
		id TypeID
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"defined", args{F_DR_TA_1}, 13, false},
		{"no defined", args{F_SG_NA_1}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetInfoObjSize(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInfoObjSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetInfoObjSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeID_String(t *testing.T) {
	tests := []struct {
		name string
		this TypeID
		want string
	}{
		{"M_SP_NA_1", M_SP_NA_1, "M_SP_NA_1"},
		{"M_SP_TB_1", M_SP_TB_1, "M_SP_TB_1"},
		{"C_SC_NA_1", C_SC_NA_1, "C_SC_NA_1"},
		{"C_SC_TA_1", C_SC_TA_1, "C_SC_TA_1"},
		{"M_EI_NA_1", M_EI_NA_1, "M_EI_NA_1"},
		{"S_CH_NA_1", S_CH_NA_1, "S_CH_NA_1"},
		{"S_US_NA_1", S_US_NA_1, "S_US_NA_1"},
		{"C_IC_NA_1", C_IC_NA_1, "C_IC_NA_1"},
		{"P_ME_NA_1", P_ME_NA_1, "P_ME_NA_1"},
		{"F_FR_NA_1", F_FR_NA_1, "F_FR_NA_1"},
		{"no defined", 0, "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.String(); got != tt.want {
				t.Errorf("TypeID.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeID_Info(t *testing.T) {
	tests := []struct {
		name string
		id   TypeID
		want TypeInfo
	}{
		{
			name: "M_SP_NA_1 — monitored, no time tag",
			id:   M_SP_NA_1,
			want: TypeInfo{
				TypeID:         M_SP_NA_1,
				Name:           "M_SP_NA_1",
				Description:    "single-point information",
				Direction:      DirectionMonitor,
				Category:       CategoryProcessInfoMonitor,
				TimeTagFormat:  TimeTagNone,
				IsCommand:      false,
				HasTimeTag:     false,
				AllowedIn104:   true,
				InfoObjectSize: 1,
			},
		},
		{
			name: "M_SP_TA_1 — CP24Time2a not allowed in 104",
			id:   M_SP_TA_1,
			want: TypeInfo{
				TypeID:         M_SP_TA_1,
				Name:           "M_SP_TA_1",
				Description:    "single-point information with time tag",
				Direction:      DirectionMonitor,
				Category:       CategoryProcessInfoMonitor,
				TimeTagFormat:  TimeTagCP24Time2a,
				IsCommand:      false,
				HasTimeTag:     true,
				AllowedIn104:   false,
				InfoObjectSize: 4,
			},
		},
		{
			name: "M_SP_TB_1 — CP56Time2a, allowed in 104",
			id:   M_SP_TB_1,
			want: TypeInfo{
				TypeID:         M_SP_TB_1,
				Name:           "M_SP_TB_1",
				Description:    "single-point information with time tag CP56Time2a",
				Direction:      DirectionMonitor,
				Category:       CategoryProcessInfoMonitor,
				TimeTagFormat:  TimeTagCP56Time2a,
				IsCommand:      false,
				HasTimeTag:     true,
				AllowedIn104:   true,
				InfoObjectSize: 8,
			},
		},
		{
			name: "S_IT_TC_1 — security stat in monitor range, CP56",
			id:   S_IT_TC_1,
			want: TypeInfo{
				TypeID:         S_IT_TC_1,
				Name:           "S_IT_TC_1",
				Description:    "integrated totals containing time-tagged security statistics",
				Direction:      DirectionMonitor,
				Category:       CategoryProcessInfoMonitor,
				TimeTagFormat:  TimeTagCP56Time2a,
				IsCommand:      false,
				HasTimeTag:     true,
				AllowedIn104:   true,
				InfoObjectSize: 0,
			},
		},
		{
			name: "C_SC_NA_1 — process control command",
			id:   C_SC_NA_1,
			want: TypeInfo{
				TypeID:         C_SC_NA_1,
				Name:           "C_SC_NA_1",
				Description:    "single command",
				Direction:      DirectionControl,
				Category:       CategoryProcessInfoControl,
				TimeTagFormat:  TimeTagNone,
				IsCommand:      true,
				HasTimeTag:     false,
				AllowedIn104:   true,
				InfoObjectSize: 1,
			},
		},
		{
			name: "C_SC_TA_1 — process control command with CP56",
			id:   C_SC_TA_1,
			want: TypeInfo{
				TypeID:         C_SC_TA_1,
				Name:           "C_SC_TA_1",
				Description:    "single command with time tag CP56Time2a",
				Direction:      DirectionControl,
				Category:       CategoryProcessInfoControl,
				TimeTagFormat:  TimeTagCP56Time2a,
				IsCommand:      true,
				HasTimeTag:     true,
				AllowedIn104:   true,
				InfoObjectSize: 0,
			},
		},
		{
			name: "M_EI_NA_1 — end of init, system info monitor direction",
			id:   M_EI_NA_1,
			want: TypeInfo{
				TypeID:         M_EI_NA_1,
				Name:           "M_EI_NA_1",
				Description:    "end of initialization",
				Direction:      DirectionMonitor,
				Category:       CategorySystemInfoMonitor,
				TimeTagFormat:  TimeTagNone,
				IsCommand:      false,
				HasTimeTag:     false,
				AllowedIn104:   true,
				InfoObjectSize: 1,
			},
		},
		{
			name: "C_IC_NA_1 — interrogation, system control",
			id:   C_IC_NA_1,
			want: TypeInfo{
				TypeID:         C_IC_NA_1,
				Name:           "C_IC_NA_1",
				Description:    "interrogation command",
				Direction:      DirectionControl,
				Category:       CategorySystemInfoControl,
				TimeTagFormat:  TimeTagNone,
				IsCommand:      true,
				HasTimeTag:     false,
				AllowedIn104:   true,
				InfoObjectSize: 1,
			},
		},
		{
			name: "C_TS_TA_1 — system command with CP56",
			id:   C_TS_TA_1,
			want: TypeInfo{
				TypeID:         C_TS_TA_1,
				Name:           "C_TS_TA_1",
				Description:    "test command with time tag CP56Time2a",
				Direction:      DirectionControl,
				Category:       CategorySystemInfoControl,
				TimeTagFormat:  TimeTagCP56Time2a,
				IsCommand:      true,
				HasTimeTag:     true,
				AllowedIn104:   true,
				InfoObjectSize: 9,
			},
		},
		{
			name: "P_ME_NA_1 — parameter, control direction, not a command",
			id:   P_ME_NA_1,
			want: TypeInfo{
				TypeID:         P_ME_NA_1,
				Name:           "P_ME_NA_1",
				Description:    "parameter of measured value, normalized value",
				Direction:      DirectionControl,
				Category:       CategoryParameterControl,
				TimeTagFormat:  TimeTagNone,
				IsCommand:      false,
				HasTimeTag:     false,
				AllowedIn104:   true,
				InfoObjectSize: 3,
			},
		},
		{
			name: "F_DR_TA_1 — directory, file transfer, embeds CP56",
			id:   F_DR_TA_1,
			want: TypeInfo{
				TypeID:         F_DR_TA_1,
				Name:           "F_DR_TA_1",
				Description:    "directory",
				Direction:      DirectionFile,
				Category:       CategoryFileTransfer,
				TimeTagFormat:  TimeTagCP56Time2a,
				IsCommand:      false,
				HasTimeTag:     true,
				AllowedIn104:   true,
				InfoObjectSize: 13,
			},
		},
		{
			name: "F_SG_NA_1 — segment, variable length",
			id:   F_SG_NA_1,
			want: TypeInfo{
				TypeID:         F_SG_NA_1,
				Name:           "F_SG_NA_1",
				Description:    "segment",
				Direction:      DirectionFile,
				Category:       CategoryFileTransfer,
				TimeTagFormat:  TimeTagNone,
				IsCommand:      false,
				HasTimeTag:     false,
				AllowedIn104:   true,
				InfoObjectSize: 0,
			},
		},
		{
			name: "undefined TypeID returns zero value",
			id:   TypeID(99),
			want: TypeInfo{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.Info()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TypeID(%d).Info() = %+v, want %+v", tt.id, got, tt.want)
			}
		})
	}
}

// TestTypeID_Info_Consistency guards against drift between the constant
// block, the Stringer, and the typeIDSpecs metadata table.
func TestTypeID_Info_Consistency(t *testing.T) {
	for i := 1; i < 256; i++ {
		id := TypeID(i)
		name := id.String()
		isDefined := name != strconv.Itoa(i)
		info := id.Info()

		if !isDefined {
			if info.Name != "" {
				t.Errorf("TypeID(%d) is undefined but Info().Name = %q", i, info.Name)
			}
			continue
		}
		if info.Name != name {
			t.Errorf("TypeID(%d): Info().Name = %q, String() = %q", i, info.Name, name)
		}
		if info.Description == "" {
			t.Errorf("TypeID(%d) %s: missing Description in typeIDSpecs", i, name)
		}
		if info.Category == CategoryUnknown {
			t.Errorf("TypeID(%d) %s: Category is Unknown", i, name)
		}
		if info.Direction == DirectionUnknown {
			t.Errorf("TypeID(%d) %s: Direction is Unknown", i, name)
		}
	}
}

func TestParseVariableStruct(t *testing.T) {
	type args struct {
		b byte
	}
	tests := []struct {
		name string
		args args
		want VariableStruct
	}{
		{"no sequence", args{0x0a}, VariableStruct{Number: 0x0a}},
		{"with sequence", args{0x8a}, VariableStruct{Number: 0x0a, IsSequence: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseVariableStruct(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseVariableStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariableStruct_Value(t *testing.T) {
	tests := []struct {
		name string
		this VariableStruct
		want byte
	}{
		{"no sequence", VariableStruct{Number: 0x0a}, 0x0a},
		{"with sequence", VariableStruct{Number: 0x0a, IsSequence: true}, 0x8a},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.Value(); got != tt.want {
				t.Errorf("VariableStruct.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariableStruct_String(t *testing.T) {
	tests := []struct {
		name string
		this VariableStruct
		want string
	}{
		{"no sequence", VariableStruct{Number: 100}, "100"},
		{"with sequence", VariableStruct{Number: 100, IsSequence: true}, "sq,100"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.String(); got != tt.want {
				t.Errorf("VariableStruct.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCauseOfTransmission(t *testing.T) {
	type args struct {
		b byte
	}
	tests := []struct {
		name string
		args args
		want CauseOfTransmission
	}{
		{"no test and neg", args{0x01}, CauseOfTransmission{Cause: Periodic}},
		{"with test", args{0x81}, CauseOfTransmission{Cause: Periodic, IsTest: true}},
		{"with neg", args{0x41}, CauseOfTransmission{Cause: Periodic, IsNegative: true}},
		{"with test and neg", args{0xc1}, CauseOfTransmission{Cause: Periodic, IsTest: true, IsNegative: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCauseOfTransmission(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseCauseOfTransmission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCauseOfTransmission_Value(t *testing.T) {
	tests := []struct {
		name string
		this CauseOfTransmission
		want byte
	}{
		{"no test and neg", CauseOfTransmission{Cause: Periodic}, 0x01},
		{"with test", CauseOfTransmission{Cause: Periodic, IsTest: true}, 0x81},
		{"with neg", CauseOfTransmission{Cause: Periodic, IsNegative: true}, 0x41},
		{"with test and neg", CauseOfTransmission{Cause: Periodic, IsTest: true, IsNegative: true}, 0xc1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.Value(); got != tt.want {
				t.Errorf("CauseOfTransmission.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCauseOfTransmission_String(t *testing.T) {
	tests := []struct {
		name string
		this CauseOfTransmission
		want string
	}{
		{"no test and neg", CauseOfTransmission{Cause: Periodic}, "Periodic"},
		{"with test", CauseOfTransmission{Cause: Periodic, IsTest: true}, "Periodic,test"},
		{"with neg", CauseOfTransmission{Cause: Periodic, IsNegative: true}, "Periodic,neg"},
		{"with test and neg", CauseOfTransmission{Cause: Periodic, IsTest: true, IsNegative: true}, "Periodic,neg,test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.this.String(); got != tt.want {
				t.Errorf("CauseOfTransmission.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

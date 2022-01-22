package cmd

import "testing"

func Test_paramSource_String(t *testing.T) {
	tests := []struct {
		name string
		s    paramSource
	}{
		{name: "S01", s: CliOpt},
		{name: "S02", s: SharedCredentials},
		{name: "S03", s: SharedCredentialsBeforeMFAProfile},
		{name: "S04", s: SharedCredentialsAfterMFAProfile},
		{name: "S05", s: SharedConfig},
		{name: "S06", s: SharedConfigBeforeMFAProfile},
		{name: "S07", s: SharedConfigAfterMFAProfile},
		{name: "S08", s: AwsmfaConfig},
		{name: "S09", s: AwsmfaBuildIn},
		{name: "S10", s: EnvAWSDefaultRegion},
		{name: "S11", s: EnvAWSRegion},
		{name: "S12", s: EnvAWSProfile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.String(); got == "unknown source" {
				t.Errorf("paramSource.String() = %v", got)
			}
		})
	}
}

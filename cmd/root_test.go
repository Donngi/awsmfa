package cmd

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"gopkg.in/ini.v1"
)

func Test_isExpired(t *testing.T) {
	type args struct {
		tokenDue   time.Time
		comparison time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "S01: active", args: args{tokenDue: time.Date(2022, 11, 23, 14, 15, 16, 10, time.UTC), comparison: time.Date(2021, 11, 23, 14, 15, 16, 10, time.UTC)}, want: false},
		{name: "S01: expired", args: args{tokenDue: time.Date(2021, 11, 23, 14, 15, 16, 10, time.UTC), comparison: time.Date(2022, 11, 23, 14, 15, 16, 10, time.UTC)}, want: true},
		{name: "S01: active - border", args: args{tokenDue: time.Date(2021, 11, 23, 14, 15, 16, 10, time.UTC), comparison: time.Date(2021, 11, 23, 14, 15, 16, 10, time.UTC)}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExpired(tt.args.tokenDue, tt.args.comparison); got != tt.want {
				t.Errorf("isExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasActiveToken(t *testing.T) {
	type args struct {
		profile string
		cred    *ini.File
	}
	tests := []struct {
		name               string
		args               args
		testDataPath       string
		wantHasActiveToken bool
	}{
		{name: "S01: active", args: args{profile: "active"}, testDataPath: "testData/hasActiveToken_credentials", wantHasActiveToken: true},
		{name: "S02: expired", args: args{profile: "expired"}, testDataPath: "testData/hasActiveToken_credentials", wantHasActiveToken: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := ini.Load(tt.testDataPath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.testDataPath)
			}

			gotHasActiveToken, _ := hasActiveToken(tt.args.profile, cred)
			if gotHasActiveToken != tt.wantHasActiveToken {
				t.Errorf("hasActiveToken() gotHasActiveToken = %v, want %v", gotHasActiveToken, tt.wantHasActiveToken)
			}
			// if !reflect.DeepEqual(gotDue, tt.wantDue) {
			// 	t.Errorf("hasActiveToken() gotDue = %v, want %v", gotDue, tt.wantDue)
			// }
		})
	}
}

func Test_setMode(t *testing.T) {
	type args struct {
		cliOpt       string
		defaultValue string
		profile      string
	}
	tests := []struct {
		name         string
		args         args
		credFilePath string
		cfgFilePath  string
		want         string
		wantErr      bool
	}{
		{name: "S01", args: args{cliOpt: "get-session-token", defaultValue: "assume-role", profile: "credhas-confighas"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", want: "get-session-token", wantErr: false},
		{name: "S02", args: args{cliOpt: "assume-role", defaultValue: "get-session-token", profile: "crednil-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", want: "assume-role", wantErr: false},
		{name: "S03", args: args{cliOpt: "", defaultValue: "get-session-token", profile: "credhas-confighas"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", want: "assume-role", wantErr: false},
		{name: "S04", args: args{cliOpt: "", defaultValue: "get-session-token", profile: "credhas-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", want: "assume-role", wantErr: false},
		{name: "S05", args: args{cliOpt: "", defaultValue: "get-session-token", profile: "crednil-confighas"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", want: "assume-role", wantErr: false},
		{name: "S05", args: args{cliOpt: "", defaultValue: "get-session-token", profile: "crednil-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", want: "get-session-token", wantErr: false},
		{name: "S06", args: args{cliOpt: "", defaultValue: "assume-role", profile: "crednil-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", want: "assume-role", wantErr: false},
		{name: "F01", args: args{cliOpt: "wrong-mode💀", defaultValue: "get-session-token", profile: "crednil-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", want: "wrong-mode", wantErr: true},
		{name: "F02", args: args{cliOpt: "", defaultValue: "wrong-mode💀", profile: "crednil-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", want: "wrong-mode", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := ini.Load(tt.credFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.credFilePath)
			}

			cfg, err := ini.Load(tt.cfgFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.cfgFilePath)
			}

			got, err := setMode(tt.args.cliOpt, tt.args.defaultValue, tt.args.profile, cred, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("setMode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("setMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setProfile(t *testing.T) {
	type args struct {
		cliOpt       string
		defaultValue string
	}
	tests := []struct {
		name      string
		args      args
		existsEnv bool
		want      string
	}{
		{name: "S01", args: args{cliOpt: "cliOpt", defaultValue: "default"}, existsEnv: false, want: "cliOpt"},
		{name: "S02", args: args{cliOpt: "cliOpt", defaultValue: "default"}, existsEnv: true, want: "cliOpt"},
		{name: "S03", args: args{cliOpt: "", defaultValue: "default"}, existsEnv: true, want: "env"},
		{name: "S04", args: args{cliOpt: "", defaultValue: "default"}, existsEnv: false, want: "default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer os.Unsetenv("AWS_PROFILE")
			if tt.existsEnv {
				os.Setenv("AWS_PROFILE", "env")
			}
			if got := setProfile(tt.args.cliOpt, tt.args.defaultValue); got != tt.want {
				t.Errorf("setProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setDurationSeconds(t *testing.T) {
	type args struct {
		cliOpt       int32
		defaultValue int32
		profile      string
	}
	tests := []struct {
		name         string
		args         args
		credFilePath string
		cfgFilePath  string
		want         int32
	}{
		{name: "S01 - Adapt cliOpt", args: args{cliOpt: 40000, defaultValue: 10000, profile: "cred30000-config20000"}, credFilePath: "testData/setDurationSeconds_credentials", cfgFilePath: "testData/setDurationSeconds_config", want: 40000},
		{name: "S02 - Adapt cred", args: args{cliOpt: 0, defaultValue: 10000, profile: "cred30000-config20000"}, credFilePath: "testData/setDurationSeconds_credentials", cfgFilePath: "testData/setDurationSeconds_config", want: 30000},
		{name: "S03 - Adapt cfg", args: args{cliOpt: 0, defaultValue: 10000, profile: "crednil-config20000"}, credFilePath: "testData/setDurationSeconds_credentials", cfgFilePath: "testData/setDurationSeconds_config", want: 20000},
		{name: "S04 - Adapt default", args: args{cliOpt: 0, defaultValue: 10000, profile: "crednil-confignil"}, credFilePath: "testData/setDurationSeconds_credentials", cfgFilePath: "testData/setDurationSeconds_config", want: 10000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := ini.Load(tt.credFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.credFilePath)
			}

			cfg, err := ini.Load(tt.cfgFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.cfgFilePath)
			}

			if got := setDurationSeconds(tt.args.cliOpt, tt.args.defaultValue, tt.args.profile, cred, cfg); got != tt.want {
				t.Errorf("setDurationSeconds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setMFASerial(t *testing.T) {
	type args struct {
		cliOpt       string
		defaultValue string
		profile      string
	}
	tests := []struct {
		name         string
		args         args
		credFilePath string
		cfgFilePath  string
		want         string
		wantErr      bool
	}{
		{name: "S01", args: args{cliOpt: "cli-serial", defaultValue: "default-serial", profile: "credhas-confighas"}, credFilePath: "testData/setMFASerial_credentials", cfgFilePath: "testData/setMFASerial_config", want: "cli-serial", wantErr: false},
		{name: "S02", args: args{cliOpt: "", defaultValue: "default-serial", profile: "credhas-confighas"}, credFilePath: "testData/setMFASerial_credentials", cfgFilePath: "testData/setMFASerial_config", want: "cred-serial", wantErr: false},
		{name: "S03", args: args{cliOpt: "", defaultValue: "default-serial", profile: "crednil-confighas"}, credFilePath: "testData/setMFASerial_credentials", cfgFilePath: "testData/setMFASerial_config", want: "config-serial", wantErr: false},
		{name: "S04", args: args{cliOpt: "", defaultValue: "default-serial", profile: "crednil-confignil"}, credFilePath: "testData/setMFASerial_credentials", cfgFilePath: "testData/setMFASerial_config", want: "default-serial", wantErr: false},
		{name: "F01", args: args{cliOpt: "", defaultValue: "unspecified", profile: "crednil-confignil"}, credFilePath: "testData/setMFASerial_credentials", cfgFilePath: "testData/setMFASerial_config", want: "error", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := ini.Load(tt.credFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.credFilePath)
			}

			cfg, err := ini.Load(tt.cfgFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.cfgFilePath)
			}

			got, err := setMFASerial(tt.args.cliOpt, tt.args.defaultValue, tt.args.profile, cred, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("setMFASerial() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("setMFASerial() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setRoleArn(t *testing.T) {
	type args struct {
		cliRoleArn string
		profile    string
	}
	tests := []struct {
		name         string
		args         args
		credFilePath string
		cfgFilePath  string
		want         string
		wantErr      bool
	}{
		{name: "S01", args: args{cliRoleArn: "cli-role-arn", profile: "credhas-confighas"}, credFilePath: "testData/setRoleArn_credentials", cfgFilePath: "testData/setRoleArn_config", want: "cli-role-arn", wantErr: false},
		{name: "S02", args: args{cliRoleArn: "", profile: "credhas-confighas"}, credFilePath: "testData/setRoleArn_credentials", cfgFilePath: "testData/setRoleArn_config", want: "cred-role-arn", wantErr: false},
		{name: "S03", args: args{cliRoleArn: "", profile: "crednil-confighas"}, credFilePath: "testData/setRoleArn_credentials", cfgFilePath: "testData/setRoleArn_config", want: "config-role-arn", wantErr: false},
		{name: "F01", args: args{cliRoleArn: "", profile: "crednil-confignil"}, credFilePath: "testData/setRoleArn_credentials", cfgFilePath: "testData/setRoleArn_config", want: "error", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := ini.Load(tt.credFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.credFilePath)
			}

			cfg, err := ini.Load(tt.cfgFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.cfgFilePath)
			}

			got, err := setRoleArn(tt.args.cliRoleArn, tt.args.profile, cred, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("setRoleArn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("setRoleArn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setRoleSessionName(t *testing.T) {
	type args struct {
		cliOpt       string
		defaultValue string
		profile      string
	}
	tests := []struct {
		name         string
		args         args
		credFilePath string
		cfgFilePath  string
		want         string
	}{
		{name: "S01", args: args{cliOpt: "cliOpt", defaultValue: "defaultValue", profile: "credhas-confighas"}, credFilePath: "testData/setRoleSessionName_credentials", cfgFilePath: "testData/setRoleSessionName_config", want: "cliOpt"},
		{name: "S02", args: args{cliOpt: "", defaultValue: "defaultValue", profile: "credhas-confighas"}, credFilePath: "testData/setRoleSessionName_credentials", cfgFilePath: "testData/setRoleSessionName_config", want: "cred-session-name"},
		{name: "S03", args: args{cliOpt: "", defaultValue: "defaultValue", profile: "crednil-confighas"}, credFilePath: "testData/setRoleSessionName_credentials", cfgFilePath: "testData/setRoleSessionName_config", want: "config-session-name"},
		{name: "S04", args: args{cliOpt: "", defaultValue: "defaultValue", profile: "crednil-confignil"}, credFilePath: "testData/setRoleSessionName_credentials", cfgFilePath: "testData/setRoleSessionName_config", want: "defaultValue"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := ini.Load(tt.credFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.credFilePath)
			}

			cfg, err := ini.Load(tt.cfgFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.cfgFilePath)
			}

			if got := setRoleSessionName(tt.args.cliOpt, tt.args.defaultValue, tt.args.profile, cred, cfg); got != tt.want {
				t.Errorf("setRoleSessionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setEndpointRegion(t *testing.T) {
	type args struct {
		cliOpt       string
		defaultValue string
		profile      string
	}
	tests := []struct {
		name                   string
		args                   args
		existsEnvREGION        bool
		existsEnvDEFAULTREGION bool
		credFilePath           string
		cfgFilePath            string
		want                   string
	}{
		{name: "S01", args: args{cliOpt: "cliOpt", defaultValue: "defaultValue", profile: "before-credhas-confighas_after-credhas-confighas"}, existsEnvREGION: true, existsEnvDEFAULTREGION: true, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", want: "cliOpt"},
		{name: "S02", args: args{cliOpt: "", defaultValue: "defaultValue", profile: "before-credhas-confighas_after-credhas-confighas"}, existsEnvREGION: true, existsEnvDEFAULTREGION: true, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", want: "env-region"},
		{name: "S03", args: args{cliOpt: "", defaultValue: "defaultValue", profile: "before-credhas-confighas_after-credhas-confighas"}, existsEnvREGION: false, existsEnvDEFAULTREGION: true, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", want: "env-default-region"},
		{name: "S04", args: args{cliOpt: "", defaultValue: "defaultValue", profile: "before-credhas-confighas_after-credhas-confighas"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", want: "before-cred"},
		{name: "S05", args: args{cliOpt: "", defaultValue: "defaultValue", profile: "before-crednil-confighas_after-credhas-confighas"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", want: "before-config"},
		{name: "S06", args: args{cliOpt: "", defaultValue: "defaultValue", profile: "before-crednil-confignil_after-credhas-confighas"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", want: "after-cred"},
		{name: "S07", args: args{cliOpt: "", defaultValue: "defaultValue", profile: "before-crednil-confignil_after-crednil-confighas"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", want: "after-config"},
		{name: "S08", args: args{cliOpt: "", defaultValue: "defaultValue", profile: "before-crednil-confignil_after-crednil-confignil"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", want: "defaultValue"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer os.Unsetenv("AWS_REGION")
			if tt.existsEnvREGION {
				os.Setenv("AWS_REGION", "env-region")
			}

			defer os.Unsetenv("AWS_DEFAULT_REGION")
			if tt.existsEnvDEFAULTREGION {
				os.Setenv("AWS_DEFAULT_REGION", "env-default-region")
			}

			cred, err := ini.Load(tt.credFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.credFilePath)
			}

			cfg, err := ini.Load(tt.cfgFilePath)
			if err != nil {
				t.Errorf("failed to load test data: %v", tt.cfgFilePath)
			}

			if got := setEndpointRegion(tt.args.cliOpt, tt.args.defaultValue, tt.args.profile, cred, cfg); got != tt.want {
				t.Errorf("setEndpointRegion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_saveTemporaryTokenFromGetSessionToken(t *testing.T) {
	type args struct {
		token               *sts.GetSessionTokenOutput
		profile             string
		credentialsFilePath string
	}

	accessKeyID := "NEWACCESSKEYID1111"
	secretAccessKey := "NEWSECRETACCESSKEY1111"
	sessionToken := "NEWSESSIONTOKEN1111"
	expiration := time.Date(2999, 11, 23, 14, 15, 16, 0, time.UTC)
	testToken := sts.GetSessionTokenOutput{
		Credentials: &types.Credentials{
			AccessKeyId:     &accessKeyID,
			Expiration:      &expiration,
			SecretAccessKey: &secretAccessKey,
			SessionToken:    &sessionToken,
		},
	}

	// Success cases
	func() {
		tests := []struct {
			name          string
			args          args
			wantFilePath  string
			fileToRestore string
			wantErr       bool
		}{
			{name: "S01", args: args{token: &testToken, profile: "existing", credentialsFilePath: "testdata/saveTemporaryTokenFromGetSessionToken_credentials"}, wantFilePath: "testdata/saveTemporaryTokenFromGetSessionToken_credentials_after_test_existing", fileToRestore: "testdata/saveTemporaryTokenFromGetSessionToken_credentials", wantErr: false},
			{name: "S02", args: args{token: &testToken, profile: "new", credentialsFilePath: "testdata/saveTemporaryTokenFromGetSessionToken_credentials"}, wantFilePath: "testdata/saveTemporaryTokenFromGetSessionToken_credentials_after_test_new", fileToRestore: "testdata/saveTemporaryTokenFromGetSessionToken_credentials", wantErr: false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				backup, err := ini.Load("testdata/saveTemporaryTokenFromGetSessionToken_credentials_before_test")
				if err != nil {
					t.Errorf("failed to load backup data: %v", tt.wantFilePath)
				}
				defer backup.SaveTo(tt.fileToRestore)

				if err := saveTemporaryTokenFromGetSessionToken(tt.args.token, tt.args.profile, tt.args.credentialsFilePath); (err != nil) != tt.wantErr {
					t.Errorf("saveTemporaryTokenFromGetSessionToken() error = %v, wantErr %v", err, tt.wantErr)
				}

				want, err := ioutil.ReadFile(tt.wantFilePath)
				if err != nil {
					t.Errorf("failed to load want data: %v", tt.wantFilePath)
				}

				got, err := ioutil.ReadFile(tt.args.credentialsFilePath)
				if err != nil {
					t.Errorf("failed to load got data: %v", tt.args.credentialsFilePath)
				}

				if string(got) != string(want) {
					t.Errorf("saveTemporaryTokenFromGetSessionToken() got = %+v, want %+v", string(got), string(want))
				}
			})
		}
	}()

	// Failure cases (failed to load credentials file)
	func() {
		tests := []struct {
			name                    string
			args                    args
			wantFilePath            string
			realCredentialsFilePath string

			wantErr bool
		}{
			{name: "F01", args: args{token: &testToken, profile: "existing", credentialsFilePath: "Should be error💀"}, wantFilePath: "testdata/saveTemporaryTokenFromGetSessionToken_credentials_before_test", realCredentialsFilePath: "testdata/saveTemporaryTokenFromGetSessionToken_credentials", wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				backup, err := ini.Load("testdata/saveTemporaryTokenFromGetSessionToken_credentials_before_test")
				if err != nil {
					t.Errorf("failed to load backup data: %v", tt.wantFilePath)
				}
				defer backup.SaveTo(tt.realCredentialsFilePath)

				if err := saveTemporaryTokenFromGetSessionToken(tt.args.token, tt.args.profile, tt.args.credentialsFilePath); (err != nil) != tt.wantErr {
					t.Errorf("saveTemporaryTokenFromGetSessionToken() error = %v, wantErr %v", err, tt.wantErr)
				}

				want, err := ioutil.ReadFile(tt.wantFilePath)
				if err != nil {
					t.Errorf("failed to load want data: %v", tt.wantFilePath)
				}

				got, err := ioutil.ReadFile(tt.realCredentialsFilePath)
				if err != nil {
					t.Errorf("failed to load got data: %v", tt.realCredentialsFilePath)
				}

				if string(got) != string(want) {
					t.Errorf("saveTemporaryTokenFromGetSessionToken() got = %+v, want %+v", string(got), string(want))
				}
			})
		}
	}()
}

func Test_saveTemporaryTokenFromAssumeRole(t *testing.T) {
	type args struct {
		token               *sts.AssumeRoleOutput
		profile             string
		credentialsFilePath string
	}

	accessKeyID := "NEWACCESSKEYID1111"
	secretAccessKey := "NEWSECRETACCESSKEY1111"
	sessionToken := "NEWSESSIONTOKEN1111"
	expiration := time.Date(2999, 11, 23, 14, 15, 16, 0, time.UTC)
	testToken := sts.AssumeRoleOutput{
		Credentials: &types.Credentials{
			AccessKeyId:     &accessKeyID,
			Expiration:      &expiration,
			SecretAccessKey: &secretAccessKey,
			SessionToken:    &sessionToken,
		},
	}

	// Success cases
	func() {
		tests := []struct {
			name          string
			args          args
			wantFilePath  string
			fileToRestore string
			wantErr       bool
		}{
			{name: "S01", args: args{token: &testToken, profile: "existing", credentialsFilePath: "testdata/saveTemporaryTokenFromAssumeRole_credentials"}, wantFilePath: "testdata/saveTemporaryTokenFromAssumeRole_credentials_after_test_existing", fileToRestore: "testdata/saveTemporaryTokenFromAssumeRole_credentials", wantErr: false},
			{name: "S02", args: args{token: &testToken, profile: "new", credentialsFilePath: "testdata/saveTemporaryTokenFromAssumeRole_credentials"}, wantFilePath: "testdata/saveTemporaryTokenFromAssumeRole_credentials_after_test_new", fileToRestore: "testdata/saveTemporaryTokenFromAssumeRole_credentials", wantErr: false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				backup, err := ini.Load("testdata/saveTemporaryTokenFromAssumeRole_credentials_before_test")
				if err != nil {
					t.Errorf("failed to load backup data: %v", tt.wantFilePath)
				}
				defer backup.SaveTo(tt.fileToRestore)

				if err := saveTemporaryTokenFromAssumeRole(tt.args.token, tt.args.profile, tt.args.credentialsFilePath); (err != nil) != tt.wantErr {
					t.Errorf("saveTemporaryTokenFromAssumeRole() error = %v, wantErr %v", err, tt.wantErr)
				}

				want, err := ioutil.ReadFile(tt.wantFilePath)
				if err != nil {
					t.Errorf("failed to load want data: %v", tt.wantFilePath)
				}

				got, err := ioutil.ReadFile(tt.args.credentialsFilePath)
				if err != nil {
					t.Errorf("failed to load got data: %v", tt.args.credentialsFilePath)
				}

				if string(got) != string(want) {
					t.Errorf("saveTemporaryTokenFromAssumeRole() got = %+v, want %+v", string(got), string(want))
				}
			})
		}
	}()

	// Failure cases (failed to load credentials file)
	func() {
		tests := []struct {
			name                    string
			args                    args
			wantFilePath            string
			realCredentialsFilePath string

			wantErr bool
		}{
			{name: "F01", args: args{token: &testToken, profile: "existing", credentialsFilePath: "Should be error💀"}, wantFilePath: "testdata/saveTemporaryTokenFromAssumeRole_credentials_before_test", realCredentialsFilePath: "testdata/saveTemporaryTokenFromAssumeRole_credentials", wantErr: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				backup, err := ini.Load("testdata/saveTemporaryTokenFromAssumeRole_credentials_before_test")
				if err != nil {
					t.Errorf("failed to load backup data: %v", tt.wantFilePath)
				}
				defer backup.SaveTo(tt.realCredentialsFilePath)

				if err := saveTemporaryTokenFromAssumeRole(tt.args.token, tt.args.profile, tt.args.credentialsFilePath); (err != nil) != tt.wantErr {
					t.Errorf("saveTemporaryTokenFromAssumeRole() error = %v, wantErr %v", err, tt.wantErr)
				}

				want, err := ioutil.ReadFile(tt.wantFilePath)
				if err != nil {
					t.Errorf("failed to load want data: %v", tt.wantFilePath)
				}

				got, err := ioutil.ReadFile(tt.realCredentialsFilePath)
				if err != nil {
					t.Errorf("failed to load got data: %v", tt.realCredentialsFilePath)
				}

				if string(got) != string(want) {
					t.Errorf("saveTemporaryTokenFromAssumeRole() got = %+v, want %+v", string(got), string(want))
				}
			})
		}
	}()
}

func Test_initDefault(t *testing.T) {
	type settableValue struct {
		credentialsFilePath                   string
		configFilePath                        string
		beforeMFASuffix                       string
		defaultMode                           string
		defaultProfile                        string
		defaultMFASerial                      string
		defaultEndpointRegion                 string
		defaultDurationSecondsGetSessionToken int32
		defaultDurationSecondsAssumeRole      int32
	}
	initial := settableValue{
		credentialsFilePath:                   credentialsFilePath,
		configFilePath:                        configFilePath,
		beforeMFASuffix:                       beforeMFASuffix,
		defaultMode:                           defaultMode,
		defaultProfile:                        defaultProfile,
		defaultMFASerial:                      defaultMFASerial,
		defaultEndpointRegion:                 defaultEndpointRegion,
		defaultDurationSecondsGetSessionToken: defaultDurationSecondsGetSessionToken,
		defaultDurationSecondsAssumeRole:      defaultDurationSecondsAssumeRole,
	}
	applied := settableValue{
		credentialsFilePath:                   "testhome/configuration_credentials_file_path",
		configFilePath:                        "testhome/configuration_config_file_path",
		beforeMFASuffix:                       "configuration_suffix_of_before_mfa_profile",
		defaultMode:                           "configuration_mode",
		defaultProfile:                        "configuration_profile",
		defaultMFASerial:                      "configuration_mfa_serial",
		defaultEndpointRegion:                 "configuration_endpoint_region",
		defaultDurationSecondsGetSessionToken: 12345,
		defaultDurationSecondsAssumeRole:      67890,
	}

	type args struct {
		configurationFilePath string
	}
	tests := []struct {
		name string
		args args
		want settableValue
	}{
		{name: "S01", args: args{configurationFilePath: "testdata/initDefault_configuration"}, want: applied},
		{name: "S02", args: args{configurationFilePath: "unspecified"}, want: initial},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				credentialsFilePath = initial.credentialsFilePath
				configFilePath = initial.configFilePath
				beforeMFASuffix = initial.beforeMFASuffix
				defaultMode = initial.defaultMode
				defaultProfile = initial.defaultProfile
				defaultMFASerial = initial.defaultMFASerial
				defaultEndpointRegion = initial.defaultEndpointRegion
				defaultDurationSecondsGetSessionToken = initial.defaultDurationSecondsGetSessionToken
				defaultDurationSecondsAssumeRole = initial.defaultDurationSecondsAssumeRole
				os.Unsetenv("TESTHOME")
			}()

			os.Setenv("TESTHOME", "testhome")

			initDefault(tt.args.configurationFilePath)

			got := settableValue{
				credentialsFilePath:                   credentialsFilePath,
				configFilePath:                        configFilePath,
				beforeMFASuffix:                       beforeMFASuffix,
				defaultMode:                           defaultMode,
				defaultProfile:                        defaultProfile,
				defaultMFASerial:                      defaultMFASerial,
				defaultEndpointRegion:                 defaultEndpointRegion,
				defaultDurationSecondsGetSessionToken: defaultDurationSecondsGetSessionToken,
				defaultDurationSecondsAssumeRole:      defaultDurationSecondsAssumeRole,
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initDefault() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_secToHMS(t *testing.T) {
	type args struct {
		seconds int32
	}
	tests := []struct {
		name     string
		args     args
		wantHour int32
		wantMin  int32
		wantSec  int32
	}{
		{name: "S01", args: args{seconds: 40000}, wantHour: 11, wantMin: 6, wantSec: 40},
		{name: "S02", args: args{seconds: 3600}, wantHour: 1, wantMin: 0, wantSec: 0},
		{name: "S03", args: args{seconds: 60}, wantHour: 0, wantMin: 1, wantSec: 0},
		{name: "S04", args: args{seconds: 1}, wantHour: 0, wantMin: 0, wantSec: 1},
		{name: "S05", args: args{seconds: 0}, wantHour: 0, wantMin: 0, wantSec: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHour, gotMin, gotSec := secToHMS(tt.args.seconds)
			if gotHour != tt.wantHour {
				t.Errorf("secToHMS() gotHour = %v, want %v", gotHour, tt.wantHour)
			}
			if gotMin != tt.wantMin {
				t.Errorf("secToHMS() gotMin = %v, want %v", gotMin, tt.wantMin)
			}
			if gotSec != tt.wantSec {
				t.Errorf("secToHMS() gotSec = %v, want %v", gotSec, tt.wantSec)
			}
		})
	}
}

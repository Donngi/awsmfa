package cmd

import (
	"os"
	"testing"

	"gopkg.in/ini.v1"
)

func Test_setMode(t *testing.T) {
	type args struct {
		cliOpt       string
		defaultValue string
		profile      string
	}
	tests := []struct {
		name              string
		args              args
		credFilePath      string
		cfgFilePath       string
		awsmfaCfgFilePath string
		wantMode          string
		wantSource        string
		wantErr           bool
	}{
		{name: "S01", args: args{cliOpt: "get-session-token", defaultValue: "assume-role", profile: "credhas-confighas"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", awsmfaCfgFilePath: "testData/setMode_awsmfaConfiguration_has", wantMode: "get-session-token", wantSource: CliOpt.String(), wantErr: false},
		{name: "S02", args: args{cliOpt: "assume-role", defaultValue: "get-session-token", profile: "crednil-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", awsmfaCfgFilePath: "testData/setMode_awsmfaConfiguration_has", wantMode: "assume-role", wantSource: CliOpt.String(), wantErr: false},
		{name: "S03", args: args{cliOpt: "", defaultValue: "get-session-token", profile: "credhas-confighas"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", awsmfaCfgFilePath: "testData/setMode_awsmfaConfiguration_has", wantMode: "assume-role", wantSource: SharedCredentials.String(), wantErr: false},
		{name: "S04", args: args{cliOpt: "", defaultValue: "get-session-token", profile: "credhas-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", awsmfaCfgFilePath: "testData/setMode_awsmfaConfiguration_has", wantMode: "assume-role", wantSource: SharedCredentials.String(), wantErr: false},
		{name: "S05", args: args{cliOpt: "", defaultValue: "get-session-token", profile: "crednil-confighas"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", awsmfaCfgFilePath: "testData/setMode_awsmfaConfiguration_has", wantMode: "assume-role", wantSource: SharedConfig.String(), wantErr: false},
		{name: "S06", args: args{cliOpt: "", defaultValue: "get-session-token", profile: "crednil-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", awsmfaCfgFilePath: "testData/setMode_awsmfaConfiguration_has", wantMode: "assume-role", wantSource: AwsmfaConfig.String(), wantErr: false},
		{name: "S07", args: args{cliOpt: "", defaultValue: "get-session-token", profile: "crednil-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", awsmfaCfgFilePath: "testData/setMode_awsmfaConfiguration_nil", wantMode: "get-session-token", wantSource: AwsmfaBuildIn.String(), wantErr: false},
		{name: "S08", args: args{cliOpt: "", defaultValue: "get-session-token", profile: "crednil-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", awsmfaCfgFilePath: "nil", wantMode: "get-session-token", wantSource: AwsmfaBuildIn.String(), wantErr: false},
		{name: "F01", args: args{cliOpt: "wrong-mode💀", defaultValue: "get-session-token", profile: "crednil-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", awsmfaCfgFilePath: "testData/setMode_awsmfaConfiguration_has", wantMode: "ERROR", wantSource: "ERROR", wantErr: true},
		{name: "F02", args: args{cliOpt: "", defaultValue: "wrong-mode💀", profile: "crednil-confignil"}, credFilePath: "testData/setMode_credentials", cfgFilePath: "testData/setMode_config", awsmfaCfgFilePath: "testData/setMode_awsmfaConfiguration_nil", wantMode: "ERROR", wantSource: "ERROR", wantErr: true},
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

			awsmfaCfg, _ := ini.Load(tt.awsmfaCfgFilePath)

			gotMode, gotSource, err := setMode(tt.args.cliOpt, tt.args.defaultValue, tt.args.profile, cred, cfg, awsmfaCfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("setMode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotMode != tt.wantMode {
				t.Errorf("setMode() mode = %v, want %v", gotMode, tt.wantMode)
			}
			if gotSource != tt.wantSource {
				t.Errorf("setSource() source = %v, want %v", gotSource, tt.wantSource)
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
		name              string
		args              args
		awsmfaCfgFilePath string
		existsEnv         bool
		wantProfile       string
		wantSource        string
	}{
		{name: "S01", args: args{cliOpt: "cliOpt", defaultValue: "default"}, awsmfaCfgFilePath: "testData/setProfile_awsmfaConfiguration_has", existsEnv: true, wantProfile: "cliOpt", wantSource: CliOpt.String()},
		{name: "S02", args: args{cliOpt: "cliOpt", defaultValue: "default"}, awsmfaCfgFilePath: "testData/setProfile_awsmfaConfiguration_has", existsEnv: true, wantProfile: "cliOpt", wantSource: CliOpt.String()},
		{name: "S03", args: args{cliOpt: "", defaultValue: "default"}, awsmfaCfgFilePath: "testData/setProfile_awsmfaConfiguration_has", existsEnv: true, wantProfile: "env", wantSource: EnvAWSProfile.String()},
		{name: "S04", args: args{cliOpt: "", defaultValue: "default"}, awsmfaCfgFilePath: "testData/setProfile_awsmfaConfiguration_has", existsEnv: false, wantProfile: "awsmfaCfg", wantSource: AwsmfaConfig.String()},
		{name: "S05", args: args{cliOpt: "", defaultValue: "default"}, awsmfaCfgFilePath: "testData/setProfile_awsmfaConfiguration_nil", existsEnv: false, wantProfile: "default", wantSource: AwsmfaBuildIn.String()},
		{name: "S06", args: args{cliOpt: "", defaultValue: "default"}, awsmfaCfgFilePath: "nil", existsEnv: false, wantProfile: "default", wantSource: AwsmfaBuildIn.String()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer os.Unsetenv("AWS_PROFILE")
			if tt.existsEnv {
				os.Setenv("AWS_PROFILE", "env")
			}
			awsmfaCfg, _ := ini.Load(tt.awsmfaCfgFilePath)

			gotProfile, gotSource := setProfile(tt.args.cliOpt, tt.args.defaultValue, awsmfaCfg)
			if gotProfile != tt.wantProfile {
				t.Errorf("setProfile() = %v, want %v", gotProfile, tt.wantSource)
			}
			if gotSource != tt.wantSource {
				t.Errorf("setProfile() = %v, want %v", gotSource, tt.wantSource)
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
		name              string
		args              args
		credFilePath      string
		cfgFilePath       string
		awsmfaCfgFilePath string
		wantDuration      int32
		wantSource        string
	}{
		{name: "S01", args: args{cliOpt: 40000, defaultValue: 5000, profile: "cred30000-config20000"}, awsmfaCfgFilePath: "testData/setDurationSeconds_awsmfaConfiguration_has", credFilePath: "testData/setDurationSeconds_credentials", cfgFilePath: "testData/setDurationSeconds_config", wantDuration: 40000, wantSource: CliOpt.String()},
		{name: "S02", args: args{cliOpt: 0, defaultValue: 5000, profile: "cred30000-config20000"}, awsmfaCfgFilePath: "testData/setDurationSeconds_awsmfaConfiguration_has", credFilePath: "testData/setDurationSeconds_credentials", cfgFilePath: "testData/setDurationSeconds_config", wantDuration: 30000, wantSource: SharedCredentials.String()},
		{name: "S03", args: args{cliOpt: 0, defaultValue: 5000, profile: "crednil-config20000"}, awsmfaCfgFilePath: "testData/setDurationSeconds_awsmfaConfiguration_has", credFilePath: "testData/setDurationSeconds_credentials", cfgFilePath: "testData/setDurationSeconds_config", wantDuration: 20000, wantSource: SharedConfig.String()},
		{name: "S04", args: args{cliOpt: 0, defaultValue: 5000, profile: "crednil-confignil"}, awsmfaCfgFilePath: "testData/setDurationSeconds_awsmfaConfiguration_has", credFilePath: "testData/setDurationSeconds_credentials", cfgFilePath: "testData/setDurationSeconds_config", wantDuration: 10000, wantSource: AwsmfaConfig.String()},
		{name: "S05", args: args{cliOpt: 0, defaultValue: 5000, profile: "crednil-confignil"}, awsmfaCfgFilePath: "testData/setDurationSeconds_awsmfaConfiguration_nil", credFilePath: "testData/setDurationSeconds_credentials", cfgFilePath: "testData/setDurationSeconds_config", wantDuration: 5000, wantSource: AwsmfaBuildIn.String()},
		{name: "S06", args: args{cliOpt: 0, defaultValue: 5000, profile: "crednil-confignil"}, awsmfaCfgFilePath: "nil", credFilePath: "testData/setDurationSeconds_credentials", cfgFilePath: "testData/setDurationSeconds_config", wantDuration: 5000, wantSource: AwsmfaBuildIn.String()},
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

			awsmfaCfg, _ := ini.Load(tt.awsmfaCfgFilePath)

			gotDuration, gotSource := setDurationSeconds(tt.args.cliOpt, tt.args.defaultValue, tt.args.profile, cred, cfg, awsmfaCfg)
			if gotDuration != tt.wantDuration {
				t.Errorf("setDurationSeconds() = %v, want %v", gotDuration, tt.wantDuration)
			}
			if gotSource != tt.wantSource {
				t.Errorf("setDurationSeconds() = %v, want %v", gotSource, tt.wantSource)
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
		name              string
		args              args
		credFilePath      string
		cfgFilePath       string
		awsmfaCfgFilePath string
		wantSerial        string
		wantSource        string
		wantErr           bool
	}{
		{name: "S01", args: args{cliOpt: "cli-serial", defaultValue: "default-serial", profile: "credhas-confighas"}, credFilePath: "testData/setMFASerial_credentials", cfgFilePath: "testData/setMFASerial_config", awsmfaCfgFilePath: "testData/setMFASerial_awsmfaConfiguration_has", wantSerial: "cli-serial", wantSource: CliOpt.String(), wantErr: false},
		{name: "S02", args: args{cliOpt: "", defaultValue: "default-serial", profile: "credhas-confighas"}, credFilePath: "testData/setMFASerial_credentials", cfgFilePath: "testData/setMFASerial_config", awsmfaCfgFilePath: "testData/setMFASerial_awsmfaConfiguration_has", wantSerial: "cred-serial", wantSource: SharedCredentials.String(), wantErr: false},
		{name: "S03", args: args{cliOpt: "", defaultValue: "default-serial", profile: "crednil-confighas"}, credFilePath: "testData/setMFASerial_credentials", cfgFilePath: "testData/setMFASerial_config", awsmfaCfgFilePath: "testData/setMFASerial_awsmfaConfiguration_has", wantSerial: "config-serial", wantSource: SharedConfig.String(), wantErr: false},
		{name: "S04", args: args{cliOpt: "", defaultValue: "default-serial", profile: "crednil-confignil"}, credFilePath: "testData/setMFASerial_credentials", cfgFilePath: "testData/setMFASerial_config", awsmfaCfgFilePath: "testData/setMFASerial_awsmfaConfiguration_has", wantSerial: "awsmfaCfg-serial", wantSource: AwsmfaConfig.String(), wantErr: false},
		{name: "F01", args: args{cliOpt: "", defaultValue: "default-serial", profile: "crednil-confignil"}, credFilePath: "testData/setMFASerial_credentials", cfgFilePath: "testData/setMFASerial_config", awsmfaCfgFilePath: "nil", wantSerial: "ERROR", wantSource: "ERROR", wantErr: true},
		{name: "F02", args: args{cliOpt: "", defaultValue: "unspecified", profile: "crednil-confignil"}, credFilePath: "testData/setMFASerial_credentials", cfgFilePath: "testData/setMFASerial_config", awsmfaCfgFilePath: "testData/setMFASerial_awsmfaConfiguration_nil", wantSerial: "ERROR", wantSource: "ERROR", wantErr: true},
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

			awsmfaCfg, _ := ini.Load(tt.awsmfaCfgFilePath)

			gotSerial, gotSource, err := setMFASerial(tt.args.cliOpt, tt.args.defaultValue, tt.args.profile, cred, cfg, awsmfaCfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("setMFASerial() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotSerial != tt.wantSerial {
				t.Errorf("setMFASerial() = %v, wantSerial %v", gotSerial, tt.wantSerial)
			}
			if gotSource != tt.wantSource {
				t.Errorf("setMFASerial() = %v, wantSource %v", gotSource, tt.wantSource)
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
		wantRoleArn  string
		wantSource   string
		wantErr      bool
	}{
		{name: "S01", args: args{cliRoleArn: "cli-role-arn", profile: "credhas-confighas"}, credFilePath: "testData/setRoleArn_credentials", cfgFilePath: "testData/setRoleArn_config", wantRoleArn: "cli-role-arn", wantSource: CliOpt.String(), wantErr: false},
		{name: "S02", args: args{cliRoleArn: "", profile: "credhas-confighas"}, credFilePath: "testData/setRoleArn_credentials", cfgFilePath: "testData/setRoleArn_config", wantRoleArn: "cred-role-arn", wantSource: SharedCredentials.String(), wantErr: false},
		{name: "S03", args: args{cliRoleArn: "", profile: "crednil-confighas"}, credFilePath: "testData/setRoleArn_credentials", cfgFilePath: "testData/setRoleArn_config", wantRoleArn: "config-role-arn", wantSource: SharedConfig.String(), wantErr: false},
		{name: "F01", args: args{cliRoleArn: "", profile: "crednil-confignil"}, credFilePath: "testData/setRoleArn_credentials", cfgFilePath: "testData/setRoleArn_config", wantRoleArn: "ERROR", wantSource: "ERROR", wantErr: true},
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

			gotRoleArn, gotSource, err := setRoleArn(tt.args.cliRoleArn, tt.args.profile, cred, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("setRoleArn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRoleArn != tt.wantRoleArn {
				t.Errorf("setRoleArn() = %v, wantRoleArn %v", gotRoleArn, tt.wantRoleArn)
			}
			if gotSource != tt.wantSource {
				t.Errorf("setRoleArn() = %v, wantSource %v", gotSource, tt.wantSource)
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
		name                string
		args                args
		credFilePath        string
		cfgFilePath         string
		awsmfaCfgFilePath   string
		wantRoleSessionName string
		wantSource          string
	}{
		{name: "S01", args: args{cliOpt: "cliOpt", defaultValue: "default-role-session-name", profile: "credhas-confighas"}, credFilePath: "testData/setRoleSessionName_credentials", cfgFilePath: "testData/setRoleSessionName_config", awsmfaCfgFilePath: "testData/setRoleSessionName_awsmfaConfiguration_has", wantSource: CliOpt.String(), wantRoleSessionName: "cliOpt"},
		{name: "S02", args: args{cliOpt: "", defaultValue: "default-role-session-name", profile: "credhas-confighas"}, credFilePath: "testData/setRoleSessionName_credentials", cfgFilePath: "testData/setRoleSessionName_config", awsmfaCfgFilePath: "testData/setRoleSessionName_awsmfaConfiguration_has", wantSource: SharedCredentials.String(), wantRoleSessionName: "cred-session-name"},
		{name: "S03", args: args{cliOpt: "", defaultValue: "default-role-session-name", profile: "crednil-confighas"}, credFilePath: "testData/setRoleSessionName_credentials", cfgFilePath: "testData/setRoleSessionName_config", awsmfaCfgFilePath: "testData/setRoleSessionName_awsmfaConfiguration_has", wantSource: SharedConfig.String(), wantRoleSessionName: "config-session-name"},
		{name: "S04", args: args{cliOpt: "", defaultValue: "default-role-session-name", profile: "crednil-confignil"}, credFilePath: "testData/setRoleSessionName_credentials", cfgFilePath: "testData/setRoleSessionName_config", awsmfaCfgFilePath: "testData/setRoleSessionName_awsmfaConfiguration_has", wantSource: AwsmfaConfig.String(), wantRoleSessionName: "awsmfaCfg-session-name"},
		{name: "S05", args: args{cliOpt: "", defaultValue: "default-role-session-name", profile: "crednil-confignil"}, credFilePath: "testData/setRoleSessionName_credentials", cfgFilePath: "testData/setRoleSessionName_config", awsmfaCfgFilePath: "testData/setRoleSessionName_awsmfaConfiguration_nil", wantSource: AwsmfaBuildIn.String(), wantRoleSessionName: "default-role-session-name"},
		{name: "S06", args: args{cliOpt: "", defaultValue: "default-role-session-name", profile: "crednil-confignil"}, credFilePath: "testData/setRoleSessionName_credentials", cfgFilePath: "testData/setRoleSessionName_config", awsmfaCfgFilePath: "nil", wantSource: AwsmfaBuildIn.String(), wantRoleSessionName: "default-role-session-name"},
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

			awsmfaCfg, _ := ini.Load(tt.awsmfaCfgFilePath)

			gotRoleSessionName, gotSource := setRoleSessionName(tt.args.cliOpt, tt.args.defaultValue, tt.args.profile, cred, cfg, awsmfaCfg)
			if gotRoleSessionName != tt.wantRoleSessionName {
				t.Errorf("setRoleSessionName() = %v, wantRoleSessionName %v", gotRoleSessionName, tt.wantRoleSessionName)
			}
			if gotSource != tt.wantSource {
				t.Errorf("setRoleSessionName() = %v, wantSource %v", gotRoleSessionName, tt.wantSource)
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
		awsmfaCfgFilePath      string
		credFilePath           string
		cfgFilePath            string
		wantEndpointRegion     string
		wantSource             string
	}{
		{name: "S01", args: args{cliOpt: "cliOpt", defaultValue: "default-region", profile: "before-credhas-confighas_after-credhas-confighas"}, existsEnvREGION: true, existsEnvDEFAULTREGION: true, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", awsmfaCfgFilePath: "testData/setEndpointRegion_awsmfaConfiguration_has", wantEndpointRegion: "cliOpt", wantSource: CliOpt.String()},
		{name: "S02", args: args{cliOpt: "", defaultValue: "default-region", profile: "before-credhas-confighas_after-credhas-confighas"}, existsEnvREGION: true, existsEnvDEFAULTREGION: true, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", awsmfaCfgFilePath: "testData/setEndpointRegion_awsmfaConfiguration_has", wantEndpointRegion: "env-region", wantSource: EnvAWSRegion.String()},
		{name: "S03", args: args{cliOpt: "", defaultValue: "default-region", profile: "before-credhas-confighas_after-credhas-confighas"}, existsEnvREGION: false, existsEnvDEFAULTREGION: true, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", awsmfaCfgFilePath: "testData/setEndpointRegion_awsmfaConfiguration_has", wantEndpointRegion: "env-default-region", wantSource: EnvAWSDefaultRegion.String()},
		{name: "S04", args: args{cliOpt: "", defaultValue: "default-region", profile: "before-credhas-confighas_after-credhas-confighas"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", awsmfaCfgFilePath: "testData/setEndpointRegion_awsmfaConfiguration_has", wantEndpointRegion: "before-cred", wantSource: SharedCredentialsBeforeMFAProfile.String()},
		{name: "S05", args: args{cliOpt: "", defaultValue: "default-region", profile: "before-crednil-confighas_after-credhas-confighas"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", awsmfaCfgFilePath: "testData/setEndpointRegion_awsmfaConfiguration_has", wantEndpointRegion: "before-config", wantSource: SharedConfigBeforeMFAProfile.String()},
		{name: "S06", args: args{cliOpt: "", defaultValue: "default-region", profile: "before-crednil-confignil_after-credhas-confighas"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", awsmfaCfgFilePath: "testData/setEndpointRegion_awsmfaConfiguration_has", wantEndpointRegion: "after-cred", wantSource: SharedCredentialsAfterMFAProfile.String()},
		{name: "S07", args: args{cliOpt: "", defaultValue: "default-region", profile: "before-crednil-confignil_after-crednil-confighas"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", awsmfaCfgFilePath: "testData/setEndpointRegion_awsmfaConfiguration_has", wantEndpointRegion: "after-config", wantSource: SharedConfigAfterMFAProfile.String()},
		{name: "S08", args: args{cliOpt: "", defaultValue: "default-region", profile: "before-crednil-confignil_after-crednil-confignil"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", awsmfaCfgFilePath: "testData/setEndpointRegion_awsmfaConfiguration_has", wantEndpointRegion: "awsmfaCfg-region", wantSource: AwsmfaConfig.String()},
		{name: "S09", args: args{cliOpt: "", defaultValue: "default-region", profile: "before-crednil-confignil_after-crednil-confignil"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", awsmfaCfgFilePath: "testData/setEndpointRegion_awsmfaConfiguration_nil", wantEndpointRegion: "default-region", wantSource: AwsmfaBuildIn.String()},
		{name: "S10", args: args{cliOpt: "", defaultValue: "default-region", profile: "before-crednil-confignil_after-crednil-confignil"}, existsEnvREGION: false, existsEnvDEFAULTREGION: false, credFilePath: "testData/setEndpointRegion_credentials", cfgFilePath: "testData/setEndpointRegion_config", awsmfaCfgFilePath: "nil", wantEndpointRegion: "default-region", wantSource: AwsmfaBuildIn.String()},
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

			awsmfaCfg, _ := ini.Load(tt.awsmfaCfgFilePath)

			gotEndpointRegion, gotSource := setEndpointRegion(tt.args.cliOpt, tt.args.defaultValue, tt.args.profile, cred, cfg, awsmfaCfg)
			if gotEndpointRegion != tt.wantEndpointRegion {
				t.Errorf("setEndpointRegion() gotEndpointRegion = %v, wantEndpointRegion %v", gotEndpointRegion, tt.wantEndpointRegion)
			}
			if gotSource != tt.wantSource {
				t.Errorf("setEndpointRegion() gotSource = %v, wantSource %v", gotSource, tt.wantSource)
			}
		})
	}
}

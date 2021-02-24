package cmd

import (
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

// setMode returns action mode to be used.
// Priority
// 1. cli option: --mode
// 2. shared credentials or config file: if given profile has a role_arn param.
// 3. awsmfa configuration file: [default-value] profile
// 4. awsmfa build in default value
// If the mode is not whether 'get-session-token' or 'assume-role', awsmfa returns an error.
func setMode(cliOpt string, defaultValue string, profile string, cred *ini.File, cfg *ini.File, awsmfaCfg *ini.File) (mode string, source string, err error) {
	if cliOpt == "get-session-token" || cliOpt == "assume-role" {
		return cliOpt, CliOpt.String(), nil
	} else if cliOpt != "" {
		return "ERROR", "ERROR", fmt.Errorf("Invalid action mode: action mode should be \"get-session-token\" or \"assume-role\"")
	}

	if cred.Section(profile).HasKey("awsmfa_role_arn") {
		return "assume-role", SharedCredentials.String(), nil
	}

	if cfg.Section("profile " + profile).HasKey("awsmfa_role_arn") {
		return "assume-role", SharedConfig.String(), nil
	}

	if awsmfaCfg != nil {
		if v := awsmfaCfg.Section("default-value").Key("mode").String(); v == "get-session-token" || v == "assume-role" {
			return v, AwsmfaConfig.String(), nil
		}
	}

	if defaultValue == "get-session-token" || defaultValue == "assume-role" {
		return defaultValue, AwsmfaBuildIn.String(), nil
	}

	return "ERROR", "ERROR", fmt.Errorf("Invalid action mode: action mode should be \"get-session-token\" or \"assume-role\"")
}

// setProfile returns a profile to be used.
// Priority
// 1. cli option: --profile
// 2. environment variable: AWS_PROFILE
// 3. awsmfa configuration file: [default-value] profile
// 4. awsmfa build in default value
func setProfile(cliOpt string, defaultValue string, awsmfaCfg *ini.File) (mode string, source string) {
	if cliOpt != "" {
		return cliOpt, CliOpt.String()
	}
	if env, exists := os.LookupEnv("AWS_PROFILE"); exists == true {
		return env, EnvAWSProfile.String()
	}
	if awsmfaCfg != nil {
		if v := awsmfaCfg.Section("default-value").Key("profile").String(); v != "" {
			return v, AwsmfaConfig.String()
		}
	}
	return defaultValue, AwsmfaBuildIn.String()
}

// setDurationSeconds returns duration seconds to be used.
// Priority
// 1. cli option: --duration-seconds
// 2. shared credentials file: ${HOME}/.aws/credentials (by default)
// 3. shared config file: ${HOME}/.aws/config (by default)
// 4. awsmfa configuration file: [default-value] duration_seconds
// 5. awsmfa build in default value
func setDurationSeconds(cliOpt int32, defaultValue int32, profile string, cred *ini.File, cfg *ini.File, awsmfaCfg *ini.File) (duration int32, source string) {
	if cliOpt != 0 {
		return cliOpt, CliOpt.String()
	}
	if v, err := cred.Section(profile).Key("duration_seconds").Int(); err == nil {
		return int32(v), SharedCredentials.String()
	}
	if v, err := cfg.Section("profile " + profile).Key("duration_seconds").Int(); err == nil {
		return int32(v), SharedConfig.String()
	}
	if awsmfaCfg != nil {
		if v, err := awsmfaCfg.Section("default-value").Key("duration_seconds").Int(); err == nil {
			return int32(v), AwsmfaConfig.String()
		}
	}
	return defaultValue, AwsmfaBuildIn.String()
}

// setMFASerial returns mfa device's serial number to be used.
// Priority
// 1. cli option: --serial-number
// 2. shared credentials file: ${HOME}/.aws/credentials (by default)
// 3. shared config file: ${HOME}/.aws/config (by default)
// 4. awsmfa configuration file: [default-value] duration_seconds
// If any serial number is not specified, setMFASerial returns error.
func setMFASerial(cliOpt string, defaultValue string, profile string, cred *ini.File, cfg *ini.File, awsmfaCfg *ini.File) (serial string, source string, err error) {
	if cliOpt != "" {
		return cliOpt, CliOpt.String(), nil
	}
	if v := cred.Section(profile).Key("mfa_serial").String(); v != "" {
		return v, SharedCredentials.String(), nil
	}
	if v := cfg.Section("profile " + profile).Key("mfa_serial").String(); v != "" {
		return v, SharedConfig.String(), nil
	}
	if awsmfaCfg != nil {
		if v := awsmfaCfg.Section("default-value").Key("mfa_serial").String(); v != "" {
			return v, AwsmfaConfig.String(), nil
		}
	}

	return "ERROR", "ERROR", fmt.Errorf("no mfa_serial specified")
}

// setRoleArn returns a role arn related to given profile.
// [CAUTION!] This function parses only custom parameter 'awsmfa_role_arn',
// not aws build in parameter 'role_arn'.
// Priority
// 1. cli option: --role-arn
// 2. shared credentials file: ${HOME}/.aws/credentials (by default)
// 3. shared config file: ${HOME}/.aws/config (by default)
// If any arn is not specified, setRoleArn returns error.
func setRoleArn(cliOpt string, profile string, cred *ini.File, cfg *ini.File) (roleArn string, source string, err error) {
	if cliOpt != "" {
		return cliOpt, CliOpt.String(), nil
	}
	if v := cred.Section(profile).Key("awsmfa_role_arn").String(); v != "" {
		return v, SharedCredentials.String(), nil
	}
	if v := cfg.Section("profile " + profile).Key("awsmfa_role_arn").String(); v != "" {
		return v, SharedConfig.String(), nil
	}

	return "ERROR", "ERROR", fmt.Errorf("no awsmfa_role_arn specified")
}

// setRoleSessionName returns a role session name to be used.
// Priority
// 1. cli option: --role-session-name
// 2. shared credentials file: ${HOME}/.aws/credentials (by default)
// 3. shared config file: ${HOME}/.aws/config (by default)
// 4. awsmfa configuration file: [default-value] role_session_name
// 5. awsmfa build in default value
func setRoleSessionName(cliOpt string, defaultValue string, profile string, cred *ini.File, cfg *ini.File, awsmfaCfg *ini.File) (roleSessionName string, source string) {
	if cliOpt != "" {
		return cliOpt, CliOpt.String()
	}
	if v := cred.Section(profile).Key("role_session_name").String(); v != "" {
		return v, SharedCredentials.String()
	}
	if v := cfg.Section("profile " + profile).Key("role_session_name").String(); v != "" {
		return v, SharedConfig.String()
	}
	if awsmfaCfg != nil {
		if v := awsmfaCfg.Section("default-value").Key("role_session_name").String(); v != "" {
			return v, AwsmfaConfig.String()
		}
	}
	return defaultValue, AwsmfaBuildIn.String()
}

// setEndpointRegion returns mfa device's serial number to be used.
// Priority
// 1. cli option: --endpoint-region
// 2. environment variable: AWS_REGION
// 3. environment variable: AWS_DEFAULT_REGION
// 4. profile-before-mfa in shared credentials file: ${HOME}/.aws/credentials (by default)
// 5. profile-before-mfa in shared config file: ${HOME}/.aws/config (by default)
// 6. profile in shared credentials file: ${HOME}/.aws/credentials (by default)
// 7. profile in shared config file: ${HOME}/.aws/config (by default)
// 6. awsmfa configuration file: [default-value] duration_seconds (Need to overwrite build in default value in advance)
// 7. awsmfa build in default value
func setEndpointRegion(cliOpt string, defaultValue string, profile string, cred *ini.File, cfg *ini.File, awsmfaCfg *ini.File) (endpointRegion string, source string) {
	if cliOpt != "" {
		return cliOpt, CliOpt.String()
	}
	if env, exists := os.LookupEnv("AWS_REGION"); exists == true {
		return env, EnvAWSRegion.String()
	}
	if env, exists := os.LookupEnv("AWS_DEFAULT_REGION"); exists == true {
		return env, EnvAWSDefaultRegion.String()
	}
	if v := cred.Section(profile + beforeMFASuffix).Key("region").String(); v != "" {
		return v, SharedCredentialsBeforeMFAProfile.String()
	}
	if v := cfg.Section("profile " + profile + beforeMFASuffix).Key("region").String(); v != "" {
		return v, SharedConfigBeforeMFAProfile.String()
	}
	if v := cred.Section(profile).Key("region").String(); v != "" {
		return v, SharedCredentialsAfterMFAProfile.String()
	}
	if v := cfg.Section("profile " + profile).Key("region").String(); v != "" {
		return v, SharedConfigAfterMFAProfile.String()
	}
	if awsmfaCfg != nil {
		if v := awsmfaCfg.Section("default-value").Key("region").String(); v != "" {
			return v, AwsmfaConfig.String()
		}
	}
	return defaultValue, AwsmfaBuildIn.String()
}

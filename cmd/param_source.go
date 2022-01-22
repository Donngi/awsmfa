package cmd

// paramSource of request params.
type paramSource int

const (
	_ paramSource = iota
	CliOpt
	SharedCredentials
	SharedCredentialsBeforeMFAProfile
	SharedCredentialsAfterMFAProfile
	SharedConfig
	SharedConfigBeforeMFAProfile
	SharedConfigAfterMFAProfile
	AwsmfaConfig
	AwsmfaBuildIn
	EnvAWSDefaultRegion
	EnvAWSRegion
	EnvAWSProfile
)

func (s paramSource) String() string {
	switch s {
	case CliOpt:
		return "cli option"
	case SharedCredentials:
		return "shared credentials file"
	case SharedCredentialsBeforeMFAProfile:
		return "shared credentials file (before-mfa profile)"
	case SharedCredentialsAfterMFAProfile:
		return "shared credentials file (after-mfa profile)"
	case SharedConfig:
		return "shared config file"
	case SharedConfigBeforeMFAProfile:
		return "shared config file (before-mfa profile)"
	case SharedConfigAfterMFAProfile:
		return "shared config file (after-mfa profile)"
	case AwsmfaConfig:
		return "awsmfa configuration file"
	case AwsmfaBuildIn:
		return "awsmfa build in default"
	case EnvAWSDefaultRegion:
		return "env AWS_DEFAULT_REGION"
	case EnvAWSRegion:
		return "env AWS_REGION"
	case EnvAWSProfile:
		return "env AWS_PROFILE"
	}
	return "unknown paramSource"
}

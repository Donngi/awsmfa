package cmd

import (
	"fmt"
	"os"
)

// generateCredentialsSkeleton returns skeleton of shared credentials file.
func generateCredentialsSkeleton(mode string) (string, error) {
	skeleton := ""
	switch mode {
	case "get-session-token":
		skeleton = `[sample-before-mfa]
aws_access_key_id     = YOUR_ACCESS_KEY_ID_HERE!!!
aws_secret_access_key = YOUR_SECRET_ACCESS_KEY_HERE!!!
	`
	case "assume-role":
		skeleton = `[sample-before-mfa]
aws_access_key_id     = YOUR_ACCESS_KEY_ID_HERE!!!
aws_secret_access_key = YOUR_SECRET_ACCESS_KEY_HERE!!!
	`
	default:
		return "ERROR", fmt.Errorf("invalid mode")
	}

	return skeleton, nil
}

// generateConfigSkeleton returns skeleton of shared credentials file.
func generateConfigSkeleton(mode string) (string, error) {
	skeleton := ""
	switch mode {
	case "get-session-token":
		skeleton = `[profile sample-before-mfa]
region     = REGION_TO_CONNECT_IN_EXECUTING_STS_GET_SESSION_TOKEN # Such as ap-northeast-1, us-east-1
output     = json
mfa_serial = YOUR_MFA_SERIAL_HERE!!! # Such as arn:aws:iam::XXXXXXXXXXX:mfa/YYYY

[profile sample]
region = REGION_TO_CONNECT_AFTER_MFA # Such as ap-northeast-1, us-east-1
output = json

# If you want to assume role (switch role), please uncomment below section.
# [profile switched-role]
# region         = REGION_TO_CONNECT_AFTER_ASSUMED_ROLE # Such as ap-northeast-1, us-east-1
# output         = json
# role_arn       = YOUR_ROLE_TO_ASSUME_HERE!!! # Such as arn:aws:iam::XXXXXXXXXXX:role/ZZZZ
# source_profile = sample-before-mfa
	`
	case "assume-role":
		skeleton = `[profile sample-before-mfa]
region          = REGION_TO_CONNECT_IN_EXECUTING_STS_ASSUME_ROLE # Such as ap-northeast-1, us-east-1
output          = json
mfa_serial      = YOUR_MFA_SERIAL_HERE!!! # Such as arn:aws:iam::XXXXXXXXXXX:mfa/YYYY
awsmfa_role_arn = YOUR_ROLE_TO_ASSUME_HERE!!! # Such as arn:aws:iam::XXXXXXXXXXX:role/ZZZZ

[profile sample]
region = REGION_TO_CONNECT_AFTER_MFA # Such as ap-northeast-1, us-east-1
output = json
	`
	default:
		return "ERROR", fmt.Errorf("invalid mode")
	}

	return skeleton, nil
}

// generateConfigurationFile create awsmfa's setting file.
func generateConfigurationFile(dir string, file string) error {
	content := `[filepath]
credentials_file_path = ${HOME}/.aws/credentials
config_file_path      = ${HOME}/.aws/config

[default-value] 
suffix_of_before_mfa_profile       = -before-mfa
mode                               = get-session-token
profile                            = default
# mfa_serial                       = YOUR_SERIAL_HERE!!!
endpoint_region                    = aws_global
duration_seconds_get_session_token = 43200
duration_seconds_assume_role       = 3600
	`

	p := dir + "/" + file

	if _, err := os.Stat(p); err == nil {
		return fmt.Errorf("The file already exists. %v", p)
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %v: %w", dir, err)
	}

	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if _, err = f.WriteString(content); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	if err = f.Sync(); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

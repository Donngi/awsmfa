package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// cli option's input value
var (
	cliMode                        string
	cliProfile                     string
	cliDurationSeconds             int32
	cliMfaSerial                   string
	cliEndpointRegion              string
	cliRoleArn                     string
	cliRoleSessionName             string
	cliGenerateCredentialsSkeleton string
	cliGenerateConfigSkeleton      string
	cliGenerateConfigurationFile   bool
	cliForce                       bool
	cliSilent                      bool
)

// Default values.
// Changeable by awsmfa's configuration file ($HOME/.awsmfa/configuration).
// The comment on the side is a corresponded parameter in the configuration file ([section-name] key-name).
var (
	credentialsFilePath                   string                       // [filepath] credentials_file_path
	configFilePath                        string                       // [filepath] config_file_path
	beforeMFASuffix                              = "-before-mfa"       // [default-value] suffix_of_before_mfa_profile
	defaultMode                                  = "get-session-token" // [default-value] mode
	defaultProfile                               = "default"           // [default-value] profile
	defaultMFASerial                             = "unspecified"       // [default-value] mfa_serial
	defaultEndpointRegion                        = "aws_global"        // [default-value] endpoint_region
	defaultDurationSecondsGetSessionToken int32  = 43200               // [default-value] duration_seconds_get_session_token
	defaultDurationSecondsAssumeRole      int32  = 3600                // [default-value] duration_seconds_assume_role
	defaultRoleSessionName                       = "awsmfa-session"    // [default-value] role_session_name
)

// Source of request params.
var (
	sourceProfile         string
	sourceDurationSeconds string
	sourceMFASerial       string
	sourceRoleArn         string
	sourceRoleSessionName string
	sourceEndpointRegion  string
	sourceAPIType         string
)

var (
	awsmfaCfgFileDir  = os.ExpandEnv("$HOME/.awsmfa")
	awsmfaCfgFileName = "configuration"
	awsmfaCfgFilePath = awsmfaCfgFileDir + "/" + awsmfaCfgFileName
)

func initBuildInDefault() {
	p, err := os.UserHomeDir()
	if err != nil {
		credentialsFilePath = "/.aws/credentials"
		configFilePath = "/.aws/config"
	} else {
		credentialsFilePath = p + "/.aws/credentials"
		configFilePath = p + "/.aws/config"
	}
}

func initUserDefault(awsmfaCfgFilePath string) {
	awsmfaCfg, err := ini.Load(awsmfaCfgFilePath)
	if err != nil {
		return
	}

	// Overwrite default values.
	credentialsFilePath = os.ExpandEnv(awsmfaCfg.Section("filepath").Key("credentials_file_path").String())
	configFilePath = os.ExpandEnv(awsmfaCfg.Section("filepath").Key("config_file_path").String())
	beforeMFASuffix = awsmfaCfg.Section("default-value").Key("suffix_of_before_mfa_profile").String()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cmd := NewCmdRoot()
	if err := cmd.Execute(); err != nil {
		printErrorRed(err)
	}
}

// NewCmdRoot returns the root command.
func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "awsmfa",
		Short:         "A simple utility command to pass the multi factor authentication (MFA) of AWS account",
		Long:          `awsmfa is a command line utility to pass the multi factor authentication (MFA) of AWS account. You could see the help page with --help or -h option.`,
		RunE:          runRootCmd,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cobra.OnInitialize(func() {
		initBuildInDefault()
		initUserDefault(awsmfaCfgFilePath)
	})

	// Flags
	cmd.PersistentFlags().StringVarP(&cliMode, "mode", "m", "", "The action mode of awsmfa, get-session-token or assume-role. The default value is get-session-token. If you specify the awsmfa_role_arn in shared credentials/config file or --role-arn option, awsmfa automatically turns the mode to assume-role.")
	cmd.PersistentFlags().StringVarP(&cliProfile, "profile", "p", "", "The profile used to get the token. You should set 'xxxx' if you have set 'xxxx-before-mfa' in the shared credentials/config file (.aws/credentials and .aws/config). The default value is 'default'")
	cmd.PersistentFlags().Int32VarP(&cliDurationSeconds, "duration-seconds", "d", 0, "The duration of the temporary security credential. Minimun value: 900 seconds (15 minutes). Max value is different depend on the authentification mode. If you try to get token of same account (with GetSessionToken), Max value is 129600 seconds (36h). In the case of assume role (with AssumeRole), Max value is 43200 seconds (12h). The default value is GetSessionToken=43200 seconds (12h), AssumeRole=3600 seconds (1h).")
	cmd.PersistentFlags().StringVar(&cliMfaSerial, "serial-number", "", "The serial number of the MFA device. The value is either an ARN of a virtual device (arn:aws:iam::123456789012:mfa/user) or the serial number of real device.")
	cmd.PersistentFlags().StringVarP(&cliEndpointRegion, "endpoint-region", "e", "", "The sts endpoint where awsmfa accesses to get a temporary credential. Such as ap-northeast-1, us-east-1.")
	cmd.PersistentFlags().StringVarP(&cliRoleArn, "role-arn", "r", "", "The ARN of the IAM role to assume. If you specify this option, awsmfa automatically turns the mode (--mode, -m) to assume-role.")
	cmd.PersistentFlags().StringVar(&cliRoleSessionName, "role-session-name", "", "The session name which will be logged to the AWS CloudTrail. The default value is awsmfa-session.")
	cmd.PersistentFlags().BoolVarP(&cliForce, "force", "f", false, "Force reflesh temporary credentials.")
	cmd.PersistentFlags().BoolVarP(&cliSilent, "silent", "s", false, "Hide source of request params.")

	cmd.PersistentFlags().StringVar(&cliGenerateCredentialsSkeleton, "generate-credentials-skeleton", "", "Generate skeleton of shared credentials file (by default, ${HOME}/.aws/credentials) for specified action mode, get-session-token or assume-role.")
	cmd.PersistentFlags().StringVar(&cliGenerateConfigSkeleton, "generate-config-skeleton", "", "Generate skeleton of shared config file (by default, ${HOME}/.aws/config). for specified action mode, get-session-token or assume-role.")
	cmd.PersistentFlags().BoolVar(&cliGenerateConfigurationFile, "generate-configuration-file", false, fmt.Sprintf("Generate awsmfa's configuration file at %v", awsmfaCfgFilePath))

	// Sub commands
	cmd.AddCommand(NewCmdCompletion())
	// TODO: Comment out if cobra v1.1.4~ is launched
	// This will hide help message of completion.
	// cmd.CompletionOptions.DisableDescriptions = true

	return cmd
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	// If --generate-xxxx-skeleton is specified, show them and terminate.
	if cliGenerateCredentialsSkeleton != "" {
		skeleton, err := generateCredentialsSkeleton(cliGenerateCredentialsSkeleton)
		if err != nil {
			return fmt.Errorf("failed to generate skeleton. Please use '--generate-credentials-skeleton get-session-token' or '--generate-credentials-skeleton assume-role' instead: %w", err)
		}

		fmt.Printf("%s", skeleton)
		return nil
	}

	if cliGenerateConfigSkeleton != "" {
		skeleton, err := generateConfigSkeleton(cliGenerateConfigSkeleton)
		if err != nil {
			return fmt.Errorf("failed to generate skeleton. Please use '--generate-config-skeleton get-session-token' or '--generate-config-skeleton assume-role' instead: %w", err)
		}

		fmt.Printf("%s", skeleton)
		return nil
	}

	if cliGenerateConfigurationFile {
		if err := generateConfigurationFile(awsmfaCfgFileDir, awsmfaCfgFileName); err != nil {
			return fmt.Errorf("failed to initialize awsmfa's configuration file: %w", err)
		}
		printCyan(fmt.Sprintf("Successfully create awsmfa's configuration file at %v\n", awsmfaCfgFilePath))
		return nil
	}

	// Load credentials, config and awsmfa's configuration files.
	cred, err := ini.Load(credentialsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load credentials file: %w", err)
	}
	cfg, err := ini.Load(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}
	awsmfaCfg, err := ini.Load(awsmfaCfgFilePath)
	if err != nil {
		printBlue(fmt.Sprintf("[Tips] There isn't an awsmfa's configuration file. You can set some default values to place the configuration file at: %v. If you would like to make it by cli, please use 'awsmfa --generate-configuration-file'\n", awsmfaCfgFilePath))
	}

	// Set target profile.
	profile, _s := setProfile(cliProfile, defaultProfile, awsmfaCfg)
	sourceProfile = _s

	// Check if initial configuration has been completed correctly.
	if _, err := cred.GetSection(profile + beforeMFASuffix); err != nil {
		return fmt.Errorf("The profile \"%v%v\" is not set to your credentials file. Please add the profile to %v. You can get template of credentials file by using '--generate-credentials-skeleton get-session-token' or '--generate-credentials-skeleton assume-role'", profile, beforeMFASuffix, credentialsFilePath)
	}
	if _, err := cfg.GetSection("profile " + profile + beforeMFASuffix); err != nil {
		return fmt.Errorf("The profile \"%v%v\" is not set to your config file. Please add the profile to %v. You can get template of config file by using '--generate-config-skeleton get-session-token' or '--generate-config-skeleton assume-role'", profile, beforeMFASuffix, configFilePath)
	}

	// Judge if reflesh is needed.
	if !cmd.Flags().Lookup("force").Changed {
		if res, due := hasActiveToken(profile, cred); res == true {
			printCyan(fmt.Sprintf("Your temporary token is still active. Expired at %v\n", due))
			return nil
		}
	}

	// Execute a handler according to action mode (GetSessionToken or AssumeRole).
	// The action mode is forcely turned to "assume-role" if --role-arn is specified or awsmfa_role_arn is specified in your shared credentials/config file.
	mode, _s, err := setMode(cliMode, defaultMode, profile+beforeMFASuffix, cred, cfg, awsmfaCfg)
	sourceAPIType = _s
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	switch mode {
	case "get-session-token":
		if err := handleGetSessionToken(profile, cred, cfg, awsmfaCfg, cmd.Flags().Lookup("silent").Changed); err != nil {
			return fmt.Errorf("failed to get-session-token: %w", err)
		}
	case "assume-role":
		if err := handleAssumeRole(profile, cred, cfg, awsmfaCfg, cmd.Flags().Lookup("silent").Changed); err != nil {
			return fmt.Errorf("failed to assume-role: %w", err)
		}
	default:
		return fmt.Errorf("invalid action mode: %v", mode)
	}

	return nil
}

func handleGetSessionToken(profile string, cred *ini.File, cfg *ini.File, awsmfaCfg *ini.File, isSilent bool) error {
	// Load long term credentials.
	// To match the priority of credentials and config params (such as access_key) to aws's default order, including environment variables,
	// awsmfa is sure to reload credentials and config file with aws-sdk-go-v2's build in loading config function before execute GetSessionToken API.
	c, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile+beforeMFASuffix),
	)
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	// Set request params.
	durationSeconds, _s := setDurationSeconds(cliDurationSeconds, defaultDurationSecondsGetSessionToken, profile+beforeMFASuffix, cred, cfg, awsmfaCfg)
	sourceDurationSeconds = _s
	mfaSerial, _s, err := setMFASerial(cliMfaSerial, defaultMFASerial, profile+beforeMFASuffix, cred, cfg, awsmfaCfg)
	sourceMFASerial = _s
	if err != nil {
		return fmt.Errorf("The mfa_serial is not specified. You can set it in %v, %v, %v or --serial-number", credentialsFilePath, configFilePath, awsmfaCfgFilePath)
	}
	endpointRegion, _s := setEndpointRegion(cliEndpointRegion, defaultEndpointRegion, profile, cred, cfg, awsmfaCfg)
	sourceEndpointRegion = _s

	// Show request params.
	h, m, s := secToHMS(durationSeconds)
	fmt.Printf("Try to get temporary token with following params ...\n")
	table := tablewriter.NewWriter(os.Stdout)
	data := [][]string{}
	if isSilent {
		data = [][]string{
			{"Profile to exec MFA", profile + beforeMFASuffix},
			// {"Credentials", fmt.Sprintf("[Only for DEBUG] %+v", c.Credentials)},
			{"Duration of token", fmt.Sprintf("%v sec (%vh %vm %vs)", durationSeconds, h, m, s)},
			{"MFA device's serial", mfaSerial},
			{"Region", endpointRegion},
			{"API Type", "AWS STS GetSessionToken"},
		}
		table.SetHeader([]string{"Parameter", "Value"})
	} else {
		data = [][]string{
			{"Profile to exec MFA", profile + beforeMFASuffix, sourceProfile},
			// {"Credentials", fmt.Sprintf("[Only for DEBUG] %+v", c.Credentials)},
			{"Duration of token", fmt.Sprintf("%v sec (%vh %vm %vs)", durationSeconds, h, m, s), sourceDurationSeconds},
			{"MFA device's serial", mfaSerial, sourceMFASerial},
			{"Region", endpointRegion, sourceEndpointRegion},
			{"API Type", "AWS STS GetSessionToken", sourceAPIType},
		}
		table.SetHeader([]string{"Parameter", "Value", "Source"})
	}
	for _, v := range data {
		table.Append(v)
	}
	table.Render()

	// Get MFA token code from Stdin.
	fmt.Print("Input your MFA token code: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	tokenCode := scanner.Text()

	// Exec GetSessionToken API.
	stsClient := sts.NewFromConfig(c)
	token, err := stsClient.GetSessionToken(context.TODO(), &sts.GetSessionTokenInput{
		DurationSeconds: &durationSeconds,
		SerialNumber:    &mfaSerial,
		TokenCode:       &tokenCode,
	})
	if err != nil {
		return fmt.Errorf("something occured in calling AWS STS GetSessionToken API: %w", err)
	}

	// Add temporary token to the credentials file.
	if err := saveTemporaryTokenFromGetSessionToken(token, profile, credentialsFilePath); err != nil {
		return fmt.Errorf("failed to save temporary credentials to file: %w", err)
	}

	printCyan(fmt.Sprintf("Success! New temporary credentials is saved as profile: %v\n", profile))
	return nil
}

func handleAssumeRole(profile string, cred *ini.File, cfg *ini.File, awsmfaCfg *ini.File, isSilent bool) error {
	// Load long term credentials.
	// To match the priority of credentials and config params (such as access_key) to aws's default order, including environment variables,
	// awsmfa is sure to reload credentials and config file with aws-sdk-go-v2's build in loading config function before execute AssumeRole API.
	c, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile+beforeMFASuffix),
	)
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}

	// Set request params.
	durationSeconds, _s := setDurationSeconds(cliDurationSeconds, defaultDurationSecondsAssumeRole, profile+beforeMFASuffix, cred, cfg, awsmfaCfg)
	sourceDurationSeconds = _s
	mfaSerial, _s, err := setMFASerial(cliMfaSerial, defaultMFASerial, profile+beforeMFASuffix, cred, cfg, awsmfaCfg)
	sourceMFASerial = _s
	if err != nil {
		return fmt.Errorf("The mfa_serial is not specified. You can set it in %v, %v, %v or --serial-number", credentialsFilePath, configFilePath, awsmfaCfgFilePath)
	}
	endpointRegion, _s := setEndpointRegion(cliEndpointRegion, defaultEndpointRegion, profile, cred, cfg, awsmfaCfg)
	sourceEndpointRegion = _s
	roleArn, _s, err := setRoleArn(cliRoleArn, profile+beforeMFASuffix, cred, cfg)
	sourceRoleArn = _s
	if err != nil {
		return fmt.Errorf("The role_arn is not specified. You can set it in %v, %v or --role-arn", credentialsFilePath, configFilePath)
	}
	roleSessionName, _s := setRoleSessionName(cliRoleSessionName, defaultRoleSessionName, profile+beforeMFASuffix, cred, cfg, awsmfaCfg)
	sourceRoleSessionName = _s

	// Show request params.
	h, m, s := secToHMS(durationSeconds)
	fmt.Printf("Try to get temporary token with following params ...\n")
	table := tablewriter.NewWriter(os.Stdout)
	data := [][]string{}
	if isSilent {
		data = [][]string{
			{"Profile to exec MFA", profile + beforeMFASuffix},
			// {"Credentials", fmt.Sprintf("[Only for DEBUG] %+v", c.Credentials)},
			{"Role arn to assume", fmt.Sprintf("%v", roleArn)},
			{"Role session name", fmt.Sprintf("%v", roleSessionName)},
			{"Duration of token", fmt.Sprintf("%v sec (%vh %vm %vs)", durationSeconds, h, m, s)},
			{"MFA device's serial", mfaSerial},
			{"Region", endpointRegion},
			{"API Type", "AWS STS AssumeRole"},
		}
		table.SetHeader([]string{"Parameter", "Value"})
	} else {
		data = [][]string{
			{"Profile to exec MFA", profile + beforeMFASuffix},
			// {"Credentials", fmt.Sprintf("[Only for DEBUG] %+v", c.Credentials)},
			{"Role arn to assume", fmt.Sprintf("%v", roleArn)},
			{"Role session name", fmt.Sprintf("%v", roleSessionName)},
			{"Duration of token", fmt.Sprintf("%v sec (%vh %vm %vs)", durationSeconds, h, m, s)},
			{"MFA device's serial", mfaSerial},
			{"Region", endpointRegion},
			{"API Type", "AWS STS AssumeRole"},
		}
		table.SetHeader([]string{"Parameter", "Value", "Source"})
	}
	for _, v := range data {
		table.Append(v)
	}
	table.Render()

	// Get MFA token code from Stdin.
	fmt.Print("Input your MFA token code: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	tokenCode := scanner.Text()

	// Exec AssumeRole API.
	stsClient := sts.NewFromConfig(c)
	token, err := stsClient.AssumeRole(context.TODO(), &sts.AssumeRoleInput{
		DurationSeconds: &durationSeconds,
		SerialNumber:    &mfaSerial,
		RoleArn:         &roleArn,
		RoleSessionName: &roleSessionName,
		TokenCode:       &tokenCode,
	})
	if err != nil {
		return fmt.Errorf("something occured in calling AWS STS AssumeRole API: %w", err)
	}

	// Add temporary token to the credentials file.
	if err := saveTemporaryTokenFromAssumeRole(token, profile, credentialsFilePath); err != nil {
		return fmt.Errorf("failed to save temporary credentials to file: %w", err)
	}

	printCyan(fmt.Sprintf("Success! New temporary credentials is saved as profile: %v\n", profile))
	return nil
}

// hasActiveToken checks if the specified profile has an active token.
func hasActiveToken(profile string, cred *ini.File) (hasActiveToken bool, due *time.Time) {
	if sec, err := cred.GetSection(profile); err == nil {
		if tokenDue, err := sec.Key("expiration").TimeFormat(time.RFC3339); err == nil {
			if !isExpired(tokenDue, time.Now().UTC()) {
				return true, &tokenDue
			}
		}
	}
	return false, nil
}

// isExpired checks if a temporary token is expired.
func isExpired(tokenDue time.Time, comparison time.Time) bool {
	if comparison.After(tokenDue) {
		return true
	}
	return false
}

// secToHMS convert seconds to hour, min and sec.
func secToHMS(seconds int32) (hour int32, min int32, sec int32) {
	h := seconds / 3600
	m := (seconds - h*3600) / 60
	s := (seconds - h*3600 - m*60)

	return h, m, s
}

// saveTemporaryTokenFromGetSessionToken writes credentials to a shared credentials file.
func saveTemporaryTokenFromGetSessionToken(token *sts.GetSessionTokenOutput, profile string, credentialsFilePath string) error {
	cred, err := ini.Load(credentialsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load credentials file: %w", err)
	}
	cred.Section(profile).Key("aws_access_key_id").SetValue(*token.Credentials.AccessKeyId)
	cred.Section(profile).Key("aws_secret_access_key").SetValue(*token.Credentials.SecretAccessKey)
	cred.Section(profile).Key("aws_session_token").SetValue(*token.Credentials.SessionToken)
	cred.Section(profile).Key("expiration").SetValue(token.Credentials.Expiration.Format(time.RFC3339))

	if err := cred.SaveTo(credentialsFilePath); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}

// saveTemporaryTokenFromAssumeRole writes credentials to a shared credentials file.
func saveTemporaryTokenFromAssumeRole(token *sts.AssumeRoleOutput, profile string, credentialsFilePath string) error {
	cred, err := ini.Load(credentialsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load credentials file: %w", err)
	}
	cred.Section(profile).Key("aws_access_key_id").SetValue(*token.Credentials.AccessKeyId)
	cred.Section(profile).Key("aws_secret_access_key").SetValue(*token.Credentials.SecretAccessKey)
	cred.Section(profile).Key("aws_session_token").SetValue(*token.Credentials.SessionToken)
	cred.Section(profile).Key("expiration").SetValue(token.Credentials.Expiration.Format(time.RFC3339))

	if err := cred.SaveTo(credentialsFilePath); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}
	return nil
}

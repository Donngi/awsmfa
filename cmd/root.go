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
)

// Default values.
// Changeable by awsmfa's configuration file ($HOME/.awsmfa/configuration).
// The comment on the side is a corresponded parameter in the configuration file ([section-name] key-name).
var (
	credentialsFilePath                         = os.ExpandEnv("$HOME/.aws/credentials") // [filepath] credentials_file_path
	configFilePath                              = os.ExpandEnv("$HOME/.aws/config")      // [filepath] config_file_path
	beforeMFASuffix                             = "-before-mfa"                          // [default-value] suffix_of_before_mfa_profile
	defaultMode                                 = "get-session-token"                    // [default-value] mode
	defaultProfile                              = "default"                              // [default-value] profile
	defaultMFASerial                            = "unspecified"                          // [default-value] mfa_serial
	defaultEndpointRegion                       = "aws_global"                           // [default-value] endpoint_region
	defaultDurationSecondsGetSessionToken int32 = 43200                                  // [default-value] duration_seconds_get_session_token
	defaultDurationSecondsAssumeRole      int32 = 3600                                   // [default-value] duration_seconds_assume_role
	defaultRoleSessionName                      = "awsmfa-session"                       // [default-value] role_session_name
)

var (
	configurationFileDir  = os.ExpandEnv("$HOME/.awsmfa")
	configurationFileName = "configuration"
	configurationFilePath = configurationFileDir + "/" + configurationFileName
)

var (
	showTipsSetting = false
)

func initDefault(configurationFilePath string) {
	configuration, err := ini.Load(configurationFilePath)
	if err != nil {
		showTipsSetting = true
		return
	}

	// Overwrite default values.
	credentialsFilePath = os.ExpandEnv(configuration.Section("filepath").Key("credentials_file_path").String())
	configFilePath = os.ExpandEnv(configuration.Section("filepath").Key("config_file_path").String())
	beforeMFASuffix = configuration.Section("default-value").Key("suffix_of_before_mfa_profile").String()
	defaultMode = configuration.Section("default-value").Key("mode").String()
	defaultProfile = configuration.Section("default-value").Key("profile").String()
	defaultMFASerial = configuration.Section("default-value").Key("mfa_serial").String()
	defaultEndpointRegion = configuration.Section("default-value").Key("endpoint_region").String()
	if value, err := configuration.Section("default-value").Key("duration_seconds_get_session_token").Int(); err == nil {
		defaultDurationSecondsGetSessionToken = int32(value)
	}
	if value, err := configuration.Section("default-value").Key("duration_seconds_assume_role").Int(); err == nil {
		defaultDurationSecondsAssumeRole = int32(value)
	}
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
		initDefault(configurationFilePath)
	})

	cmd.PersistentFlags().StringVarP(&cliMode, "mode", "m", "", "The action mode of awsmfa, get-session-token or assume-role. The default value is get-session-token. If you specify the awsmfa_role_arn in shared credentials/config file or --role-arn option, awsmfa automatically turns the mode to assume-role.")
	cmd.PersistentFlags().StringVarP(&cliProfile, "profile", "p", "", "The profile used to get the token. You should set 'xxxx' if you have set 'xxxx-before-mfa' in the shared credentials/config file (.aws/credentials and .aws/config). The default value is 'default'")
	cmd.PersistentFlags().Int32VarP(&cliDurationSeconds, "duration-seconds", "d", 0, "The duration of the temporary security credential. Minimun value: 900 seconds (15 minutes). Max value is different depend on the authentification mode. If you try to get token of same account (with GetSessionToken), Max value is 129600 seconds (36h). In the case of assume role (with AssumeRole), Max value is 43200 seconds (12h). The default value is GetSessionToken=43200 seconds (12h), AssumeRole=3600 seconds (1h).")
	cmd.PersistentFlags().StringVar(&cliMfaSerial, "serial-number", "", "The serial number of the MFA device. The value is either an ARN of a virtual device (arn:aws:iam::123456789012:mfa/user) or the serial number of real device.")
	cmd.PersistentFlags().StringVarP(&cliEndpointRegion, "endpoint-region", "e", "", "The sts endpoint where awsmfa accesses to get a temporary credential. Such as ap-northeast-1, us-east-1.")
	cmd.PersistentFlags().StringVarP(&cliRoleArn, "role-arn", "r", "", "The ARN of the IAM role to assume. If you specify this option, awsmfa automatically turns the mode (--mode, -m) to assume-role.")
	cmd.PersistentFlags().StringVar(&cliRoleSessionName, "role-session-name", "", "The session name which will be logged to the AWS CloudTrail. The default value is awsmfa-session.")
	cmd.PersistentFlags().BoolVarP(&cliForce, "force", "f", false, "Force reflesh temporary credentials.")

	cmd.PersistentFlags().StringVar(&cliGenerateCredentialsSkeleton, "generate-credentials-skeleton", "", "Generate skeleton of shared credentials file (by default, ${HOME}/.aws/credentials) for specified action mode, get-session-token or assume-role.")
	cmd.PersistentFlags().StringVar(&cliGenerateConfigSkeleton, "generate-config-skeleton", "", "Generate skeleton of shared config file (by default, ${HOME}/.aws/config). for specified action mode, get-session-token or assume-role.")
	cmd.PersistentFlags().BoolVar(&cliGenerateConfigurationFile, "generate-configuration-file", false, fmt.Sprintf("Generate awsmfa's configuration file at %v", configurationFilePath))

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
		if err := generateConfigurationFile(configurationFileDir, configurationFileName); err != nil {
			return fmt.Errorf("failed to initialize awsmfa's configuration file: %w", err)
		}
		printCyan(fmt.Sprintf("Successfully create awsmfa's configuration file at %v\n", configurationFilePath))
		return nil
	}

	// Show tips.
	// To avoid including tips in the output of skeletons, all tips will be printed here.
	if showTipsSetting {
		printBlue(fmt.Sprintf("[Tips] There isn't an awsmfa's setting file. You can set some default values to place the setting file at: %v. If you would like to make it by cli, please use 'awsmfa --generate-configuration-file true'\n", configurationFilePath))
	}

	// Load credentials and config files.
	cred, err := ini.Load(credentialsFilePath)
	if err != nil {
		return fmt.Errorf("failed to load credentials file: %w", err)
	}
	cfg, err := ini.Load(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	// Set target profile.
	profile := setProfile(cliProfile, defaultProfile)

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
	mode, err := setMode(cliMode, defaultMode, profile+beforeMFASuffix, cred, cfg)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	switch mode {
	case "get-session-token":
		if err := handleGetSessionToken(profile, cred, cfg); err != nil {
			return fmt.Errorf("failed to get-session-token: %w", err)
		}
	case "assume-role":
		if err := handleAssumeRole(profile, cred, cfg); err != nil {
			return fmt.Errorf("failed to assume-role: %w", err)
		}
	default:
		return fmt.Errorf("invalid action mode: %v", mode)
	}

	return nil
}

func handleGetSessionToken(profile string, cred *ini.File, cfg *ini.File) error {
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
	durationSeconds := setDurationSeconds(cliDurationSeconds, defaultDurationSecondsGetSessionToken, profile+beforeMFASuffix, cred, cfg)
	mfaSerial, err := setMFASerial(cliMfaSerial, defaultMFASerial, profile+beforeMFASuffix, cred, cfg)
	if err != nil {
		return fmt.Errorf("The mfa_serial is not specified. You can set it in %v, %v, %v or --serial-number", credentialsFilePath, configFilePath, configurationFilePath)
	}
	endpointRegion := setEndpointRegion(cliEndpointRegion, defaultEndpointRegion, profile, cred, cfg)

	// Show request params.
	h, m, s := secToHMS(durationSeconds)
	data := [][]string{
		{"Profile to exec MFA", profile + beforeMFASuffix},
		// {"Credentials", fmt.Sprintf("[Only for DEBUG] %+v", c.Credentials)},
		{"Duration of token", fmt.Sprintf("%v sec (%vh %vm %vs)", durationSeconds, h, m, s)},
		{"MFA device's serial", mfaSerial},
		{"Region", endpointRegion},
		{"API Type", "AWS STS GetSessionToken"},
	}
	fmt.Printf("Try to get temporary token with following params ...\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Parameter", "Value"})
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

func handleAssumeRole(profile string, cred *ini.File, cfg *ini.File) error {
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
	durationSeconds := setDurationSeconds(cliDurationSeconds, defaultDurationSecondsAssumeRole, profile+beforeMFASuffix, cred, cfg)
	mfaSerial, err := setMFASerial(cliMfaSerial, defaultMFASerial, profile+beforeMFASuffix, cred, cfg)
	if err != nil {
		return fmt.Errorf("The mfa_serial is not specified. You can set it in %v, %v, %v or --serial-number", credentialsFilePath, configFilePath, configurationFilePath)
	}
	endpointRegion := setEndpointRegion(cliEndpointRegion, defaultEndpointRegion, profile, cred, cfg)
	roleArn, err := setRoleArn(cliRoleArn, profile+beforeMFASuffix, cred, cfg)
	if err != nil {
		return fmt.Errorf("The role_arn is not specified. You can set it in %v, %v or --role-arn", credentialsFilePath, configFilePath)
	}
	roleSessionName := setRoleSessionName(cliRoleSessionName, defaultRoleSessionName, profile+beforeMFASuffix, cred, cfg)

	// Show request params.
	h, m, s := secToHMS(durationSeconds)
	data := [][]string{
		{"Profile to exec MFA", profile + beforeMFASuffix},
		// {"Credentials", fmt.Sprintf("[Only for DEBUG] %+v", c.Credentials)},
		{"Role arn to assume", fmt.Sprintf("%v", roleArn)},
		{"Role session name", fmt.Sprintf("%v", roleSessionName)},
		{"Duration of token", fmt.Sprintf("%v sec (%vh %vm %vs)", durationSeconds, h, m, s)},
		{"MFA device's serial", mfaSerial},
		{"Region", endpointRegion},
		{"API Type", "AWS STS AssumeRole"},
	}
	fmt.Printf("Try to get temporary token with following params ...\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Parameter", "Value"})
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

// setMode returns action mode to be used.
// Priority
// 1. cli option: --mode
// 2. shared credentials or config file: if given profile has a role_arn param.
// 3. awsmfa configuration file: [default-value] profile (Need to overwrite build in default value in advance)
// 4. awsmfa build in default value
// If the mode is not whether 'get-session-token' or 'assume-role', awsmfa returns an error.
func setMode(cliOpt string, defaultValue string, profile string, cred *ini.File, cfg *ini.File) (string, error) {
	if cliOpt == "get-session-token" || cliOpt == "assume-role" {
		return cliOpt, nil
	} else if cliOpt != "" {
		return "wrong-mode", fmt.Errorf("Invalid action mode: action mode should be \"get-session-token\" or \"assume-role\"")
	}

	if cred.Section(profile).HasKey("awsmfa_role_arn") || cfg.Section("profile "+profile).HasKey("awsmfa_role_arn") {
		return "assume-role", nil
	}

	if defaultValue == "get-session-token" || defaultValue == "assume-role" {
		return defaultValue, nil
	}

	return "wrong-mode", fmt.Errorf("Invalid action mode: action mode should be \"get-session-token\" or \"assume-role\"")
}

// setProfile returns a profile to be used.
// Priority
// 1. cli option: --profile
// 2. environment variable: AWS_PROFILE
// 3. awsmfa configuration file: [default-value] profile (Need to overwrite build in default value in advance)
// 4. awsmfa build in default value
func setProfile(cliOpt string, defaultValue string) string {
	if cliOpt != "" {
		return cliOpt
	}
	if env, exists := os.LookupEnv("AWS_PROFILE"); exists == true {
		return env
	}
	return defaultValue
}

// setDurationSeconds returns duration seconds to be used.
// Priority
// 1. cli option: --duration-seconds
// 2. shared credentials file: ${HOME}/.aws/credentials (by default)
// 3. shared config file: ${HOME}/.aws/config (by default)
// 4. awsmfa configuration file: [default-value] duration_seconds (Need to overwrite build in default value in advance)
// 5. awsmfa build in default value
func setDurationSeconds(cliOpt int32, defaultValue int32, profile string, cred *ini.File, cfg *ini.File) int32 {
	if cliOpt != 0 {
		return cliOpt
	}
	if v, err := cred.Section(profile).Key("duration_seconds").Int(); err == nil {
		return int32(v)
	}
	if v, err := cfg.Section("profile " + profile).Key("duration_seconds").Int(); err == nil {
		return int32(v)
	}
	return defaultValue
}

// setMFASerial returns mfa device's serial number to be used.
// Priority
// 1. cli option: --serial-number
// 2. shared credentials file: ${HOME}/.aws/credentials (by default)
// 3. shared config file: ${HOME}/.aws/config (by default)
// 4. awsmfa configuration file: [default-value] duration_seconds (Need to overwrite build in default value in advance)
// If any serial number is not specified, setMFASerial returns error.
func setMFASerial(cliOpt string, defaultValue string, profile string, cred *ini.File, cfg *ini.File) (string, error) {
	if cliOpt != "" {
		return cliOpt, nil
	}
	if v := cred.Section(profile).Key("mfa_serial").String(); v != "" {
		return v, nil
	}
	if v := cfg.Section("profile " + profile).Key("mfa_serial").String(); v != "" {
		return v, nil
	}
	if defaultValue != "unspecified" {
		return defaultValue, nil
	}

	return "error", fmt.Errorf("no mfa_serial specified")
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
func setEndpointRegion(cliOpt string, defaultValue string, profile string, cred *ini.File, cfg *ini.File) string {
	if cliOpt != "" {
		return cliOpt
	}
	if env, exists := os.LookupEnv("AWS_REGION"); exists == true {
		return env
	}
	if env, exists := os.LookupEnv("AWS_DEFAULT_REGION"); exists == true {
		return env
	}
	if v := cred.Section(profile + beforeMFASuffix).Key("region").String(); v != "" {
		return v
	}
	if v := cfg.Section("profile " + profile + beforeMFASuffix).Key("region").String(); v != "" {
		return v
	}
	if v := cred.Section(profile).Key("region").String(); v != "" {
		return v
	}
	if v := cfg.Section("profile " + profile).Key("region").String(); v != "" {
		return v
	}
	return defaultValue
}

// setRoleArn returns a role arn related to given profile.
// [CAUTION!] This function parses only custom parameter 'awsmfa_role_arn',
// not aws build in parameter 'role_arn'.
// Priority
// 1. cli option: --role-arn
// 2. shared credentials file: ${HOME}/.aws/credentials (by default)
// 3. shared config file: ${HOME}/.aws/config (by default)
// If any arn is not specified, setRoleArn returns error.
func setRoleArn(cliOpt string, profile string, cred *ini.File, cfg *ini.File) (string, error) {
	if cliOpt != "" {
		return cliOpt, nil
	}
	if v := cred.Section(profile).Key("awsmfa_role_arn").String(); v != "" {
		return v, nil
	}
	if v := cfg.Section("profile " + profile).Key("awsmfa_role_arn").String(); v != "" {
		return v, nil
	}

	return "error", fmt.Errorf("no awsmfa_role_arn specified")
}

// setRoleSessionName returns a role session name to be used.
// Priority
// 1. cli option: --role-session-name
// 2. shared credentials file: ${HOME}/.aws/credentials (by default)
// 3. shared config file: ${HOME}/.aws/config (by default)
// 4. awsmfa configuration file: [default-value] role_session_name (Need to overwrite build in default value in advance)
// 5. awsmfa build in default value
func setRoleSessionName(cliOpt string, defaultValue string, profile string, cred *ini.File, cfg *ini.File) string {
	if cliOpt != "" {
		return cliOpt
	}
	if v := cred.Section(profile).Key("role_session_name").String(); v != "" {
		return v
	}
	if v := cfg.Section("profile " + profile).Key("role_session_name").String(); v != "" {
		return v
	}
	return defaultValue
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

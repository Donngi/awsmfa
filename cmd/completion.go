package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// NewCmdCompletion returns the completion command
func NewCmdCompletion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

	$ source <(awsmfa completion bash)

	# To load completions for each session, execute once:
	# Linux:
	$ awsmfa completion bash > /etc/bash_completion.d/awsmfa
	# macOS:
	$ awsmfa completion bash > /usr/local/etc/bash_completion.d/awsmfa

Zsh:

	# If shell completion is not already enabled in your environment,
	# you will need to enable it.  You can execute the following once:

	$ echo "autoload -U compinit; compinit" >> ~/.zshrc

	# To load completions for each session, execute once:
	$ awsmfa completion zsh > "${fpath[1]}/_awsmfa"

	# You will need to start a new shell for this setup to take effect.

fish:

	$ awsmfa completion fish | source

	# To load completions for each session, execute once:
	$ awsmfa completion fish > ~/.config/fish/completions/awsmfa.fish

PowerShell:

	PS> awsmfa completion powershell | Out-String | Invoke-Expression

	# To load completions for every new session, run:
	PS> awsmfa completion powershell > awsmfa.ps1
	# and source this file from your PowerShell profile.
`,
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}

	return cmd
}

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `This command is used to generate shell completion scripts.

For Bash:

	$ source <(albatross completion bash)

To load completions for each session, execute once:

	# Linux:
	$ albatross completion bash > /etc/bash_completion.d/albatross
	
	# MacOS:
	$ albatross completion bash > /usr/local/etc/bash_completion.d/albatross

For Zsh:

	# If shell completion is not already enabled in your environment you will need
	# to enable it.  You can execute the following once:
	$ echo "autoload -U compinit; compinit" >> ~/.zshrc

	# To load completions for each session, execute once:
	$ albatross completion zsh > "${fpath[1]}/_albatross"

	# You will need to start a new shell for this setup to take effect.

For Fish:

	$ albatross completion fish | source

	# To load completions for each session, execute once:
	$ albatross completion fish > ~/.config/fish/completions/albatross.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

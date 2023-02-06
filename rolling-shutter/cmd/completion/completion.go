package completion

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  $ source <(rolling-shutter completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ rolling-shutter completion bash > /etc/bash_completion.d/rolling-shutter
  # macOS:
  $ rolling-shutter completion bash > /usr/local/etc/bash_completion.d/rolling-shutter

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ rolling-shutter completion zsh > "${fpath[1]}/_rolling-shutter"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ rolling-shutter completion fish | source

  # To load completions for each session, execute once:
  $ rolling-shutter completion fish > ~/.config/fish/completions/rolling-shutter.fish

PowerShell:

  PS> rolling-shutter completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> rolling-shutter completion powershell > rolling-shutter.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return errors.Errorf("illegal argument: %s", args[0]) // should never happen
		},
	}
	return cmd
}

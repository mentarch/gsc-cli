package cmd

import (
	"fmt"
	"os"

	"github.com/sivori/gsc-cli/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	siteURL    string
	jsonOutput bool
	noColor    bool

	// Version info (set at build time)
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gsc",
		Short: "Google Search Console CLI",
		Long: `gsc-cli - Query Google Search Console from the terminal.

Pull top queries, compare date ranges, export CSVs, and flag ranking drops.

To get started:
  1. Create a Google Cloud project and enable the Search Console API
  2. Create OAuth2 Desktop credentials and download client_secret.json
  3. Run: gsc auth login`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip config init for auth and version commands
			if cmd.Name() == "login" || cmd.Name() == "version" || cmd.Name() == "completion" {
				return nil
			}
			if cmd.Parent() != nil && cmd.Parent().Name() == "auth" {
				return nil
			}

			if err := config.Init(); err != nil {
				return fmt.Errorf("could not initialize config: %w", err)
			}

			// Apply global flags
			if noColor {
				color.NoColor = true
			}

			// Override site URL if provided
			if siteURL == "" {
				siteURL = config.GetSiteURL()
			}

			return nil
		},
	}

	// Global flags
	cmd.PersistentFlags().StringVarP(&siteURL, "site", "s", "", "Search Console site URL (e.g., sc-domain:example.com)")
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Add commands
	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newSitesCmd())
	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newQueriesCmd())
	cmd.AddCommand(newCompareCmd())
	cmd.AddCommand(newDropsCmd())
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newCompletionCmd())

	return cmd
}

// Execute runs the root command
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("Error: %v", err))
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gsc-cli %s\n", Version)
			fmt.Printf("  commit: %s\n", Commit)
			fmt.Printf("  built:  %s\n", Date)
		},
	}
}

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for gsc-cli.

To load completions:

Bash:
  $ source <(gsc completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ gsc completion bash > /etc/bash_completion.d/gsc
  # macOS:
  $ gsc completion bash > /usr/local/etc/bash_completion.d/gsc

Zsh:
  $ source <(gsc completion zsh)
  # To load completions for each session, execute once:
  $ gsc completion zsh > "${fpath[1]}/_gsc"

Fish:
  $ gsc completion fish | source
  # To load completions for each session, execute once:
  $ gsc completion fish > ~/.config/fish/completions/gsc.fish

PowerShell:
  PS> gsc completion powershell | Out-String | Invoke-Expression
`,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		DisableFlagsInUseLine: true,
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
			return nil
		},
	}
}

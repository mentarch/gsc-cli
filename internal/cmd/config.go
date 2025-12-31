package cmd

import (
	"fmt"

	"github.com/sivori/gsc-cli/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Commands for managing gsc-cli configuration.",
	}

	cmd.AddCommand(newConfigSetSiteCmd())
	cmd.AddCommand(newConfigShowCmd())

	return cmd
}

func newConfigSetSiteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-site <site-url>",
		Short: "Set the default Search Console site",
		Long: `Set the default site URL for Search Console queries.

Examples:
  gsc config set-site sc-domain:example.com
  gsc config set-site https://example.com/`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Init(); err != nil {
				return fmt.Errorf("could not initialize config: %w", err)
			}

			siteURL := args[0]

			if err := config.SetSiteURL(siteURL); err != nil {
				return fmt.Errorf("could not save site URL: %w", err)
			}

			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("%s Default site set to: %s\n", green("✓"), siteURL)

			return nil
		},
	}
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Init(); err != nil {
				return fmt.Errorf("could not initialize config: %w", err)
			}

			fmt.Println("Current configuration:")
			fmt.Printf("  Site URL:           %s\n", config.GetSiteURL())
			fmt.Printf("  Client secret path: %s\n", config.GetClientSecretPath())

			return nil
		},
	}
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gsc-cli/internal/auth"
	"gsc-cli/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Commands for managing Google OAuth2 authentication.",
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthStatusCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var clientSecretPath string
	var site string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Google Search Console",
		Long: `Authenticate with Google Search Console using OAuth2.

You need to:
1. Create a Google Cloud project
2. Enable the Search Console API
3. Create OAuth2 Desktop credentials
4. Download the client_secret.json file

Then run: gsc auth login --client-secret /path/to/client_secret.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize config
			if err := config.Init(); err != nil {
				return fmt.Errorf("could not initialize config: %w", err)
			}

			// Get client secret path
			if clientSecretPath == "" {
				clientSecretPath = config.GetClientSecretPath()
			}

			if clientSecretPath == "" {
				fmt.Print("Path to client_secret.json: ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("could not read input: %w", err)
				}
				clientSecretPath = strings.TrimSpace(input)
			}

			// Verify file exists
			if _, err := os.Stat(clientSecretPath); os.IsNotExist(err) {
				return fmt.Errorf("client secret file not found: %s", clientSecretPath)
			}

			// Get site URL
			if site == "" {
				site = config.GetSiteURL()
			}

			if site == "" {
				fmt.Println()
				fmt.Println("Enter your Search Console site URL.")
				fmt.Println("Examples:")
				fmt.Println("  - sc-domain:example.com (domain property)")
				fmt.Println("  - https://example.com/ (URL prefix property)")
				fmt.Print("Site URL: ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("could not read input: %w", err)
				}
				site = strings.TrimSpace(input)
			}

			fmt.Println()

			// Perform OAuth login flow
			token, err := auth.LoginFlow(clientSecretPath)
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			// Store token
			if err := auth.SetToken(token); err != nil {
				return fmt.Errorf("could not save token: %w", err)
			}

			// Save config
			if err := config.SetClientSecretPath(clientSecretPath); err != nil {
				return fmt.Errorf("could not save client secret path: %w", err)
			}

			if err := config.SetSiteURL(site); err != nil {
				return fmt.Errorf("could not save site URL: %w", err)
			}

			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("%s Successfully authenticated!\n", green("✓"))
			fmt.Printf("  Site: %s\n", site)

			return nil
		},
	}

	cmd.Flags().StringVar(&clientSecretPath, "client-secret", "", "Path to client_secret.json")
	cmd.Flags().StringVar(&site, "site", "", "Search Console site URL")

	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.DeleteToken(); err != nil {
				return fmt.Errorf("could not delete token: %w", err)
			}

			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("%s Logged out successfully\n", green("✓"))

			return nil
		},
	}
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Init(); err != nil {
				return fmt.Errorf("could not initialize config: %w", err)
			}

			info, err := auth.GetTokenInfo()
			if err != nil {
				return fmt.Errorf("could not get token info: %w", err)
			}

			if !info.HasToken {
				yellow := color.New(color.FgYellow).SprintFunc()
				fmt.Printf("%s Not logged in\n", yellow("!"))
				fmt.Println("Run 'gsc auth login' to authenticate")
				return nil
			}

			green := color.New(color.FgGreen).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()

			fmt.Println("Authentication Status:")
			fmt.Printf("  Logged in: %s\n", green("Yes"))

			if info.IsExpired {
				fmt.Printf("  Token:     %s (will refresh on next use)\n", red("Expired"))
			} else {
				fmt.Printf("  Token:     %s\n", green("Valid"))
				fmt.Printf("  Expires:   %s\n", info.Expiry.Local().Format("2006-01-02 15:04:05"))
			}

			siteURL := config.GetSiteURL()
			if siteURL != "" {
				fmt.Printf("  Site:      %s\n", siteURL)
			}

			return nil
		},
	}
}

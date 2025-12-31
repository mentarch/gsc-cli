package cmd

import (
	"fmt"

	"gsc-cli/internal/api"
	"gsc-cli/internal/config"
	"gsc-cli/internal/output"

	"github.com/spf13/cobra"
)

func newSitesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sites",
		Short: "List available Search Console sites",
		Long:  "List all sites you have access to in Google Search Console.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// We need a dummy site to create the client, but ListSites doesn't use it
			client, err := api.NewClientForSites()
			if err != nil {
				return err
			}

			sites, err := client.ListSites()
			if err != nil {
				return fmt.Errorf("could not list sites: %w", err)
			}

			if len(sites) == 0 {
				fmt.Println("No sites found. Add a site at https://search.google.com/search-console")
				return nil
			}

			currentSite := config.GetSiteURL()

			fmt.Println("Available sites:")
			fmt.Println()

			table := output.NewTable()
			table.SetHeaders("SITE URL", "PERMISSION", "")

			for _, site := range sites {
				marker := ""
				if site.SiteUrl == currentSite {
					marker = output.Green("(current)")
				}
				table.Append([]string{
					site.SiteUrl,
					site.PermissionLevel,
					marker,
				})
			}

			table.Render()

			fmt.Println()
			fmt.Println("To change your default site:")
			fmt.Println("  gsc config set-site <site-url>")

			return nil
		},
	}
}

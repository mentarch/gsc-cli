package cmd

import (
	"fmt"
	"sort"

	"github.com/sivori/gsc-cli/internal/api"
	"github.com/sivori/gsc-cli/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newDropsCmd() *cobra.Command {
	var (
		threshold  float64
		minClicks  float64
		days       int
		limit      int
		csvFile    string
	)

	cmd := &cobra.Command{
		Use:   "drops",
		Short: "Find queries with ranking drops",
		Long: `Identify queries that have dropped in ranking position.

Examples:
  gsc drops                         # Drops > 5 positions, last 7 days vs prior 7
  gsc drops --threshold 3           # Drops > 3 positions
  gsc drops --min-clicks 10         # Only queries with 10+ clicks
  gsc drops --days 14               # Compare 14-day periods
  gsc drops --csv drops.csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if siteURL == "" {
				return fmt.Errorf("no site configured - run 'gsc auth login' or use --site")
			}

			client, err := api.NewClient(siteURL)
			if err != nil {
				return err
			}

			// Calculate date ranges
			p := api.GetComparisonPeriod("week")
			if days > 7 {
				// Custom period based on days
				currentStart, currentEnd := api.DateRangeForDays(days)
				prevStart, prevEnd := api.DateRangeForDays(days * 2)
				p = api.ComparisonPeriod{
					CurrentStart:  currentStart,
					CurrentEnd:    currentEnd,
					PreviousStart: prevStart,
					PreviousEnd:   prevEnd,
				}
			}

			// Query current period
			currentResult, err := client.Query(api.QueryRequest{
				StartDate:  p.CurrentStart,
				EndDate:    p.CurrentEnd,
				Dimensions: []string{"query"},
				RowLimit:   25000,
			})
			if err != nil {
				return fmt.Errorf("could not query current period: %w", err)
			}

			// Query previous period
			prevResult, err := client.Query(api.QueryRequest{
				StartDate:  p.PreviousStart,
				EndDate:    p.PreviousEnd,
				Dimensions: []string{"query"},
				RowLimit:   25000,
			})
			if err != nil {
				return fmt.Errorf("could not query previous period: %w", err)
			}

			// Find drops
			drops := findDrops(currentResult.Rows, prevResult.Rows, threshold, minClicks)

			// Sort by drop magnitude
			sort.Slice(drops, func(i, j int) bool {
				return drops[i].PositionDrop > drops[j].PositionDrop
			})

			// Limit results
			if len(drops) > limit {
				drops = drops[:limit]
			}

			// Output
			if csvFile != "" {
				if err := output.WriteDropsCSV(csvFile, drops); err != nil {
					return err
				}
				green := color.New(color.FgGreen).SprintFunc()
				fmt.Printf("%s Exported %d drops to %s\n", green("✓"), len(drops), csvFile)
				return nil
			}

			if jsonOutput {
				return output.PrintDropsJSON(threshold, drops)
			}

			// Print header
			fmt.Printf("Ranking drops for %s\n", output.Cyan(siteURL))
			fmt.Printf("Current:  %s to %s\n", p.CurrentStart, p.CurrentEnd)
			fmt.Printf("Previous: %s to %s\n", p.PreviousStart, p.PreviousEnd)
			fmt.Printf("Threshold: >%.1f positions\n", threshold)
			if minClicks > 0 {
				fmt.Printf("Min clicks: %.0f\n", minClicks)
			}
			fmt.Println()

			if len(drops) == 0 {
				green := color.New(color.FgGreen).SprintFunc()
				fmt.Printf("%s No significant ranking drops found!\n", green("✓"))
				return nil
			}

			red := color.New(color.FgRed).SprintFunc()
			fmt.Printf("%s Found %d queries with ranking drops\n\n", red("!"), len(drops))

			// Print table
			table := output.NewTable()
			table.SetHeaders("QUERY", "DROP", "NOW", "WAS", "CLICKS", "IMPR")

			for _, row := range drops {
				table.Append([]string{
					output.TruncateString(row.Query, 40),
					output.Red(fmt.Sprintf("+%.1f", row.PositionDrop)),
					output.FormatPosition(row.CurrentPosition),
					output.FormatPosition(row.PreviousPosition),
					output.FormatNumber(row.CurrentClicks),
					output.FormatNumber(row.CurrentImpressions),
				})
			}

			table.Render()
			return nil
		},
	}

	cmd.Flags().Float64Var(&threshold, "threshold", 5, "Minimum position drop to flag")
	cmd.Flags().Float64Var(&minClicks, "min-clicks", 0, "Minimum clicks in previous period")
	cmd.Flags().IntVar(&days, "days", 7, "Number of days per period")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of results")
	cmd.Flags().StringVar(&csvFile, "csv", "", "Export to CSV file")

	return cmd
}

func findDrops(current, previous []api.QueryRow, threshold, minClicks float64) []output.DropsRow {
	// Build lookup maps
	currentMap := make(map[string]api.QueryRow)
	for _, row := range current {
		currentMap[row.Query] = row
	}

	prevMap := make(map[string]api.QueryRow)
	for _, row := range previous {
		prevMap[row.Query] = row
	}

	var drops []output.DropsRow

	// Find queries that exist in both periods and have dropped
	for query, prev := range prevMap {
		// Skip if below minimum clicks
		if minClicks > 0 && prev.Clicks < minClicks {
			continue
		}

		curr, exists := currentMap[query]
		if !exists {
			// Query disappeared - could be a complete drop
			// Only include if it had significant previous position
			if prev.Position <= 20 && prev.Impressions >= 100 {
				drops = append(drops, output.DropsRow{
					Query:            query,
					CurrentPosition:  100, // Treat as dropped out
					PreviousPosition: prev.Position,
					PositionDrop:     100 - prev.Position,
					CurrentClicks:    0,
					PreviousClicks:   prev.Clicks,
					CurrentImpressions: 0,
				})
			}
			continue
		}

		// Calculate position drop (higher position number = worse ranking)
		drop := curr.Position - prev.Position

		if drop >= threshold {
			drops = append(drops, output.DropsRow{
				Query:            query,
				CurrentPosition:  curr.Position,
				PreviousPosition: prev.Position,
				PositionDrop:     drop,
				CurrentClicks:    curr.Clicks,
				PreviousClicks:   prev.Clicks,
				CurrentImpressions: curr.Impressions,
			})
		}
	}

	return drops
}

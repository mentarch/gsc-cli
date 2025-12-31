package cmd

import (
	"fmt"
	"sort"

	"gsc-cli/internal/api"
	"gsc-cli/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newCompareCmd() *cobra.Command {
	var (
		period      string
		fromStart   string
		fromEnd     string
		toStart     string
		toEnd       string
		limit       int
		csvFile     string
		sortBy      string
	)

	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare metrics across date ranges",
		Long: `Compare search metrics between two date ranges.

Examples:
  gsc compare --period week         # This week vs last week
  gsc compare --period month        # This month vs last month
  gsc compare --from-start 2025-01-01 --from-end 2025-01-15 \
              --to-start 2024-12-15 --to-end 2024-12-31
  gsc compare --csv comparison.csv
  gsc compare --sort clicks         # Sort by clicks delta`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if siteURL == "" {
				return fmt.Errorf("no site configured - run 'gsc auth login' or use --site")
			}

			client, err := api.NewClient(siteURL)
			if err != nil {
				return err
			}

			// Determine date ranges
			var currentStart, currentEnd, prevStart, prevEnd string
			if fromStart != "" && fromEnd != "" && toStart != "" && toEnd != "" {
				currentStart, currentEnd = fromStart, fromEnd
				prevStart, prevEnd = toStart, toEnd
			} else {
				p := api.GetComparisonPeriod(period)
				currentStart, currentEnd = p.CurrentStart, p.CurrentEnd
				prevStart, prevEnd = p.PreviousStart, p.PreviousEnd
			}

			// Query current period
			currentResult, err := client.Query(api.QueryRequest{
				StartDate:  currentStart,
				EndDate:    currentEnd,
				Dimensions: []string{"query"},
				RowLimit:   int64(limit * 2), // Fetch more to account for new queries
			})
			if err != nil {
				return fmt.Errorf("could not query current period: %w", err)
			}

			// Query previous period
			prevResult, err := client.Query(api.QueryRequest{
				StartDate:  prevStart,
				EndDate:    prevEnd,
				Dimensions: []string{"query"},
				RowLimit:   int64(limit * 2),
			})
			if err != nil {
				return fmt.Errorf("could not query previous period: %w", err)
			}

			// Build comparison
			rows := buildComparison(currentResult.Rows, prevResult.Rows)

			// Sort
			sortComparison(rows, sortBy)

			// Limit results
			if len(rows) > limit {
				rows = rows[:limit]
			}

			// Output
			if csvFile != "" {
				if err := output.WriteComparisonCSV(csvFile, rows); err != nil {
					return err
				}
				green := color.New(color.FgGreen).SprintFunc()
				fmt.Printf("%s Exported %d rows to %s\n", green("✓"), len(rows), csvFile)
				return nil
			}

			if jsonOutput {
				return output.PrintComparisonJSON(
					output.Period{Start: currentStart, End: currentEnd},
					output.Period{Start: prevStart, End: prevEnd},
					rows,
				)
			}

			// Print header
			fmt.Printf("Comparison for %s\n", output.Cyan(siteURL))
			fmt.Printf("Current:  %s to %s\n", currentStart, currentEnd)
			fmt.Printf("Previous: %s to %s\n", prevStart, prevEnd)
			fmt.Println()

			if len(rows) == 0 {
				fmt.Println("No data found for comparison.")
				return nil
			}

			// Print table
			table := output.NewTable()
			table.SetHeaders("QUERY", "CLICKS", "Δ", "IMPR", "Δ", "POS", "Δ")

			for _, row := range rows {
				table.Append([]string{
					output.TruncateString(row.Query, 40),
					output.FormatNumber(row.CurrentClicks),
					output.FormatDelta(row.ClicksDelta, true),
					output.FormatNumber(row.CurrentImpressions),
					output.FormatDelta(row.ImpressionsDelta, true),
					output.FormatPosition(row.CurrentPosition),
					output.FormatDelta(row.PositionDelta, false),
				})
			}

			table.Render()
			return nil
		},
	}

	cmd.Flags().StringVar(&period, "period", "week", "Comparison period (week, month)")
	cmd.Flags().StringVar(&fromStart, "from-start", "", "Current period start (YYYY-MM-DD)")
	cmd.Flags().StringVar(&fromEnd, "from-end", "", "Current period end (YYYY-MM-DD)")
	cmd.Flags().StringVar(&toStart, "to-start", "", "Previous period start (YYYY-MM-DD)")
	cmd.Flags().StringVar(&toEnd, "to-end", "", "Previous period end (YYYY-MM-DD)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of results")
	cmd.Flags().StringVar(&csvFile, "csv", "", "Export to CSV file")
	cmd.Flags().StringVar(&sortBy, "sort", "clicks", "Sort by (clicks, impressions, position)")

	return cmd
}

func buildComparison(current, previous []api.QueryRow) []output.ComparisonRow {
	// Build lookup maps
	currentMap := make(map[string]api.QueryRow)
	for _, row := range current {
		currentMap[row.Query] = row
	}

	prevMap := make(map[string]api.QueryRow)
	for _, row := range previous {
		prevMap[row.Query] = row
	}

	// Collect all unique queries
	queries := make(map[string]bool)
	for q := range currentMap {
		queries[q] = true
	}
	for q := range prevMap {
		queries[q] = true
	}

	// Build comparison rows
	var rows []output.ComparisonRow
	for query := range queries {
		curr := currentMap[query]
		prev := prevMap[query]

		row := output.ComparisonRow{
			Query:              query,
			CurrentClicks:      curr.Clicks,
			PreviousClicks:     prev.Clicks,
			ClicksDelta:        curr.Clicks - prev.Clicks,
			CurrentImpressions: curr.Impressions,
			PreviousImpressions: prev.Impressions,
			ImpressionsDelta:   curr.Impressions - prev.Impressions,
			CurrentPosition:    curr.Position,
			PreviousPosition:   prev.Position,
			PositionDelta:      curr.Position - prev.Position,
		}

		// Calculate percentage changes
		if prev.Clicks > 0 {
			row.ClicksPercent = ((curr.Clicks - prev.Clicks) / prev.Clicks) * 100
		}
		if prev.Impressions > 0 {
			row.ImpressionsPercent = ((curr.Impressions - prev.Impressions) / prev.Impressions) * 100
		}

		rows = append(rows, row)
	}

	return rows
}

func sortComparison(rows []output.ComparisonRow, sortBy string) {
	switch sortBy {
	case "impressions":
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].CurrentImpressions > rows[j].CurrentImpressions
		})
	case "position":
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].CurrentPosition < rows[j].CurrentPosition
		})
	default: // clicks
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].CurrentClicks > rows[j].CurrentClicks
		})
	}
}

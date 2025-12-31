package cmd

import (
	"fmt"
	"strings"

	"gsc-cli/internal/api"
	"gsc-cli/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newQueriesCmd() *cobra.Command {
	var (
		days      int
		startDate string
		endDate   string
		limit     int
		filter    string
		csvFile   string
		dimension string
	)

	cmd := &cobra.Command{
		Use:   "queries",
		Short: "Fetch top search queries",
		Long: `Fetch top search queries from Google Search Console.

Examples:
  gsc queries                       # Last 28 days, top 100
  gsc queries --days 7              # Last 7 days
  gsc queries --start 2025-01-01 --end 2025-01-15
  gsc queries --limit 500           # Top 500 queries
  gsc queries --filter "page:*/blog/*"
  gsc queries --csv output.csv      # Export to CSV
  gsc queries --json                # JSON output
  gsc queries --dimension page      # Group by page instead of query`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if siteURL == "" {
				return fmt.Errorf("no site configured - run 'gsc auth login' or use --site")
			}

			client, err := api.NewClient(siteURL)
			if err != nil {
				return err
			}

			// Determine date range
			var start, end string
			if startDate != "" && endDate != "" {
				start, end = startDate, endDate
			} else if days > 0 {
				start, end = api.DateRangeForDays(days)
			} else {
				start, end = api.DefaultDateRange()
			}

			// Build dimensions
			dimensions := []string{"query"}
			if dimension != "" {
				dimensions = []string{dimension}
			}

			// Build filters
			var filters []api.Filter
			if filter != "" {
				f, err := parseFilter(filter)
				if err != nil {
					return err
				}
				filters = append(filters, f)
			}

			// Execute query
			result, err := client.Query(api.QueryRequest{
				StartDate:  start,
				EndDate:    end,
				Dimensions: dimensions,
				RowLimit:   int64(limit),
				Filters:    filters,
			})
			if err != nil {
				return err
			}

			// Output
			if csvFile != "" {
				if err := output.WriteQueryResultCSV(csvFile, result, dimensions); err != nil {
					return err
				}
				green := color.New(color.FgGreen).SprintFunc()
				fmt.Printf("%s Exported %d rows to %s\n", green("✓"), len(result.Rows), csvFile)
				return nil
			}

			if jsonOutput {
				return output.PrintQueryResultJSON(result)
			}

			// Print header
			fmt.Printf("Search queries for %s\n", output.Cyan(siteURL))
			fmt.Printf("Date range: %s to %s\n", start, end)
			fmt.Printf("Total results: %d\n\n", result.TotalRows)

			if len(result.Rows) == 0 {
				fmt.Println("No data found for this period.")
				return nil
			}

			// Print table
			table := output.NewTable()

			// Set headers based on dimension
			switch dimensions[0] {
			case "page":
				table.SetHeaders("PAGE", "CLICKS", "IMPR", "CTR", "POS")
			default:
				table.SetHeaders("QUERY", "CLICKS", "IMPR", "CTR", "POS")
			}

			for _, row := range result.Rows {
				var label string
				switch dimensions[0] {
				case "page":
					label = output.TruncateString(row.Page, 60)
				case "country":
					label = row.Country
				case "device":
					label = row.Device
				default:
					label = output.TruncateString(row.Query, 50)
				}

				table.Append([]string{
					label,
					output.FormatNumber(row.Clicks),
					output.FormatNumber(row.Impressions),
					output.FormatCTR(row.CTR),
					output.FormatPosition(row.Position),
				})
			}

			table.Render()
			return nil
		},
	}

	cmd.Flags().IntVar(&days, "days", 0, "Number of days to query (default: 28)")
	cmd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Maximum number of results")
	cmd.Flags().StringVar(&filter, "filter", "", "Filter (e.g., page:*/blog/*, query:keyword)")
	cmd.Flags().StringVar(&csvFile, "csv", "", "Export to CSV file")
	cmd.Flags().StringVar(&dimension, "dimension", "", "Dimension to group by (query, page, country, device)")

	return cmd
}

// parseFilter parses a filter string like "page:*/blog/*" or "query:keyword"
func parseFilter(s string) (api.Filter, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return api.Filter{}, fmt.Errorf("invalid filter format: %s (expected dimension:expression)", s)
	}

	dimension := strings.ToLower(parts[0])
	expression := parts[1]

	// Validate dimension
	validDimensions := map[string]bool{
		"query": true, "page": true, "country": true, "device": true,
	}
	if !validDimensions[dimension] {
		return api.Filter{}, fmt.Errorf("invalid dimension: %s (valid: query, page, country, device)", dimension)
	}

	// Determine operator based on expression
	operator := "equals"
	if strings.Contains(expression, "*") {
		operator = "includingRegex"
		// Convert glob to regex
		expression = strings.ReplaceAll(expression, "*", ".*")
	}

	return api.Filter{
		Dimension:  dimension,
		Operator:   operator,
		Expression: expression,
	}, nil
}

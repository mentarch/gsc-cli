package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sivori/gsc-cli/internal/api"
	"github.com/sivori/gsc-cli/internal/output"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newPagesCmd() *cobra.Command {
	var (
		days      int
		startDate string
		endDate   string
		limit     int
		filter    string
		csvFile   string
		query     string
		fullURL   bool
	)

	cmd := &cobra.Command{
		Use:   "pages",
		Short: "Fetch top pages by performance",
		Long: `Fetch top pages from Google Search Console.

Shows which pages are driving traffic and their search performance metrics.

Examples:
  gsc pages                       # Last 28 days, top 100 pages
  gsc pages --days 7              # Last 7 days
  gsc pages --limit 50            # Top 50 pages
  gsc pages --query "keyword"     # Pages ranking for a specific query
  gsc pages --filter "page:*/blog/*"  # Filter to blog pages only
  gsc pages --full                # Show full URLs (not truncated)
  gsc pages --csv output.csv      # Export to CSV
  gsc pages --json                # JSON output`,
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

			// Build filters
			var filters []api.Filter
			if filter != "" {
				f, err := parseFilter(filter)
				if err != nil {
					return err
				}
				filters = append(filters, f)
			}

			// Add query filter if specified
			if query != "" {
				filters = append(filters, api.Filter{
					Dimension:  "query",
					Operator:   "contains",
					Expression: query,
				})
			}

			// Execute query with page dimension
			result, err := client.Query(api.QueryRequest{
				StartDate:  start,
				EndDate:    end,
				Dimensions: []string{"page"},
				RowLimit:   int64(limit),
				Filters:    filters,
			})
			if err != nil {
				return err
			}

			// Output
			if csvFile != "" {
				if err := output.WriteQueryResultCSV(csvFile, result, []string{"page"}); err != nil {
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
			fmt.Printf("Top pages for %s\n", output.Cyan(siteURL))
			fmt.Printf("Date range: %s to %s\n", start, end)
			if query != "" {
				fmt.Printf("Filtered by query: %s\n", output.Cyan(query))
			}
			fmt.Printf("Total results: %d\n\n", result.TotalRows)

			if len(result.Rows) == 0 {
				fmt.Println("No data found for this period.")
				return nil
			}

			// Print table
			table := output.NewTable()
			table.SetHeaders("PAGE", "CLICKS", "IMPR", "CTR", "POS")

			for _, row := range result.Rows {
				pageDisplay := row.Page
				if !fullURL {
					pageDisplay = formatPageURL(row.Page)
				}

				table.Append([]string{
					pageDisplay,
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
	cmd.Flags().StringVar(&filter, "filter", "", "Filter (e.g., page:*/blog/*)")
	cmd.Flags().StringVar(&query, "query", "", "Query to filter pages by (shows pages ranking for this query)")
	cmd.Flags().StringVar(&csvFile, "csv", "", "Export to CSV file")
	cmd.Flags().BoolVar(&fullURL, "full", false, "Show full URLs instead of paths")

	return cmd
}

// formatPageURL extracts and truncates the path from a full URL
func formatPageURL(fullURL string) string {
	parsed, err := url.Parse(fullURL)
	if err != nil {
		return output.TruncateString(fullURL, 60)
	}

	path := parsed.Path
	if path == "" {
		path = "/"
	}

	// Add query string if present
	if parsed.RawQuery != "" {
		path = path + "?" + parsed.RawQuery
	}

	// Truncate long paths
	if len(path) > 60 {
		return path[:57] + "..."
	}

	return path
}

// stripDomain removes the domain from a URL, keeping just the path
func stripDomain(pageURL string) string {
	if idx := strings.Index(pageURL, "://"); idx != -1 {
		rest := pageURL[idx+3:]
		if slashIdx := strings.Index(rest, "/"); slashIdx != -1 {
			return rest[slashIdx:]
		}
		return "/"
	}
	return pageURL
}

package api

import (
	"fmt"
	"time"

	"google.golang.org/api/searchconsole/v1"
)

// QueryRequest represents parameters for a Search Analytics query
type QueryRequest struct {
	StartDate  string
	EndDate    string
	Dimensions []string // query, page, country, device, date
	RowLimit   int64
	StartRow   int64
	Filters    []Filter
}

// Filter represents a dimension filter
type Filter struct {
	Dimension  string // query, page, country, device
	Operator   string // equals, contains, notContains, includingRegex, excludingRegex
	Expression string
}

// QueryRow represents a single row of query results
type QueryRow struct {
	Query       string
	Page        string
	Country     string
	Device      string
	Date        string
	Clicks      float64
	Impressions float64
	CTR         float64
	Position    float64
}

// QueryResult represents the result of a Search Analytics query
type QueryResult struct {
	Rows       []QueryRow
	TotalRows  int
	StartDate  string
	EndDate    string
}

// Query executes a Search Analytics query
func (c *Client) Query(req QueryRequest) (*QueryResult, error) {
	// Build the API request
	apiReq := &searchconsole.SearchAnalyticsQueryRequest{
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Dimensions: req.Dimensions,
		RowLimit:   req.RowLimit,
		StartRow:   req.StartRow,
	}

	// Set defaults
	if apiReq.RowLimit == 0 {
		apiReq.RowLimit = 1000
	}
	if len(apiReq.Dimensions) == 0 {
		apiReq.Dimensions = []string{"query"}
	}

	// Add filters
	if len(req.Filters) > 0 {
		var dimensionFilters []*searchconsole.ApiDimensionFilter
		for _, f := range req.Filters {
			dimensionFilters = append(dimensionFilters, &searchconsole.ApiDimensionFilter{
				Dimension:  f.Dimension,
				Operator:   f.Operator,
				Expression: f.Expression,
			})
		}
		apiReq.DimensionFilterGroups = []*searchconsole.ApiDimensionFilterGroup{
			{
				GroupType: "and",
				Filters:   dimensionFilters,
			},
		}
	}

	// Execute query
	resp, err := c.service.Searchanalytics.Query(c.siteURL, apiReq).Do()
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// Parse results
	result := &QueryResult{
		TotalRows: len(resp.Rows),
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	for _, row := range resp.Rows {
		qr := QueryRow{
			Clicks:      row.Clicks,
			Impressions: row.Impressions,
			CTR:         row.Ctr,
			Position:    row.Position,
		}

		// Map dimensions to fields
		for i, dim := range req.Dimensions {
			if i < len(row.Keys) {
				switch dim {
				case "query":
					qr.Query = row.Keys[i]
				case "page":
					qr.Page = row.Keys[i]
				case "country":
					qr.Country = row.Keys[i]
				case "device":
					qr.Device = row.Keys[i]
				case "date":
					qr.Date = row.Keys[i]
				}
			}
		}

		result.Rows = append(result.Rows, qr)
	}

	return result, nil
}

// QueryAll fetches all results with pagination
func (c *Client) QueryAll(req QueryRequest) (*QueryResult, error) {
	var allRows []QueryRow
	startRow := int64(0)
	batchSize := int64(25000) // Max allowed by API

	for {
		req.StartRow = startRow
		req.RowLimit = batchSize

		result, err := c.Query(req)
		if err != nil {
			return nil, err
		}

		allRows = append(allRows, result.Rows...)

		if len(result.Rows) < int(batchSize) {
			break
		}

		startRow += batchSize
	}

	return &QueryResult{
		Rows:      allRows,
		TotalRows: len(allRows),
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}, nil
}

// DefaultDateRange returns the default date range (last 28 days)
func DefaultDateRange() (start, end string) {
	now := time.Now()
	end = now.AddDate(0, 0, -3).Format("2006-01-02") // 3 days ago (data delay)
	start = now.AddDate(0, 0, -30).Format("2006-01-02") // 30 days ago
	return
}

// DateRangeForDays returns a date range for the last N days
func DateRangeForDays(days int) (start, end string) {
	now := time.Now()
	end = now.AddDate(0, 0, -3).Format("2006-01-02") // 3 days ago (data delay)
	start = now.AddDate(0, 0, -3-days).Format("2006-01-02")
	return
}

// ComparisonPeriod represents two date ranges for comparison
type ComparisonPeriod struct {
	CurrentStart string
	CurrentEnd   string
	PreviousStart string
	PreviousEnd  string
}

// GetComparisonPeriod returns date ranges for week-over-week or month-over-month comparison
func GetComparisonPeriod(period string) ComparisonPeriod {
	now := time.Now()
	dataDelay := 3 // GSC data is typically 3 days behind

	switch period {
	case "week":
		currentEnd := now.AddDate(0, 0, -dataDelay)
		currentStart := currentEnd.AddDate(0, 0, -6)
		previousEnd := currentStart.AddDate(0, 0, -1)
		previousStart := previousEnd.AddDate(0, 0, -6)

		return ComparisonPeriod{
			CurrentStart:  currentStart.Format("2006-01-02"),
			CurrentEnd:    currentEnd.Format("2006-01-02"),
			PreviousStart: previousStart.Format("2006-01-02"),
			PreviousEnd:   previousEnd.Format("2006-01-02"),
		}

	case "month":
		currentEnd := now.AddDate(0, 0, -dataDelay)
		currentStart := currentEnd.AddDate(0, 0, -29)
		previousEnd := currentStart.AddDate(0, 0, -1)
		previousStart := previousEnd.AddDate(0, 0, -29)

		return ComparisonPeriod{
			CurrentStart:  currentStart.Format("2006-01-02"),
			CurrentEnd:    currentEnd.Format("2006-01-02"),
			PreviousStart: previousStart.Format("2006-01-02"),
			PreviousEnd:   previousEnd.Format("2006-01-02"),
		}

	default:
		// Default to week
		return GetComparisonPeriod("week")
	}
}

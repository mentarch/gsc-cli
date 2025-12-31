package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sivori/gsc-cli/internal/api"
)

// JSONQueryResult represents query results in JSON format
type JSONQueryResult struct {
	StartDate string         `json:"start_date"`
	EndDate   string         `json:"end_date"`
	TotalRows int            `json:"total_rows"`
	Rows      []JSONQueryRow `json:"rows"`
}

// JSONQueryRow represents a single query row in JSON format
type JSONQueryRow struct {
	Query       string  `json:"query,omitempty"`
	Page        string  `json:"page,omitempty"`
	Country     string  `json:"country,omitempty"`
	Device      string  `json:"device,omitempty"`
	Date        string  `json:"date,omitempty"`
	Clicks      float64 `json:"clicks"`
	Impressions float64 `json:"impressions"`
	CTR         float64 `json:"ctr"`
	Position    float64 `json:"position"`
}

// PrintQueryResultJSON prints query results as JSON
func PrintQueryResultJSON(result *api.QueryResult) error {
	output := JSONQueryResult{
		StartDate: result.StartDate,
		EndDate:   result.EndDate,
		TotalRows: result.TotalRows,
		Rows:      make([]JSONQueryRow, len(result.Rows)),
	}

	for i, row := range result.Rows {
		output.Rows[i] = JSONQueryRow{
			Query:       row.Query,
			Page:        row.Page,
			Country:     row.Country,
			Device:      row.Device,
			Date:        row.Date,
			Clicks:      row.Clicks,
			Impressions: row.Impressions,
			CTR:         row.CTR,
			Position:    row.Position,
		}
	}

	return printJSON(output)
}

// JSONComparisonResult represents comparison results in JSON format
type JSONComparisonResult struct {
	CurrentPeriod  Period               `json:"current_period"`
	PreviousPeriod Period               `json:"previous_period"`
	Rows           []JSONComparisonRow  `json:"rows"`
}

// Period represents a date range
type Period struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// JSONComparisonRow represents a comparison row in JSON format
type JSONComparisonRow struct {
	Query              string  `json:"query"`
	CurrentClicks      float64 `json:"current_clicks"`
	PreviousClicks     float64 `json:"previous_clicks"`
	ClicksDelta        float64 `json:"clicks_delta"`
	ClicksPercent      float64 `json:"clicks_percent"`
	CurrentImpressions float64 `json:"current_impressions"`
	PreviousImpressions float64 `json:"previous_impressions"`
	ImpressionsDelta   float64 `json:"impressions_delta"`
	ImpressionsPercent float64 `json:"impressions_percent"`
	CurrentPosition    float64 `json:"current_position"`
	PreviousPosition   float64 `json:"previous_position"`
	PositionDelta      float64 `json:"position_delta"`
}

// PrintComparisonJSON prints comparison results as JSON
func PrintComparisonJSON(currentPeriod, previousPeriod Period, rows []ComparisonRow) error {
	output := JSONComparisonResult{
		CurrentPeriod:  currentPeriod,
		PreviousPeriod: previousPeriod,
		Rows:           make([]JSONComparisonRow, len(rows)),
	}

	for i, row := range rows {
		output.Rows[i] = JSONComparisonRow{
			Query:              row.Query,
			CurrentClicks:      row.CurrentClicks,
			PreviousClicks:     row.PreviousClicks,
			ClicksDelta:        row.ClicksDelta,
			ClicksPercent:      row.ClicksPercent,
			CurrentImpressions: row.CurrentImpressions,
			PreviousImpressions: row.PreviousImpressions,
			ImpressionsDelta:   row.ImpressionsDelta,
			ImpressionsPercent: row.ImpressionsPercent,
			CurrentPosition:    row.CurrentPosition,
			PreviousPosition:   row.PreviousPosition,
			PositionDelta:      row.PositionDelta,
		}
	}

	return printJSON(output)
}

// JSONDropsResult represents drops results in JSON format
type JSONDropsResult struct {
	Threshold float64        `json:"threshold"`
	Count     int            `json:"count"`
	Rows      []JSONDropsRow `json:"rows"`
}

// JSONDropsRow represents a drops row in JSON format
type JSONDropsRow struct {
	Query              string  `json:"query"`
	PositionDrop       float64 `json:"position_drop"`
	CurrentPosition    float64 `json:"current_position"`
	PreviousPosition   float64 `json:"previous_position"`
	CurrentClicks      float64 `json:"current_clicks"`
	PreviousClicks     float64 `json:"previous_clicks"`
	CurrentImpressions float64 `json:"current_impressions"`
}

// PrintDropsJSON prints drops results as JSON
func PrintDropsJSON(threshold float64, rows []DropsRow) error {
	output := JSONDropsResult{
		Threshold: threshold,
		Count:     len(rows),
		Rows:      make([]JSONDropsRow, len(rows)),
	}

	for i, row := range rows {
		output.Rows[i] = JSONDropsRow{
			Query:              row.Query,
			PositionDrop:       row.PositionDrop,
			CurrentPosition:    row.CurrentPosition,
			PreviousPosition:   row.PreviousPosition,
			CurrentClicks:      row.CurrentClicks,
			PreviousClicks:     row.PreviousClicks,
			CurrentImpressions: row.CurrentImpressions,
		}
	}

	return printJSON(output)
}

func printJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return fmt.Errorf("could not encode JSON: %w", err)
	}
	return nil
}

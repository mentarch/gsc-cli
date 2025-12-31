package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"gsc-cli/internal/api"
)

// WriteQueryResultCSV writes query results to a CSV file
func WriteQueryResultCSV(filename string, result *api.QueryResult, dimensions []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Build header based on dimensions
	header := []string{}
	for _, dim := range dimensions {
		switch dim {
		case "query":
			header = append(header, "Query")
		case "page":
			header = append(header, "Page")
		case "country":
			header = append(header, "Country")
		case "device":
			header = append(header, "Device")
		case "date":
			header = append(header, "Date")
		}
	}
	header = append(header, "Clicks", "Impressions", "CTR", "Position")

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("could not write header: %w", err)
	}

	// Write rows
	for _, row := range result.Rows {
		record := []string{}
		for _, dim := range dimensions {
			switch dim {
			case "query":
				record = append(record, row.Query)
			case "page":
				record = append(record, row.Page)
			case "country":
				record = append(record, row.Country)
			case "device":
				record = append(record, row.Device)
			case "date":
				record = append(record, row.Date)
			}
		}
		record = append(record,
			strconv.FormatFloat(row.Clicks, 'f', 0, 64),
			strconv.FormatFloat(row.Impressions, 'f', 0, 64),
			strconv.FormatFloat(row.CTR*100, 'f', 2, 64)+"%",
			strconv.FormatFloat(row.Position, 'f', 1, 64),
		)

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("could not write row: %w", err)
		}
	}

	return nil
}

// ComparisonRow represents a comparison between two periods
type ComparisonRow struct {
	Query             string
	CurrentClicks     float64
	PreviousClicks    float64
	ClicksDelta       float64
	ClicksPercent     float64
	CurrentImpressions float64
	PreviousImpressions float64
	ImpressionsDelta   float64
	ImpressionsPercent float64
	CurrentPosition   float64
	PreviousPosition  float64
	PositionDelta     float64
}

// WriteComparisonCSV writes comparison results to a CSV file
func WriteComparisonCSV(filename string, rows []ComparisonRow) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"Query",
		"Clicks (Current)", "Clicks (Previous)", "Clicks Delta", "Clicks %",
		"Impressions (Current)", "Impressions (Previous)", "Impressions Delta", "Impressions %",
		"Position (Current)", "Position (Previous)", "Position Delta",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("could not write header: %w", err)
	}

	for _, row := range rows {
		record := []string{
			row.Query,
			strconv.FormatFloat(row.CurrentClicks, 'f', 0, 64),
			strconv.FormatFloat(row.PreviousClicks, 'f', 0, 64),
			strconv.FormatFloat(row.ClicksDelta, 'f', 0, 64),
			strconv.FormatFloat(row.ClicksPercent, 'f', 1, 64) + "%",
			strconv.FormatFloat(row.CurrentImpressions, 'f', 0, 64),
			strconv.FormatFloat(row.PreviousImpressions, 'f', 0, 64),
			strconv.FormatFloat(row.ImpressionsDelta, 'f', 0, 64),
			strconv.FormatFloat(row.ImpressionsPercent, 'f', 1, 64) + "%",
			strconv.FormatFloat(row.CurrentPosition, 'f', 1, 64),
			strconv.FormatFloat(row.PreviousPosition, 'f', 1, 64),
			strconv.FormatFloat(row.PositionDelta, 'f', 1, 64),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("could not write row: %w", err)
		}
	}

	return nil
}

// DropsRow represents a query with ranking drops
type DropsRow struct {
	Query           string
	CurrentPosition float64
	PreviousPosition float64
	PositionDrop    float64
	CurrentClicks   float64
	PreviousClicks  float64
	CurrentImpressions float64
}

// WriteDropsCSV writes ranking drops to a CSV file
func WriteDropsCSV(filename string, rows []DropsRow) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"Query",
		"Position Drop",
		"Current Position", "Previous Position",
		"Current Clicks", "Previous Clicks",
		"Current Impressions",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("could not write header: %w", err)
	}

	for _, row := range rows {
		record := []string{
			row.Query,
			strconv.FormatFloat(row.PositionDrop, 'f', 1, 64),
			strconv.FormatFloat(row.CurrentPosition, 'f', 1, 64),
			strconv.FormatFloat(row.PreviousPosition, 'f', 1, 64),
			strconv.FormatFloat(row.CurrentClicks, 'f', 0, 64),
			strconv.FormatFloat(row.PreviousClicks, 'f', 0, 64),
			strconv.FormatFloat(row.CurrentImpressions, 'f', 0, 64),
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("could not write row: %w", err)
		}
	}

	return nil
}

package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// Table wraps tablewriter with sensible defaults
type Table struct {
	*tablewriter.Table
}

// NewTable creates a new table with default styling
func NewTable() *Table {
	t := tablewriter.NewWriter(os.Stdout)
	t.SetBorder(false)
	t.SetColumnSeparator(" ")
	t.SetHeaderLine(false)
	t.SetAutoWrapText(false)
	t.SetAutoFormatHeaders(false)
	t.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.SetTablePadding("  ")
	t.SetNoWhiteSpace(true)
	return &Table{Table: t}
}

// SetHeaders sets bold headers
func (t *Table) SetHeaders(headers ...string) {
	boldHeaders := make([]string, len(headers))
	for i, h := range headers {
		boldHeaders[i] = Bold(h)
	}
	t.SetHeader(boldHeaders)
}

// Color helpers
var (
	boldColor   = color.New(color.Bold)
	greenColor  = color.New(color.FgGreen)
	redColor    = color.New(color.FgRed)
	yellowColor = color.New(color.FgYellow)
	cyanColor   = color.New(color.FgCyan)
	dimColor    = color.New(color.Faint)
)

// Bold returns bold text
func Bold(s string) string {
	return boldColor.Sprint(s)
}

// Green returns green text
func Green(s string) string {
	return greenColor.Sprint(s)
}

// Red returns red text
func Red(s string) string {
	return redColor.Sprint(s)
}

// Yellow returns yellow text
func Yellow(s string) string {
	return yellowColor.Sprint(s)
}

// Cyan returns cyan text
func Cyan(s string) string {
	return cyanColor.Sprint(s)
}

// Dim returns dimmed text
func Dim(s string) string {
	return dimColor.Sprint(s)
}

// FormatNumber formats a number with commas
func FormatNumber(n float64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", n/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", n/1000)
	}
	return fmt.Sprintf("%.0f", n)
}

// FormatCTR formats CTR as a percentage
func FormatCTR(ctr float64) string {
	return fmt.Sprintf("%.1f%%", ctr*100)
}

// FormatPosition formats position with 1 decimal
func FormatPosition(pos float64) string {
	return fmt.Sprintf("%.1f", pos)
}

// FormatDelta formats a delta value with color and sign
func FormatDelta(delta float64, higherIsBetter bool) string {
	if delta == 0 {
		return Dim("—")
	}

	sign := "+"
	if delta < 0 {
		sign = ""
	}

	formatted := fmt.Sprintf("%s%.1f", sign, delta)

	if higherIsBetter {
		if delta > 0 {
			return Green(formatted)
		}
		return Red(formatted)
	} else {
		// Lower is better (e.g., position)
		if delta < 0 {
			return Green(formatted)
		}
		return Red(formatted)
	}
}

// FormatPercentDelta formats a percentage delta with color
func FormatPercentDelta(delta float64, higherIsBetter bool) string {
	if delta == 0 {
		return Dim("—")
	}

	sign := "+"
	if delta < 0 {
		sign = ""
	}

	formatted := fmt.Sprintf("%s%.1f%%", sign, delta)

	if higherIsBetter {
		if delta > 0 {
			return Green(formatted)
		}
		return Red(formatted)
	} else {
		if delta < 0 {
			return Green(formatted)
		}
		return Red(formatted)
	}
}

// TruncateString truncates a string to max length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

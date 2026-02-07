// Package output provides formatters for CLI output
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

// Format represents output format type
type Format int

const (
	FormatTable Format = iota
	FormatJSON
	FormatQuiet
)

// Formatter interface for output formatting
type Formatter interface {
	// Print outputs data to stdout
	Print(data interface{}) error
	
	// PrintTable outputs tabular data
	PrintTable(headers []string, rows [][]string) error
	
	// PrintSuccess prints a success message
	PrintSuccess(msg string)
	
	// PrintError prints an error message
	PrintError(msg string)
	
	// PrintWarning prints a warning message
	PrintWarning(msg string)
	
	// PrintInfo prints an info message
	PrintInfo(msg string)
}

// NewFormatter creates a formatter based on format type
func NewFormatter(format Format) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{Writer: os.Stdout}
	case FormatQuiet:
		return &QuietFormatter{Writer: os.Stdout}
	default:
		return &TableFormatter{Writer: os.Stdout}
	}
}

// TableFormatter outputs data as colored tables
type TableFormatter struct {
	Writer io.Writer
}

func (f *TableFormatter) Print(data interface{}) error {
	return json.NewEncoder(f.Writer).Encode(data)
}

func (f *TableFormatter) PrintTable(headers []string, rows [][]string) error {
	w := tabwriter.NewWriter(f.Writer, 0, 0, 2, ' ', 0)
	
	// Print header with bold
	boldHeader := make([]string, len(headers))
	for i, h := range headers {
		boldHeader[i] = color.New(color.Bold).Sprint(h)
	}
	fmt.Fprintln(w, strings.Join(boldHeader, "\t"))
	
	// Print rows
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	
	return w.Flush()
}

func (f *TableFormatter) PrintSuccess(msg string) {
	color.Green("✓ %s", msg)
}

func (f *TableFormatter) PrintError(msg string) {
	color.Red("✗ %s", msg)
}

func (f *TableFormatter) PrintWarning(msg string) {
	color.Yellow("⚠ %s", msg)
}

func (f *TableFormatter) PrintInfo(msg string) {
	color.Cyan("ℹ %s", msg)
}

// JSONFormatter outputs data as JSON
type JSONFormatter struct {
	Writer io.Writer
}

func (f *JSONFormatter) Print(data interface{}) error {
	encoder := json.NewEncoder(f.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *JSONFormatter) PrintTable(headers []string, rows [][]string) error {
	// Convert table to JSON array of objects
	var result []map[string]string
	for _, row := range rows {
		obj := make(map[string]string)
		for i, header := range headers {
			if i < len(row) {
				obj[header] = row[i]
			}
		}
		result = append(result, obj)
	}
	return f.Print(result)
}

func (f *JSONFormatter) PrintSuccess(msg string) {
	f.Print(map[string]string{"status": "success", "message": msg})
}

func (f *JSONFormatter) PrintError(msg string) {
	f.Print(map[string]string{"status": "error", "message": msg})
}

func (f *JSONFormatter) PrintWarning(msg string) {
	f.Print(map[string]string{"status": "warning", "message": msg})
}

func (f *JSONFormatter) PrintInfo(msg string) {
	f.Print(map[string]string{"status": "info", "message": msg})
}

// QuietFormatter outputs minimal data (IDs only)
type QuietFormatter struct {
	Writer io.Writer
}

func (f *QuietFormatter) Print(data interface{}) error {
	fmt.Fprintln(f.Writer, data)
	return nil
}

func (f *QuietFormatter) PrintTable(headers []string, rows [][]string) error {
	// Print only first column (assumed to be ID)
	for _, row := range rows {
		if len(row) > 0 {
			fmt.Fprintln(f.Writer, row[0])
		}
	}
	return nil
}

func (f *QuietFormatter) PrintSuccess(msg string) {}
func (f *QuietFormatter) PrintError(msg string)   { fmt.Fprintln(f.Writer, msg) }
func (f *QuietFormatter) PrintWarning(msg string) {}
func (f *QuietFormatter) PrintInfo(msg string)    {}

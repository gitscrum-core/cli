package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		format   Format
		wantType string
	}{
		{FormatTable, "*output.TableFormatter"},
		{FormatJSON, "*output.JSONFormatter"},
		{FormatQuiet, "*output.QuietFormatter"},
	}

	for _, tt := range tests {
		t.Run(tt.wantType, func(t *testing.T) {
			f := NewFormatter(tt.format)
			if f == nil {
				t.Fatal("NewFormatter returned nil")
			}
		})
	}
}

func TestTableFormatter_PrintTable(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{Writer: &buf}

	headers := []string{"CODE", "TITLE", "STATUS"}
	rows := [][]string{
		{"GS-123", "Fix bug", "done"},
		{"GS-456", "Add feature", "in progress"},
	}

	err := f.PrintTable(headers, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Note: pterm renders to os.Stdout, not to the Writer field
	// This test verifies no error occurs during rendering
}

func TestTableFormatter_PrintSuccess(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{Writer: &buf}
	
	// Note: color output goes to os.Stdout, not the formatter's Writer
	// This test just ensures no panic
	f.PrintSuccess("Task created")
}

func TestTableFormatter_PrintError(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{Writer: &buf}
	f.PrintError("Something failed")
}

func TestTableFormatter_PrintWarning(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{Writer: &buf}
	f.PrintWarning("Be careful")
}

func TestTableFormatter_PrintInfo(t *testing.T) {
	var buf bytes.Buffer
	f := &TableFormatter{Writer: &buf}
	f.PrintInfo("FYI")
}

func TestJSONFormatter_Print(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{Writer: &buf}

	data := map[string]interface{}{
		"id":    "123",
		"title": "Test Task",
		"done":  true,
	}

	err := f.Print(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse output as JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if result["id"] != "123" {
		t.Errorf("id = %v, want %v", result["id"], "123")
	}
	if result["title"] != "Test Task" {
		t.Errorf("title = %v, want %v", result["title"], "Test Task")
	}
}

func TestJSONFormatter_PrintTable(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{Writer: &buf}

	headers := []string{"code", "title"}
	rows := [][]string{
		{"GS-123", "First task"},
		{"GS-456", "Second task"},
	}

	err := f.PrintTable(headers, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse output as JSON array
	var result []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON array: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}

	if result[0]["code"] != "GS-123" {
		t.Errorf("first item code = %q, want %q", result[0]["code"], "GS-123")
	}
	if result[1]["title"] != "Second task" {
		t.Errorf("second item title = %q, want %q", result[1]["title"], "Second task")
	}
}

func TestJSONFormatter_PrintSuccess(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{Writer: &buf}

	f.PrintSuccess("Done!")

	var result map[string]string
	json.Unmarshal(buf.Bytes(), &result)

	if result["status"] != "success" {
		t.Errorf("status = %q, want %q", result["status"], "success")
	}
	if result["message"] != "Done!" {
		t.Errorf("message = %q, want %q", result["message"], "Done!")
	}
}

func TestJSONFormatter_PrintError(t *testing.T) {
	var buf bytes.Buffer
	f := &JSONFormatter{Writer: &buf}

	f.PrintError("Failed!")

	var result map[string]string
	json.Unmarshal(buf.Bytes(), &result)

	if result["status"] != "error" {
		t.Errorf("status = %q, want %q", result["status"], "error")
	}
}

func TestQuietFormatter_Print(t *testing.T) {
	var buf bytes.Buffer
	f := &QuietFormatter{Writer: &buf}

	f.Print("GS-123")

	output := strings.TrimSpace(buf.String())
	if output != "GS-123" {
		t.Errorf("output = %q, want %q", output, "GS-123")
	}
}

func TestQuietFormatter_PrintTable(t *testing.T) {
	var buf bytes.Buffer
	f := &QuietFormatter{Writer: &buf}

	headers := []string{"id", "title"}
	rows := [][]string{
		{"GS-123", "First"},
		{"GS-456", "Second"},
	}

	err := f.PrintTable(headers, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Quiet mode should only output first column (IDs)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "GS-123" {
		t.Errorf("line 0 = %q, want %q", lines[0], "GS-123")
	}
	if lines[1] != "GS-456" {
		t.Errorf("line 1 = %q, want %q", lines[1], "GS-456")
	}
}

func TestQuietFormatter_SilentMethods(t *testing.T) {
	var buf bytes.Buffer
	f := &QuietFormatter{Writer: &buf}

	// These should not produce output
	f.PrintSuccess("test")
	f.PrintWarning("test")
	f.PrintInfo("test")

	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}

func TestQuietFormatter_PrintError(t *testing.T) {
	var buf bytes.Buffer
	f := &QuietFormatter{Writer: &buf}

	// Errors should still be printed
	f.PrintError("something failed")

	if !strings.Contains(buf.String(), "something failed") {
		t.Error("PrintError should output error message")
	}
}

func TestFormat_Constants(t *testing.T) {
	// Ensure format constants are distinct
	if FormatTable == FormatJSON {
		t.Error("FormatTable should not equal FormatJSON")
	}
	if FormatJSON == FormatQuiet {
		t.Error("FormatJSON should not equal FormatQuiet")
	}
	if FormatTable == FormatQuiet {
		t.Error("FormatTable should not equal FormatQuiet")
	}
}

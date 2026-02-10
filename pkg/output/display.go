// Package output provides display helpers for consistent CLI output
package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Standard color definitions used across all commands
var (
	headerColor    = color.New(color.FgWhite, color.Bold)
	subHeaderColor = color.New(color.FgWhite)
	successColor   = color.New(color.FgGreen)
	warningColor   = color.New(color.FgYellow)
	errorColor     = color.New(color.FgRed, color.Bold)
	infoColor      = color.New(color.FgCyan)
	dimColor       = color.New(color.FgHiBlack)
	keyColor       = color.New(color.FgHiBlack)
	valueColor     = color.New(color.FgWhite)
	alertColor     = color.New(color.FgYellow, color.Bold)
)

// Header prints a bold section header with a separator line.
//
//	┃ CRM Dashboard
//	┃ ────────────────────────────
func Header(title string) {
	fmt.Println()
	headerColor.Println(title)
	dimColor.Println(strings.Repeat("─", 50))
}

// SubHeader prints a section sub-header (dimmer than Header).
//
//	Summary
func SubHeader(title string) {
	fmt.Println()
	subHeaderColor.Println(title)
}

// Separator prints a thin separator line.
func Separator() {
	dimColor.Println(strings.Repeat("─", 50))
}

// Success prints a green success message: ✓ message
func Success(msg string) {
	successColor.Printf("✓ %s\n", msg)
}

// Successf prints a green formatted success message: ✓ message
func Successf(format string, a ...interface{}) {
	Success(fmt.Sprintf(format, a...))
}

// Warning prints a yellow warning message: ⚠ message
func Warning(msg string) {
	warningColor.Printf("⚠ %s\n", msg)
}

// Warningf prints a yellow formatted warning message: ⚠ message
func Warningf(format string, a ...interface{}) {
	Warning(fmt.Sprintf(format, a...))
}

// Error prints a red error message: ✗ message
func Error(msg string) {
	errorColor.Printf("✗ %s\n", msg)
}

// Errorf prints a red formatted error message: ✗ message
func Errorf(format string, a ...interface{}) {
	Error(fmt.Sprintf(format, a...))
}

// Info prints a cyan info message: ● message
func Info(msg string) {
	infoColor.Printf("● %s\n", msg)
}

// Infof prints a cyan formatted info message: ● message
func Infof(format string, a ...interface{}) {
	Info(fmt.Sprintf(format, a...))
}

// Alert prints a bold yellow alert message: ⚠ message
func Alert(msg string) {
	alertColor.Printf("  ⚠ %s\n", msg)
}

// Alertf prints a bold yellow formatted alert message: ⚠ message
func Alertf(format string, a ...interface{}) {
	Alert(fmt.Sprintf(format, a...))
}

// KeyValue prints an aligned key-value pair.
//
//	  Workspace: my-workspace
func KeyValue(key, value string) {
	keyColor.Printf("  %s: ", key)
	valueColor.Println(value)
}

// KeyValuef prints a formatted key-value pair.
func KeyValuef(key, format string, a ...interface{}) {
	KeyValue(key, fmt.Sprintf(format, a...))
}

// Stat prints an inline stat with label and value.
//
//	Active: 42
func Stat(label string, value interface{}) {
	dimColor.Printf("%s: ", label)
	valueColor.Printf("%v", value)
}

// StatLine prints a line of multiple stats separated by " │ "
func StatLine(stats ...string) {
	fmt.Printf("  %s\n", strings.Join(stats, dimColor.Sprint(" │ ")))
}

// Empty prints a helpful empty-state message with a suggested next action.
//
//	No clients found
//	  Add one with: gitscrum clients create "Name"
func Empty(msg, suggestion string) {
	dimColor.Printf("  %s\n", msg)
	if suggestion != "" {
		fmt.Println()
		infoColor.Printf("  %s\n", suggestion)
	}
}

// EmptyContext prints an empty-state message with workspace/project context.
// Use this for project-scoped resources like tasks, sprints, wiki pages.
//
//	No tasks found in project "my-project" (workspace: my-workspace)
//	  Create one with: gitscrum tasks create "Task title"
func EmptyContext(resource, workspace, project, suggestion string) {
	var context string
	if project != "" && workspace != "" {
		context = fmt.Sprintf(" in project \"%s\" (workspace: %s)", project, workspace)
	} else if workspace != "" {
		context = fmt.Sprintf(" in workspace \"%s\"", workspace)
	}
	dimColor.Printf("  No %s found%s\n", resource, context)
	if suggestion != "" {
		fmt.Println()
		infoColor.Printf("  %s\n", suggestion)
	}
}

// Bullet prints a bulleted list item.
//
//	• item text
func Bullet(text string) {
	fmt.Printf("  • %s\n", text)
}

// Bulletf prints a formatted bulleted list item.
func Bulletf(format string, a ...interface{}) {
	Bullet(fmt.Sprintf(format, a...))
}

// Dim prints dimmed/secondary text.
func Dim(text string) {
	dimColor.Printf("    %s\n", text)
}

// Dimf prints formatted dimmed/secondary text.
func Dimf(format string, a ...interface{}) {
	Dim(fmt.Sprintf(format, a...))
}

// FormatStatInline formats a label:value pair for use in StatLine.
func FormatStatInline(label string, value interface{}) string {
	return fmt.Sprintf("%s: %s",
		dimColor.Sprint(label),
		valueColor.Sprintf("%v", value))
}

// FormatMoney formats a monetary amount with currency symbol.
func FormatMoney(symbol string, amount float64) string {
	return fmt.Sprintf("%s%.2f", symbol, amount)
}

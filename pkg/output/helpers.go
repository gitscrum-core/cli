// Package output provides shared formatting helpers for CLI output
package output

import (
	"regexp"
	"strings"
	"time"
)

// StripHTML removes HTML tags and converts to plain text for terminal display
func StripHTML(html string) string {
	if html == "" {
		return ""
	}
	// Replace <br>, <br/>, <br /> with newlines
	brRe := regexp.MustCompile(`(?i)<br\s*/?>`)
	html = brRe.ReplaceAllString(html, "\n")
	// Replace </p>, </div>, </li> with newlines
	blockRe := regexp.MustCompile(`(?i)</(?:p|div|li|h[1-6])>`)
	html = blockRe.ReplaceAllString(html, "\n")
	// Strip all remaining HTML tags
	tagRe := regexp.MustCompile(`<[^>]*>`)
	html = tagRe.ReplaceAllString(html, "")
	// Decode common HTML entities
	html = strings.ReplaceAll(html, "&amp;", "&")
	html = strings.ReplaceAll(html, "&lt;", "<")
	html = strings.ReplaceAll(html, "&gt;", ">")
	html = strings.ReplaceAll(html, "&quot;", "\"")
	html = strings.ReplaceAll(html, "&#39;", "'")
	html = strings.ReplaceAll(html, "&nbsp;", " ")
	// Collapse multiple blank lines
	multiNewline := regexp.MustCompile(`\n{3,}`)
	html = multiNewline.ReplaceAllString(html, "\n\n")
	return strings.TrimSpace(html)
}

// StatusIcon returns a text-based status icon for display
func StatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "done", "completed", "closed", "merged":
		return "[x]"
	case "in progress", "in_progress", "doing", "active":
		return "[>]"
	case "todo", "to do", "to_do", "open", "new":
		return "[ ]"
	case "blocked":
		return "[!]"
	case "review", "in review", "in_review", "pending":
		return "[~]"
	default:
		return "[-]"
	}
}

// Truncate shortens a string to the specified length with ellipsis
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// FormatDate formats an ISO 8601 date string for display
func FormatDate(dateStr string) string {
	if dateStr == "" {
		return "-"
	}

	// Try common date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	var t time.Time
	var err error
	for _, format := range formats {
		t, err = time.Parse(format, dateStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return dateStr // Return original if unparseable
	}

	// Format as relative or short date
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return strings.ReplaceAll(string(rune(mins))+" mins ago", string(rune(mins)), formatInt(mins))
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return formatInt(hours) + " hours ago"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return formatInt(days) + " days ago"
	default:
		return t.Format("Jan 2, 2006")
	}
}

// FormatDuration formats a duration in hours/minutes for display
func FormatDuration(hours float64) string {
	if hours < 0.01 {
		return "-"
	}
	
	totalMinutes := int(hours * 60)
	h := totalMinutes / 60
	m := totalMinutes % 60

	if h == 0 {
		return formatInt(m) + "m"
	}
	if m == 0 {
		return formatInt(h) + "h"
	}
	return formatInt(h) + "h " + formatInt(m) + "m"
}

// formatInt is a simple int to string helper
func formatInt(n int) string {
	if n < 0 {
		return "-" + formatInt(-n)
	}
	
	digits := []byte{}
	if n == 0 {
		return "0"
	}
	
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	
	return string(digits)
}

// PriorityLabel returns a human-readable priority label
func PriorityLabel(priority int) string {
	switch priority {
	case 1:
		return "Critical"
	case 2:
		return "High"
	case 3:
		return "Medium"
	case 4:
		return "Low"
	default:
		return "-"
	}
}

// BoolYesNo returns "Yes" or "No" for boolean display
func BoolYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// EmptyDefault returns the default value if the string is empty
func EmptyDefault(s, defaultVal string) string {
	if strings.TrimSpace(s) == "" {
		return defaultVal
	}
	return s
}

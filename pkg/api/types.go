// Package api provides shared types for the GitScrum API
package api

// DateResource represents a date object from the API (DateResource.php)
// All dates in the API are returned in this format with timezone-aware fields.
type DateResource struct {
	DateForHumans string `json:"date_for_humans"`
	ISO8601       string `json:"iso8601"`
	Timestamp     int64  `json:"timestamp"`
}

// FormatDate returns the human-readable date or ISO8601 fallback
func (d *DateResource) FormatDate() string {
	if d == nil {
		return ""
	}
	if d.DateForHumans != "" {
		return d.DateForHumans
	}
	if d.ISO8601 != "" {
		return d.ISO8601
	}
	return ""
}

// FormatISO returns the ISO8601 date string
func (d *DateResource) FormatISO() string {
	if d == nil {
		return ""
	}
	return d.ISO8601
}

// ISODate returns just the date portion (YYYY-MM-DD) from ISO8601
func (d *DateResource) ISODate() string {
	if d == nil || d.ISO8601 == "" {
		return ""
	}
	// ISO8601 is like "2026-02-08T02:37:54+00:00", take first 10 chars
	if len(d.ISO8601) >= 10 {
		return d.ISO8601[:10]
	}
	return d.ISO8601
}

// DateTime returns date and time (YYYY-MM-DD HH:MM) from ISO8601
func (d *DateResource) DateTime() string {
	if d == nil || d.ISO8601 == "" {
		return ""
	}
	// ISO8601 is like "2026-02-08T02:37:54+00:00", take first 16 chars and replace T with space
	if len(d.ISO8601) >= 16 {
		return d.ISO8601[:10] + " " + d.ISO8601[11:16]
	}
	return d.ISO8601
}

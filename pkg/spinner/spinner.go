// Package spinner provides loading indicators
package spinner

import (
	"time"

	"github.com/briandowns/spinner"
)

// Spinner wraps briandowns/spinner with defaults
type Spinner struct {
	s *spinner.Spinner
}

// New creates a new spinner with GitScrum styling
func New(suffix string) *Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + suffix
	s.Color("cyan")
	return &Spinner{s: s}
}

// Start shows the spinner
func (sp *Spinner) Start() {
	sp.s.Start()
}

// Stop hides the spinner
func (sp *Spinner) Stop() {
	sp.s.Stop()
}

// UpdateSuffix changes the spinner text
func (sp *Spinner) UpdateSuffix(suffix string) {
	sp.s.Suffix = " " + suffix
}

// Success stops spinner with success message
func (sp *Spinner) Success(msg string) {
	sp.s.FinalMSG = "✓ " + msg + "\n"
	sp.s.Stop()
}

// Error stops spinner with error message
func (sp *Spinner) Error(msg string) {
	sp.s.FinalMSG = "✗ " + msg + "\n"
	sp.s.Stop()
}

// WithSpinner runs a function with a spinner
func WithSpinner(msg string, fn func() error) error {
	sp := New(msg)
	sp.Start()
	defer sp.Stop()
	return fn()
}

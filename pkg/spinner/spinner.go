// Package spinner provides modern PTerm loading indicators
package spinner

import (
	"github.com/pterm/pterm"
)

// Spinner wraps pterm spinner with GitScrum styling
type Spinner struct {
	s       *pterm.SpinnerPrinter
	message string
}

// New creates a new spinner with GitScrum styling
func New(message string) *Spinner {
	return &Spinner{
		message: message,
	}
}

// Start shows the spinner
func (sp *Spinner) Start() {
	spinner, _ := pterm.DefaultSpinner.
		WithSequence("⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷").
		Start(sp.message)
	sp.s = spinner
}

// Stop hides the spinner
func (sp *Spinner) Stop() {
	if sp.s != nil {
		sp.s.Stop()
	}
}

// UpdateSuffix changes the spinner text
func (sp *Spinner) UpdateSuffix(message string) {
	if sp.s != nil {
		sp.s.UpdateText(message)
	}
}

// Success stops spinner with success message
func (sp *Spinner) Success(msg string) {
	if sp.s != nil {
		sp.s.Success(msg)
	}
}

// Error stops spinner with error message
func (sp *Spinner) Error(msg string) {
	if sp.s != nil {
		sp.s.Fail(msg)
	}
}

// Warning stops spinner with warning message
func (sp *Spinner) Warning(msg string) {
	if sp.s != nil {
		sp.s.Warning(msg)
	}
}

// WithSpinner runs a function with a spinner
func WithSpinner(msg string, fn func() error) error {
	sp := New(msg)
	sp.Start()
	err := fn()
	if err != nil {
		sp.Error(err.Error())
		return err
	}
	sp.Stop()
	return nil
}

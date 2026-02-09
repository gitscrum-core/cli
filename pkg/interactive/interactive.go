// Package interactive provides modern interactive CLI components using PTerm
package interactive

import (
	"github.com/pterm/pterm"
)

// Option represents an item in a selection list
type Option struct {
	Label string
	Value string
}

// Select shows an interactive selection menu and returns the selected option
// Users can navigate with arrow keys, type to filter, and press Enter to select
func Select(message string, options []Option) (Option, error) {
	labels := make([]string, len(options))
	for i, opt := range options {
		labels[i] = opt.Label
	}

	selectedLabel, err := pterm.DefaultInteractiveSelect.
		WithOptions(labels).
		Show(message)
	if err != nil {
		return Option{}, err
	}

	// Find the selected option
	for _, opt := range options {
		if opt.Label == selectedLabel {
			return opt, nil
		}
	}

	return Option{}, nil
}

// SelectString shows an interactive selection menu and returns the selected string
func SelectString(message string, options []string) (string, error) {
	return pterm.DefaultInteractiveSelect.
		WithOptions(options).
		Show(message)
}

// MultiSelect shows an interactive multi-selection menu
func MultiSelect(message string, options []Option) ([]Option, error) {
	labels := make([]string, len(options))
	for i, opt := range options {
		labels[i] = opt.Label
	}

	selectedLabels, err := pterm.DefaultInteractiveMultiselect.
		WithOptions(labels).
		Show(message)
	if err != nil {
		return nil, err
	}

	// Find the selected options
	var selected []Option
	for _, label := range selectedLabels {
		for _, opt := range options {
			if opt.Label == label {
				selected = append(selected, opt)
				break
			}
		}
	}

	return selected, nil
}

// Confirm shows a confirmation prompt
func Confirm(message string) (bool, error) {
	return pterm.DefaultInteractiveConfirm.Show(message)
}

// TextInput shows an interactive text input prompt
func TextInput(message string, defaultValue string) (string, error) {
	return pterm.DefaultInteractiveTextInput.
		WithDefaultValue(defaultValue).
		Show(message)
}

// PasswordInput shows a masked password input prompt
func PasswordInput(message string) (string, error) {
	return pterm.DefaultInteractiveTextInput.
		WithMask("*").
		Show(message)
}

// Package i18n provides internationalization support for the CLI
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

var (
	bundle     *i18n.Bundle
	localizer  *i18n.Localizer
	currentLang = "en"
	mu         sync.RWMutex
)

// SupportedLanguages lists all available languages
var SupportedLanguages = []string{"en", "pt", "fr", "es"}

// LanguageNames maps language codes to display names
var LanguageNames = map[string]string{
	"en": "English",
	"pt": "Português",
	"fr": "Français",
	"es": "Español",
}

// Init initializes the i18n system with the specified language
func Init(lang string) error {
	mu.Lock()
	defer mu.Unlock()

	if lang == "" {
		lang = "en"
	}

	// Validate language
	valid := false
	for _, l := range SupportedLanguages {
		if l == lang {
			valid = true
			break
		}
	}
	if !valid {
		lang = "en"
	}

	currentLang = lang

	// Create bundle with English as default
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load all locale files
	for _, l := range SupportedLanguages {
		data, err := localeFS.ReadFile(fmt.Sprintf("locales/%s.json", l))
		if err != nil {
			continue // Skip missing files
		}
		bundle.MustParseMessageFileBytes(data, fmt.Sprintf("%s.json", l))
	}

	// Create localizer for current language with English fallback
	localizer = i18n.NewLocalizer(bundle, lang, "en")

	return nil
}

// T translates a message by key
func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if localizer == nil {
		return key // Return key if not initialized
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})
	if err != nil {
		return key // Return key if translation not found
	}
	return msg
}

// Tf translates a message with template data
func Tf(key string, data map[string]interface{}) string {
	mu.RLock()
	defer mu.RUnlock()

	if localizer == nil {
		return key
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
	})
	if err != nil {
		return key
	}
	return msg
}

// CurrentLanguage returns the current language code
func CurrentLanguage() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// IsSupported checks if a language is supported
func IsSupported(lang string) bool {
	for _, l := range SupportedLanguages {
		if l == lang {
			return true
		}
	}
	return false
}

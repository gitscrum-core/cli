// Package config provides config commands for GitScrum CLI
package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/config"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
)

// NewCmdConfig creates the config command group
func NewCmdConfig(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage GitScrum CLI configuration",
		Long: `View and modify GitScrum CLI configuration.

Configuration is stored in ~/.gitscrum/config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow(f)
		},
	}

	cmd.AddCommand(NewCmdConfigSet(f))
	cmd.AddCommand(NewCmdConfigGet(f))
	cmd.AddCommand(NewCmdConfigReset(f))
	cmd.AddCommand(NewCmdConfigLanguage(f))

	return cmd
}

func runConfigShow(f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	configPath, _ := config.ConfigPath()

	output.Header("GitScrum CLI Configuration")

	output.KeyValue("Config file", configPath)
	output.KeyValue("API URL", cfg.APIURL)
	output.KeyValue("Workspace", valueOrNone(cfg.Workspace))
	output.KeyValue("Project", valueOrNone(cfg.Project))
	lang := config.GetLanguage()
	langName := i18n.LanguageNames[lang]
	output.KeyValuef("Language", "%s (%s)", lang, langName)

	fmt.Println()
	return nil
}

func valueOrNone(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}

// NewCmdConfigSet creates the config set command
func NewCmdConfigSet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value.

Available keys:
  workspace    Default workspace slug
  project      Default project slug
  language     Language (en, pt, fr, es)
  api_url      API URL (advanced)`,
		Example: `  gitscrum config set workspace my-company
  gitscrum config set project my-project`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			return runConfigSet(f, key, value)
		},
	}

	return cmd
}

func runConfigSet(f *factory.Factory, key, value string) error {
	switch key {
	case "workspace":
		if err := config.SetWorkspace(value); err != nil {
			return err
		}
		output.Successf("Set workspace to '%s'", value)

	case "project":
		if err := config.SetProject(value); err != nil {
			return err
		}
		output.Successf("Set project to '%s'", value)

	case "api_url":
		viper.Set("api_url", value)
		cfg, _ := f.Config()
		cfg.APIURL = value
		if err := config.Save(cfg); err != nil {
			return err
		}
		output.Successf("Set API URL to '%s'", value)

	case "language":
		if !i18n.IsSupported(value) {
			return fmt.Errorf("invalid language: %s. Supported: en, pt, fr, es", value)
		}
		if err := config.SetLanguage(value); err != nil {
			return err
		}
		i18n.Init(value)
		output.Successf("Set language to '%s' (%s)", value, i18n.LanguageNames[value])

	default:
		return fmt.Errorf("unknown config key: %s. Available: workspace, project, language, api_url", key)
	}

	return nil
}

// NewCmdConfigGet creates the config get command
func NewCmdConfigGet(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "get <key>",
		Short:   "Get a configuration value",
		Example: "  gitscrum config get workspace",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			key := args[0]
			switch key {
			case "workspace":
				fmt.Println(cfg.Workspace)
			case "project":
				fmt.Println(cfg.Project)
			case "language":
				fmt.Println(config.GetLanguage())
			case "api_url":
				fmt.Println(cfg.APIURL)
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}

			return nil
		},
	}
}

// NewCmdConfigReset creates the config reset command
func NewCmdConfigReset(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &config.Config{
				APIURL: config.DefaultAPIURL,
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			output.Success("Configuration reset to defaults")
			return nil
		},
	}
}

// NewCmdConfigLanguage creates the config language command
func NewCmdConfigLanguage(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "language [lang]",
		Short: "Get or set CLI language",
		Long: `Get or set the CLI language.

Supported languages:
  en   English
  pt   Português
  fr   Français
  es   Español

Without arguments, shows current language.
With a language code, sets the new language.`,
		Example: `  # Show current language
  gitscrum config language

  # Set language to Portuguese
  gitscrum config language pt`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				// Show current language
				lang := config.GetLanguage()
				langName := i18n.LanguageNames[lang]
				output.Infof("Current language: %s (%s)", lang, langName)
				fmt.Println()
				output.Dim("Available languages:")
				for _, l := range i18n.SupportedLanguages {
					marker := "  "
					if l == lang {
						marker = "→ "
					}
					fmt.Printf("  %s%s  %s\n", marker, l, i18n.LanguageNames[l])
				}
				return nil
			}

			// Set language
			lang := args[0]
			if !i18n.IsSupported(lang) {
				return fmt.Errorf("invalid language: %s. Supported: en, pt, fr, es", lang)
			}
			if err := config.SetLanguage(lang); err != nil {
				return err
			}
			i18n.Init(lang)
			output.Successf("Language changed to %s (%s)", lang, i18n.LanguageNames[lang])
			return nil
		},
	}
}

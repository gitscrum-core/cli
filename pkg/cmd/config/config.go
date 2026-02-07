// Package config provides config commands for GitScrum CLI
package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/config"
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

	return cmd
}

func runConfigShow(f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	configPath, _ := config.ConfigPath()

	fmt.Println("GitScrum CLI Configuration")
	fmt.Println()
	fmt.Printf("  Config file: %s\n", configPath)
	fmt.Println()
	fmt.Printf("  API URL:     %s\n", cfg.APIURL)
	fmt.Printf("  Workspace:   %s\n", valueOrNone(cfg.Workspace))
	fmt.Printf("  Project:     %s\n", valueOrNone(cfg.Project))

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
		fmt.Printf("✓ Set workspace to '%s'\n", value)

	case "project":
		if err := config.SetProject(value); err != nil {
			return err
		}
		fmt.Printf("✓ Set project to '%s'\n", value)

	case "api_url":
		viper.Set("api_url", value)
		cfg, _ := f.Config()
		cfg.APIURL = value
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("✓ Set API URL to '%s'\n", value)

	default:
		return fmt.Errorf("unknown config key: %s. Available: workspace, project, api_url", key)
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
			fmt.Println("✓ Configuration reset to defaults")
			return nil
		},
	}
}

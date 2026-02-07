// Package hooks provides git hooks commands for GitScrum CLI
package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/git"
)

// NewCmdHooks creates the hooks command group
func NewCmdHooks(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Manage git hooks",
		Long:  `Install and manage git hooks for GitScrum integration.`,
	}

	cmd.AddCommand(NewCmdHooksInstall(f))
	cmd.AddCommand(NewCmdHooksUninstall(f))

	return cmd
}

// NewCmdHooksInstall creates the hooks install command
func NewCmdHooksInstall(f *factory.Factory) *cobra.Command {
	var commitMsg bool
	var postCommit bool
	var prePush bool
	var all bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install git hooks",
		Long: `Install git hooks for GitScrum integration.

Available hooks:
- commit-msg: Detects task code in branch, adds to commit message
- post-commit: Updates task status after commit (optional)
- pre-push: Alerts about unassigned tasks before pushing`,
		Example: `  gitscrum hooks install --all
  gitscrum hooks install --commit-msg
  gitscrum hooks install --commit-msg --pre-push`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHooksInstall(f, all, commitMsg, postCommit, prePush)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Install all hooks")
	cmd.Flags().BoolVar(&commitMsg, "commit-msg", false, "Install commit-msg hook")
	cmd.Flags().BoolVar(&postCommit, "post-commit", false, "Install post-commit hook")
	cmd.Flags().BoolVar(&prePush, "pre-push", false, "Install pre-push hook")

	return cmd
}

func runHooksInstall(f *factory.Factory, all, commitMsg, postCommit, prePush bool) error {
	// Check if in git repo
	gitCtx, err := git.NewContext(".")
	if err != nil {
		return fmt.Errorf("not in a git repository")
	}

	hooksDir := filepath.Join(gitCtx.RootPath, ".git", "hooks")

	if !all && !commitMsg && !postCommit && !prePush {
		return fmt.Errorf("specify at least one hook: --all, --commit-msg, --post-commit, --pre-push")
	}

	if all {
		commitMsg = true
		postCommit = true
		prePush = true
	}

	fmt.Println("🪝 Installing git hooks...")
	fmt.Println()

	installed := 0

	if commitMsg {
		if err := installHook(hooksDir, "commit-msg", commitMsgHook); err != nil {
			fmt.Printf("  ⚠️  commit-msg: %v\n", err)
		} else {
			fmt.Println("  ✓ commit-msg")
			installed++
		}
	}

	if postCommit {
		if err := installHook(hooksDir, "post-commit", postCommitHook); err != nil {
			fmt.Printf("  ⚠️  post-commit: %v\n", err)
		} else {
			fmt.Println("  ✓ post-commit")
			installed++
		}
	}

	if prePush {
		if err := installHook(hooksDir, "pre-push", prePushHook); err != nil {
			fmt.Printf("  ⚠️  pre-push: %v\n", err)
		} else {
			fmt.Println("  ✓ pre-push")
			installed++
		}
	}

	fmt.Println()
	if installed > 0 {
		fmt.Printf("✅ Installed %d hook(s)\n", installed)
	}

	return nil
}

func installHook(hooksDir, name, content string) error {
	hookPath := filepath.Join(hooksDir, name)

	// Check if hook already exists
	if _, err := os.Stat(hookPath); err == nil {
		// Read existing hook
		existing, _ := os.ReadFile(hookPath)
		if strings.Contains(string(existing), "gitscrum") {
			return fmt.Errorf("already installed")
		}
		// Append to existing hook
		content = string(existing) + "\n\n# GitScrum hook\n" + content
	}

	return os.WriteFile(hookPath, []byte(content), 0755)
}

// NewCmdHooksUninstall creates the hooks uninstall command
func NewCmdHooksUninstall(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "uninstall",
		Short:   "Uninstall git hooks",
		Example: "  gitscrum hooks uninstall",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHooksUninstall(f)
		},
	}
}

func runHooksUninstall(f *factory.Factory) error {
	gitCtx, err := git.NewContext(".")
	if err != nil {
		return fmt.Errorf("not in a git repository")
	}

	hooksDir := filepath.Join(gitCtx.RootPath, ".git", "hooks")
	hooks := []string{"commit-msg", "post-commit", "pre-push"}

	fmt.Println("🪝 Uninstalling git hooks...")
	fmt.Println()

	removed := 0
	for _, name := range hooks {
		hookPath := filepath.Join(hooksDir, name)
		content, err := os.ReadFile(hookPath)
		if err != nil {
			continue
		}

		if strings.Contains(string(content), "gitscrum") {
			// Remove GitScrum portion
			lines := strings.Split(string(content), "\n")
			var newLines []string
			skip := false
			for _, line := range lines {
				if strings.Contains(line, "# GitScrum hook") {
					skip = true
					continue
				}
				if skip && (strings.HasPrefix(line, "#") || line == "" || strings.HasPrefix(line, "gitscrum")) {
					continue
				}
				skip = false
				newLines = append(newLines, line)
			}

			if len(newLines) == 0 || (len(newLines) == 1 && newLines[0] == "") {
				os.Remove(hookPath)
			} else {
				os.WriteFile(hookPath, []byte(strings.Join(newLines, "\n")), 0755)
			}
			fmt.Printf("  ✓ %s removed\n", name)
			removed++
		}
	}

	fmt.Println()
	if removed > 0 {
		fmt.Printf("✅ Removed %d hook(s)\n", removed)
	} else {
		fmt.Println("No GitScrum hooks found")
	}

	return nil
}

// Hook templates
const commitMsgHook = `#!/bin/sh
# GitScrum commit-msg hook
# Prepends task code from branch name to commit message

BRANCH=$(git symbolic-ref --short HEAD 2>/dev/null)
TASK_CODE=$(echo "$BRANCH" | grep -oE '[A-Z]+-[0-9]+' | head -1)

if [ -n "$TASK_CODE" ]; then
    COMMIT_MSG=$(cat "$1")
    # Only add if not already present
    if ! echo "$COMMIT_MSG" | grep -q "$TASK_CODE"; then
        echo "[$TASK_CODE] $COMMIT_MSG" > "$1"
    fi
fi
`

const postCommitHook = `#!/bin/sh
# GitScrum post-commit hook
# Optional: Updates task status after commit

BRANCH=$(git symbolic-ref --short HEAD 2>/dev/null)
TASK_CODE=$(echo "$BRANCH" | grep -oE '[A-Z]+-[0-9]+' | head -1)

if [ -n "$TASK_CODE" ]; then
    # Uncomment to auto-update task status
    # gitscrum tasks update "$TASK_CODE" --column="In Progress" 2>/dev/null || true
    echo "GitScrum: Committed for task $TASK_CODE"
fi
`

const prePushHook = `#!/bin/sh
# GitScrum pre-push hook
# Alerts about unassigned tasks before pushing

BRANCH=$(git symbolic-ref --short HEAD 2>/dev/null)
TASK_CODE=$(echo "$BRANCH" | grep -oE '[A-Z]+-[0-9]+' | head -1)

if [ -n "$TASK_CODE" ]; then
    echo "GitScrum: Pushing changes for task $TASK_CODE"
fi
`

#!/bin/bash
# onboarding-setup.sh - Quick setup script for new team members
#
# This script helps new team members get started with GitScrum CLI:
# - Installs the CLI
# - Authenticates with GitScrum
# - Configures workspace and project
# - Sets up git hooks
# - Shows helpful commands
#
# Usage:
#   curl -sL https://cli.gitscrum.com/onboarding.sh | bash
#   # Or run locally:
#   ./onboarding-setup.sh

set -e

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🚀 GitScrum CLI - Team Onboarding"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# ============================================
# Step 1: Install CLI
# ============================================
echo "📦 Step 1: Installing GitScrum CLI..."

if command -v gitscrum &> /dev/null; then
    VERSION=$(gitscrum --version 2>/dev/null || echo "unknown")
    echo "   ✅ Already installed: $VERSION"
else
    if command -v go &> /dev/null; then
        go install github.com/gitscrum-core/cli/cmd/gitscrum@latest
    else
        curl -sL https://raw.githubusercontent.com/gitscrum-core/cli/main/install.sh | sh
    fi
    echo "   ✅ Installed!"
fi
echo ""

# ============================================
# Step 2: Authenticate
# ============================================
echo "🔐 Step 2: Authentication..."

if gitscrum auth status &> /dev/null; then
    USER=$(gitscrum auth whoami --format json | jq -r '.name // .username')
    echo "   ✅ Already logged in as: $USER"
else
    echo "   Starting authentication..."
    gitscrum auth login
fi
echo ""

# ============================================
# Step 3: Configure Workspace
# ============================================
echo "⚙️  Step 3: Workspace Configuration..."

CURRENT_WORKSPACE=$(gitscrum config get workspace 2>/dev/null || echo "")

if [ -n "$CURRENT_WORKSPACE" ]; then
    echo "   Current workspace: $CURRENT_WORKSPACE"
    read -p "   Keep this workspace? [Y/n] " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        CURRENT_WORKSPACE=""
    fi
fi

if [ -z "$CURRENT_WORKSPACE" ]; then
    echo ""
    echo "   Available workspaces:"
    gitscrum workspaces list 2>/dev/null || echo "   (unable to fetch)"
    echo ""
    read -p "   Enter workspace slug: " WORKSPACE_SLUG
    gitscrum config set workspace "$WORKSPACE_SLUG"
    echo "   ✅ Workspace set: $WORKSPACE_SLUG"
fi
echo ""

# ============================================
# Step 4: Configure Project
# ============================================
echo "📁 Step 4: Project Configuration..."

CURRENT_PROJECT=$(gitscrum config get project 2>/dev/null || echo "")

if [ -n "$CURRENT_PROJECT" ]; then
    echo "   Current project: $CURRENT_PROJECT"
    read -p "   Keep this project? [Y/n] " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        CURRENT_PROJECT=""
    fi
fi

if [ -z "$CURRENT_PROJECT" ]; then
    echo ""
    echo "   Available projects:"
    gitscrum projects list 2>/dev/null || echo "   (unable to fetch)"
    echo ""
    read -p "   Enter project slug: " PROJECT_SLUG
    gitscrum config set project "$PROJECT_SLUG"
    echo "   ✅ Project set: $PROJECT_SLUG"
fi
echo ""

# ============================================
# Step 5: Git Hooks (Optional)
# ============================================
echo "🪝 Step 5: Git Hooks (Optional)..."

if [ -d ".git" ]; then
    read -p "   Install git hooks? [y/N] " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Download and install hooks
        HOOKS_URL="https://raw.githubusercontent.com/gitscrum-core/cli/main/docs/examples/git-hooks"
        
        for hook in commit-msg prepare-commit-msg post-checkout; do
            curl -sL "$HOOKS_URL/${hook}.sh" -o ".git/hooks/$hook" 2>/dev/null || true
            chmod +x ".git/hooks/$hook" 2>/dev/null || true
        done
        
        echo "   ✅ Hooks installed!"
        
        # Enable auto-timer config
        read -p "   Auto-start timer on branch checkout? [y/N] " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git config gitscrum.autoTimer true
            git config gitscrum.showTask true
            echo "   ✅ Auto-timer enabled!"
        fi
    else
        echo "   Skipped"
    fi
else
    echo "   Not in a git repository, skipping."
fi
echo ""

# ============================================
# Summary
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Setup Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Quick Commands:"
echo ""
echo "  gitscrum tasks              # List your tasks"
echo "  gitscrum tasks today        # Today's assigned tasks"
echo "  gitscrum tasks view GS-123  # View task details"
echo "  gitscrum tasks branch GS-1  # Create branch for task"
echo ""
echo "  gitscrum timer start GS-1   # Start time tracking"
echo "  gitscrum timer stop         # Stop timer"
echo "  gitscrum timer              # View active timer"
echo ""
echo "  gitscrum sprints current    # View current sprint"
echo "  gitscrum sprints burndown   # Sprint burndown chart"
echo ""
echo "Need help? Run: gitscrum --help"
echo ""

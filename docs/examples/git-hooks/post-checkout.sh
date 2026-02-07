#!/bin/bash
# post-checkout hook - Auto-start timer when switching to task branch
#
# When you checkout a branch with a task code, automatically:
# 1. Shows task info
# 2. Optionally starts a timer
#
# Installation:
#   cp post-checkout.sh .git/hooks/post-checkout
#   chmod +x .git/hooks/post-checkout
#
# Configuration:
#   git config gitscrum.autoTimer true  # Enable auto-timer
#   git config gitscrum.showTask true   # Show task info on checkout

PREV_HEAD=$1
NEW_HEAD=$2
BRANCH_CHECKOUT=$3

# Only run on branch checkout (not file checkout)
if [ "$BRANCH_CHECKOUT" != "1" ]; then
    exit 0
fi

# Get new branch name
BRANCH_NAME=$(git symbolic-ref --short HEAD 2>/dev/null)

# Extract task code
TASK_CODE=$(echo "$BRANCH_NAME" | grep -oE '[A-Z]{2,5}-[0-9]+' | head -1)

if [ -z "$TASK_CODE" ]; then
    exit 0
fi

# Check if GitScrum CLI is available
if ! command -v gitscrum &> /dev/null; then
    exit 0
fi

# Check configuration
SHOW_TASK=$(git config --bool gitscrum.showTask 2>/dev/null)
AUTO_TIMER=$(git config --bool gitscrum.autoTimer 2>/dev/null)

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📌 Task: $TASK_CODE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Show task info if enabled
if [ "$SHOW_TASK" = "true" ]; then
    gitscrum tasks view "$TASK_CODE" 2>/dev/null
    echo ""
fi

# Auto-start timer if enabled
if [ "$AUTO_TIMER" = "true" ]; then
    # Check if there's already a running timer
    ACTIVE_TIMER=$(gitscrum timer --format json 2>/dev/null | jq -r '.data[0].uuid // empty' 2>/dev/null)
    
    if [ -z "$ACTIVE_TIMER" ]; then
        echo "⏱️  Starting timer for $TASK_CODE..."
        gitscrum timer start "$TASK_CODE" -m "Switched to branch $BRANCH_NAME"
    else
        echo "⏱️  Timer already running"
    fi
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

exit 0

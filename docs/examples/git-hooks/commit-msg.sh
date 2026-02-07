#!/bin/bash
# commit-msg hook - Prepends task code to commit message
#
# This hook extracts the task code from the current branch name
# and prepends it to the commit message if not already present.
#
# Example:
#   Branch: feature/GS-123-login-page
#   Input:  "Add login form"
#   Output: "[GS-123] Add login form"
#
# Installation:
#   cp commit-msg.sh .git/hooks/commit-msg
#   chmod +x .git/hooks/commit-msg

COMMIT_MSG_FILE=$1

# Get current branch name
BRANCH_NAME=$(git symbolic-ref --short HEAD 2>/dev/null)

# Extract task code (e.g., GS-123, TK-456, PROJ-789)
TASK_CODE=$(echo "$BRANCH_NAME" | grep -oE '[A-Z]{2,5}-[0-9]+' | head -1)

if [ -n "$TASK_CODE" ]; then
    # Read current commit message
    COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")
    
    # Check if task code is already in the message
    if ! echo "$COMMIT_MSG" | grep -qE "^\[?$TASK_CODE\]?"; then
        # Prepend task code
        echo "[$TASK_CODE] $COMMIT_MSG" > "$COMMIT_MSG_FILE"
        echo "✅ Prepended task code: [$TASK_CODE]"
    fi
fi

exit 0

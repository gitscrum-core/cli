#!/bin/bash
# prepare-commit-msg hook - Adds task code template to commit message
#
# This hook prepares the commit message template with task information.
# Shows task title and code as a comment for reference.
#
# Installation:
#   cp prepare-commit-msg.sh .git/hooks/prepare-commit-msg
#   chmod +x .git/hooks/prepare-commit-msg

COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2

# Only run for regular commits (not merges, squashes, etc.)
if [ "$COMMIT_SOURCE" != "" ] && [ "$COMMIT_SOURCE" != "message" ]; then
    exit 0
fi

# Get current branch name
BRANCH_NAME=$(git symbolic-ref --short HEAD 2>/dev/null)

# Extract task code
TASK_CODE=$(echo "$BRANCH_NAME" | grep -oE '[A-Z]{2,5}-[0-9]+' | head -1)

if [ -n "$TASK_CODE" ]; then
    # Try to get task info from GitScrum
    TASK_INFO=$(gitscrum tasks view "$TASK_CODE" --format json 2>/dev/null)
    
    if [ -n "$TASK_INFO" ]; then
        TASK_TITLE=$(echo "$TASK_INFO" | jq -r '.title // empty' 2>/dev/null)
    fi
    
    # Create commit message template
    {
        echo "[$TASK_CODE] "
        echo ""
        echo "# Task: $TASK_CODE"
        if [ -n "$TASK_TITLE" ]; then
            echo "# Title: $TASK_TITLE"
        fi
        echo "#"
        echo "# Conventional commit types:"
        echo "#   feat:     A new feature"
        echo "#   fix:      A bug fix"
        echo "#   docs:     Documentation only"
        echo "#   style:    Formatting, missing semicolons, etc."
        echo "#   refactor: Code change that neither fixes nor adds"
        echo "#   test:     Adding missing tests"
        echo "#   chore:    Maintenance tasks"
        echo ""
        cat "$COMMIT_MSG_FILE"
    } > "$COMMIT_MSG_FILE.tmp"
    
    mv "$COMMIT_MSG_FILE.tmp" "$COMMIT_MSG_FILE"
fi

exit 0

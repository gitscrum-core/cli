#!/bin/bash
# post-merge hook - Complete task when merged to main
#
# Automatically completes the task when the branch is merged.
# Also stops any running timer for the task.
#
# Installation:
#   cp post-merge.sh .git/hooks/post-merge
#   chmod +x .git/hooks/post-merge

SQUASH_MERGE=$1

# Get the branch that was merged (before the merge)
MERGED_BRANCH=$(git reflog -1 | grep -oE 'checkout: moving from [^ ]+' | cut -d' ' -f4)

# Extract task code from the merged branch
TASK_CODE=$(echo "$MERGED_BRANCH" | grep -oE '[A-Z]{2,5}-[0-9]+' | head -1)

# Get current branch
CURRENT_BRANCH=$(git symbolic-ref --short HEAD 2>/dev/null)

# Only act on merges to main/develop
if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "develop" ]; then
    exit 0
fi

if [ -z "$TASK_CODE" ]; then
    exit 0
fi

if ! command -v gitscrum &> /dev/null; then
    exit 0
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 Merged: $TASK_CODE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Stop timer if running for this task
gitscrum timer stop 2>/dev/null

# Complete the task
gitscrum tasks complete "$TASK_CODE" 2>/dev/null
echo "✅ Task $TASK_CODE marked as complete"

# Add merge note
COMMIT_SHA=$(git rev-parse HEAD)
gitscrum tasks update "$TASK_CODE" \
    --note "Merged to $CURRENT_BRANCH (${COMMIT_SHA:0:8})" \
    2>/dev/null

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

exit 0

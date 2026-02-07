#!/bin/bash
# pre-push hook - Validate task exists before pushing
#
# Ensures the branch's task code is valid before allowing push.
# Prevents orphaned branches that don't link to any task.
#
# Installation:
#   cp pre-push.sh .git/hooks/pre-push
#   chmod +x .git/hooks/pre-push

REMOTE=$1
URL=$2

# Get current branch name
BRANCH_NAME=$(git symbolic-ref --short HEAD 2>/dev/null)

# Some branches don't need task codes
EXEMPT_BRANCHES="main master develop staging production release hotfix"
for exempt in $EXEMPT_BRANCHES; do
    if [ "$BRANCH_NAME" = "$exempt" ]; then
        exit 0
    fi
done

# Extract task code
TASK_CODE=$(echo "$BRANCH_NAME" | grep -oE '[A-Z]{2,5}-[0-9]+' | head -1)

if [ -z "$TASK_CODE" ]; then
    echo "⚠️  Warning: Branch '$BRANCH_NAME' has no task code"
    echo "   Consider using format: feature/GS-123-description"
    echo ""
    # Allow push but warn (change to exit 1 to block)
    exit 0
fi

# Verify task exists in GitScrum
if command -v gitscrum &> /dev/null; then
    TASK_EXISTS=$(gitscrum tasks view "$TASK_CODE" --format json 2>/dev/null | jq -r '.uuid // empty' 2>/dev/null)
    
    if [ -z "$TASK_EXISTS" ]; then
        echo "❌ Error: Task $TASK_CODE not found in GitScrum"
        echo "   Create the task first or check the task code"
        echo ""
        echo "   To create: gitscrum tasks create -t \"Task title\" -p project-slug"
        exit 1
    fi
    
    echo "✅ Task $TASK_CODE validated"
fi

exit 0

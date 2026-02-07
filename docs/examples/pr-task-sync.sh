#!/bin/bash
# pr-task-sync.sh - Bidirectional sync between PRs and GitScrum tasks
#
# This script synchronizes GitHub/GitLab PRs with GitScrum tasks:
# - Updates task status based on PR state
# - Adds PR links to task
# - Comments on PR with task info
# - Auto-detects task code from branch name
#
# Usage:
#   ./pr-task-sync.sh --pr-url https://github.com/org/repo/pull/123
#   ./pr-task-sync.sh --scan-repo    # Scan all open PRs
#
# Environment Variables:
#   GITSCRUM_ACCESS_TOKEN     - GitScrum API token
#   GITHUB_TOKEN       - GitHub API token (for commenting)
#
# Note: This script is typically run by CI/CD, not manually.

set -e

# Parse arguments
PR_URL=""
SCAN_REPO=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --pr-url)    PR_URL="$2"; shift 2 ;;
        --scan-repo) SCAN_REPO=true; shift ;;
        *)           shift ;;
    esac
done

# ============================================
# Process Single PR
# ============================================
process_pr() {
    local PR_URL=$1
    local PR_STATE=$2  # open, closed, merged
    local PR_TITLE=$3
    local SOURCE_BRANCH=$4
    
    echo "Processing PR: $PR_URL"
    
    # Extract task code from branch
    TASK_CODE=$(echo "$SOURCE_BRANCH" | grep -oE '[A-Z]{2,5}-[0-9]+' | head -1)
    
    if [ -z "$TASK_CODE" ]; then
        echo "  ⚠️  No task code found in branch: $SOURCE_BRANCH"
        return 0
    fi
    
    echo "  📌 Task: $TASK_CODE"
    
    # Verify task exists
    TASK_EXISTS=$(gitscrum tasks view "$TASK_CODE" --format json 2>/dev/null | jq -r '.uuid // empty')
    if [ -z "$TASK_EXISTS" ]; then
        echo "  ❌ Task $TASK_CODE not found"
        return 1
    fi
    
    # Link PR to task
    gitscrum tasks pr "$TASK_CODE" --url "$PR_URL" --title "$PR_TITLE" 2>/dev/null || true
    
    # Update task status based on PR state
    case $PR_STATE in
        open)
            echo "  → Moving to In Progress"
            gitscrum tasks move "$TASK_CODE" --column "in-progress" 2>/dev/null || true
            ;;
        review)
            echo "  → Moving to Review"
            gitscrum tasks move "$TASK_CODE" --column "review" 2>/dev/null || true
            ;;
        approved)
            echo "  → Moving to Approved"
            gitscrum tasks move "$TASK_CODE" --column "approved" 2>/dev/null || true
            ;;
        merged)
            echo "  → Completing task"
            gitscrum tasks complete "$TASK_CODE" 2>/dev/null || true
            ;;
        closed)
            # PR closed without merge - add note
            gitscrum tasks update "$TASK_CODE" \
                --note "PR closed without merge: $PR_URL" 2>/dev/null || true
            ;;
    esac
    
    echo "  ✅ Synced"
}

# ============================================
# Scan Repository
# ============================================
scan_repo() {
    echo "Scanning repository for open PRs/MRs..."
    
    # Get origin URL to determine platform
    REMOTE_URL=$(git remote get-url origin 2>/dev/null || echo "")
    
    if [[ "$REMOTE_URL" == *"github.com"* ]]; then
        # ============================================
        # GitHub
        # ============================================
        REPO=$(echo "$REMOTE_URL" | sed -E 's/.*github.com[:/](.+)(\.git)?$/\1/' | sed 's/.git$//')
        
        if [ -z "$GITHUB_TOKEN" ]; then
            echo "❌ GITHUB_TOKEN required for scanning"
            exit 1
        fi
        
        echo "Platform: GitHub ($REPO)"
        
        # Get open PRs
        PRS=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
            "https://api.github.com/repos/$REPO/pulls?state=open")
        
        echo "$PRS" | jq -c '.[]' | while read -r pr; do
            PR_URL=$(echo "$pr" | jq -r '.html_url')
            PR_TITLE=$(echo "$pr" | jq -r '.title')
            SOURCE_BRANCH=$(echo "$pr" | jq -r '.head.ref')
            
            process_pr "$PR_URL" "open" "$PR_TITLE" "$SOURCE_BRANCH"
        done
        
    elif [[ "$REMOTE_URL" == *"gitlab"* ]]; then
        # ============================================
        # GitLab
        # ============================================
        # Extract project path from URL
        PROJECT_PATH=$(echo "$REMOTE_URL" | sed -E 's/.*gitlab[^/]*[:/](.+)(\.git)?$/\1/' | sed 's/.git$//')
        PROJECT_PATH_ENCODED=$(echo "$PROJECT_PATH" | sed 's/\//%2F/g')
        
        # Determine GitLab host (gitlab.com or self-hosted)
        if [[ "$REMOTE_URL" == *"gitlab.com"* ]]; then
            GITLAB_HOST="https://gitlab.com"
        else
            GITLAB_HOST=$(echo "$REMOTE_URL" | grep -oE 'https?://[^/]+')
        fi
        
        if [ -z "$GITLAB_TOKEN" ]; then
            echo "❌ GITLAB_TOKEN required for scanning"
            exit 1
        fi
        
        echo "Platform: GitLab ($PROJECT_PATH)"
        
        # Get open MRs
        MRS=$(curl -s -H "PRIVATE-TOKEN: $GITLAB_TOKEN" \
            "$GITLAB_HOST/api/v4/projects/$PROJECT_PATH_ENCODED/merge_requests?state=opened")
        
        echo "$MRS" | jq -c '.[]' | while read -r mr; do
            MR_URL=$(echo "$mr" | jq -r '.web_url')
            MR_TITLE=$(echo "$mr" | jq -r '.title')
            SOURCE_BRANCH=$(echo "$mr" | jq -r '.source_branch')
            
            process_pr "$MR_URL" "open" "$MR_TITLE" "$SOURCE_BRANCH"
        done
        
    elif [[ "$REMOTE_URL" == *"bitbucket"* ]]; then
        # ============================================
        # Bitbucket
        # ============================================
        # Extract workspace and repo from URL
        WORKSPACE=$(echo "$REMOTE_URL" | sed -E 's/.*bitbucket.org[:/]([^/]+).*/\1/')
        REPO_SLUG=$(echo "$REMOTE_URL" | sed -E 's/.*bitbucket.org[:/][^/]+\/([^.]+).*/\1/')
        
        if [ -z "$BITBUCKET_TOKEN" ]; then
            echo "❌ BITBUCKET_TOKEN required for scanning"
            echo "   Create App Password at: https://bitbucket.org/account/settings/app-passwords/"
            exit 1
        fi
        
        echo "Platform: Bitbucket ($WORKSPACE/$REPO_SLUG)"
        
        # Get open PRs (using App Password auth)
        PRS=$(curl -s -u "$BITBUCKET_USER:$BITBUCKET_TOKEN" \
            "https://api.bitbucket.org/2.0/repositories/$WORKSPACE/$REPO_SLUG/pullrequests?state=OPEN")
        
        echo "$PRS" | jq -c '.values[]' | while read -r pr; do
            PR_URL=$(echo "$pr" | jq -r '.links.html.href')
            PR_TITLE=$(echo "$pr" | jq -r '.title')
            SOURCE_BRANCH=$(echo "$pr" | jq -r '.source.branch.name')
            
            process_pr "$PR_URL" "open" "$PR_TITLE" "$SOURCE_BRANCH"
        done
        
    else
        echo "❌ Unknown git platform: $REMOTE_URL"
        echo ""
        echo "Supported platforms:"
        echo "  - GitHub (github.com)"
        echo "  - GitLab (gitlab.com or self-hosted)"
        echo "  - Bitbucket (bitbucket.org)"
        exit 1
    fi
}

# ============================================
# Main
# ============================================

if [ "$SCAN_REPO" = true ]; then
    scan_repo
elif [ -n "$PR_URL" ]; then
    # For single PR, we need to fetch its info
    # This is typically called from CI with full context
    echo "Single PR processing requires CI context"
    echo "Use GitHub Actions or GitLab CI integration instead"
    echo "See: docs/examples/github-actions.yml"
else
    echo "Usage:"
    echo "  ./pr-task-sync.sh --pr-url <url>  # Process single PR"
    echo "  ./pr-task-sync.sh --scan-repo     # Scan all open PRs"
fi

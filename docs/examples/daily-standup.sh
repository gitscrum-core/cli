#!/bin/bash
# daily-standup.sh - Generate daily standup report from GitScrum
#
# This script generates a standup report showing:
# - What you did yesterday
# - What you're working on today
# - Any blockers
#
# Usage:
#   ./daily-standup.sh                    # Output to terminal
#   ./daily-standup.sh --slack            # Post to Slack
#   ./daily-standup.sh --email            # Send via email
#   ./daily-standup.sh --copy             # Copy to clipboard
#
# Environment Variables:
#   GITSCRUM_ACCESS_TOKEN        - API token (required)
#   GITSCRUM_WORKSPACE    - Workspace slug
#   SLACK_WEBHOOK_URL     - Slack webhook for --slack
#
# Cron example (9am every weekday):
#   0 9 * * 1-5 /path/to/daily-standup.sh --slack

set -e

# Configuration
YESTERDAY=$(date -d "yesterday" +%Y-%m-%d 2>/dev/null || date -v-1d +%Y-%m-%d)
TODAY=$(date +%Y-%m-%d)

# Parse arguments
OUTPUT_MODE="terminal"
while [[ $# -gt 0 ]]; do
    case $1 in
        --slack) OUTPUT_MODE="slack"; shift ;;
        --email) OUTPUT_MODE="email"; shift ;;
        --copy)  OUTPUT_MODE="clipboard"; shift ;;
        --json)  OUTPUT_MODE="json"; shift ;;
        *)       shift ;;
    esac
done

# Check GitScrum CLI
if ! command -v gitscrum &> /dev/null; then
    echo "Error: gitscrum CLI not found"
    echo "Install: curl -sL https://raw.githubusercontent.com/gitscrum-core/cli/main/install.sh | sh"
    exit 1
fi

# Get user info
USER_NAME=$(gitscrum auth whoami --format json | jq -r '.name // .username')

echo "Generating standup report for $USER_NAME..."
echo ""

# ============================================
# Gather Data
# ============================================

# Tasks completed yesterday
COMPLETED_YESTERDAY=$(gitscrum tasks list \
    --assignee @me \
    --completed-after "$YESTERDAY" \
    --completed-before "$TODAY" \
    --format json 2>/dev/null || echo '{"data":[]}')

# Time logged yesterday
TIME_YESTERDAY=$(gitscrum timer log \
    --date "$YESTERDAY" \
    --format json 2>/dev/null || echo '{"data":[]}')

# Tasks in progress
IN_PROGRESS=$(gitscrum tasks list \
    --assignee @me \
    --status "in-progress,review" \
    --format json 2>/dev/null || echo '{"data":[]}')

# Blockers
BLOCKERS=$(gitscrum tasks list \
    --assignee @me \
    --filter blocker \
    --format json 2>/dev/null || echo '{"data":[]}')

# ============================================
# Format Report
# ============================================

generate_report() {
    echo "# 📋 Daily Standup - $TODAY"
    echo ""
    echo "**$USER_NAME**"
    echo ""
    
    # Yesterday
    echo "## ✅ Yesterday"
    COMPLETED_COUNT=$(echo "$COMPLETED_YESTERDAY" | jq '.data | length')
    if [ "$COMPLETED_COUNT" -gt 0 ]; then
        echo "$COMPLETED_YESTERDAY" | jq -r '.data[] | "- [\(.code)] \(.title)"'
    else
        echo "_No tasks completed_"
    fi
    
    # Time summary
    TOTAL_MINUTES=$(echo "$TIME_YESTERDAY" | jq '[.data[].time.duration_minutes // 0] | add // 0')
    HOURS=$((TOTAL_MINUTES / 60))
    MINS=$((TOTAL_MINUTES % 60))
    if [ "$TOTAL_MINUTES" -gt 0 ]; then
        echo ""
        echo "⏱️ Time logged: ${HOURS}h ${MINS}m"
    fi
    echo ""
    
    # Today
    echo "## 🎯 Today"
    IN_PROGRESS_COUNT=$(echo "$IN_PROGRESS" | jq '.data | length')
    if [ "$IN_PROGRESS_COUNT" -gt 0 ]; then
        echo "$IN_PROGRESS" | jq -r '.data[] | "- [\(.code)] \(.title) _(\(.workflow.title))_"'
    else
        echo "_No tasks in progress_"
    fi
    echo ""
    
    # Blockers
    BLOCKER_COUNT=$(echo "$BLOCKERS" | jq '.data | length')
    if [ "$BLOCKER_COUNT" -gt 0 ]; then
        echo "## 🚨 Blockers"
        echo "$BLOCKERS" | jq -r '.data[] | "- [\(.code)] \(.title)"'
        echo ""
    fi
}

REPORT=$(generate_report)

# ============================================
# Output
# ============================================

case $OUTPUT_MODE in
    terminal)
        echo "$REPORT"
        ;;
    
    json)
        jq -n \
            --arg user "$USER_NAME" \
            --arg date "$TODAY" \
            --argjson completed "$COMPLETED_YESTERDAY" \
            --argjson inProgress "$IN_PROGRESS" \
            --argjson blockers "$BLOCKERS" \
            '{
                user: $user,
                date: $date,
                yesterday: $completed.data,
                today: $inProgress.data,
                blockers: $blockers.data
            }'
        ;;
    
    slack)
        if [ -z "$SLACK_WEBHOOK_URL" ]; then
            echo "Error: SLACK_WEBHOOK_URL not set"
            exit 1
        fi
        
        # Convert to Slack mrkdwn format
        SLACK_TEXT=$(echo "$REPORT" | sed 's/^## /\n*/' | sed 's/^# /*/' | sed 's/$/*\n/' | tr '\n' ' ')
        
        curl -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-type: application/json' \
            -d "{\"text\": \"$SLACK_TEXT\"}" \
            --silent
        
        echo "✅ Posted to Slack"
        ;;
    
    clipboard)
        if command -v pbcopy &> /dev/null; then
            echo "$REPORT" | pbcopy
        elif command -v xclip &> /dev/null; then
            echo "$REPORT" | xclip -selection clipboard
        elif command -v clip.exe &> /dev/null; then
            echo "$REPORT" | clip.exe
        else
            echo "Error: No clipboard command found"
            exit 1
        fi
        echo "✅ Copied to clipboard"
        ;;
    
    email)
        echo "Email output not implemented. Use with mail command:"
        echo "  ./daily-standup.sh | mail -s 'Standup $TODAY' team@company.com"
        ;;
esac

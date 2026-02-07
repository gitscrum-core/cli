#!/bin/bash
# time-report.sh - Generate weekly time tracking report
#
# This script generates a detailed time tracking report including:
# - Total hours per project
# - Daily breakdown
# - Task-level details
#
# Usage:
#   ./time-report.sh                    # This week
#   ./time-report.sh --last-week        # Previous week
#   ./time-report.sh --email            # Send via email
#   ./time-report.sh --csv              # Export to CSV
#
# Environment Variables:
#   GITSCRUM_ACCESS_TOKEN        - API token (required)
#   GITSCRUM_WORKSPACE    - Workspace slug
#   SMTP_HOST/USER/PASS   - For email sending
#
# Cron example (Friday 5pm):
#   0 17 * * 5 /path/to/time-report.sh --email

set -e

# Parse arguments
WEEK="current"
OUTPUT_MODE="terminal"
while [[ $# -gt 0 ]]; do
    case $1 in
        --last-week) WEEK="last"; shift ;;
        --email)     OUTPUT_MODE="email"; shift ;;
        --csv)       OUTPUT_MODE="csv"; shift ;;
        --json)      OUTPUT_MODE="json"; shift ;;
        --slack)     OUTPUT_MODE="slack"; shift ;;
        *)           shift ;;
    esac
done

# Calculate date range
if [ "$WEEK" = "last" ]; then
    # Last week: Monday to Sunday
    START_DATE=$(date -d "last monday -1 week" +%Y-%m-%d 2>/dev/null || date -v-1w -v-mon +%Y-%m-%d)
    END_DATE=$(date -d "last sunday" +%Y-%m-%d 2>/dev/null || date -v-sun +%Y-%m-%d)
else
    # This week: Monday to today
    START_DATE=$(date -d "last monday" +%Y-%m-%d 2>/dev/null || date -v-mon +%Y-%m-%d)
    END_DATE=$(date +%Y-%m-%d)
fi

echo "Time Report: $START_DATE to $END_DATE"
echo "==========================================="
echo ""

# Get user info
USER_NAME=$(gitscrum auth whoami --format json | jq -r '.name // .username')

# ============================================
# Fetch Time Data
# ============================================

TIME_DATA=$(gitscrum timer log \
    --from "$START_DATE" \
    --to "$END_DATE" \
    --format json 2>/dev/null || echo '{"data":[]}')

ENTRIES_COUNT=$(echo "$TIME_DATA" | jq '.data | length')

if [ "$ENTRIES_COUNT" -eq 0 ]; then
    echo "No time entries found for this period."
    exit 0
fi

# ============================================
# Generate Report
# ============================================

generate_report() {
    echo "# ⏱️ Weekly Time Report"
    echo "**$USER_NAME** | $START_DATE to $END_DATE"
    echo ""
    
    # Total hours
    TOTAL_MINUTES=$(echo "$TIME_DATA" | jq '[.data[].time.duration_minutes // 0] | add')
    TOTAL_HOURS=$((TOTAL_MINUTES / 60))
    TOTAL_MINS=$((TOTAL_MINUTES % 60))
    echo "## Summary"
    echo "- **Total Time:** ${TOTAL_HOURS}h ${TOTAL_MINS}m"
    echo "- **Entries:** $ENTRIES_COUNT"
    echo ""
    
    # By Project
    echo "## By Project"
    echo "$TIME_DATA" | jq -r '
        .data 
        | group_by(.project.slug) 
        | map({
            project: .[0].project.name,
            minutes: [.[].time.duration_minutes] | add
        })
        | sort_by(.minutes) | reverse
        | .[]
        | "- **\(.project):** \(.minutes / 60 | floor)h \(.minutes % 60)m"
    '
    echo ""
    
    # By Day
    echo "## By Day"
    echo "$TIME_DATA" | jq -r '
        .data 
        | group_by(.time.start.date | split("T")[0])
        | map({
            date: .[0].time.start.date | split("T")[0],
            minutes: [.[].time.duration_minutes] | add
        })
        | sort_by(.date)
        | .[]
        | "- \(.date): \(.minutes / 60 | floor)h \(.minutes % 60)m"
    '
    echo ""
    
    # Top Tasks
    echo "## Top Tasks (by time)"
    echo "$TIME_DATA" | jq -r '
        .data
        | group_by(.task.code)
        | map({
            code: .[0].task.code,
            title: .[0].task.title,
            minutes: [.[].time.duration_minutes] | add
        })
        | sort_by(.minutes) | reverse
        | .[0:10]
        | .[]
        | "- [\(.code)] \(.title) - \(.minutes / 60 | floor)h \(.minutes % 60)m"
    '
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
        echo "$TIME_DATA" | jq '{
            period: {start: $start, end: $end},
            total_minutes: [.data[].time.duration_minutes] | add,
            entries: .data | length,
            by_project: (.data | group_by(.project.slug) | map({
                project: .[0].project.name,
                minutes: [.[].time.duration_minutes] | add
            })),
            entries: .data
        }' --arg start "$START_DATE" --arg end "$END_DATE"
        ;;
    
    csv)
        echo "Date,Project,Task Code,Task Title,Duration (min),Comment"
        echo "$TIME_DATA" | jq -r '.data[] | 
            "\(.time.start.date | split("T")[0]),\(.project.name // ""),\(.task.code // ""),\"\(.task.title // "")\",\(.time.duration_minutes),\"\(.comment // "")\""
        '
        ;;
    
    slack)
        if [ -z "$SLACK_WEBHOOK_URL" ]; then
            echo "Error: SLACK_WEBHOOK_URL not set"
            exit 1
        fi
        
        TOTAL_HOURS=$(($(echo "$TIME_DATA" | jq '[.data[].time.duration_minutes] | add') / 60))
        
        PAYLOAD=$(jq -n \
            --arg title "⏱️ Weekly Time Report" \
            --arg user "$USER_NAME" \
            --arg period "$START_DATE to $END_DATE" \
            --arg hours "${TOTAL_HOURS}h" \
            '{
                "blocks": [
                    {"type": "header", "text": {"type": "plain_text", "text": $title}},
                    {"type": "section", "fields": [
                        {"type": "mrkdwn", "text": "*User:*\n\($user)"},
                        {"type": "mrkdwn", "text": "*Period:*\n\($period)"},
                        {"type": "mrkdwn", "text": "*Total:*\n\($hours)"}
                    ]}
                ]
            }')
        
        curl -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-type: application/json' \
            -d "$PAYLOAD" \
            --silent
        
        echo "✅ Posted to Slack"
        ;;
    
    email)
        if [ -z "$EMAIL_TO" ]; then
            echo "Error: EMAIL_TO not set"
            echo "Usage: EMAIL_TO=you@company.com ./time-report.sh --email"
            exit 1
        fi
        
        echo "$REPORT" | mail -s "Time Report: $START_DATE to $END_DATE" "$EMAIL_TO"
        echo "✅ Sent to $EMAIL_TO"
        ;;
esac

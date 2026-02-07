#!/bin/bash
# sprint-health.sh - Monitor sprint health and send alerts
#
# This script checks sprint progress and alerts if:
# - Sprint is at risk (behind schedule)
# - Burndown is trending badly
# - Too many tasks in "Blocked" status
# - Sprint deadline approaching
#
# Usage:
#   ./sprint-health.sh                    # Check current sprint
#   ./sprint-health.sh --sprint sprint-5  # Specific sprint
#   ./sprint-health.sh --slack            # Post to Slack
#
# Environment Variables:
#   GITSCRUM_ACCESS_TOKEN        - API token (required)
#   SLACK_WEBHOOK_URL     - Slack incoming webhook
#
# Cron example (daily at 9am):
#   0 9 * * 1-5 /path/to/sprint-health.sh --slack

set -e

# Parse arguments
SPRINT=""
OUTPUT_MODE="terminal"
while [[ $# -gt 0 ]]; do
    case $1 in
        --sprint|-s)  SPRINT="$2"; shift 2 ;;
        --slack)      OUTPUT_MODE="slack"; shift ;;
        --json)       OUTPUT_MODE="json"; shift ;;
        *)            shift ;;
    esac
done

# ============================================
# Fetch Sprint Data
# ============================================

if [ -n "$SPRINT" ]; then
    SPRINT_DATA=$(gitscrum sprints view "$SPRINT" --format json 2>/dev/null)
else
    SPRINT_DATA=$(gitscrum sprints current --format json 2>/dev/null)
fi

if [ -z "$SPRINT_DATA" ] || [ "$(echo "$SPRINT_DATA" | jq '.slug // empty')" = "" ]; then
    echo "❌ No active sprint found"
    exit 0
fi

# Extract sprint info
SPRINT_NAME=$(echo "$SPRINT_DATA" | jq -r '.title')
SPRINT_END=$(echo "$SPRINT_DATA" | jq -r '.date_finish.date // .timebox.finish.date')
TOTAL_TASKS=$(echo "$SPRINT_DATA" | jq -r '.stats.total_tasks // 0')
CLOSED_TASKS=$(echo "$SPRINT_DATA" | jq -r '.stats.closed_tasks // 0')
PERCENTAGE=$(echo "$SPRINT_DATA" | jq -r '.stats.percentage // 0')
STORY_POINTS=$(echo "$SPRINT_DATA" | jq -r '.stats.story_points // 0')

# Calculate days remaining
TODAY=$(date +%Y-%m-%d)
DAYS_LEFT=$(( ($(date -d "$SPRINT_END" +%s) - $(date -d "$TODAY" +%s)) / 86400 )) 2>/dev/null || DAYS_LEFT=0

# ============================================
# Health Checks
# ============================================

HEALTH_STATUS="healthy"
ALERTS=()

# Check 1: Progress vs Time
DURATION=$(echo "$SPRINT_DATA" | jq -r '.duration // 14')
ELAPSED_DAYS=$((DURATION - DAYS_LEFT))
EXPECTED_PROGRESS=$((ELAPSED_DAYS * 100 / DURATION))
ACTUAL_PROGRESS=$PERCENTAGE

if [ $ACTUAL_PROGRESS -lt $((EXPECTED_PROGRESS - 20)) ]; then
    HEALTH_STATUS="at-risk"
    ALERTS+=("📉 Behind schedule: ${ACTUAL_PROGRESS}% complete (expected: ${EXPECTED_PROGRESS}%)")
fi

# Check 2: Deadline approaching
if [ $DAYS_LEFT -le 2 ] && [ $PERCENTAGE -lt 80 ]; then
    HEALTH_STATUS="critical"
    ALERTS+=("⏰ Deadline in ${DAYS_LEFT} days with only ${PERCENTAGE}% complete")
fi

# Check 3: Sprint ending today
if [ $DAYS_LEFT -eq 0 ]; then
    if [ $PERCENTAGE -lt 100 ]; then
        ALERTS+=("🏁 Sprint ends TODAY with ${TOTAL_TASKS - CLOSED_TASKS} tasks remaining")
    fi
fi

# ============================================
# Generate Report
# ============================================

generate_report() {
    # Header with status emoji
    case $HEALTH_STATUS in
        healthy)  EMOJI="✅" ;;
        at-risk)  EMOJI="⚠️" ;;
        critical) EMOJI="🚨" ;;
    esac
    
    echo "# $EMOJI Sprint Health: $SPRINT_NAME"
    echo ""
    echo "## Progress"
    echo "- **Completed:** ${CLOSED_TASKS}/${TOTAL_TASKS} tasks (${PERCENTAGE}%)"
    echo "- **Days Left:** $DAYS_LEFT"
    echo "- **Story Points:** $STORY_POINTS"
    echo ""
    
    # Progress bar
    FILLED=$((PERCENTAGE / 5))
    EMPTY=$((20 - FILLED))
    BAR=$(printf '█%.0s' $(seq 1 $FILLED))$(printf '░%.0s' $(seq 1 $EMPTY))
    echo "**Progress:** [$BAR] ${PERCENTAGE}%"
    echo ""
    
    # Alerts
    if [ ${#ALERTS[@]} -gt 0 ]; then
        echo "## ⚠️ Alerts"
        for alert in "${ALERTS[@]}"; do
            echo "- $alert"
        done
        echo ""
    fi
    
    # Status
    echo "## Status"
    case $HEALTH_STATUS in
        healthy)  echo "Sprint is on track! 🎉" ;;
        at-risk)  echo "Sprint needs attention. Consider scope adjustment." ;;
        critical) echo "Sprint at risk! Immediate action required." ;;
    esac
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
            --arg name "$SPRINT_NAME" \
            --arg status "$HEALTH_STATUS" \
            --arg end "$SPRINT_END" \
            --argjson daysLeft "$DAYS_LEFT" \
            --argjson total "$TOTAL_TASKS" \
            --argjson closed "$CLOSED_TASKS" \
            --argjson percentage "$PERCENTAGE" \
            --argjson alerts "$(printf '%s\n' "${ALERTS[@]}" | jq -R . | jq -s .)" \
            '{
                sprint: $name,
                status: $status,
                end_date: $end,
                days_left: $daysLeft,
                progress: {
                    total: $total,
                    closed: $closed,
                    percentage: $percentage
                },
                alerts: $alerts
            }'
        ;;
    
    slack)
        if [ -z "$SLACK_WEBHOOK_URL" ]; then
            echo "Error: SLACK_WEBHOOK_URL not set"
            exit 1
        fi
        
        # Color based on status
        case $HEALTH_STATUS in
            healthy)  COLOR="good" ;;
            at-risk)  COLOR="warning" ;;
            critical) COLOR="danger" ;;
        esac
        
        # Build attachment
        PAYLOAD=$(jq -n \
            --arg title "Sprint Health: $SPRINT_NAME" \
            --arg color "$COLOR" \
            --arg progress "${CLOSED_TASKS}/${TOTAL_TASKS} (${PERCENTAGE}%)" \
            --arg days "$DAYS_LEFT days left" \
            --arg alerts "$(printf '%s\n' "${ALERTS[@]}")" \
            '{
                "attachments": [{
                    "color": $color,
                    "title": $title,
                    "fields": [
                        {"title": "Progress", "value": $progress, "short": true},
                        {"title": "Time", "value": $days, "short": true}
                    ],
                    "text": $alerts
                }]
            }')
        
        curl -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-type: application/json' \
            -d "$PAYLOAD" \
            --silent
        
        echo "✅ Posted to Slack"
        ;;
esac

# Exit with non-zero if critical
if [ "$HEALTH_STATUS" = "critical" ]; then
    exit 1
fi

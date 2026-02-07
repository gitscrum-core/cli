#!/bin/bash
# blocker-alert.sh - Send Slack/Discord alerts for blocker tasks
#
# This script monitors for blocker tasks and sends notifications.
# Can run on a schedule to alert the team about blockers.
#
# Usage:
#   ./blocker-alert.sh                    # Check and notify
#   ./blocker-alert.sh --project my-proj  # Specific project
#   ./blocker-alert.sh --dry-run          # Preview without sending
#
# Environment Variables:
#   GITSCRUM_ACCESS_TOKEN        - API token (required)
#   SLACK_WEBHOOK_URL     - Slack incoming webhook
#   DISCORD_WEBHOOK_URL   - Discord webhook (alternative)
#   TEAMS_WEBHOOK_URL     - MS Teams webhook (alternative)
#
# Cron example (every 2 hours during work hours):
#   0 9-18/2 * * 1-5 /path/to/blocker-alert.sh

set -e

# Parse arguments
PROJECT=""
DRY_RUN=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --project|-p) PROJECT="$2"; shift 2 ;;
        --dry-run)    DRY_RUN=true; shift ;;
        *)            shift ;;
    esac
done

# ============================================
# Fetch Blockers
# ============================================

if [ -n "$PROJECT" ]; then
    BLOCKERS=$(gitscrum tasks list --filter blocker --project "$PROJECT" --format json 2>/dev/null)
else
    BLOCKERS=$(gitscrum tasks list --filter blocker --format json 2>/dev/null)
fi

BLOCKER_COUNT=$(echo "$BLOCKERS" | jq '.data | length')

if [ "$BLOCKER_COUNT" -eq 0 ]; then
    echo "✅ No blockers found"
    exit 0
fi

echo "🚨 Found $BLOCKER_COUNT blocker(s)"

# ============================================
# Format Message
# ============================================

format_slack_message() {
    BLOCKS=$(echo "$BLOCKERS" | jq '
        {
            "blocks": [
                {
                    "type": "header",
                    "text": {
                        "type": "plain_text",
                        "text": "🚨 Blocker Alert",
                        "emoji": true
                    }
                },
                {
                    "type": "section",
                    "text": {
                        "type": "mrkdwn",
                        "text": "There are \(.data | length) blocker task(s) requiring attention:"
                    }
                }
            ] + [
                .data[] | {
                    "type": "section",
                    "text": {
                        "type": "mrkdwn",
                        "text": "*[\(.code)]* \(.title)\n:bust_in_silhouette: \(.assignee.name // "Unassigned") | :file_folder: \(.project.name)"
                    },
                    "accessory": {
                        "type": "button",
                        "text": {
                            "type": "plain_text",
                            "text": "View Task"
                        },
                        "url": .url
                    }
                }
            ] + [
                {
                    "type": "divider"
                },
                {
                    "type": "context",
                    "elements": [
                        {
                            "type": "mrkdwn",
                            "text": "Resolve blockers ASAP to unblock the team 🏃"
                        }
                    ]
                }
            ]
        }
    ')
    echo "$BLOCKS"
}

format_discord_message() {
    EMBEDS=$(echo "$BLOCKERS" | jq '[
        .data[] | {
            "title": "[\(.code)] \(.title)",
            "description": "Assignee: \(.assignee.name // "Unassigned")\nProject: \(.project.name)",
            "color": 15158332,
            "url": .url
        }
    ]')
    
    jq -n \
        --argjson embeds "$EMBEDS" \
        '{
            "content": "🚨 **Blocker Alert** - \($embeds | length) task(s) need attention!",
            "embeds": $embeds
        }'
}

format_teams_message() {
    FACTS=$(echo "$BLOCKERS" | jq '[
        .data[] | {
            "name": "[\(.code)]",
            "value": "\(.title) - \(.assignee.name // "Unassigned")"
        }
    ]')
    
    jq -n \
        --argjson facts "$FACTS" \
        '{
            "@type": "MessageCard",
            "@context": "http://schema.org/extensions",
            "themeColor": "FF0000",
            "summary": "Blocker Alert",
            "sections": [{
                "activityTitle": "🚨 Blocker Tasks",
                "facts": $facts
            }]
        }'
}

# ============================================
# Send Notification
# ============================================

if $DRY_RUN; then
    echo ""
    echo "--- DRY RUN (not sending) ---"
    echo ""
    echo "$BLOCKERS" | jq -r '.data[] | "🔴 [\(.code)] \(.title) - \(.assignee.name // "Unassigned")"'
    exit 0
fi

# Slack
if [ -n "$SLACK_WEBHOOK_URL" ]; then
    PAYLOAD=$(format_slack_message)
    curl -X POST "$SLACK_WEBHOOK_URL" \
        -H 'Content-type: application/json' \
        -d "$PAYLOAD" \
        --silent
    echo "✅ Sent to Slack"
fi

# Discord
if [ -n "$DISCORD_WEBHOOK_URL" ]; then
    PAYLOAD=$(format_discord_message)
    curl -X POST "$DISCORD_WEBHOOK_URL" \
        -H 'Content-type: application/json' \
        -d "$PAYLOAD" \
        --silent
    echo "✅ Sent to Discord"
fi

# Teams
if [ -n "$TEAMS_WEBHOOK_URL" ]; then
    PAYLOAD=$(format_teams_message)
    curl -X POST "$TEAMS_WEBHOOK_URL" \
        -H 'Content-type: application/json' \
        -d "$PAYLOAD" \
        --silent
    echo "✅ Sent to Teams"
fi

if [ -z "$SLACK_WEBHOOK_URL" ] && [ -z "$DISCORD_WEBHOOK_URL" ] && [ -z "$TEAMS_WEBHOOK_URL" ]; then
    echo "⚠️  No webhook configured. Set SLACK_WEBHOOK_URL, DISCORD_WEBHOOK_URL, or TEAMS_WEBHOOK_URL"
    echo ""
    echo "$BLOCKERS" | jq -r '.data[] | "🔴 [\(.code)] \(.title) - \(.assignee.name // "Unassigned")"'
fi

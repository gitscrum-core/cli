#!/bin/bash
# task-import.sh - Bulk import tasks from CSV or Jira
#
# This script imports tasks from external sources:
# - CSV files
# - Jira export
# - Markdown task lists
#
# Usage:
#   ./task-import.sh --csv tasks.csv --project my-project
#   ./task-import.sh --jira export.json --project my-project
#   ./task-import.sh --md TODO.md --project my-project
#
# CSV Format:
#   Title,Description,Priority,Labels
#   "Fix login bug","Users can't login",high,"bug,urgent"
#
# Environment Variables:
#   GITSCRUM_ACCESS_TOKEN        - API token (required)
#   GITSCRUM_WORKSPACE    - Workspace slug
#   GITSCRUM_PROJECT      - Default project

set -e

# Parse arguments
CSV_FILE=""
JIRA_FILE=""
MD_FILE=""
PROJECT=""
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --csv)      CSV_FILE="$2"; shift 2 ;;
        --jira)     JIRA_FILE="$2"; shift 2 ;;
        --md)       MD_FILE="$2"; shift 2 ;;
        --project|-p) PROJECT="$2"; shift 2 ;;
        --dry-run)  DRY_RUN=true; shift ;;
        *)          shift ;;
    esac
done

if [ -z "$PROJECT" ]; then
    PROJECT="$GITSCRUM_PROJECT"
fi

if [ -z "$PROJECT" ]; then
    echo "❌ Project required: --project <slug> or GITSCRUM_PROJECT env"
    exit 1
fi

# ============================================
# Import from CSV
# ============================================
import_csv() {
    local file=$1
    echo "Importing from CSV: $file"
    echo ""
    
    # Skip header line
    tail -n +2 "$file" | while IFS=',' read -r title description priority labels; do
        # Remove quotes
        title=$(echo "$title" | tr -d '"')
        description=$(echo "$description" | tr -d '"')
        priority=$(echo "$priority" | tr -d '"')
        labels=$(echo "$labels" | tr -d '"')
        
        if [ -z "$title" ]; then
            continue
        fi
        
        echo "📝 $title"
        
        if $DRY_RUN; then
            echo "   [DRY RUN] Would create task"
            continue
        fi
        
        # Build command
        CMD="gitscrum tasks create -t \"$title\" -p \"$PROJECT\""
        
        if [ -n "$description" ]; then
            CMD="$CMD -d \"$description\""
        fi
        
        if [ -n "$priority" ]; then
            CMD="$CMD --priority \"$priority\""
        fi
        
        if [ -n "$labels" ]; then
            # Split labels and add each
            IFS=',' read -ra LABEL_ARRAY <<< "$labels"
            for label in "${LABEL_ARRAY[@]}"; do
                CMD="$CMD --label \"$(echo $label | xargs)\""
            done
        fi
        
        eval "$CMD" 2>/dev/null && echo "   ✅ Created" || echo "   ❌ Failed"
    done
}

# ============================================
# Import from Jira Export
# ============================================
import_jira() {
    local file=$1
    echo "Importing from Jira: $file"
    echo ""
    
    # Parse Jira export (assumes JSON export from Jira)
    jq -c '.issues[]' "$file" 2>/dev/null | while read -r issue; do
        title=$(echo "$issue" | jq -r '.fields.summary')
        description=$(echo "$issue" | jq -r '.fields.description // ""')
        priority=$(echo "$issue" | jq -r '.fields.priority.name // "medium"' | tr '[:upper:]' '[:lower:]')
        jira_key=$(echo "$issue" | jq -r '.key')
        
        echo "📝 [$jira_key] $title"
        
        if $DRY_RUN; then
            echo "   [DRY RUN] Would create task"
            continue
        fi
        
        # Map Jira priority to GitScrum
        case $priority in
            highest|blocker) priority="blocker" ;;
            high)            priority="high" ;;
            medium|normal)   priority="medium" ;;
            low|lowest)      priority="low" ;;
        esac
        
        gitscrum tasks create \
            -t "$title" \
            -d "$description" \
            -p "$PROJECT" \
            --priority "$priority" \
            --note "Imported from Jira: $jira_key" \
            2>/dev/null && echo "   ✅ Created" || echo "   ❌ Failed"
    done
}

# ============================================
# Import from Markdown
# ============================================
import_markdown() {
    local file=$1
    echo "Importing from Markdown: $file"
    echo ""
    
    # Parse markdown task lists
    # Format: - [ ] Task title
    #         - [x] Completed task (skip)
    grep -E '^\s*-\s*\[\s*\]' "$file" | while read -r line; do
        # Extract task title
        title=$(echo "$line" | sed -E 's/^\s*-\s*\[\s*\]\s*//')
        
        if [ -z "$title" ]; then
            continue
        fi
        
        echo "📝 $title"
        
        if $DRY_RUN; then
            echo "   [DRY RUN] Would create task"
            continue
        fi
        
        gitscrum tasks create \
            -t "$title" \
            -p "$PROJECT" \
            2>/dev/null && echo "   ✅ Created" || echo "   ❌ Failed"
    done
}

# ============================================
# Main
# ============================================

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📥 Task Import - Project: $PROJECT"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

if [ -n "$CSV_FILE" ]; then
    import_csv "$CSV_FILE"
elif [ -n "$JIRA_FILE" ]; then
    import_jira "$JIRA_FILE"
elif [ -n "$MD_FILE" ]; then
    import_markdown "$MD_FILE"
else
    echo "Usage:"
    echo "  ./task-import.sh --csv tasks.csv --project my-project"
    echo "  ./task-import.sh --jira export.json --project my-project"
    echo "  ./task-import.sh --md TODO.md --project my-project"
    echo ""
    echo "Options:"
    echo "  --dry-run    Preview without creating tasks"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

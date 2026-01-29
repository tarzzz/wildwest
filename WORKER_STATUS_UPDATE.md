# Worker Status Update Guide

## Overview

Workers (Claude instances) should update their `current_work` status in `session.json` every 10 seconds so the TUI shows real-time activity.

## Session JSON Structure

```json
{
  "id": "software-engineer-1769706081538",
  "persona_type": "software-engineer",
  "persona_name": "watt",
  "start_time": "2026-01-29T12:01:21.538601-05:00",
  "status": "active",
  "workspace_id": "ws-1769706081",
  "pid": 82678,
  "current_work": "Implementing user authentication API endpoints"
}
```

## How Workers Update Status

### Option 1: Using SessionManager API (Go)

```go
import "github.com/tarzzz/wildwest/pkg/session"

sm, _ := session.NewSessionManager(".database")
err := sm.UpdateCurrentWork("software-engineer-1769706081538",
    "Writing unit tests for auth module")
```

### Option 2: Direct JSON Update (Bash)

Workers can update their own session.json directly:

```bash
# Get your session ID from environment or directory name
SESSION_ID="software-engineer-1769706081538"
SESSION_FILE=".database/$SESSION_ID/session.json"

# Update current_work field
jq '.current_work = "Implementing REST API endpoints"' "$SESSION_FILE" > tmp.json
mv tmp.json "$SESSION_FILE"
```

### Option 3: Python Script

```python
import json
import time

def update_status(session_id, status_message):
    session_file = f".database/{session_id}/session.json"

    with open(session_file, 'r') as f:
        session = json.load(f)

    session['current_work'] = status_message

    with open(session_file, 'w') as f:
        json.dump(session, f, indent=2)

# Update every 10 seconds
while True:
    update_status("your-session-id", "Current task description")
    time.sleep(10)
```

## Best Practices

### Status Message Guidelines

- **Keep it concise**: Max 3 lines
- **Be specific**: Clear, actionable items
- **Current action**: What you're doing RIGHT NOW
- **Update regularly**: Every 10 seconds while working

### Examples

✅ **Good (1-3 lines):**
```
Implementing user auth API
Writing login endpoint
Testing with JWT tokens
```

```
Debugging CORS issue in API gateway
Checking nginx config
```

```
Reviewing architect's database schema
```

❌ **Bad:**
- "Working" (too vague)
- Multiple paragraphs (too long)
- "Doing stuff" (not specific)

## Auto-Update Script for Workers

Create a background task that updates status every 10 seconds:

```bash
#!/bin/bash
# Auto-update worker status every 10s

SESSION_ID=$(basename $(pwd))
SESSION_FILE="session.json"

while true; do
    # Get current task from tasks.md or other source
    CURRENT_TASK=$(head -1 tasks.md | sed 's/^#* //')

    # Update session.json
    jq --arg status "$CURRENT_TASK" '.current_work = $status' \
        "$SESSION_FILE" > tmp.json && mv tmp.json "$SESSION_FILE"

    sleep 10
done
```

## Integration with Orchestrator

The orchestrator's persona instructions should include:

```markdown
## Status Updates

Update your session.json every 10 seconds with what you're currently working on:

```bash
# Add this to your workflow
while true; do
    jq '.current_work = "Your current task here"' \
        .database/your-session-id/session.json > tmp.json
    mv tmp.json .database/your-session-id/session.json
    sleep 10
done &
```

This keeps the TUI updated with real-time status.
```

## TUI Display

The TUI reads `current_work` from session.json and displays it:

```
▶ 🔄  watt (Engineering)
  └─ Implementing user authentication API endpoints
─────────────────────────────────────────
  🔄  berners-lee (Engineering)
  └─ Writing unit tests for auth module
```

Updates appear in the TUI within 6 seconds (next poll cycle).

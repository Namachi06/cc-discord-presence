#!/bin/bash
# Statusline wrapper for Discord Rich Presence
# Saves statusline data for the Discord daemon, then pipes to original statusline (if exists)

DATA_FILE="$HOME/.claude/discord-presence-data.json"
SESSION_DIR="$HOME/.claude"
ORIGINAL_STATUSLINE="$HOME/.claude/statusline.sh"

# Read JSON from stdin
read -r json_data

# Save for Discord presence - shared file (atomic write, backward compat)
echo "$json_data" > "${DATA_FILE}.tmp" && mv "${DATA_FILE}.tmp" "$DATA_FILE"

# Also write per-session file for session focus feature
session_id=$(echo "$json_data" | grep -o '"session_id":"[^"]*"' | head -1 | sed 's/"session_id":"//;s/"//')
if [[ -n "$session_id" ]]; then
    SESSION_FILE="${SESSION_DIR}/discord-presence-session-${session_id}.json"
    echo "$json_data" > "${SESSION_FILE}.tmp" && mv "${SESSION_FILE}.tmp" "$SESSION_FILE"
fi

# If there's an original statusline, pass the data to it
if [[ -x "$ORIGINAL_STATUSLINE" ]]; then
    echo "$json_data" | "$ORIGINAL_STATUSLINE"
fi

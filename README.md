# cc-discord-presence

Discord Rich Presence for Claude Code. Show your coding session on Discord in real-time.

```
┌─────────────────────────────────────┐
│ Coding...                           │  App name (customizable)
│ Working on: my-project (main)       │  details_prefix, project_name, git_branch
│ Opus 4.5 | 1.5M tokens | $0.12      │  model_name, tokens, cost
│ 00:45:30 elapsed                    │  duration
│ [GitHub] [Portfolio]                │  buttons (visible to others only)
└─────────────────────────────────────┘
```

## Table of Contents

- [Quick Start](#quick-start)
- [Features](#features)
- [Configuration](#configuration)
  - [Examples](#examples)
  - [Format Templates](#format-templates)
  - [Field Reference](#field-reference)
- [Data Sources](#data-sources)
  - [Statusline Setup](#statusline-setup)
  - [Verify Your Data Source](#verify-your-data-source)
- [Advanced Features](#advanced-features)
  - [Idle Detection](#idle-detection)
  - [Auto-disable on Extended Idle](#auto-disable-on-extended-idle)
  - [Privacy Mode](#privacy-mode)
  - [Cost Alert](#cost-alert)
  - [Model Icons](#model-icons)
  - [Project Exclusion](#project-exclusion)
  - [Project Name Aliases](#project-name-aliases)
  - [Session Focus](#session-focus)
  - [Discord Buttons](#discord-buttons)
  - [Custom Discord App](#custom-discord-app)
- [Token Pricing](#token-pricing)
- [Platform Support](#platform-support)
- [Troubleshooting](#troubleshooting)
- [File Paths](#file-paths)
- [Building from Source](#building-from-source)
- [Uninstallation](#uninstallation)
- [Privacy](#privacy)
- [Contributing](#contributing)
- [License](#license)
- [Acknowledgments](#acknowledgments)

## Quick Start

1. Open Discord desktop app
2. In Claude Code:
   ```
   /plugin marketplace add Namachi06/cc-discord-presence
   /plugin install cc-discord-presence@cc-discord-presence
   /reload-plugins
   ```
3. Start a new Claude Code session

The plugin automatically downloads the binary and starts a background daemon. Your Discord status updates within seconds.

## Features

- Real-time session display: project, branch, model, tokens, cost, duration
- Works out of the box with zero configuration
- Show or hide any field individually
- Custom prefixes, separators, and cost precision
- Format templates for full layout control (`{project}`, `{model}`, `{tokens}`, etc.)
- Clickable buttons on your presence (GitHub, portfolio, etc.)
- Split token display (input/output separately)
- Idle detection with configurable timeout
- Session focus for multi-session setups (shows most recently active)
- Custom Discord app support (your own icon and name)
- Live config reload — edit and save, changes apply instantly
- Cross-platform: macOS (arm64/amd64), Linux (amd64/arm64), Windows (amd64)

## Configuration

Create `~/.claude/discord-presence.json` to customize. All fields are optional — omitted fields use defaults.

### Examples

**Hide git branch:**

```json
{
  "show": {
    "git_branch": false
  }
}
```

**Custom display:**

```json
{
  "show": {
    "split_tokens": true,
    "cost_in_details": true
  },
  "display": {
    "details_prefix": "Coding",
    "cost_precision": 2
  }
}
```

**Full real-world config:**

```json
{
  "show": {
    "project_name": true,
    "git_branch": false,
    "cost": true,
    "cost_in_details": true,
    "split_tokens": false,
    "session_focus": true
  },
  "display": {
    "details_prefix": "Coding",
    "cost_precision": 2,
    "idle_timeout": 300,
    "large_image": "logo",
    "large_text": "AI-assisted coding session",
    "discord_app_id": "YOUR_APP_ID"
  },
  "buttons": [
    {"label": "GitHub", "url": "https://github.com/username"}
  ]
}
```

### Format Templates

Set `details_format` or `state_format` to override the `show.*` system entirely for that line.

```json
{
  "display": {
    "details_format": "{project} | ${cost}",
    "state_format": "{model} | {in_tokens} in | {out_tokens} out"
  }
}
```

| Variable | Description | Example |
|----------|-------------|---------|
| `{project}` | Project directory name | `my-app` |
| `{branch}` | Git branch | `main` |
| `{model}` | Model display name | `Opus 4.5` |
| `{tokens}` | Total tokens (formatted) | `1.5M` |
| `{in_tokens}` | Input tokens | `1.2M` |
| `{out_tokens}` | Output tokens | `300K` |
| `{cost}` | Cost without `$` sign | `0.12` |
| `{duration}` | Session duration | `1h30m0s` |
| `{separator}` | Configured separator | ` \| ` |

> `{cost}` returns the number only. Write `${cost}` to get `$0.12`.

### Field Reference

**What to show:**

| Field | Default | Affects | Description |
|-------|---------|---------|-------------|
| `project_name` | `true` | Details | Project directory name |
| `git_branch` | `true` | Details | Git branch in parentheses |
| `model_name` | `true` | State | Claude model name |
| `tokens` | `true` | State | Token count |
| `split_tokens` | `false` | State | Show input/output separately instead of total |
| `cost` | `true` | State | Session cost |
| `cost_in_details` | `false` | Details | Move cost from State to Details line |
| `duration` | `true` | Timer | Elapsed time |
| `session_focus` | `false` | Behavior | Display most recently active session (requires statusline) |
| `privacy_mode` | `false` | Behavior | Show only "Using Claude Code" with no session details |

**How it looks:**

| Field | Default | Description |
|-------|---------|-------------|
| `details_prefix` | `"Working on"` | Text before project name |
| `details_format` | `""` | Template override for Details line |
| `state_format` | `""` | Template override for State line |
| `separator` | `" \| "` | Separator between State fields |
| `cost_precision` | `4` | Decimal places for cost (0-10) |
| `idle_timeout` | `0` | Seconds before showing "Idle" (0 = disabled, max 3600) |
| `idle_disable` | `0` | Seconds of idle before clearing presence entirely (0 = disabled, max 86400) |
| `cost_alert` | `0` | Cost threshold to prepend ⚠ warning (0 = disabled) |
| `exclude_projects` | `[]` | List of path glob patterns to mask as "Private Project" |
| `project_names` | `{}` | Map of path glob pattern to custom display name |
| `model_icons` | `{}` | Map of model keyword to Discord asset key (requires custom app) |
| `large_image` | `""` | Asset key for icon (requires custom Discord app) |
| `large_text` | `"Clawd Code - ..."` | Tooltip on icon hover (requires `large_image`) |
| `discord_app_id` | `""` | Custom Discord Application ID |

**Buttons:**

Array of up to 2 objects with `label` (max 32 chars) and `url` (http/https).

> - Config changes apply automatically via live reload
> - Changing `discord_app_id` requires a daemon restart
> - Buttons are only visible to **other Discord users**, not yourself

## Data Sources

The plugin reads session data two ways. Both work out of the box, but statusline is more accurate.

| | JSONL (default) | Statusline |
|---|---|---|
| **Setup** | None | One-time script |
| **Token accuracy** | Estimated from transcripts | Direct from Claude Code |
| **Cost accuracy** | Calculated from pricing table | Direct from Claude Code |
| **Session focus** | Not supported | Supported |

### Statusline Setup

**Automatic:**

```bash
# Find and run the setup script from the plugin cache:
~/.claude/plugins/cache/*/cc-discord-presence/*/scripts/setup-statusline.sh
```

Requires `jq`. Restart Claude Code after setup.

**Manual:**

1. Copy the wrapper script:
   ```bash
   cp ~/.claude/plugins/cache/*/cc-discord-presence/*/scripts/statusline-wrapper.sh ~/.claude/
   chmod +x ~/.claude/statusline-wrapper.sh
   ```

2. Add to `~/.claude/settings.json`:
   ```json
   {
     "statusLine": {
       "type": "command",
       "command": "~/.claude/statusline-wrapper.sh"
     }
   }
   ```

3. Restart Claude Code

### Verify Your Data Source

```bash
cat ~/.claude/discord-presence.log
```

Look for:
- `using statusline data` — Best accuracy
- `using JSONL fallback` — Working, but consider setting up statusline

## Advanced Features

### Idle Detection

When `idle_timeout` is set, the State line changes to **"Idle"** if no session activity is detected within the timeout. Resumes automatically when you interact with Claude Code again.

```json
{"display": {"idle_timeout": 300}}
```

Set to `0` to disable. Maximum: `3600` seconds (1 hour).

### Auto-disable on Extended Idle

When `idle_disable` is set, the presence is **completely cleared** from Discord after being idle for the specified duration. Re-enables automatically when activity resumes. The timer starts when idle begins (after `idle_timeout`).

```json
{"display": {"idle_timeout": 300, "idle_disable": 5400}}
```

Example: 5 min inactivity → "Idle" → after 1h30 more → presence disappears → activity resumes → presence reappears.

Requires `idle_timeout` to be set (otherwise the session never goes idle).

### Privacy Mode

Hide all session details. Only shows "Using Claude Code" with elapsed time. No project name, tokens, cost, or model visible. Buttons and icon are preserved.

```json
{"show": {"privacy_mode": true}}
```

Toggle via live reload — useful during screen shares or meetings.

### Cost Alert

Prepend a ⚠ warning to the State line when session cost exceeds a threshold.

```json
{"display": {"cost_alert": 50}}
```

Suppressed when idle or in privacy mode. Set to `0` to disable.

### Model Icons

Display a model-specific small icon in the corner of the presence. Requires a custom Discord app with uploaded assets.

```json
{
  "display": {
    "model_icons": {
      "opus": "icon-opus",
      "sonnet": "icon-sonnet",
      "haiku": "icon-haiku"
    }
  }
}
```

Matches model name by case-insensitive substring. Suppressed in privacy mode.

### Project Exclusion

Mask specific projects from Discord. The project name is replaced with "Private Project" and the git branch is hidden, but model, tokens, cost, and duration remain visible.

```json
{
  "display": {
    "exclude_projects": [
      "/Users/me/work/*",
      "/Users/me/clients/*"
    ]
  }
}
```

Uses glob patterns on the full project path (`*` matches one directory level). Live reload supported.

### Project Name Aliases

Map project paths to custom display names. The custom name replaces the directory name while git branch and all other fields remain visible.

```json
{
  "display": {
    "project_names": {
      "/Users/me/projects/my-long-project-name": "Short Name",
      "/Users/me/work/*": "Work Project"
    }
  }
}
```

Uses glob patterns on the full project path, same as `exclude_projects`. If a project matches both `exclude_projects` and `project_names`, exclusion takes priority (privacy wins). Live reload supported.

### Session Focus

When running multiple Claude Code sessions simultaneously, the presence normally flickers between sessions. Enable `session_focus` to always display the **most recently active session** — the one you're currently interacting with.

```json
{"show": {"session_focus": true}}
```

Requires statusline integration. Per-session data files are automatically cleaned up after 10 minutes of inactivity.

### Discord Buttons

Add up to 2 clickable buttons. Labels max 32 characters, URLs must be http or https.

```json
{
  "buttons": [
    {"label": "GitHub", "url": "https://github.com/username"},
    {"label": "Portfolio", "url": "https://example.com"}
  ]
}
```

**Important:** Buttons are only visible to other Discord users viewing your profile, not to yourself.

### Custom Discord App

By default, the plugin uses a shared Discord app. Create your own for a custom icon and name:

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click **New Application** and name it
   > Discord blocks trademarked names like "Claude Code". Use alternatives like "Coding..." or "Dev Session".
3. Set an app icon in **General Information**
4. For icon tooltip support: go to **Rich Presence > Art Assets**, upload an image, and note the asset key name
5. Copy the **Application ID** and add to your config:
   ```json
   {
     "display": {
       "discord_app_id": "YOUR_APPLICATION_ID",
       "large_image": "your-asset-key",
       "large_text": "Your tooltip text"
     }
   }
   ```
6. Restart the daemon (exit and reopen Claude Code)

## Token Pricing

Used for cost estimation in JSONL fallback mode. Statusline mode gets cost directly from Claude Code.

| Model | Input ($/1M tokens) | Output ($/1M tokens) |
|-------|--------------------:|---------------------:|
| Opus 4.5 | $15.00 | $75.00 |
| Sonnet 4.5 | $3.00 | $15.00 |
| Sonnet 4 | $3.00 | $15.00 |
| Haiku 4.5 | $1.00 | $5.00 |

Unknown models default to Sonnet 4 pricing. See [Anthropic pricing](https://www.anthropic.com/pricing) for updates.

## Platform Support

| Platform | Status | Notes |
|----------|--------|-------|
| macOS (Apple Silicon) | Tested | Primary development platform |
| macOS (Intel) | Supported | |
| Linux (x86_64) | Supported | |
| Linux (ARM64) | Supported | |
| Windows (x86_64) | Supported | Requires Git Bash for shell scripts |

## Troubleshooting

| Problem | Solution |
|---------|----------|
| No presence showing | Ensure Discord **desktop app** is running (not web). Check `~/.claude/discord-presence.log` for errors. |
| Tokens or cost seem wrong | Set up statusline integration. JSONL fallback estimates from a pricing table. |
| "Idle" showing when coding | Increase `idle_timeout` or set to `0` to disable. |
| Config changes not applying | Changes auto-reload. Exception: `discord_app_id` requires daemon restart. |
| Binary download fails | Check internet. Manual download: get binary from [Releases](https://github.com/Namachi06/cc-discord-presence/releases) and place in `~/.claude/bin/`. |
| Buttons not showing | Buttons are only visible to **other** Discord users. Ask a friend to check. |
| Wrong project in multi-session | Enable `session_focus` in config. Requires statusline. |
| Windows: not working | Ensure Git Bash is installed. WSL is not supported (Discord runs on Windows host). |

## File Paths

| File | Purpose |
|------|---------|
| `~/.claude/discord-presence.json` | User configuration (optional) |
| `~/.claude/discord-presence.log` | Daemon log output |
| `~/.claude/discord-presence.pid` | Daemon process ID |
| `~/.claude/discord-presence-data.json` | Statusline data (shared) |
| `~/.claude/discord-presence-session-<id>.json` | Per-session data (session focus) |
| `~/.claude/discord-presence-sessions/` | Active session PID tracking |
| `~/.claude/bin/cc-discord-presence-<os>-<arch>` | Auto-downloaded binary |
| `~/.claude/statusline-wrapper.sh` | Statusline wrapper (after setup) |

## Building from Source

Requires Go 1.25+.

```bash
git clone https://github.com/Namachi06/cc-discord-presence.git
cd cc-discord-presence
go build -o cc-discord-presence .

# Cross-compile for all platforms:
./scripts/build.sh

# Run tests:
go test -v ./...
```

## Uninstallation

### Plugin Removal

```
/plugin uninstall cc-discord-presence@cc-discord-presence
/reload-plugins
```

### Manual Cleanup

```bash
# Remove daemon files
rm -f ~/.claude/discord-presence.pid
rm -f ~/.claude/discord-presence.log
rm -f ~/.claude/discord-presence-data.json
rm -f ~/.claude/discord-presence-session-*.json
rm -rf ~/.claude/discord-presence-sessions
rm -f ~/.claude/discord-presence.json

# Remove binary
rm -f ~/.claude/bin/cc-discord-presence-*

# Remove statusline wrapper (if installed)
rm -f ~/.claude/statusline-wrapper.sh
```

If you set up statusline integration, restore your original statusline in `~/.claude/settings.json`.

## Privacy

No data is collected or sent to external servers. All processing is local. See [PRIVACY.md](PRIVACY.md).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Run `go test -v ./...` before submitting PRs.

## License

[MIT](LICENSE)

## Acknowledgments

- [Anthropic](https://anthropic.com) for Claude
- [tsanva](https://github.com/tsanva) for the original [cc-discord-presence](https://github.com/tsanva/cc-discord-presence)
- [fsnotify](https://github.com/fsnotify/fsnotify) for file watching
- [go-winio](https://github.com/Microsoft/go-winio) for Windows named pipe support

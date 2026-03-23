# Clawd Code - Discord Rich Presence for Claude Code

Show your Claude Code session on Discord! Display your current project, git branch, model, session time, token usage, and cost in real-time.

## Platform Support

| Platform | Status |
|----------|--------|
| macOS (Apple Silicon) | ✅ Tested |
| macOS (Intel) | ⚠️ Untested |
| Linux (x64) | ⚠️ Untested |
| Linux (ARM64) | ⚠️ Untested |
| Windows (x64) | ✅ Tested |
| Windows (ARM64) | ⚠️ Untested |

> **Note**: macOS Intel and Linux should work but haven't been verified. Please [report problems](https://github.com/tsanva/cc-discord-presence/issues).
>
> **Windows users**: Requires [Git Bash](https://git-scm.com/downloads) (included with Git for Windows) for automatic plugin hooks. Alternatively, run the PowerShell scripts manually (`scripts/start.ps1` and `scripts/stop.ps1`). WSL won't work as Discord runs on the Windows host.

## Features

- **Session Time** - Shows how long you've been coding with Claude
- **Project Name** - Displays the current project you're working on
- **Git Branch** - Shows your current git branch
- **Model Name** - Shows which Claude model you're using (Opus 4.5, Sonnet 4.5, Haiku 4.5)
- **Total Tokens** - Token usage counter (input + output)
- **Total Cost** - Real-time cost tracking for your session

## Installation

### As a Claude Code Plugin (Recommended)

```bash
# Add the marketplace
claude plugin marketplace add tsanva/cc-discord-presence

# Install the plugin
claude plugin install cc-discord-presence@cc-discord-presence
```

That's it! The plugin will automatically start when you begin a Claude Code session and stop when you exit.

### Manual Installation

```bash
# Clone and build
git clone https://github.com/tsanva/cc-discord-presence.git
cd cc-discord-presence
go build -o cc-discord-presence .

# Run manually
./cc-discord-presence
```

## How It Works

The app reads session data from Claude Code in two ways:

### 1. JSONL Fallback (Zero Config)

By default, the app parses Claude Code's session files from `~/.claude/projects/`. This works out of the box with no configuration needed.

### 2. Statusline Integration (More Accurate)

For the most accurate token/cost data, you can configure the statusline integration. This uses Claude Code's own calculations instead of estimating from JSONL.

<a name="statusline-setup"></a>
#### Statusline Setup

**Automatic Setup (Recommended)**:

Run the setup script (requires `jq`):
```bash
# Find your plugin directory and run setup
~/.claude/plugins/cache/*/cc-discord-presence/*/scripts/setup-statusline.sh
```

Or if you have the repo cloned:
```bash
./scripts/setup-statusline.sh
```

The setup script will:
- Copy `statusline-wrapper.sh` to `~/.claude/`
- Update your `~/.claude/settings.json` automatically
- Back up any existing statusline to `~/.claude/statusline.sh`

**Manual Setup**: If you prefer, edit `~/.claude/settings.json`:
```json
{
  "statusLine": {
    "command": "~/.claude/statusline-wrapper.sh",
    "type": "command"
  }
}
```

Then copy `scripts/statusline-wrapper.sh` to `~/.claude/statusline-wrapper.sh`.

**Note**: Restart Claude Code after setup for changes to take effect.

#### Verifying Your Setup

Check which data source is being used by viewing the daemon log:
```bash
cat ~/.claude/discord-presence.log
```

You'll see one of:
- `✓ Found active session: project-name (using statusline data)` - Best accuracy
- `✓ Found active session: project-name (using JSONL fallback)` - Working, but consider setting up statusline

## Configuration

You can customize which fields are displayed on Discord by creating a config file at `~/.claude/discord-presence.json`.

**All fields are optional** — if the file doesn't exist or a field is omitted, defaults are used (everything visible).

### Minimal Example (hide git branch)

```json
{
  "show": {
    "git_branch": false
  }
}
```

### Full Example

```json
{
  "show": {
    "project_name": true,
    "git_branch": false,
    "model_name": true,
    "tokens": true,
    "cost": false,
    "duration": true
  },
  "display": {
    "details_prefix": "Coding",
    "separator": " - ",
    "cost_precision": 2,
    "large_text": "My Custom Presence Text",
    "discord_app_id": "YOUR_APP_ID"
  }
}
```

### Field Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `show.project_name` | bool | `true` | Show project name in Details line |
| `show.git_branch` | bool | `true` | Show git branch in Details line |
| `show.model_name` | bool | `true` | Show Claude model in State line |
| `show.tokens` | bool | `true` | Show token count in State line |
| `show.cost` | bool | `true` | Show cost in State line |
| `show.cost_in_tooltip` | bool | `false` | Move cost to icon tooltip (hover) instead of State line |
| `show.duration` | bool | `true` | Show elapsed time |
| `display.details_prefix` | string | `"Working on"` | Prefix before project/branch |
| `display.separator` | string | `" \| "` | Separator between State parts |
| `display.cost_precision` | int | `4` | Decimal places for cost (0-10) |
| `display.large_text` | string | `"Clawd Code - ..."` | Tooltip text on the large icon |
| `display.discord_app_id` | string | `""` | Custom Discord Application ID |

> **Live reload**: Changes to the config file are picked up automatically — no need to restart the daemon.
>
> **Note**: Changing `discord_app_id` requires a daemon restart.

## Discord Presence Display

```
┌─────────────────────────────────┐
│ Clawd Code                      │
│ Working on: my-project (main)   │
│ Opus 4.5 | 1.5M tokens | $0.1234│
│ 00:45:30 elapsed                │
└─────────────────────────────────┘
```

## Requirements

- [Discord](https://discord.com) desktop app running
- [Claude Code](https://claude.ai/code) installed
- Go 1.25+ (only for building from source)

## Building from Source

```bash
# Build for current platform
go build -o cc-discord-presence .

# Cross-compile for all platforms
mkdir -p bin
GOOS=darwin GOARCH=arm64 go build -o bin/cc-discord-presence-darwin-arm64 .
GOOS=darwin GOARCH=amd64 go build -o bin/cc-discord-presence-darwin-amd64 .
GOOS=linux GOARCH=amd64 go build -o bin/cc-discord-presence-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -o bin/cc-discord-presence-linux-arm64 .
GOOS=windows GOARCH=amd64 go build -o bin/cc-discord-presence-windows-amd64.exe .
```

## Token Pricing

Cost is calculated using current Claude API pricing (Dec 2025):

| Model | Input (per 1M tokens) | Output (per 1M tokens) |
|-------|----------------------|------------------------|
| Opus 4.5 | $15.00 | $75.00 |
| Sonnet 4.5 | $3.00 | $15.00 |
| Sonnet 4 | $3.00 | $15.00 |
| Haiku 4.5 | $1.00 | $5.00 |

## Advanced: Custom Discord App

By default, this uses a shared Discord application ("Clawd Code"). If you want to use your own:

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application" and name it
   > ⚠️ **Note**: Discord blocks trademarked names like "Claude Code"
3. Set an app icon in "General Information" (this appears in Rich Presence)
4. Copy the **Application ID** and set it in your config file:
   ```json
   {
     "display": {
       "discord_app_id": "YOUR_APPLICATION_ID"
     }
   }
   ```
5. Restart the daemon

## Uninstallation

### Plugin Removal

```bash
claude plugin uninstall cc-discord-presence@cc-discord-presence
```

### Statusline Cleanup (if configured)

If you set up statusline integration, restore your original settings:

```bash
# Remove the wrapper script
rm ~/.claude/statusline-wrapper.sh

# Restore your original statusline in settings.json:
# Option 1: Point back to the default statusline.sh
jq '.statusLine.command = "~/.claude/statusline.sh"' ~/.claude/settings.json > ~/.claude/settings.json.tmp \
  && mv ~/.claude/settings.json.tmp ~/.claude/settings.json

# Option 2: Remove statusline config entirely
jq 'del(.statusLine)' ~/.claude/settings.json > ~/.claude/settings.json.tmp \
  && mv ~/.claude/settings.json.tmp ~/.claude/settings.json
```

Restart Claude Code after making changes.

## Privacy

This application runs entirely locally and does not collect any data. See [PRIVACY.md](PRIVACY.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Before submitting a PR**, make sure to run the test suite:
```bash
go test -v ./...
```

## Acknowledgments

- [Anthropic](https://anthropic.com) for Claude
- [fsnotify](https://github.com/fsnotify/fsnotify) for file watching
- [go-winio](https://github.com/Microsoft/go-winio) for Windows named pipe support
- The Claude Code community

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/tsanva/cc-discord-presence/discord"
)

const (
	// Discord Application ID for "Clawd Code"
	ClientID = "1455326944060248250"

	// Polling interval as fallback
	PollInterval = 3 * time.Second
)

// Model pricing per million tokens (December 2025)
// Update these when new models are released: https://www.anthropic.com/pricing
var modelPricing = map[string]struct{ Input, Output float64 }{
	"claude-opus-4-5-20251101":   {15.0, 75.0},
	"claude-sonnet-4-5-20241022": {3.0, 15.0},
	"claude-sonnet-4-20250514":   {3.0, 15.0},
	"claude-haiku-4-5-20241022":  {1.0, 5.0},
}

// Model display names - add new model IDs here when released
var modelDisplayNames = map[string]string{
	"claude-opus-4-5-20251101":   "Opus 4.5",
	"claude-sonnet-4-5-20241022": "Sonnet 4.5",
	"claude-sonnet-4-20250514":   "Sonnet 4",
	"claude-haiku-4-5-20241022":  "Haiku 4.5",
}

// StatusLineData matches Claude Code's statusline JSON structure
type StatusLineData struct {
	SessionID string `json:"session_id"`
	Cwd       string `json:"cwd"`
	Model     struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"model"`
	Workspace struct {
		CurrentDir string `json:"current_dir"`
		ProjectDir string `json:"project_dir"`
	} `json:"workspace"`
	Cost struct {
		TotalCostUSD       float64 `json:"total_cost_usd"`
		TotalDurationMS    int64   `json:"total_duration_ms"`
		TotalAPIDurationMS int64   `json:"total_api_duration_ms"`
	} `json:"cost"`
	ContextWindow struct {
		TotalInputTokens  int64 `json:"total_input_tokens"`
		TotalOutputTokens int64 `json:"total_output_tokens"`
	} `json:"context_window"`
}

// SessionData holds parsed session information
type SessionData struct {
	ProjectName  string
	ProjectPath  string
	GitBranch    string
	ModelName    string
	TotalTokens  int64
	InputTokens  int64
	OutputTokens int64
	TotalCost    float64
	StartTime    time.Time
}

// JSONLMessage represents a message entry in JSONL files
type JSONLMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Cwd       string `json:"cwd"`
	Message   struct {
		Model string `json:"model"`
		Usage struct {
			InputTokens  int64 `json:"input_tokens"`
			OutputTokens int64 `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

var (
	claudeDir        string
	projectsDir      string
	dataFilePath     string
	sessionStartTime = time.Now()
	discordClient    *discord.Client
	usingFallback    bool
	nudgeShown       bool
	lastSessionData  *SessionData
	lastDataChange   time.Time
	isIdle           bool
	isDisabled           bool
	idleStartTime        time.Time
	lastPresenceDetails  string
	lastPresenceState    string
	gitBranchCache       = make(map[string]struct {
		branch  string
		expires time.Time
	})
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}
	claudeDir = filepath.Join(home, ".claude")
	projectsDir = filepath.Join(claudeDir, "projects")
	dataFilePath = filepath.Join(claudeDir, "discord-presence-data.json")
	configFilePath = filepath.Join(claudeDir, "discord-presence.json")
}

func main() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║     Clawd Code - Discord Rich Presence                    ║
║     Show your Claude Code session on Discord!             ║
╚═══════════════════════════════════════════════════════════╝`)

	// Load configuration
	currentConfig = LoadConfig()
	fmt.Println("✓ Configuration loaded")

	// Determine client ID (allow custom Discord App ID)
	clientID := ClientID
	if currentConfig.Display.DiscordAppID != "" {
		clientID = currentConfig.Display.DiscordAppID
		fmt.Printf("  Using custom Discord App ID: %s\n", clientID)
	}

	// Connect to Discord
	fmt.Println("🔗 Connecting to Discord...")
	discordClient = discord.NewClient(clientID)
	if err := discordClient.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect to Discord: %v\n", err)
		fmt.Fprintln(os.Stderr, "   Make sure Discord is running and try again.")
		os.Exit(1)
	}
	fmt.Println("✓ Discord RPC connected!")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n⏹ Shutting down...")
		discordClient.Close()
		os.Exit(0)
	}()

	// Initialize idle detection
	lastDataChange = time.Now()

	// Try initial read and show data source
	if session := readSessionData(); session != nil {
		processSessionUpdate(session)
		if usingFallback {
			fmt.Printf("✓ Found active session: %s (using JSONL fallback)\n", session.ProjectName)
		} else {
			fmt.Printf("✓ Found active session: %s (using statusline data)\n", session.ProjectName)
		}
	} else {
		fmt.Println("⏳ Waiting for Claude Code session...")
	}

	fmt.Println("🎮 Discord Rich Presence is now active! Press Ctrl+C to stop.")

	// Start watching for changes
	watchForChanges()
}

// statusLineToSessionData converts StatusLineData to SessionData.
func statusLineToSessionData(sl *StatusLineData, startTime time.Time) *SessionData {
	projectPath := sl.Workspace.ProjectDir
	if projectPath == "" {
		projectPath = sl.Cwd
	}

	return &SessionData{
		ProjectName:  projectNameFromPath(projectPath),
		ProjectPath:  projectPath,
		GitBranch:    getGitBranch(projectPath),
		ModelName:    sl.Model.DisplayName,
		InputTokens:  sl.ContextWindow.TotalInputTokens,
		OutputTokens: sl.ContextWindow.TotalOutputTokens,
		TotalTokens:  sl.ContextWindow.TotalInputTokens + sl.ContextWindow.TotalOutputTokens,
		TotalCost:    sl.Cost.TotalCostUSD,
		StartTime:    startTime,
	}
}

func readStatusLineData() *SessionData {
	data, err := os.ReadFile(dataFilePath)
	if err != nil {
		return nil
	}

	var statusLine StatusLineData
	if err := json.Unmarshal(data, &statusLine); err != nil {
		return nil
	}

	if statusLine.SessionID == "" {
		return nil
	}

	return statusLineToSessionData(&statusLine, sessionStartTime)
}

// sessionFilePattern returns the glob pattern for per-session data files.
func sessionFilePattern() string {
	return filepath.Join(claudeDir, "discord-presence-session-*.json")
}

// readSessionFile parses a per-session JSON file.
func readSessionFile(path string) (*StatusLineData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var statusLine StatusLineData
	if err := json.Unmarshal(data, &statusLine); err != nil {
		return nil, err
	}
	if statusLine.SessionID == "" {
		return nil, fmt.Errorf("missing session_id")
	}
	return &statusLine, nil
}

// readFocusedSessionData reads the most recently modified per-session file.
func readFocusedSessionData() *SessionData {
	matches, err := filepath.Glob(sessionFilePattern())
	if err != nil || len(matches) == 0 {
		return nil
	}

	var newestPath string
	var newestTime time.Time
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.ModTime().After(newestTime) {
			newestTime = info.ModTime()
			newestPath = path
		}
	}

	if newestPath == "" {
		return nil
	}

	sl, err := readSessionFile(newestPath)
	if err != nil {
		return nil
	}
	return statusLineToSessionData(sl, sessionStartTime)
}

// cleanupStaleSessions removes per-session files not modified in the last 10 minutes.
func cleanupStaleSessions() {
	matches, err := filepath.Glob(sessionFilePattern())
	if err != nil {
		return
	}
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if time.Since(info.ModTime()) > 10*time.Minute {
			os.Remove(path)
		}
	}
}

func projectNameFromPath(path string) string {
	name := filepath.Base(path)
	if name == "" || name == "." {
		return "Unknown Project"
	}
	return name
}

func getGitBranch(projectPath string) string {
	if projectPath == "" {
		return ""
	}

	// Check cache (30s TTL)
	if cached, ok := gitBranchCache[projectPath]; ok && time.Now().Before(cached.expires) {
		return cached.branch
	}

	cmd := exec.Command("git", "-C", projectPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	branch := strings.TrimSpace(string(output))

	// If HEAD (no commits yet), try to get the branch name from symbolic-ref
	if branch == "HEAD" {
		cmd = exec.Command("git", "-C", projectPath, "symbolic-ref", "--short", "HEAD")
		output, err = cmd.Output()
		if err == nil {
			branch = strings.TrimSpace(string(output))
		}
	}

	// Cache the result for 30 seconds
	gitBranchCache[projectPath] = struct {
		branch  string
		expires time.Time
	}{branch, time.Now().Add(30 * time.Second)}

	return branch
}

// findMostRecentJSONL finds the most recently modified JSONL file in ~/.claude/projects/
func findMostRecentJSONL() (string, string, error) {
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		return "", "", fmt.Errorf("projects directory does not exist")
	}

	type jsonlFile struct {
		path        string
		projectPath string
		modTime     time.Time
	}

	var files []jsonlFile

	err := filepath.WalkDir(projectsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if d.IsDir() || !strings.HasSuffix(path, ".jsonl") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Extract project path from the directory structure
		// ~/.claude/projects/<encoded-path>/<session>.jsonl
		// Encoded path uses dashes: -Users-vasantpns-Developer-project
		relPath, _ := filepath.Rel(projectsDir, path)
		parts := strings.SplitN(relPath, string(filepath.Separator), 2)
		if len(parts) < 1 {
			return nil
		}

		// Decode the project path
		// Claude Code encodes paths: / becomes -, and literal - becomes --
		// Example: /Users/foo/my-project -> -Users-foo-my--project
		// Must decode -- to - FIRST, then decode single - to /
		encodedPath := parts[0]
		// Use a placeholder for double dashes (escaped literal dashes)
		projectPath := strings.ReplaceAll(encodedPath, "--", "\x00")
		// Convert single dashes to path separators
		projectPath = strings.ReplaceAll(projectPath, "-", "/")
		// Restore literal dashes from placeholder
		projectPath = strings.ReplaceAll(projectPath, "\x00", "-")

		files = append(files, jsonlFile{
			path:        path,
			projectPath: projectPath,
			modTime:     info.ModTime(),
		})

		return nil
	})

	if err != nil {
		return "", "", err
	}

	if len(files) == 0 {
		return "", "", fmt.Errorf("no JSONL files found")
	}

	// Sort by modification time, most recent first
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})

	return files[0].path, files[0].projectPath, nil
}

// parseJSONLSession parses a JSONL file and extracts session data
func parseJSONLSession(jsonlPath, _ string) *SessionData {
	file, err := os.Open(jsonlPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var (
		totalInputTokens  int64
		totalOutputTokens int64
		lastModel         string
		projectPath       string
	)

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		var msg JSONLMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		// Extract cwd from any message that has it (usually first message)
		if msg.Cwd != "" && projectPath == "" {
			projectPath = msg.Cwd
		}

		// Only process assistant messages with usage data
		if msg.Type == "assistant" && msg.Message.Model != "" {
			lastModel = msg.Message.Model
			totalInputTokens += msg.Message.Usage.InputTokens
			totalOutputTokens += msg.Message.Usage.OutputTokens
		}
	}

	if lastModel == "" {
		return nil
	}

	// Calculate cost based on model pricing
	totalCost := calculateCost(lastModel, totalInputTokens, totalOutputTokens)

	// Get display name for model
	modelName := formatModelName(lastModel)

	return &SessionData{
		ProjectName:  projectNameFromPath(projectPath),
		ProjectPath:  projectPath,
		GitBranch:    getGitBranch(projectPath),
		ModelName:    modelName,
		InputTokens:  totalInputTokens,
		OutputTokens: totalOutputTokens,
		TotalTokens:  totalInputTokens + totalOutputTokens,
		TotalCost:    totalCost,
		StartTime:    sessionStartTime,
	}
}

// calculateCost calculates the cost based on token usage and model pricing
func calculateCost(modelID string, inputTokens, outputTokens int64) float64 {
	pricing, ok := modelPricing[modelID]
	if !ok {
		// Default to Sonnet 4 pricing if unknown model
		pricing = modelPricing["claude-sonnet-4-20250514"]
	}

	inputCost := float64(inputTokens) / 1_000_000 * pricing.Input
	outputCost := float64(outputTokens) / 1_000_000 * pricing.Output

	return inputCost + outputCost
}

// formatModelName converts model ID to display name
func formatModelName(modelID string) string {
	if name, ok := modelDisplayNames[modelID]; ok {
		return name
	}

	// Try to extract a reasonable name from the model ID
	if strings.Contains(modelID, "opus") {
		return "Opus"
	}
	if strings.Contains(modelID, "sonnet") {
		return "Sonnet"
	}
	if strings.Contains(modelID, "haiku") {
		return "Haiku"
	}

	return "Claude"
}

// readSessionData tries statusline data first, then falls back to JSONL parsing
func readSessionData() *SessionData {
	cfg := currentConfig

	// Session focus mode: pick most recently active per-session file
	if showFieldDefault(cfg.Show.SessionFocus, false) {
		if data := readFocusedSessionData(); data != nil {
			if usingFallback {
				usingFallback = false
			}
			return data
		}
	}

	// First try statusline data (most accurate)
	if data := readStatusLineData(); data != nil {
		if usingFallback {
			usingFallback = false
			fmt.Println("📊 Now using statusline data (more accurate)")
		}
		return data
	}

	// Fall back to JSONL parsing
	jsonlPath, projectPath, err := findMostRecentJSONL()
	if err != nil {
		return nil
	}

	if !usingFallback && !nudgeShown {
		usingFallback = true
		nudgeShown = true
		fmt.Println("\n💡 Tip: For more accurate token/cost data, configure the statusline wrapper.")
		fmt.Println("   See: https://github.com/Namachi06/cc-discord-presence#statusline-setup")
	}

	return parseJSONLSession(jsonlPath, projectPath)
}

// sessionDataChanged checks if session data has meaningfully changed.
func sessionDataChanged(old, new *SessionData) bool {
	if old == nil {
		return true
	}
	return old.TotalTokens != new.TotalTokens ||
		old.TotalCost != new.TotalCost ||
		old.ProjectName != new.ProjectName ||
		old.ModelName != new.ModelName ||
		old.GitBranch != new.GitBranch
}

// checkIdle returns whether the session should be considered idle.
func checkIdle(old, new *SessionData, lastChange time.Time, timeoutSecs int) (idle bool, newLastChange time.Time) {
	if timeoutSecs <= 0 {
		return false, lastChange
	}
	if sessionDataChanged(old, new) {
		return false, time.Now()
	}
	if time.Since(lastChange) >= time.Duration(timeoutSecs)*time.Second {
		return true, lastChange
	}
	return false, lastChange
}

// processSessionUpdate handles idle detection then updates the presence.
func processSessionUpdate(session *SessionData) {
	if session == nil {
		return
	}

	cfg := currentConfig

	wasIdle := isIdle
	isIdle, lastDataChange = checkIdle(lastSessionData, session, lastDataChange, intDefault(cfg.Display.IdleTimeout, 0))

	if isIdle && !wasIdle {
		idleStartTime = time.Now()
		fmt.Println("💤 Session idle")
	} else if !isIdle && wasIdle {
		fmt.Println("🔄 Session active again")
		if isDisabled {
			isDisabled = false
			fmt.Println("✅ Presence re-enabled")
		}
	}

	// Auto-disable after extended idle
	if checkIdleDisable(isIdle, idleStartTime, intDefault(cfg.Display.IdleDisable, 0)) {
		if !isDisabled {
			isDisabled = true
			fmt.Println("🛑 Presence cleared (idle too long)")
			clearPresence()
		}
		lastSessionData = session
		return
	}

	lastSessionData = session
	updatePresence(session)
}

func clearPresence() {
	if err := discordClient.ClearActivity(); err != nil {
		fmt.Fprintf(os.Stderr, "Error clearing presence: %v\n", err)
	}
}

func updatePresence(session *SessionData) {
	cfg := currentConfig
	privacy := showFieldDefault(cfg.Show.PrivacyMode, false)

	var details, state string

	if privacy {
		details = "Using Claude Code"
		state = ""
	} else {
		if cfg.Display.DetailsFormat != "" {
			details = formatTemplate(cfg.Display.DetailsFormat, session, cfg)
		} else {
			details = buildDetailsLine(session, cfg)
		}

		if cfg.Display.StateFormat != "" {
			state = formatTemplate(cfg.Display.StateFormat, session, cfg)
		} else {
			state = buildStateLine(session, cfg)
		}

		if isIdle {
			state = "Idle"
		}

		state = applyCostAlert(state, session.TotalCost, cfg.Display.CostAlert, isIdle, privacy)
	}

	activity := discord.Activity{
		Details:    details,
		State:      state,
		LargeImage: cfg.Display.LargeImage,
		LargeText:  cfg.Display.LargeText,
	}

	// Model icons (suppressed in privacy mode)
	if !privacy {
		if icon, found := matchModelIcon(session.ModelName, cfg.Display.ModelIcons); found {
			activity.SmallImage = icon
			activity.SmallText = session.ModelName
		}
	}

	if showField(cfg.Show.Duration) {
		activity.StartTime = &session.StartTime
	}
	for _, b := range cfg.Buttons {
		activity.Buttons = append(activity.Buttons, discord.Button{
			Label: b.Label,
			URL:   b.URL,
		})
	}

	// Skip no-op updates to avoid unnecessary IPC writes
	if details == lastPresenceDetails && state == lastPresenceState {
		return
	}
	lastPresenceDetails = details
	lastPresenceState = state

	if err := discordClient.SetActivity(activity); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating presence: %v\n", err)
	}
}

func formatNumber(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	} else if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func watchForChanges() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Using polling mode for session tracking")
		pollForChanges()
		return
	}
	defer watcher.Close()

	// Watch both the main claude dir (for statusline data) and projects dir (for JSONL fallback)
	if err := watcher.Add(claudeDir); err != nil {
		fmt.Println("Using polling mode for session tracking")
		pollForChanges()
		return
	}

	// Also poll as backup (especially important for JSONL which is in subdirs)
	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			baseName := filepath.Base(event.Name)
			// Respond to config file changes
			if baseName == "discord-presence.json" {
				newConfig := LoadConfig()
				oldAppID := currentConfig.Display.DiscordAppID
				currentConfig = newConfig
				fmt.Println("✓ Configuration reloaded")
				if newConfig.Display.DiscordAppID != oldAppID {
					fmt.Println("⚠ discord_app_id change requires restart to take effect")
				}
				if session := readSessionData(); session != nil {
					processSessionUpdate(session)
				}
			}
			// Respond to statusline data file changes (shared or per-session)
			if baseName == "discord-presence-data.json" || strings.HasPrefix(baseName, "discord-presence-session-") {
				if session := readSessionData(); session != nil {
					processSessionUpdate(session)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)
		case <-ticker.C:
			// Poll reads from either statusline or JSONL fallback
			if session := readSessionData(); session != nil {
				processSessionUpdate(session)
			}
		case <-cleanupTicker.C:
			if showFieldDefault(currentConfig.Show.SessionFocus, false) {
				cleanupStaleSessions()
			}
		}
	}
}

func pollForChanges() {
	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	cleanupTicker := time.NewTicker(5 * time.Minute)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ticker.C:
			if session := readSessionData(); session != nil {
				processSessionUpdate(session)
			}
		case <-cleanupTicker.C:
			if showFieldDefault(currentConfig.Show.SessionFocus, false) {
				cleanupStaleSessions()
			}
		}
	}
}

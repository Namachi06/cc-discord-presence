package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ButtonConfig represents a clickable button in Discord Rich Presence.
type ButtonConfig struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// Config represents the user configuration for Discord Rich Presence display.
type Config struct {
	Show    ShowConfig     `json:"show"`
	Display DisplayConfig  `json:"display"`
	Buttons []ButtonConfig `json:"buttons"`
}

// ShowConfig controls which fields are visible in the Discord presence.
type ShowConfig struct {
	ProjectName   *bool `json:"project_name"`
	GitBranch     *bool `json:"git_branch"`
	ModelName     *bool `json:"model_name"`
	Tokens        *bool `json:"tokens"`
	SplitTokens   *bool `json:"split_tokens"`
	Cost          *bool `json:"cost"`
	CostInDetails *bool `json:"cost_in_details"`
	Duration      *bool `json:"duration"`
	SessionFocus  *bool `json:"session_focus"`
	PrivacyMode   *bool `json:"privacy_mode"`
}

// DisplayConfig controls formatting and customization of the presence.
type DisplayConfig struct {
	DetailsPrefix string `json:"details_prefix"`
	DetailsFormat string `json:"details_format"`
	StateFormat   string `json:"state_format"`
	Separator     string `json:"separator"`
	CostPrecision *int   `json:"cost_precision"`
	IdleTimeout   *int   `json:"idle_timeout"`
	IdleDisable   *int                `json:"idle_disable"`
	CostAlert     *float64            `json:"cost_alert"`
	ExcludeProjects []string            `json:"exclude_projects"`
	ModelIcons      map[string]string   `json:"model_icons"`
	LargeImage    string              `json:"large_image"`
	LargeText     string              `json:"large_text"`
	DiscordAppID  string              `json:"discord_app_id"`
}

var (
	configFilePath string
	currentConfig  *Config
)

func boolPtr(b bool) *bool          { return &b }
func intPtr(i int) *int             { return &i }
func float64Ptr(f float64) *float64 { return &f }

func intDefault(ptr *int, def int) int {
	if ptr == nil {
		return def
	}
	return *ptr
}

func clampInt(ptr *int, min, max int) *int {
	if ptr == nil {
		return nil
	}
	v := *ptr
	if v < min {
		v = min
	}
	if v > max {
		v = max
	}
	return intPtr(v)
}

func clampFloat64(ptr *float64, min, max float64) *float64 {
	if ptr == nil {
		return nil
	}
	v := *ptr
	if v < min {
		v = min
	}
	if v > max {
		v = max
	}
	return float64Ptr(v)
}

// showField safely dereferences a *bool, defaulting to true if nil.
func showField(ptr *bool) bool {
	if ptr == nil {
		return true
	}
	return *ptr
}

// showFieldDefault safely dereferences a *bool with a custom default for nil.
func showFieldDefault(ptr *bool, defaultVal bool) bool {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// DefaultConfig returns a config with all fields set to their default values.
func DefaultConfig() *Config {
	return &Config{
		Show: ShowConfig{
			ProjectName:   boolPtr(true),
			GitBranch:     boolPtr(true),
			ModelName:     boolPtr(true),
			Tokens:        boolPtr(true),
			SplitTokens:   boolPtr(false),
			Cost:          boolPtr(true),
			CostInDetails: boolPtr(false),
			Duration:      boolPtr(true),
			SessionFocus:  boolPtr(false),
		},
		Display: DisplayConfig{
			DetailsPrefix: "Working on",
			Separator:     " | ",
			CostPrecision: intPtr(4),
			IdleTimeout:   intPtr(0),
			LargeText:     "Clawd Code - Discord Rich Presence for Claude Code",
		},
	}
}

// LoadConfig reads the config file and merges with defaults.
// Returns DefaultConfig() if the file does not exist or cannot be parsed.
func LoadConfig() *Config {
	defaults := DefaultConfig()

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return defaults
	}

	var userConfig Config
	if err := json.Unmarshal(data, &userConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: invalid config file %s: %v (using defaults)\n", configFilePath, err)
		return defaults
	}

	return mergeConfig(defaults, &userConfig)
}

// mergeConfig applies user-provided values on top of defaults.
func mergeConfig(defaults, user *Config) *Config {
	result := *defaults

	// Show fields
	if user.Show.ProjectName != nil {
		result.Show.ProjectName = user.Show.ProjectName
	}
	if user.Show.GitBranch != nil {
		result.Show.GitBranch = user.Show.GitBranch
	}
	if user.Show.ModelName != nil {
		result.Show.ModelName = user.Show.ModelName
	}
	if user.Show.Tokens != nil {
		result.Show.Tokens = user.Show.Tokens
	}
	if user.Show.SplitTokens != nil {
		result.Show.SplitTokens = user.Show.SplitTokens
	}
	if user.Show.Cost != nil {
		result.Show.Cost = user.Show.Cost
	}
	if user.Show.CostInDetails != nil {
		result.Show.CostInDetails = user.Show.CostInDetails
	}
	if user.Show.Duration != nil {
		result.Show.Duration = user.Show.Duration
	}
	if user.Show.SessionFocus != nil {
		result.Show.SessionFocus = user.Show.SessionFocus
	}
	if user.Show.PrivacyMode != nil {
		result.Show.PrivacyMode = user.Show.PrivacyMode
	}

	// Display fields
	if user.Display.DetailsPrefix != "" {
		result.Display.DetailsPrefix = user.Display.DetailsPrefix
	}
	if user.Display.DetailsFormat != "" {
		result.Display.DetailsFormat = user.Display.DetailsFormat
	}
	if user.Display.StateFormat != "" {
		result.Display.StateFormat = user.Display.StateFormat
	}
	if user.Display.Separator != "" {
		result.Display.Separator = user.Display.Separator
	}
	if v := clampInt(user.Display.CostPrecision, 0, 10); v != nil {
		result.Display.CostPrecision = v
	}
	if v := clampInt(user.Display.IdleTimeout, 0, 3600); v != nil {
		result.Display.IdleTimeout = v
	}
	if v := clampInt(user.Display.IdleDisable, 0, 86400); v != nil {
		result.Display.IdleDisable = v
	}
	if v := clampFloat64(user.Display.CostAlert, 0, 100000); v != nil {
		result.Display.CostAlert = v
	}
	if len(user.Display.ExcludeProjects) > 0 {
		result.Display.ExcludeProjects = user.Display.ExcludeProjects
	}
	if len(user.Display.ModelIcons) > 0 {
		result.Display.ModelIcons = user.Display.ModelIcons
	}
	if user.Display.LargeImage != "" {
		result.Display.LargeImage = user.Display.LargeImage
	}
	if user.Display.LargeText != "" {
		result.Display.LargeText = user.Display.LargeText
	}
	if user.Display.DiscordAppID != "" {
		result.Display.DiscordAppID = user.Display.DiscordAppID
	}

	// Buttons
	if len(user.Buttons) > 0 {
		result.Buttons = validateButtons(user.Buttons)
	}

	return &result
}

// validateButtons filters and validates button configs.
func validateButtons(buttons []ButtonConfig) []ButtonConfig {
	var valid []ButtonConfig
	for _, b := range buttons {
		if b.Label == "" || b.URL == "" {
			continue
		}
		if len(b.Label) > 32 {
			b.Label = b.Label[:32]
		}
		if !strings.HasPrefix(b.URL, "http://") && !strings.HasPrefix(b.URL, "https://") {
			continue
		}
		valid = append(valid, b)
		if len(valid) >= 2 {
			break
		}
	}
	return valid
}

// buildDetailsLine constructs the Details field for Discord presence.
func buildDetailsLine(session *SessionData, cfg *Config) string {
	showProject := showField(cfg.Show.ProjectName)
	showBranch := showField(cfg.Show.GitBranch) && session.GitBranch != ""

	if !showProject && !showBranch {
		return ""
	}

	prefix := cfg.Display.DetailsPrefix
	var base string

	if showProject && showBranch {
		base = fmt.Sprintf("%s: %s (%s)", prefix, session.ProjectName, session.GitBranch)
	} else if showProject {
		base = fmt.Sprintf("%s: %s", prefix, session.ProjectName)
	} else {
		base = fmt.Sprintf("%s: %s", prefix, session.GitBranch)
	}

	if showField(cfg.Show.Cost) && showField(cfg.Show.CostInDetails) {
		base = fmt.Sprintf("%s | $%.*f", base, intDefault(cfg.Display.CostPrecision, 4), session.TotalCost)
	}

	return truncate(base, 128)
}

// buildStateLine constructs the State field for Discord presence.
func buildStateLine(session *SessionData, cfg *Config) string {
	sep := cfg.Display.Separator
	precision := intDefault(cfg.Display.CostPrecision, 4)

	var parts []string

	if showField(cfg.Show.ModelName) && session.ModelName != "" {
		parts = append(parts, session.ModelName)
	}
	if showField(cfg.Show.Tokens) {
		if showFieldDefault(cfg.Show.SplitTokens, false) {
			parts = append(parts, fmt.Sprintf("%s in | %s out",
				formatNumber(session.InputTokens),
				formatNumber(session.OutputTokens)))
		} else {
			parts = append(parts, fmt.Sprintf("%s tokens", formatNumber(session.TotalTokens)))
		}
	}
	if showField(cfg.Show.Cost) && !showField(cfg.Show.CostInDetails) {
		parts = append(parts, fmt.Sprintf("$%.*f", precision, session.TotalCost))
	}

	return truncate(strings.Join(parts, sep), 128)
}

// formatTemplate replaces template variables with session data values.
// Available variables: {project}, {branch}, {model}, {tokens},
// {in_tokens}, {out_tokens}, {cost}, {duration}, {separator}
func formatTemplate(template string, session *SessionData, cfg *Config) string {
	precision := intDefault(cfg.Display.CostPrecision, 4)

	duration := time.Since(session.StartTime).Truncate(time.Second).String()

	replacements := map[string]string{
		"{project}":    session.ProjectName,
		"{branch}":     session.GitBranch,
		"{model}":      session.ModelName,
		"{tokens}":     formatNumber(session.TotalTokens),
		"{in_tokens}":  formatNumber(session.InputTokens),
		"{out_tokens}": formatNumber(session.OutputTokens),
		"{cost}":       fmt.Sprintf("%.*f", precision, session.TotalCost),
		"{duration}":   duration,
		"{separator}":  cfg.Display.Separator,
	}

	result := template
	for key, value := range replacements {
		result = strings.Replace(result, key, value, -1)
	}

	return truncate(result, 128)
}

// applyCostAlert prepends a warning symbol if cost exceeds the threshold.
func applyCostAlert(state string, cost float64, threshold *float64, idle, privacy bool) string {
	if idle || privacy || threshold == nil || *threshold <= 0 {
		return state
	}
	if cost >= *threshold {
		return truncate("\u26a0 "+state, 128)
	}
	return state
}

// matchModelIcon finds the matching icon key for a model name (case-insensitive substring).
func matchModelIcon(modelName string, icons map[string]string) (string, bool) {
	if len(icons) == 0 || modelName == "" {
		return "", false
	}
	lower := strings.ToLower(modelName)
	for key, iconKey := range icons {
		if strings.Contains(lower, strings.ToLower(key)) {
			return iconKey, true
		}
	}
	return "", false
}

// checkIdleDisable returns true if presence should be disabled due to extended idle.
func checkIdleDisable(idle bool, idleStart time.Time, secs int) bool {
	if secs <= 0 || !idle || idleStart.IsZero() {
		return false
	}
	return time.Since(idleStart) >= time.Duration(secs)*time.Second
}

// isProjectExcluded checks if a project path matches any exclusion pattern.
func isProjectExcluded(projectPath string, patterns []string) bool {
	if len(patterns) == 0 || projectPath == "" {
		return false
	}
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, projectPath); matched {
			return true
		}
	}
	return false
}

// truncate shortens a string to maxLen, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

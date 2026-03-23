package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Config represents the user configuration for Discord Rich Presence display.
type Config struct {
	Show    ShowConfig    `json:"show"`
	Display DisplayConfig `json:"display"`
}

// ShowConfig controls which fields are visible in the Discord presence.
type ShowConfig struct {
	ProjectName   *bool `json:"project_name"`
	GitBranch     *bool `json:"git_branch"`
	ModelName     *bool `json:"model_name"`
	Tokens        *bool `json:"tokens"`
	Cost          *bool `json:"cost"`
	CostInTooltip *bool `json:"cost_in_tooltip"`
	Duration      *bool `json:"duration"`
}

// DisplayConfig controls formatting and customization of the presence.
type DisplayConfig struct {
	DetailsPrefix string `json:"details_prefix"`
	Separator     string `json:"separator"`
	CostPrecision *int   `json:"cost_precision"`
	LargeText     string `json:"large_text"`
	DiscordAppID  string `json:"discord_app_id"`
}

var (
	configFilePath string
	currentConfig  *Config
)

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }

// showField safely dereferences a *bool, defaulting to true if nil.
func showField(ptr *bool) bool {
	if ptr == nil {
		return true
	}
	return *ptr
}

// DefaultConfig returns a config with all fields set to their default values.
func DefaultConfig() *Config {
	return &Config{
		Show: ShowConfig{
			ProjectName: boolPtr(true),
			GitBranch:   boolPtr(true),
			ModelName:   boolPtr(true),
			Tokens:      boolPtr(true),
			Cost:          boolPtr(true),
			CostInTooltip: boolPtr(false),
			Duration:      boolPtr(true),
		},
		Display: DisplayConfig{
			DetailsPrefix: "Working on",
			Separator:     " | ",
			CostPrecision: intPtr(4),
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
	if user.Show.Cost != nil {
		result.Show.Cost = user.Show.Cost
	}
	if user.Show.CostInTooltip != nil {
		result.Show.CostInTooltip = user.Show.CostInTooltip
	}
	if user.Show.Duration != nil {
		result.Show.Duration = user.Show.Duration
	}

	if user.Display.DetailsPrefix != "" {
		result.Display.DetailsPrefix = user.Display.DetailsPrefix
	}
	if user.Display.Separator != "" {
		result.Display.Separator = user.Display.Separator
	}
	if user.Display.CostPrecision != nil {
		p := *user.Display.CostPrecision
		if p < 0 {
			p = 0
		}
		if p > 10 {
			p = 10
		}
		result.Display.CostPrecision = intPtr(p)
	}
	if user.Display.LargeText != "" {
		result.Display.LargeText = user.Display.LargeText
	}
	if user.Display.DiscordAppID != "" {
		result.Display.DiscordAppID = user.Display.DiscordAppID
	}

	return &result
}

// buildDetailsLine constructs the Details field for Discord presence.
func buildDetailsLine(session *SessionData, cfg *Config) string {
	showProject := showField(cfg.Show.ProjectName)
	showBranch := showField(cfg.Show.GitBranch) && session.GitBranch != ""

	if !showProject && !showBranch {
		return ""
	}

	prefix := cfg.Display.DetailsPrefix

	if showProject && showBranch {
		return truncate(fmt.Sprintf("%s: %s (%s)", prefix, session.ProjectName, session.GitBranch), 128)
	}
	if showProject {
		return truncate(fmt.Sprintf("%s: %s", prefix, session.ProjectName), 128)
	}
	return truncate(fmt.Sprintf("%s: %s", prefix, session.GitBranch), 128)
}

// buildStateLine constructs the State field for Discord presence.
func buildStateLine(session *SessionData, cfg *Config) string {
	sep := cfg.Display.Separator
	precision := 4
	if cfg.Display.CostPrecision != nil {
		precision = *cfg.Display.CostPrecision
	}

	var parts []string

	if showField(cfg.Show.ModelName) && session.ModelName != "" {
		parts = append(parts, session.ModelName)
	}
	if showField(cfg.Show.Tokens) {
		parts = append(parts, fmt.Sprintf("%s tokens", formatNumber(session.TotalTokens)))
	}
	if showField(cfg.Show.Cost) && !showField(cfg.Show.CostInTooltip) {
		parts = append(parts, fmt.Sprintf("$%.*f", precision, session.TotalCost))
	}

	return truncate(strings.Join(parts, sep), 128)
}

// buildLargeText constructs the tooltip text, optionally including cost.
func buildLargeText(session *SessionData, cfg *Config) string {
	base := cfg.Display.LargeText
	if showField(cfg.Show.Cost) && showField(cfg.Show.CostInTooltip) {
		precision := 4
		if cfg.Display.CostPrecision != nil {
			precision = *cfg.Display.CostPrecision
		}
		return fmt.Sprintf("%s — $%.*f", base, precision, session.TotalCost)
	}
	return base
}

// truncate shortens a string to maxLen, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

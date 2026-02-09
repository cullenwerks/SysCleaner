package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"syscleaner/pkg/cleaner"
)

// RAMMonitorSettings holds threshold configuration for RAM monitoring.
type RAMMonitorSettings struct {
	FreeThresholdPercent    float64 `json:"free_threshold_percent"`
	StandbyThresholdPercent float64 `json:"standby_threshold_percent"`
}

// UIPreferences stores persistent UI state.
type UIPreferences struct {
	LastActiveTab string `json:"last_active_tab"`
}

// Config is the top-level application configuration.
type Config struct {
	ProcessWhitelist    []string
	DefaultCleanOptions cleaner.CleanOptions
	RAMMonitor          RAMMonitorSettings
	UIPreferences       UIPreferences
	ActiveProfile       string
}

// ConfigDir returns the path to the SysCleaner configuration directory.
// The directory is derived from os.UserConfigDir() with "SysCleaner" appended.
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine user config directory: %w", err)
	}
	return filepath.Join(base, "SysCleaner"), nil
}

// configFilePath returns the full path to the configuration file.
func configFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// LoadConfig reads the configuration from disk. If the file does not exist,
// a default configuration is returned without error.
func LoadConfig() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var d configData
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	return fromConfigData(d), nil
}

// SaveConfig writes the configuration to disk, creating the config directory
// if it does not already exist.
func SaveConfig(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(toConfigData(cfg), "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	return nil
}

// DefaultConfig returns a Config populated with sensible default values.
func DefaultConfig() *Config {
	return &Config{
		ProcessWhitelist: []string{},
		DefaultCleanOptions: cleaner.CleanOptions{
			WindowsTemp:    true,
			UserTemp:       true,
			Prefetch:       true,
			ThumbnailCache: true,
			DNSCache:       true,
			ChromeCache:    true,
			FirefoxCache:   true,
			EdgeCache:      true,
		},
		RAMMonitor: RAMMonitorSettings{
			FreeThresholdPercent:    15.0,
			StandbyThresholdPercent: 50.0,
		},
		UIPreferences: UIPreferences{
			LastActiveTab: "dashboard",
		},
		ActiveProfile: "default",
	}
}

// ---------------------------------------------------------------------------
// JSON serialization helpers
//
// cleaner.CleanOptions contains a ProgressFunc field (function type) that
// encoding/json cannot marshal. The types below mirror the serializable
// fields so that Config can be round-tripped through JSON transparently.
// ---------------------------------------------------------------------------

// cleanOptionsData is a JSON-safe mirror of cleaner.CleanOptions.
type cleanOptionsData struct {
	// System categories
	WindowsTemp          bool `json:"windows_temp"`
	UserTemp             bool `json:"user_temp"`
	WindowsUpdate        bool `json:"windows_update"`
	WindowsInstaller     bool `json:"windows_installer"`
	Prefetch             bool `json:"prefetch"`
	CrashDumps           bool `json:"crash_dumps"`
	ErrorReports         bool `json:"error_reports"`
	ThumbnailCache       bool `json:"thumbnail_cache"`
	IconCache            bool `json:"icon_cache"`
	FontCache            bool `json:"font_cache"`
	ShaderCache          bool `json:"shader_cache"`
	DNSCache             bool `json:"dns_cache"`
	WindowsLogs          bool `json:"windows_logs"`
	EventLogs            bool `json:"event_logs"`
	DeliveryOptimization bool `json:"delivery_optimization"`
	RecycleBin           bool `json:"recycle_bin"`

	// Application categories
	ChromeCache  bool `json:"chrome_cache"`
	FirefoxCache bool `json:"firefox_cache"`
	EdgeCache    bool `json:"edge_cache"`
	BraveCache   bool `json:"brave_cache"`
	OperaCache   bool `json:"opera_cache"`
	DiscordCache bool `json:"discord_cache"`
	SpotifyCache bool `json:"spotify_cache"`
	SteamCache   bool `json:"steam_cache"`
	TeamsCache   bool `json:"teams_cache"`
	VSCodeCache  bool `json:"vscode_cache"`
	JavaCache    bool `json:"java_cache"`

	// Execution options
	DryRun bool `json:"dry_run"`
}

// configData is the JSON-serializable representation of Config.
type configData struct {
	ProcessWhitelist    []string           `json:"process_whitelist"`
	DefaultCleanOptions cleanOptionsData   `json:"default_clean_options"`
	RAMMonitor          RAMMonitorSettings `json:"ram_monitor"`
	UIPreferences       UIPreferences      `json:"ui_preferences"`
	ActiveProfile       string             `json:"active_profile"`
}

func toCleanOptionsData(o cleaner.CleanOptions) cleanOptionsData {
	return cleanOptionsData{
		WindowsTemp:          o.WindowsTemp,
		UserTemp:             o.UserTemp,
		WindowsUpdate:        o.WindowsUpdate,
		WindowsInstaller:     o.WindowsInstaller,
		Prefetch:             o.Prefetch,
		CrashDumps:           o.CrashDumps,
		ErrorReports:         o.ErrorReports,
		ThumbnailCache:       o.ThumbnailCache,
		IconCache:            o.IconCache,
		FontCache:            o.FontCache,
		ShaderCache:          o.ShaderCache,
		DNSCache:             o.DNSCache,
		WindowsLogs:          o.WindowsLogs,
		EventLogs:            o.EventLogs,
		DeliveryOptimization: o.DeliveryOptimization,
		RecycleBin:           o.RecycleBin,
		ChromeCache:          o.ChromeCache,
		FirefoxCache:         o.FirefoxCache,
		EdgeCache:            o.EdgeCache,
		BraveCache:           o.BraveCache,
		OperaCache:           o.OperaCache,
		DiscordCache:         o.DiscordCache,
		SpotifyCache:         o.SpotifyCache,
		SteamCache:           o.SteamCache,
		TeamsCache:           o.TeamsCache,
		VSCodeCache:          o.VSCodeCache,
		JavaCache:            o.JavaCache,
		DryRun:               o.DryRun,
	}
}

func fromCleanOptionsData(d cleanOptionsData) cleaner.CleanOptions {
	return cleaner.CleanOptions{
		WindowsTemp:          d.WindowsTemp,
		UserTemp:             d.UserTemp,
		WindowsUpdate:        d.WindowsUpdate,
		WindowsInstaller:     d.WindowsInstaller,
		Prefetch:             d.Prefetch,
		CrashDumps:           d.CrashDumps,
		ErrorReports:         d.ErrorReports,
		ThumbnailCache:       d.ThumbnailCache,
		IconCache:            d.IconCache,
		FontCache:            d.FontCache,
		ShaderCache:          d.ShaderCache,
		DNSCache:             d.DNSCache,
		WindowsLogs:          d.WindowsLogs,
		EventLogs:            d.EventLogs,
		DeliveryOptimization: d.DeliveryOptimization,
		RecycleBin:           d.RecycleBin,
		ChromeCache:          d.ChromeCache,
		FirefoxCache:         d.FirefoxCache,
		EdgeCache:            d.EdgeCache,
		BraveCache:           d.BraveCache,
		OperaCache:           d.OperaCache,
		DiscordCache:         d.DiscordCache,
		SpotifyCache:         d.SpotifyCache,
		SteamCache:           d.SteamCache,
		TeamsCache:           d.TeamsCache,
		VSCodeCache:          d.VSCodeCache,
		JavaCache:            d.JavaCache,
		DryRun:               d.DryRun,
	}
}

func toConfigData(c *Config) configData {
	return configData{
		ProcessWhitelist:    c.ProcessWhitelist,
		DefaultCleanOptions: toCleanOptionsData(c.DefaultCleanOptions),
		RAMMonitor:          c.RAMMonitor,
		UIPreferences:       c.UIPreferences,
		ActiveProfile:       c.ActiveProfile,
	}
}

func fromConfigData(d configData) *Config {
	return &Config{
		ProcessWhitelist:    d.ProcessWhitelist,
		DefaultCleanOptions: fromCleanOptionsData(d.DefaultCleanOptions),
		RAMMonitor:          d.RAMMonitor,
		UIPreferences:       d.UIPreferences,
		ActiveProfile:       d.ActiveProfile,
	}
}

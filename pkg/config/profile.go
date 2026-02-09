package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProfileCleanOptions mirrors cleaner.CleanOptions with only the
// serializable fields. A separate type is used to avoid a circular
// dependency between the config and cleaner packages when profiles
// are loaded independently of the cleaner runtime.
type ProfileCleanOptions struct {
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

// GamingConfig holds gaming-mode specific settings for a profile.
type GamingConfig struct {
	UseExtremeMode bool `json:"use_extreme_mode"`
	CPUBoost       int  `json:"cpu_boost"`
	RAMReserveGB   int  `json:"ram_reserve_gb"`
}

// Profile represents a named collection of settings that can be
// switched between at runtime.
type Profile struct {
	Name             string              `json:"name"`
	ProcessWhitelist []string            `json:"process_whitelist"`
	CleanOptions     ProfileCleanOptions `json:"clean_options"`
	GamingConfig     GamingConfig        `json:"gaming_config"`
}

// profilesDir returns the path to the profiles directory, which is
// ConfigDir()/profiles/.
func profilesDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "profiles"), nil
}

// profilePath returns the file path for the given profile name.
// The name is sanitised to prevent directory traversal.
func profilePath(name string) (string, error) {
	dir, err := profilesDir()
	if err != nil {
		return "", err
	}
	safe := filepath.Base(name)
	if safe == "." || safe == ".." || safe == string(filepath.Separator) {
		return "", fmt.Errorf("invalid profile name: %q", name)
	}
	return filepath.Join(dir, safe+".json"), nil
}

// LoadProfile reads a profile by name from the profiles directory.
func LoadProfile(name string) (*Profile, error) {
	path, err := profilePath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile %q not found", name)
		}
		return nil, fmt.Errorf("reading profile %q: %w", name, err)
	}

	p := &Profile{}
	if err := json.Unmarshal(data, p); err != nil {
		return nil, fmt.Errorf("parsing profile %q: %w", name, err)
	}
	return p, nil
}

// SaveProfile writes a profile to the profiles directory, creating
// the directory if it does not already exist. The profile name is
// used to derive the file name.
func SaveProfile(p *Profile) error {
	dir, err := profilesDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating profiles directory: %w", err)
	}

	path, err := profilePath(p.Name)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling profile %q: %w", p.Name, err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing profile %q: %w", p.Name, err)
	}
	return nil
}

// ListProfiles returns the names of all saved profiles by scanning the
// profiles directory for JSON files.
func ListProfiles() ([]string, error) {
	dir, err := profilesDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading profiles directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".json") {
			names = append(names, strings.TrimSuffix(name, ".json"))
		}
	}
	return names, nil
}

// DeleteProfile removes a saved profile by name.
func DeleteProfile(name string) error {
	path, err := profilePath(name)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("profile %q not found", name)
		}
		return fmt.Errorf("deleting profile %q: %w", name, err)
	}
	return nil
}

// DefaultProfile returns a Profile populated with sensible default values.
func DefaultProfile() *Profile {
	return &Profile{
		Name:             "default",
		ProcessWhitelist: []string{},
		CleanOptions: ProfileCleanOptions{
			WindowsTemp:    true,
			UserTemp:       true,
			Prefetch:       true,
			ThumbnailCache: true,
			DNSCache:       true,
			ChromeCache:    true,
			FirefoxCache:   true,
			EdgeCache:      true,
		},
		GamingConfig: GamingConfig{
			UseExtremeMode: false,
			CPUBoost:       0,
			RAMReserveGB:   0,
		},
	}
}

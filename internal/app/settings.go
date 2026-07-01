package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	BytesPerMB             int64 = 1024 * 1024
	MinimumSplitPartSizeMB int64 = 5
	NoSplitMaxPartSize     int64 = 1 << 62

	SourceCleanupKeep   = "keep"
	SourceCleanupDelete = "delete"

	NoiseFileIgnoreEverywhere              = "ignore_everywhere"
	NoiseFileIgnoreDuringVerifyAndMatching = "ignore_during_verify_and_matching"
	NoiseFileDoNotIgnore                   = "do_not_ignore"

	FailureHandlingAbort    = "abort"
	FailureHandlingContinue = "continue"

	ThemeSystem = "system"
	ThemeLight  = "light"
	ThemeDark   = "dark"

	LanguageSystem = "system"
	LanguageENUS   = "en-US"
	LanguageZHCN   = "zh-CN"
)

type Settings struct {
	DefaultMaxPartSize int64  `json:"defaultMaxPartSize"`
	SourceCleanupMode  string `json:"sourceCleanupMode"`
	NoiseFileHandling  string `json:"noiseFileHandling"`
	Theme              string `json:"theme"`
	Language           string `json:"language"`
	BackupRetention    int    `json:"backupRetention"`
	FailureHandling    string `json:"failureHandling"`
}

func DefaultSettings() Settings {
	return Settings{
		DefaultMaxPartSize: 0,
		SourceCleanupMode:  SourceCleanupDelete,
		NoiseFileHandling:  NoiseFileIgnoreEverywhere,
		Theme:              ThemeSystem,
		Language:           LanguageSystem,
		BackupRetention:    DefaultBackupRetention,
		FailureHandling:    FailureHandlingAbort,
	}
}

func (s Service) SettingsPath() string {
	return filepath.Join(s.DataDir, "settings.json")
}

func (s Service) ReadSettings() (Settings, error) {
	if err := s.EnsureDataDir(); err != nil {
		return Settings{}, err
	}

	path := s.SettingsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultSettings(), nil
		}
		return Settings{}, fmt.Errorf("read settings: %w", err)
	}

	settings := DefaultSettings()
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}, fmt.Errorf("decode settings: %w", err)
	}
	return normalizeSettings(settings)
}

func (s Service) SaveSettings(settings Settings) (Settings, error) {
	if err := s.EnsureDataDir(); err != nil {
		return Settings{}, err
	}

	normalized, err := normalizeSettings(settings)
	if err != nil {
		return Settings{}, err
	}

	data, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return Settings{}, fmt.Errorf("encode settings: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.SettingsPath(), data, 0o600); err != nil {
		return Settings{}, fmt.Errorf("write settings: %w", err)
	}
	return normalized, nil
}

func normalizeSettings(settings Settings) (Settings, error) {
	if settings.DefaultMaxPartSize < 0 {
		return Settings{}, fmt.Errorf("default max part size must not be negative")
	}
	if settings.DefaultMaxPartSize < MinimumSplitPartSizeMB*BytesPerMB {
		settings.DefaultMaxPartSize = 0
	}

	switch settings.SourceCleanupMode {
	case "":
		settings.SourceCleanupMode = SourceCleanupDelete
	case SourceCleanupKeep, SourceCleanupDelete:
	default:
		return Settings{}, fmt.Errorf("unsupported source cleanup mode %q", settings.SourceCleanupMode)
	}

	switch settings.NoiseFileHandling {
	case "":
		settings.NoiseFileHandling = NoiseFileIgnoreEverywhere
	case NoiseFileIgnoreEverywhere, NoiseFileIgnoreDuringVerifyAndMatching, NoiseFileDoNotIgnore:
	default:
		return Settings{}, fmt.Errorf("unsupported noise file handling mode %q", settings.NoiseFileHandling)
	}

	switch settings.Theme {
	case "", ThemeSystem:
		settings.Theme = ThemeSystem
	case ThemeLight, ThemeDark:
	default:
		return Settings{}, fmt.Errorf("unsupported theme %q", settings.Theme)
	}

	switch settings.Language {
	case "", LanguageSystem:
		settings.Language = LanguageSystem
	case LanguageENUS, LanguageZHCN:
	default:
		return Settings{}, fmt.Errorf("unsupported language %q", settings.Language)
	}

	if settings.BackupRetention < 0 {
		return Settings{}, fmt.Errorf("backup retention must not be negative")
	}
	if settings.BackupRetention == 0 {
		settings.BackupRetention = DefaultBackupRetention
	}

	switch settings.FailureHandling {
	case "":
		settings.FailureHandling = FailureHandlingAbort
	case FailureHandlingAbort, FailureHandlingContinue:
	default:
		return Settings{}, fmt.Errorf("unsupported failure handling mode %q", settings.FailureHandling)
	}

	return settings, nil
}

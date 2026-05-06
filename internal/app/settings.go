package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	GuideFormatTXT = "txt"
	GuideFormatMD  = "md"

	SourceCleanupAsk    = "ask"
	SourceCleanupKeep   = "keep"
	SourceCleanupDelete = "delete"

	ThemeSystem = "system"
	ThemeLight  = "light"
	ThemeDark   = "dark"

	LanguageSystem = "system"
	LanguageENUS   = "en-US"
	LanguageZHCN   = "zh-CN"
)

type Settings struct {
	OperationGuideFormat   string   `json:"operationGuideFormat"`
	DefaultMaxPartSize     int64    `json:"defaultMaxPartSize"`
	SourceCleanupMode      string   `json:"sourceCleanupMode"`
	RememberRecentPaths    bool     `json:"rememberRecentPaths"`
	RecentPaths            []string `json:"recentPaths"`
	WindowStatePersistence bool     `json:"windowStatePersistence"`
	Theme                  string   `json:"theme"`
	Language               string   `json:"language"`
}

func DefaultSettings() Settings {
	return Settings{
		OperationGuideFormat:   GuideFormatTXT,
		DefaultMaxPartSize:     0,
		SourceCleanupMode:      SourceCleanupAsk,
		RememberRecentPaths:    true,
		RecentPaths:            []string{},
		WindowStatePersistence: true,
		Theme:                  ThemeSystem,
		Language:               LanguageSystem,
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

func (s Service) ClearRecentPaths() (Settings, error) {
	settings, err := s.ReadSettings()
	if err != nil {
		return Settings{}, err
	}
	settings.RecentPaths = []string{}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return Settings{}, fmt.Errorf("encode settings: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.SettingsPath(), data, 0o600); err != nil {
		return Settings{}, fmt.Errorf("write settings: %w", err)
	}
	return settings, nil
}

func normalizeSettings(settings Settings) (Settings, error) {
	switch settings.OperationGuideFormat {
	case "", GuideFormatTXT:
		settings.OperationGuideFormat = GuideFormatTXT
	case GuideFormatMD:
	default:
		return Settings{}, fmt.Errorf("unsupported operation guide format %q", settings.OperationGuideFormat)
	}

	if settings.DefaultMaxPartSize < 0 {
		return Settings{}, fmt.Errorf("default max part size must not be negative")
	}

	switch settings.SourceCleanupMode {
	case "", SourceCleanupAsk:
		settings.SourceCleanupMode = SourceCleanupAsk
	case SourceCleanupKeep, SourceCleanupDelete:
	default:
		return Settings{}, fmt.Errorf("unsupported source cleanup mode %q", settings.SourceCleanupMode)
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

	if settings.RecentPaths == nil {
		settings.RecentPaths = []string{}
	}
	return settings, nil
}

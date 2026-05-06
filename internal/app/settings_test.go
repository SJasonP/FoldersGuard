package app

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadSettingsReturnsDefaultsWhenMissing(t *testing.T) {
	service, err := NewService(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}

	settings, err := service.ReadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(settings, DefaultSettings()) {
		t.Fatalf("settings = %+v, want %+v", settings, DefaultSettings())
	}
}

func TestSaveSettingsPersistsNormalizedValues(t *testing.T) {
	service, err := NewService(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}

	saved, err := service.SaveSettings(Settings{
		OperationGuideFormat:   GuideFormatMD,
		DefaultMaxPartSize:     4096,
		SourceCleanupMode:      SourceCleanupKeep,
		RememberRecentPaths:    false,
		WindowStatePersistence: false,
		Theme:                  ThemeDark,
		Language:               LanguageZHCN,
	})
	if err != nil {
		t.Fatal(err)
	}
	if saved.OperationGuideFormat != GuideFormatMD || saved.DefaultMaxPartSize != 4096 || saved.Theme != ThemeDark || saved.Language != LanguageZHCN {
		t.Fatalf("saved settings = %+v", saved)
	}

	read, err := service.ReadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(read, saved) {
		t.Fatalf("read settings = %+v, want %+v", read, saved)
	}
}

func TestClearRecentPaths(t *testing.T) {
	service, err := NewService(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.SaveSettings(Settings{
		OperationGuideFormat:   GuideFormatTXT,
		DefaultMaxPartSize:     0,
		SourceCleanupMode:      SourceCleanupAsk,
		RememberRecentPaths:    true,
		RecentPaths:            []string{"a", "b"},
		WindowStatePersistence: true,
		Theme:                  ThemeSystem,
		Language:               LanguageSystem,
	})
	if err != nil {
		t.Fatal(err)
	}

	cleared, err := service.ClearRecentPaths()
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.RecentPaths) != 0 {
		t.Fatalf("recent paths = %v, want empty", cleared.RecentPaths)
	}
}

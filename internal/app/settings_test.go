package app

import (
	"errors"
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
		DefaultMaxPartSize: 8 * BytesPerMB,
		SourceCleanupMode:  SourceCleanupAfterPart,
		NoiseFileHandling:  NoiseFileDoNotIgnore,
		Theme:              ThemeDark,
		Language:           LanguageZHCN,
	})
	if err != nil {
		t.Fatal(err)
	}
	if saved.DefaultMaxPartSize != 8*BytesPerMB || saved.SourceCleanupMode != SourceCleanupAfterPart || saved.IncrementalSourceCleanup || saved.NoiseFileHandling != NoiseFileDoNotIgnore || saved.Theme != ThemeDark || saved.Language != LanguageZHCN {
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

func TestSaveSettingsDisablesSmallDefaultMaxPartSize(t *testing.T) {
	service, err := NewService(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}

	saved, err := service.SaveSettings(Settings{
		DefaultMaxPartSize: 4 * BytesPerMB,
	})
	if err != nil {
		t.Fatal(err)
	}
	if saved.DefaultMaxPartSize != 0 {
		t.Fatalf("default max part size = %d, want disabled", saved.DefaultMaxPartSize)
	}
}

func TestSaveSettingsRequiresSplittingForIncrementalCleanup(t *testing.T) {
	service, err := NewService(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.SaveSettings(Settings{SourceCleanupMode: SourceCleanupAfterPart})
	if !errors.Is(err, ErrIncrementalRequiresSplit) {
		t.Fatalf("error = %v, want incremental cleanup split requirement", err)
	}
}

func TestNormalizeSettingsMigratesLegacyIncrementalCleanup(t *testing.T) {
	settings, err := normalizeSettings(Settings{
		DefaultMaxPartSize:       8 * BytesPerMB,
		SourceCleanupMode:        SourceCleanupDelete,
		IncrementalSourceCleanup: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if settings.SourceCleanupMode != SourceCleanupAfterPart || settings.IncrementalSourceCleanup {
		t.Fatalf("migrated settings = %+v", settings)
	}
}

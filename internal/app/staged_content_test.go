package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStagedContentDirPrefersDesktop(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, "Desktop"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	service := Service{DataDir: filepath.Join(t.TempDir(), "data")}
	want := filepath.Join(home, "Desktop")
	if got := service.StagedContentDir(); got != want {
		t.Fatalf("staged content dir = %q, want %q", got, want)
	}
}

func TestStagedContentDirFallsBackToDataDirWithoutDesktop(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir := filepath.Join(t.TempDir(), "data")
	service := Service{DataDir: dataDir}
	want := filepath.Join(dataDir, "staged-content")
	if got := service.StagedContentDir(); got != want {
		t.Fatalf("staged content dir = %q, want %q", got, want)
	}
}

func TestStagedContentDirUsesConfiguredLocation(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, "Desktop"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	custom := t.TempDir()
	service := Service{DataDir: filepath.Join(t.TempDir(), "data")}
	if _, err := service.SaveSettings(Settings{StagedContentLocation: custom}); err != nil {
		t.Fatal(err)
	}
	// The configured location wins over the Desktop.
	if got := service.StagedContentDir(); got != custom {
		t.Fatalf("staged content dir = %q, want configured %q", got, custom)
	}
}

func TestSaveSettingsRejectsRelativeStagedContentLocation(t *testing.T) {
	service := Service{DataDir: filepath.Join(t.TempDir(), "data")}
	if _, err := service.SaveSettings(Settings{StagedContentLocation: "relative/dir"}); err == nil {
		t.Fatal("expected a relative staged content location to be rejected")
	}
}

func TestStagedContentDirectoryNameUsesProjectNameAndLocalMinute(t *testing.T) {
	createdAt := time.Date(2026, 5, 12, 10, 1, 0, 0, time.Local)
	got := stagedContentDirectoryName("  My:Vault/Archive?*  ", createdAt)
	want := "My-Vault-Archive 2026-05-12 10.01"
	if got != want {
		t.Fatalf("staged directory name = %q, want %q", got, want)
	}
}

func TestPrepareProjectChangeStagingUsesLocalProjectName(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, "Desktop"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	service := Service{DataDir: filepath.Join(t.TempDir(), "data")}
	if err := service.writeProjectNames(map[string]string{"project-id": "My Vault"}); err != nil {
		t.Fatal(err)
	}

	staging, err := service.prepareProjectChangeStaging("project-id")
	if err != nil {
		t.Fatal(err)
	}
	if !staging.OnDesktop {
		t.Fatalf("staging should be marked as desktop: %+v", staging)
	}
	if !strings.HasPrefix(staging.Name, "My Vault ") {
		t.Fatalf("staging name = %q", staging.Name)
	}
	if filepath.Dir(staging.Path) != filepath.Join(home, "Desktop") {
		t.Fatalf("staging path = %q", staging.Path)
	}
	assertExists(t, staging.Path)
}

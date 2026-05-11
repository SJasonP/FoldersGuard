package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRenderOperationGuideUsesSimplifiedChinese(t *testing.T) {
	createdAt := time.Date(2026, 5, 11, 8, 0, 0, 0, time.UTC)
	guide := renderOperationGuide("project-id", []ProjectContentOperation{
		{Type: "upload", SourcePath: "staged/new", TargetPath: "encrypted/new"},
		{Type: "move", SourcePath: "encrypted/old", TargetPath: "encrypted/archive/old"},
		{Type: "delete", TargetPath: "encrypted/remove"},
	}, createdAt, GuideFormatMD, LanguageZHCN)

	for _, want := range []string{
		"# FoldersGuard 操作指南",
		"请按以下步骤手动处理加密内容",
		"项目 ID",
		"上传 `staged/new` -> `encrypted/new`",
		"移动 `encrypted/old` -> `encrypted/archive/old`",
		"删除 `encrypted/remove`",
	} {
		if !strings.Contains(guide, want) {
			t.Fatalf("guide missing %q:\n%s", want, guide)
		}
	}
}

func TestOperationGuideLanguageFallsBackToSettings(t *testing.T) {
	if got := operationGuideLanguage("", LanguageZHCN); got != LanguageZHCN {
		t.Fatalf("language = %q, want %q", got, LanguageZHCN)
	}
	if got := operationGuideLanguage("", LanguageSystem); got != LanguageENUS {
		t.Fatalf("language = %q, want %q", got, LanguageENUS)
	}
	if got := operationGuideLanguage(LanguageENUS, LanguageZHCN); got != LanguageENUS {
		t.Fatalf("language = %q, want request language", got)
	}
}

func TestOperationGuidesDirPrefersDesktop(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, "Desktop"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	service := Service{DataDir: filepath.Join(t.TempDir(), "data")}
	want := filepath.Join(home, "Desktop", "FoldersGuard", "operation-guides")
	if got := service.OperationGuidesDir(); got != want {
		t.Fatalf("operation guides dir = %q, want %q", got, want)
	}
}

func TestStagedContentDirPrefersDesktop(t *testing.T) {
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, "Desktop"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	service := Service{DataDir: filepath.Join(t.TempDir(), "data")}
	want := filepath.Join(home, "Desktop", "FoldersGuard", "staged-content")
	if got := service.StagedContentDir(); got != want {
		t.Fatalf("staged content dir = %q, want %q", got, want)
	}
}

func TestOperationGuidesDirFallsBackToDataDirWithoutDesktop(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dataDir := filepath.Join(t.TempDir(), "data")
	service := Service{DataDir: dataDir}
	want := filepath.Join(dataDir, "operation-guides")
	if got := service.OperationGuidesDir(); got != want {
		t.Fatalf("operation guides dir = %q, want %q", got, want)
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

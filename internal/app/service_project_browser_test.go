package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestServiceOpenProjectBrowser(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "root.txt"), []byte("root"), 0o600); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateProject(ctx, CreateProjectInput{
		SourcePath:    source,
		ContentOutput: encrypted,
		Password:      password,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}

	state, err := service.OpenProjectBrowser(ctx, OpenProjectBrowserInput{
		ProjectID: created.ProjectID,
		Password:  password,
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.ProjectID != created.ProjectID || state.ProjectName != "source" || state.RootFolderName != "source" {
		t.Fatalf("browser state = %+v", state)
	}
	if state.ContentConnected {
		t.Fatalf("content connected = true, want false")
	}
	if state.Files != 2 || state.Folders != 2 || state.Parts != 0 {
		t.Fatalf("browser counts = %+v", state)
	}
	if len(state.Items) != 4 {
		t.Fatalf("browser items = %d, want 4", len(state.Items))
	}
	wantSizes := map[string]int64{
		"source":               9,
		"source/docs":          5,
		"source/docs/note.txt": 5,
		"source/root.txt":      4,
	}
	for _, item := range state.Items {
		if !item.MetadataCaptured {
			t.Fatalf("item metadata captured = false for %s", item.Path)
		}
		if !item.ContentAvailable {
			t.Fatalf("item content available = false for %s without content connection", item.Path)
		}
		if item.Size != wantSizes[item.Path] {
			t.Fatalf("item %s size = %d, want %d", item.Path, item.Size, wantSizes[item.Path])
		}
	}
}

func TestServiceOpenProjectBrowserWithConnectedContent(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateProject(ctx, CreateProjectInput{
		SourcePath:    source,
		ContentOutput: encrypted,
		Password:      password,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}

	state, err := service.OpenProjectBrowser(ctx, OpenProjectBrowserInput{
		ProjectID:     created.ProjectID,
		Password:      password,
		EncryptedRoot: encrypted,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !state.ContentConnected || state.EncryptedRoot != encrypted {
		t.Fatalf("browser state = %+v", state)
	}
	for _, item := range state.Items {
		if !item.ContentAvailable {
			t.Fatalf("item content available = false for %s with connected content", item.Path)
		}
	}
}

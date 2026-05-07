package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"foldersguard/internal/format"
)

func TestServiceListShareableItemsAndCreateShare(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	sharePath := filepath.Join(root, "share"+format.SetExtension)
	dataDir := filepath.Join(root, "data")
	projectPassword := "project-password"
	sharePassword := "share-password"

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
		Password:      projectPassword,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}

	items, err := service.ListShareableItems(ctx, DatabaseOpen{
		ProjectRef: created.ProjectID,
		Password:   projectPassword,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 4 {
		t.Fatalf("shareable items = %d, want 4", len(items))
	}
	if items[0].Path != "source" || items[0].Type != "folder" {
		t.Fatalf("root shareable item = %+v", items[0])
	}

	result, err := service.CreateShare(ctx, CreateShareInput{
		ProjectID:         created.ProjectID,
		ProjectPassword:   projectPassword,
		ItemPaths:         []string{"source/docs"},
		OutputPath:        sharePath,
		PasswordProtected: true,
		SharePassword:     sharePassword,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ProjectID != created.ProjectID || result.ShareID == "" || result.OutputPath != sharePath {
		t.Fatalf("create share result = %+v", result)
	}
	if result.TopLevelItems != 1 || result.Files != 1 || result.Folders != 1 || !result.PasswordProtected {
		t.Fatalf("create share counts = %+v", result)
	}
	if len(result.ContentLocations) != 1 {
		t.Fatalf("content locations = %d, want 1", len(result.ContentLocations))
	}

	summary, err := service.InspectShare(ctx, ShareOpen{
		DatabasePath: sharePath,
		Password:     sharePassword,
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary.ShareID != result.ShareID || summary.Files != 1 || !summary.PasswordProtected {
		t.Fatalf("share summary = %+v", summary)
	}
}

func TestServiceCreateUnprotectedShare(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	sharePath := filepath.Join(root, "share"+format.SetExtension)
	dataDir := filepath.Join(root, "data")
	projectPassword := "project-password"

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.CreateProject(ctx, CreateProjectInput{
		SourcePath:    source,
		ContentOutput: encrypted,
		Password:      projectPassword,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.CreateShare(ctx, CreateShareInput{
		ProjectID:         created.ProjectID,
		ProjectPassword:   projectPassword,
		ItemPaths:         []string{"source/note.txt"},
		OutputPath:        sharePath,
		PasswordProtected: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.PasswordProtected {
		t.Fatalf("create share result = %+v, want unprotected", result)
	}

	summary, err := service.InspectShare(ctx, ShareOpen{
		DatabasePath: sharePath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary.PasswordProtected {
		t.Fatalf("share summary = %+v, want unprotected", summary)
	}
}

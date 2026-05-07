package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestServiceApplyProjectRenameChange(t *testing.T) {
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

	result, err := service.ApplyProjectChanges(ctx, ApplyProjectChangesInput{
		ProjectID: created.ProjectID,
		Password:  password,
		RenameChanges: []ProjectRenameChange{{
			ItemPath: "source/docs",
			NewName:  "papers",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ProjectID != created.ProjectID || result.AppliedRenames != 1 {
		t.Fatalf("apply result = %+v", result)
	}
	if !browserHasPath(result.BrowserState, "source/papers") {
		t.Fatalf("browser state missing renamed path: %+v", result.BrowserState.Items)
	}
	if browserHasPath(result.BrowserState, "source/docs") {
		t.Fatalf("browser state still has old path: %+v", result.BrowserState.Items)
	}
}

func TestServiceApplyProjectRenameRejectsRoot(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "project-password"

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
		Password:      password,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = service.ApplyProjectChanges(ctx, ApplyProjectChangesInput{
		ProjectID: created.ProjectID,
		Password:  password,
		RenameChanges: []ProjectRenameChange{{
			ItemPath: "source",
			NewName:  "renamed-source",
		}},
	})
	if err == nil {
		t.Fatal("expected root rename to be rejected")
	}
}

func browserHasPath(state ProjectBrowserState, path string) bool {
	for _, item := range state.Items {
		if item.Path == path {
			return true
		}
	}
	return false
}

package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/project"
)

func TestServicePathsUseConfiguredDataDir(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatal(err)
	}

	activePath, err := service.ActiveProjectDatabasePath("project-id")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dataDir, "projects", "project-id"+format.ProjectExtension)
	if activePath != want {
		t.Fatalf("active project path = %q, want %q", activePath, want)
	}

	sharePath := filepath.Join(dataDir, "share"+format.SetExtension)
	resolved, err := service.DatabasePathFromProjectRef(sharePath)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != sharePath {
		t.Fatalf("share path = %q, want %q", resolved, sharePath)
	}

	_, err = service.DatabasePathFromProjectRef(filepath.Join(dataDir, "export"+format.ProjectExtension))
	if err == nil || !strings.Contains(err.Error(), "must be imported before use") {
		t.Fatalf("exported project path error = %v, want import requirement", err)
	}
}

func TestDefaultDataDirUsesFoldersGuardName(t *testing.T) {
	dataDir, err := DefaultDataDir()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(dataDir) != format.DataDirName {
		t.Fatalf("default data dir = %q, want base %q", dataDir, format.DataDirName)
	}
}

func TestServiceInspectAndVerify(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	databasePath := filepath.Join(root, "data", "projects", "project-id"+format.ProjectExtension)
	password := "test-password"

	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "note.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	scan, err := fswalk.ScanTopFolder(source)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := project.Planner{MaxPartSize: 1024}.Plan(scan)
	if err != nil {
		t.Fatal(err)
	}
	if err := (project.Executor{OutputRoot: encrypted}).EncryptContent(ctx, plan); err != nil {
		t.Fatal(err)
	}
	if err := WriteProjectDatabase(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	}, plan); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.Inspect(ctx, DatabaseOpen{
		ProjectRef: "project-id",
		Password:   password,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ProjectID != plan.Project.ID.String() {
		t.Fatalf("project id = %q, want %q", result.ProjectID, plan.Project.ID)
	}
	if result.DatabaseType != "project" || result.RootName != "source" || result.Files != 1 || result.Folders != 2 {
		t.Fatalf("inspect result = %+v", result)
	}

	verify, err := service.Verify(ctx, DatabaseOpen{
		ProjectRef: "project-id",
		Password:   password,
	}, encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if verify.Status != "ok" || verify.MissingObjects != 0 || verify.TamperedObjects != 0 || verify.ExtraObjects != 0 {
		t.Fatalf("verify result = %+v", verify)
	}
}

func TestServiceEnsureDataDirAndListActiveProjects(t *testing.T) {
	root := t.TempDir()
	service, err := NewService(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}

	if err := service.EnsureDataDir(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(service.ProjectsDir()); err != nil {
		t.Fatalf("projects dir stat error = %v", err)
	}

	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	projectPath := filepath.Join(service.ProjectsDir(), "alpha"+format.ProjectExtension)
	if err := os.WriteFile(projectPath, []byte("project"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(projectPath, now, now); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(service.ProjectsDir(), "ignore"+format.SetExtension), []byte("share"), 0o600); err != nil {
		t.Fatal(err)
	}

	projects, err := service.ListActiveProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Fatalf("project count = %d, want 1", len(projects))
	}
	if projects[0].ProjectID != "alpha" || projects[0].FileName != "alpha"+format.ProjectExtension || projects[0].Availability != "available" {
		t.Fatalf("project summary = %+v", projects[0])
	}
	if !projects[0].ModifiedAt.Equal(now) {
		t.Fatalf("modified at = %s, want %s", projects[0].ModifiedAt, now)
	}
}

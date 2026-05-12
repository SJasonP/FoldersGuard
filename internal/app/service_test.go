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

func TestValidateOutputOutsideSourceRejectsAncestorOutput(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}

	err := ValidateOutputOutsideSource(source, root)
	if err == nil {
		t.Fatal("expected source parent output to be rejected")
	}
	if !strings.Contains(err.Error(), "must not contain the source folder") {
		t.Fatalf("error = %v, want containment rejection", err)
	}
}

func TestValidateOutputOutsideSourceAllowsSiblingOutput(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	output := filepath.Join(root, "output")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := ValidateOutputOutsideSource(source, output); err != nil {
		t.Fatalf("sibling output rejected: %v", err)
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
	if projects[0].ProjectID != "alpha" || projects[0].ProjectName != "alpha" || projects[0].FileName != "alpha"+format.ProjectExtension || projects[0].Availability != "available" {
		t.Fatalf("project summary = %+v", projects[0])
	}
	if !projects[0].ModifiedAt.Equal(now) {
		t.Fatalf("modified at = %s, want %s", projects[0].ModifiedAt, now)
	}
}

func TestServiceSaveLocalProjectName(t *testing.T) {
	root := t.TempDir()
	service := Service{DataDir: filepath.Join(root, "data")}
	if err := service.EnsureDataDir(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(service.ProjectsDir(), "alpha"+format.ProjectExtension), []byte("project"), 0o600); err != nil {
		t.Fatal(err)
	}

	saved, err := service.SaveLocalProjectName(SaveLocalProjectNameInput{
		ProjectID:   "alpha",
		ProjectName: "  Personal Archive  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if saved.ProjectName != "Personal Archive" {
		t.Fatalf("project name = %q", saved.ProjectName)
	}

	projects, err := service.ListActiveProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].ProjectName != "Personal Archive" {
		t.Fatalf("projects = %+v", projects)
	}
}

func TestServiceExportAndDeleteProject(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	dataDir := filepath.Join(root, "data")
	databasePath := filepath.Join(dataDir, "projects", "project-id"+format.ProjectExtension)
	exportPath := filepath.Join(root, "exported"+format.ProjectExtension)
	password := "test-password"

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o600); err != nil {
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
	if err := WriteProjectDatabase(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	}, plan); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	exported, err := service.ExportProject(ctx, ExportProjectInput{
		ProjectID:  "project-id",
		Password:   password,
		OutputPath: exportPath,
		Force:      false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if exported.ProjectID != plan.Project.ID.String() || exported.OutputPath != exportPath {
		t.Fatalf("export result = %+v", exported)
	}
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("exported path stat error = %v", err)
	}

	deleted, err := service.DeleteProject(ctx, DeleteProjectInput{
		ProjectID: "project-id",
		Password:  password,
	})
	if err != nil {
		t.Fatal(err)
	}
	if deleted.ProjectID != "project-id" {
		t.Fatalf("delete result = %+v", deleted)
	}
	if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
		t.Fatalf("database stat error = %v, want not exist", err)
	}
}

func TestServiceCreateProject(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "test-password"

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
	if _, err := service.SaveSettings(Settings{
		DefaultMaxPartSize: 0,
		SourceCleanupMode:  SourceCleanupDelete,
		Theme:              ThemeSystem,
		Language:           LanguageSystem,
	}); err != nil {
		t.Fatal(err)
	}

	result, err := service.CreateProject(ctx, CreateProjectInput{
		SourcePath:    source,
		ContentOutput: encrypted,
		Password:      password,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ProjectID == "" || result.ProjectName != "source" {
		t.Fatalf("create result = %+v", result)
	}
	if result.EncryptedFiles != 1 || result.EncryptedFolders != 2 || result.DeletedCleartextFiles != 1 {
		t.Fatalf("create result counts = %+v", result)
	}

	activePath, err := service.ActiveProjectDatabasePath(result.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("active database stat error = %v", err)
	}

	var encryptedFiles int
	err = filepath.Walk(encrypted, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() {
			encryptedFiles++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if encryptedFiles != 1 {
		t.Fatalf("encrypted files = %d, want 1", encryptedFiles)
	}

	if _, err := os.Stat(filepath.Join(source, "docs", "note.txt")); !os.IsNotExist(err) {
		t.Fatalf("source file stat error = %v, want not exist", err)
	}
}

func TestServiceCreateProjectRejectsAncestorOutputWithoutDeletingSource(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	dataDir := filepath.Join(root, "data")
	password := "test-password"

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
	_, err = service.CreateProject(ctx, CreateProjectInput{
		SourcePath:    source,
		ContentOutput: root,
		Password:      password,
		Force:         true,
		SourceCleanup: SourceCleanupKeep,
	})
	if err == nil {
		t.Fatal("expected ancestor output to be rejected")
	}
	if _, statErr := os.Stat(filepath.Join(source, "note.txt")); statErr != nil {
		t.Fatalf("source was modified after rejected create: %v", statErr)
	}
}

func TestServiceCreateProjectKeepPreservesEmptySourceFolders(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	encrypted := filepath.Join(root, "encrypted")
	dataDir := filepath.Join(root, "data")
	password := "test-password"
	emptyFolder := filepath.Join(source, "docs", "empty")

	if err := os.MkdirAll(emptyFolder, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.CreateProject(ctx, CreateProjectInput{
		SourcePath:    source,
		ContentOutput: encrypted,
		Password:      password,
		SourceCleanup: SourceCleanupKeep,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.DeletedCleartextFiles != 0 || result.DeletedCleartextFolders != 0 {
		t.Fatalf("create result = %+v, want no source cleanup", result)
	}
	if _, err := os.Stat(emptyFolder); err != nil {
		t.Fatalf("empty source folder was not preserved: %v", err)
	}
}

func TestServiceImportProject(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	importPath := filepath.Join(root, "incoming"+format.ProjectExtension)
	dataDir := filepath.Join(root, "data")
	password := "test-password"

	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o600); err != nil {
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
	if err := WriteProjectDatabase(ctx, db.Config{
		Path:       importPath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	}, plan); err != nil {
		t.Fatal(err)
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	imported, err := service.ImportProject(ctx, ImportProjectInput{
		InputPath: importPath,
		Password:  password,
		Force:     false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if imported.ProjectID != plan.Project.ID.String() {
		t.Fatalf("import result = %+v", imported)
	}
	activePath, err := service.ActiveProjectDatabasePath(plan.Project.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("active database stat error = %v", err)
	}
}

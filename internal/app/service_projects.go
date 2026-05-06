package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
	"foldersguard/internal/project"
)

func (s Service) ListActiveProjects() ([]ActiveProjectSummary, error) {
	if err := s.EnsureDataDir(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.ProjectsDir())
	if err != nil {
		return nil, fmt.Errorf("read projects directory: %w", err)
	}

	var projects []ActiveProjectSummary
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != format.ProjectExtension {
			continue
		}

		project := ActiveProjectSummary{
			ProjectID:    entry.Name()[:len(entry.Name())-len(format.ProjectExtension)],
			FileName:     entry.Name(),
			Availability: "available",
		}

		info, err := entry.Info()
		if err != nil || info.IsDir() {
			project.Availability = "unavailable"
			projects = append(projects, project)
			continue
		}

		project.ModifiedAt = info.ModTime().UTC()
		projects = append(projects, project)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].ProjectID < projects[j].ProjectID
	})
	return projects, nil
}

func (s Service) Inspect(ctx context.Context, input DatabaseOpen) (InspectResult, error) {
	plan, meta, err := s.ReadDatabase(ctx, input)
	if err != nil {
		return InspectResult{}, err
	}
	return inspectResult(plan, meta), nil
}

func (s Service) Verify(ctx context.Context, input DatabaseOpen, encryptedRoot string) (VerifyResult, error) {
	if err := ValidateExistingDirectory(encryptedRoot, "content"); err != nil {
		return VerifyResult{}, err
	}
	plan, _, err := s.ReadDatabase(ctx, input)
	if err != nil {
		return VerifyResult{}, err
	}
	report, err := (project.Verifier{EncryptedRoot: encryptedRoot}).VerifyContent(ctx, plan)
	if err != nil {
		return VerifyResult{}, err
	}
	status := "ok"
	if !report.OK() {
		status = "failed"
	}
	return VerifyResult{
		ProjectID:       plan.Project.ID.String(),
		CheckedObjects:  report.CheckedObjects,
		MissingObjects:  report.MissingObjects,
		TamperedObjects: report.TamperedObjects,
		ExtraObjects:    report.ExtraObjects,
		Status:          status,
	}, nil
}

func (s Service) InspectShare(ctx context.Context, input ShareOpen) (ShareSummary, error) {
	plan, meta, passwordProtected, err := s.ReadShareDatabase(ctx, input)
	if err != nil {
		return ShareSummary{}, err
	}
	if meta["database_type"] != "share" {
		return ShareSummary{}, fmt.Errorf("database type = %q, want share", meta["database_type"])
	}
	return shareSummary(plan, meta, passwordProtected), nil
}

func (s Service) VerifyShare(ctx context.Context, input ShareOpen, encryptedRoot string) (VerifyResult, error) {
	if err := ValidateExistingDirectory(encryptedRoot, "content"); err != nil {
		return VerifyResult{}, err
	}
	plan, meta, _, err := s.ReadShareDatabase(ctx, input)
	if err != nil {
		return VerifyResult{}, err
	}
	if meta["database_type"] != "share" {
		return VerifyResult{}, fmt.Errorf("database type = %q, want share", meta["database_type"])
	}
	report, err := (project.Verifier{EncryptedRoot: encryptedRoot}).VerifyContent(ctx, plan)
	if err != nil {
		return VerifyResult{}, err
	}
	status := "ok"
	if !report.OK() {
		status = "failed"
	}
	return VerifyResult{
		ProjectID:       plan.Project.ID.String(),
		CheckedObjects:  report.CheckedObjects,
		MissingObjects:  report.MissingObjects,
		TamperedObjects: report.TamperedObjects,
		ExtraObjects:    report.ExtraObjects,
		Status:          status,
	}, nil
}

func (s Service) ExportProject(ctx context.Context, input ExportProjectInput) (ExportProjectResult, error) {
	if !format.IsProjectExtension(input.OutputPath) {
		return ExportProjectResult{}, fmt.Errorf("database output must use %s extension", format.ProjectExtension)
	}
	sourcePath, err := s.ActiveProjectDatabasePath(input.ProjectID)
	if err != nil {
		return ExportProjectResult{}, err
	}
	plan, meta, err := ReadDatabase(ctx, db.Config{
		Path:       sourcePath,
		DriverName: db.SQLCipherDriver,
		Password:   input.Password,
	})
	if err != nil {
		return ExportProjectResult{}, err
	}
	if meta["database_type"] != "project" {
		return ExportProjectResult{}, fmt.Errorf("database type = %q, want project", meta["database_type"])
	}
	if err := ValidateDistinctPaths(sourcePath, input.OutputPath); err != nil {
		return ExportProjectResult{}, err
	}
	if err := PrepareFileOutput(input.OutputPath, input.Force); err != nil {
		return ExportProjectResult{}, err
	}
	if err := CopyFile(sourcePath, input.OutputPath); err != nil {
		return ExportProjectResult{}, err
	}
	return ExportProjectResult{
		ProjectID:  plan.Project.ID.String(),
		OutputPath: input.OutputPath,
	}, nil
}

func (s Service) DeleteProject(ctx context.Context, input DeleteProjectInput) (DeleteProjectResult, error) {
	sourcePath, err := s.ActiveProjectDatabasePath(input.ProjectID)
	if err != nil {
		return DeleteProjectResult{}, err
	}
	_, meta, err := ReadDatabase(ctx, db.Config{
		Path:       sourcePath,
		DriverName: db.SQLCipherDriver,
		Password:   input.Password,
	})
	if err != nil {
		return DeleteProjectResult{}, err
	}
	if meta["database_type"] != "project" {
		return DeleteProjectResult{}, fmt.Errorf("database type = %q, want project", meta["database_type"])
	}
	if err := os.Remove(sourcePath); err != nil {
		return DeleteProjectResult{}, fmt.Errorf("delete active project database: %w", err)
	}
	return DeleteProjectResult{ProjectID: input.ProjectID}, nil
}

func (s Service) ImportProject(ctx context.Context, input ImportProjectInput) (ImportProjectResult, error) {
	if !format.IsProjectExtension(input.InputPath) {
		return ImportProjectResult{}, fmt.Errorf("input must use %s extension", format.ProjectExtension)
	}

	plan, meta, err := ReadDatabase(ctx, db.Config{
		Path:       input.InputPath,
		DriverName: db.SQLCipherDriver,
		Password:   input.Password,
	})
	if err != nil {
		return ImportProjectResult{}, err
	}
	if meta["database_type"] != "project" {
		return ImportProjectResult{}, fmt.Errorf("database type = %q, want project", meta["database_type"])
	}

	activePath, err := s.ActiveProjectDatabasePath(plan.Project.ID.String())
	if err != nil {
		return ImportProjectResult{}, err
	}
	if err := ValidateDistinctPaths(input.InputPath, activePath); err != nil {
		return ImportProjectResult{}, err
	}
	if err := PrepareFileOutput(activePath, input.Force); err != nil {
		return ImportProjectResult{}, err
	}
	if err := CopyFile(input.InputPath, activePath); err != nil {
		return ImportProjectResult{}, err
	}
	return ImportProjectResult{
		ProjectID: plan.Project.ID.String(),
	}, nil
}

func (s Service) CreateProject(ctx context.Context, input CreateProjectInput) (CreateProjectResult, error) {
	if input.Password == "" {
		return CreateProjectResult{}, fmt.Errorf("password is required")
	}
	if err := ValidateExistingDirectory(input.SourcePath, "source"); err != nil {
		return CreateProjectResult{}, err
	}
	if err := ValidateOutputOutsideSource(input.SourcePath, input.ContentOutput); err != nil {
		return CreateProjectResult{}, err
	}
	if err := ValidateDistinctPaths(input.SourcePath, input.ContentOutput); err != nil {
		return CreateProjectResult{}, err
	}
	if err := PrepareContentOutput(input.ContentOutput, input.Force); err != nil {
		return CreateProjectResult{}, err
	}
	if input.DatabaseExport != "" {
		if err := ValidateOutputOutsideSource(input.SourcePath, input.DatabaseExport); err != nil {
			return CreateProjectResult{}, err
		}
		if !format.IsProjectExtension(input.DatabaseExport) {
			return CreateProjectResult{}, fmt.Errorf("database export must use %s extension", format.ProjectExtension)
		}
		if err := PrepareFileOutput(input.DatabaseExport, input.Force); err != nil {
			return CreateProjectResult{}, err
		}
	}

	maxPartSize, err := s.resolveMaxPartSize(input.MaxPartSize)
	if err != nil {
		return CreateProjectResult{}, err
	}
	sourceCleanup, err := s.resolveSourceCleanupMode(input.SourceCleanup)
	if err != nil {
		return CreateProjectResult{}, err
	}

	scan, err := fswalk.ScanTopFolder(input.SourcePath)
	if err != nil {
		return CreateProjectResult{}, err
	}
	plan, err := project.Planner{MaxPartSize: maxPartSize}.Plan(scan)
	if err != nil {
		return CreateProjectResult{}, err
	}

	activeDatabase, err := s.ActiveProjectDatabasePath(plan.Project.ID.String())
	if err != nil {
		return CreateProjectResult{}, err
	}
	if err := ValidateOutputOutsideSource(input.SourcePath, activeDatabase); err != nil {
		return CreateProjectResult{}, err
	}
	if err := PrepareFileOutput(activeDatabase, input.Force); err != nil {
		return CreateProjectResult{}, err
	}

	if err := WriteProjectDatabase(ctx, db.Config{
		Path:       activeDatabase,
		DriverName: db.SQLCipherDriver,
		Password:   input.Password,
	}, plan); err != nil {
		return CreateProjectResult{}, err
	}
	if input.DatabaseExport != "" {
		if err := WriteProjectDatabase(ctx, db.Config{
			Path:       input.DatabaseExport,
			DriverName: db.SQLCipherDriver,
			Password:   input.Password,
		}, plan); err != nil {
			return CreateProjectResult{}, err
		}
	}

	var deletedFiles int
	afterFile := func(file model.File) error {
		if sourceCleanup != SourceCleanupDelete {
			return nil
		}
		if err := os.Remove(file.SourcePath); err != nil {
			return fmt.Errorf("delete source file: %w", err)
		}
		deletedFiles++
		return nil
	}

	if err := (project.Executor{
		OutputRoot: input.ContentOutput,
		AfterFile:  afterFile,
	}).EncryptContent(ctx, plan); err != nil {
		return CreateProjectResult{}, err
	}

	deletedFolders, err := removeEmptyFoldersUnderRoot(input.SourcePath)
	if err != nil {
		return CreateProjectResult{}, err
	}

	return CreateProjectResult{
		ProjectID:               plan.Project.ID.String(),
		ProjectName:             plan.RootItem.RealName,
		ContentOutput:           input.ContentOutput,
		DatabaseExport:          input.DatabaseExport,
		EncryptedFiles:          len(plan.Files),
		EncryptedFolders:        CountFolders(plan),
		EncryptedParts:          len(plan.Parts),
		DeletedCleartextFiles:   deletedFiles,
		DeletedCleartextFolders: deletedFolders,
		FailedFiles:             0,
	}, nil
}

func (s Service) resolveMaxPartSize(override int64) (int64, error) {
	if override > 0 {
		return override, nil
	}

	settings, err := s.ReadSettings()
	if err != nil {
		return 0, err
	}
	if settings.DefaultMaxPartSize > 0 {
		return settings.DefaultMaxPartSize, nil
	}
	return (1 << 62), nil
}

func (s Service) resolveSourceCleanupMode(requested string) (string, error) {
	if requested == "" {
		settings, err := s.ReadSettings()
		if err != nil {
			return "", err
		}
		requested = settings.SourceCleanupMode
	}

	switch requested {
	case "", SourceCleanupAsk:
		return SourceCleanupKeep, nil
	case SourceCleanupKeep, SourceCleanupDelete:
		return requested, nil
	default:
		return "", fmt.Errorf("unsupported source cleanup mode %q", requested)
	}
}

func removeEmptyFoldersUnderRoot(root string) (int, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return 0, fmt.Errorf("resolve source root: %w", err)
	}

	var directories []string
	err = filepath.WalkDir(rootAbs, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		directories = append(directories, path)
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("scan source directories: %w", err)
	}

	var removed int
	sort.Slice(directories, func(i, j int) bool {
		return len(directories[i]) > len(directories[j])
	})

	for _, dir := range directories {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return removed, fmt.Errorf("read source directory: %w", err)
		}
		if len(entries) != 0 {
			continue
		}
		if err := os.Remove(dir); err != nil {
			return removed, fmt.Errorf("delete empty source directory: %w", err)
		}
		removed++
	}
	return removed, nil
}

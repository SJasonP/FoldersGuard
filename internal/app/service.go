package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/model"
	"foldersguard/internal/project"
	"foldersguard/internal/storage"
)

type Service struct {
	DataDir string
}

type DatabaseOpen struct {
	ProjectRef string
	Password   string
}

type InspectResult struct {
	ProjectID      string
	DatabaseType   string
	RootFolderID   string
	RootName       string
	FormatVersion  string
	SchemaVersion  string
	Items          int
	Folders        int
	Files          int
	Parts          int
	StorageObjects int
}

type VerifyResult struct {
	ProjectID       string
	CheckedObjects  int
	MissingObjects  int
	TamperedObjects int
	ExtraObjects    int
	Status          string
}

type ActiveProjectSummary struct {
	ProjectID    string
	FileName     string
	ModifiedAt   time.Time
	Availability string
}

func NewService(dataDir string) (Service, error) {
	if dataDir == "" {
		resolved, err := DefaultDataDir()
		if err != nil {
			return Service{}, err
		}
		dataDir = resolved
	}
	return Service{DataDir: dataDir}, nil
}

func DefaultDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(configDir, format.DataDirName), nil
}

func (s Service) ProjectsDir() string {
	return filepath.Join(s.DataDir, "projects")
}

func (s Service) EnsureDataDir() error {
	if err := os.MkdirAll(s.ProjectsDir(), 0o755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	return nil
}

func (s Service) ActiveProjectDatabasePath(projectID string) (string, error) {
	if projectID == "" {
		return "", fmt.Errorf("project id is required")
	}
	if format.IsProjectExtension(projectID) || format.IsSetExtension(projectID) {
		return "", fmt.Errorf("project id must reference an active project, not a database path")
	}
	return filepath.Join(s.ProjectsDir(), projectID+format.ProjectExtension), nil
}

func (s Service) DatabasePathFromProjectRef(projectRef string) (string, error) {
	if projectRef == "" {
		return "", fmt.Errorf("project reference is required")
	}
	if format.IsProjectExtension(projectRef) {
		return "", fmt.Errorf("exported %s databases must be imported before use", format.ProjectExtension)
	}
	if format.IsSetExtension(projectRef) {
		return projectRef, nil
	}
	return s.ActiveProjectDatabasePath(projectRef)
}

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

func (s Service) ReadDatabase(ctx context.Context, input DatabaseOpen) (model.PlannedProject, map[string]string, error) {
	path, err := s.DatabasePathFromProjectRef(input.ProjectRef)
	if err != nil {
		return model.PlannedProject{}, nil, err
	}
	return ReadDatabase(ctx, db.Config{
		Path:       path,
		DriverName: db.SQLCipherDriver,
		Password:   input.Password,
	})
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

func WriteProjectDatabase(ctx context.Context, config db.Config, plan model.PlannedProject) error {
	return WriteDatabase(ctx, config, plan, "project")
}

func WriteShareDatabase(ctx context.Context, config db.Config, plan model.PlannedProject) error {
	return WriteDatabase(ctx, config, plan, "share")
}

func WriteDatabase(ctx context.Context, config db.Config, plan model.PlannedProject, databaseType string) error {
	database, err := db.OpenProject(ctx, config)
	if err != nil {
		return err
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		return err
	}
	if err := store.InitProject(ctx, storage.ProjectSpec{
		ProjectID:       plan.Project.ID,
		RootFolderID:    plan.Project.RootFolderID,
		RootVisibleName: plan.RootItem.VisibleName,
		RootRealName:    plan.RootItem.RealName,
		RootFolderKey:   plan.RootFolder.Key,
		DatabaseType:    databaseType,
		CreatedAt:       plan.Project.CreatedAt,
	}); err != nil {
		return err
	}
	if err := store.WritePlannedProject(ctx, plan); err != nil {
		return err
	}
	return nil
}

func ReadDatabase(ctx context.Context, config db.Config) (model.PlannedProject, map[string]string, error) {
	if err := ValidateDatabasePath(config.Path); err != nil {
		return model.PlannedProject{}, nil, err
	}

	database, err := db.OpenProject(ctx, config)
	if err != nil {
		return model.PlannedProject{}, nil, err
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		return model.PlannedProject{}, nil, err
	}
	meta, err := store.Meta(ctx)
	if err != nil {
		return model.PlannedProject{}, nil, err
	}
	plan, err := store.ReadPlannedProject(ctx)
	if err != nil {
		return model.PlannedProject{}, nil, err
	}
	return plan, meta, nil
}

func CopyFile(source, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open source database: %w", err)
	}
	defer input.Close()

	output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create target database: %w", err)
	}
	committed := false
	defer func() {
		_ = output.Close()
		if !committed {
			_ = os.Remove(target)
		}
	}()

	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("copy database: %w", err)
	}
	if err := output.Sync(); err != nil {
		return fmt.Errorf("sync target database: %w", err)
	}
	if err := output.Close(); err != nil {
		return fmt.Errorf("close target database: %w", err)
	}
	committed = true
	return nil
}

func inspectResult(plan model.PlannedProject, meta map[string]string) InspectResult {
	return InspectResult{
		ProjectID:      plan.Project.ID.String(),
		DatabaseType:   meta["database_type"],
		RootFolderID:   plan.Project.RootFolderID.String(),
		RootName:       plan.RootItem.RealName,
		FormatVersion:  meta["format_version"],
		SchemaVersion:  meta["schema_version"],
		Items:          len(plan.Items) + 1,
		Folders:        CountFolders(plan),
		Files:          len(plan.Files),
		Parts:          len(plan.Parts),
		StorageObjects: len(plan.StorageObjects),
	}
}

func CountFolders(plan model.PlannedProject) int {
	if plan.Project.DatabaseType == "share" && plan.RootItem.RealName == "" {
		return len(plan.Folders)
	}
	return len(plan.Folders) + 1
}

package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/model"
)

type Service struct {
	DataDir string
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

func (s Service) OperationGuidesDir() string {
	return filepath.Join(s.DataDir, "operation-guides")
}

func (s Service) StagedContentDir() string {
	return filepath.Join(s.DataDir, "staged-content")
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

func (s Service) ReadShareDatabase(ctx context.Context, input ShareOpen) (model.PlannedProject, map[string]string, bool, error) {
	if !format.IsSetExtension(input.DatabasePath) {
		return model.PlannedProject{}, nil, false, fmt.Errorf("share database must use %s extension", format.SetExtension)
	}
	if err := ValidateDatabasePath(input.DatabasePath); err != nil {
		return model.PlannedProject{}, nil, false, err
	}

	if input.Password == "" {
		plan, meta, err := ReadDatabase(ctx, db.Config{
			Path:       input.DatabasePath,
			DriverName: db.SQLCipherDriver,
			Password:   db.UnprotectedSharePassword,
		})
		if err == nil {
			return plan, meta, false, nil
		}
	}

	if input.Password == "" {
		return model.PlannedProject{}, nil, false, fmt.Errorf("share password is required")
	}

	plan, meta, err := ReadDatabase(ctx, db.Config{
		Path:       input.DatabasePath,
		DriverName: db.SQLCipherDriver,
		Password:   input.Password,
	})
	if err != nil {
		return model.PlannedProject{}, nil, false, err
	}
	return plan, meta, true, nil
}

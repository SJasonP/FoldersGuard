package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/model"
	"foldersguard/internal/storage"
)

func activeProjectDatabasePath(projectID string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(configDir, format.AppID, "projects", projectID+format.ProjectExtension), nil
}

func databasePathFromProjectRef(projectRef string) (string, error) {
	if projectRef == "" {
		return "", fmt.Errorf("project reference is required")
	}
	if format.IsProjectExtension(projectRef) || format.IsSetExtension(projectRef) {
		return projectRef, nil
	}
	return activeProjectDatabasePath(projectRef)
}

func projectDatabasePathFromProjectRef(projectRef string) (string, error) {
	if format.IsSetExtension(projectRef) {
		return "", fmt.Errorf("project editing commands do not accept %s databases", format.SetExtension)
	}
	return databasePathFromProjectRef(projectRef)
}

func writeProjectDatabase(ctx context.Context, config db.Config, plan model.PlannedProject) error {
	return writeDatabase(ctx, config, plan, "project")
}

func writeShareDatabase(ctx context.Context, config db.Config, plan model.PlannedProject) error {
	return writeDatabase(ctx, config, plan, "share")
}

func writeDatabase(ctx context.Context, config db.Config, plan model.PlannedProject, databaseType string) error {
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

func readProjectDatabase(ctx context.Context, config db.Config) (model.PlannedProject, error) {
	plan, _, err := readProjectDatabaseWithMeta(ctx, config)
	return plan, err
}

func (c cli) readDatabaseFromProjectRef(ctx context.Context, projectRef string, options passwordOptions) (model.PlannedProject, error) {
	plan, _, err := c.readDatabaseWithMetaFromProjectRef(ctx, projectRef, options)
	return plan, err
}

func (c cli) readDatabaseWithMetaFromProjectRef(ctx context.Context, projectRef string, options passwordOptions) (model.PlannedProject, map[string]string, error) {
	databasePath, err := databasePathFromProjectRef(projectRef)
	if err != nil {
		return model.PlannedProject{}, nil, err
	}
	if err := validateDatabasePath(databasePath); err != nil {
		return model.PlannedProject{}, nil, err
	}
	if format.IsSetExtension(projectRef) && !hasPasswordInput(options) {
		plan, meta, err := readProjectDatabaseWithMeta(ctx, db.Config{
			Path:       databasePath,
			DriverName: db.SQLCipherDriver,
			Password:   db.UnprotectedSharePassword,
		})
		if err == nil {
			return plan, meta, nil
		}
	}

	password, err := c.readDatabasePassword(projectRef, options)
	if err != nil {
		return model.PlannedProject{}, nil, err
	}
	plan, meta, err := readProjectDatabaseWithMeta(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return model.PlannedProject{}, nil, err
	}
	return plan, meta, nil
}

func validateDatabasePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat database: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("database path is a directory")
	}
	return nil
}

func readProjectDatabaseWithMeta(ctx context.Context, config db.Config) (model.PlannedProject, map[string]string, error) {
	if err := validateDatabasePath(config.Path); err != nil {
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

func copyFile(source, target string) error {
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

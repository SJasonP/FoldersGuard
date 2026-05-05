package cli

import (
	"context"
	"fmt"
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

func writeProjectDatabase(ctx context.Context, config db.Config, plan model.PlannedProject) error {
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
		DatabaseType:    "project",
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
	database, err := db.OpenProject(ctx, config)
	if err != nil {
		return model.PlannedProject{}, err
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		return model.PlannedProject{}, err
	}
	return store.ReadPlannedProject(ctx)
}

package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"foldersguard/internal/db"
	"foldersguard/internal/model"
	"foldersguard/internal/storage"
)

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
		ProjectName:    plan.RootItem.RealName,
		RootFolderID:   plan.Project.RootFolderID.String(),
		RootName:       plan.RootItem.RealName,
		FormatVersion:  meta["format_version"],
		SchemaVersion:  meta["schema_version"],
		CreatedAt:      plan.Project.CreatedAt.UTC(),
		UpdatedAt:      plan.Project.UpdatedAt.UTC(),
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

func shareSummary(plan model.PlannedProject, meta map[string]string, passwordProtected bool) ShareSummary {
	return ShareSummary{
		ShareID:           plan.Project.ID.String(),
		DatabaseType:      meta["database_type"],
		FormatVersion:     meta["format_version"],
		SchemaVersion:     meta["schema_version"],
		TopLevelItems:     countTopLevelItems(plan),
		Files:             len(plan.Files),
		Folders:           CountFolders(plan),
		Parts:             len(plan.Parts),
		StorageObjects:    len(plan.StorageObjects),
		PasswordProtected: passwordProtected,
	}
}

func countTopLevelItems(plan model.PlannedProject) int {
	count := 0
	for _, item := range plan.Items {
		if item.ParentID == nil {
			continue
		}
		if item.ParentID.String() == plan.RootItem.ID.String() {
			count++
		}
	}
	return count
}

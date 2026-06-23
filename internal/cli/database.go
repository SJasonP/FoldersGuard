package cli

import (
	"context"

	"foldersguard/internal/app"
	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/model"
)

func activeProjectDatabasePath(projectID string) (string, error) {
	service, err := app.NewService("")
	if err != nil {
		return "", err
	}
	return service.ActiveProjectDatabasePath(projectID)
}

func databasePathFromProjectRef(projectRef string) (string, error) {
	service, err := app.NewService("")
	if err != nil {
		return "", err
	}
	return service.DatabasePathFromProjectRef(projectRef)
}

func activeProjectDatabasePathFromID(projectID string) (string, error) {
	service, err := app.NewService("")
	if err != nil {
		return "", err
	}
	return service.ActiveProjectDatabasePath(projectID)
}

func readNoiseFileHandling() (string, error) {
	service, err := app.NewService("")
	if err != nil {
		return "", err
	}
	settings, err := service.ReadSettings()
	if err != nil {
		return "", err
	}
	return settings.NoiseFileHandling, nil
}

func writeProjectDatabase(ctx context.Context, config db.Config, plan model.PlannedProject) error {
	return app.WriteProjectDatabase(ctx, config, plan)
}

func writeShareDatabase(ctx context.Context, config db.Config, plan model.PlannedProject) error {
	return app.WriteShareDatabase(ctx, config, plan)
}

func writeDatabase(ctx context.Context, config db.Config, plan model.PlannedProject, databaseType string) error {
	return app.WriteDatabase(ctx, config, plan, databaseType)
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
	return app.ValidateDatabasePath(path)
}

func readProjectDatabaseWithMeta(ctx context.Context, config db.Config) (model.PlannedProject, map[string]string, error) {
	return app.ReadDatabase(ctx, config)
}

func copyFile(source, target string) error {
	return app.CopyFile(source, target)
}

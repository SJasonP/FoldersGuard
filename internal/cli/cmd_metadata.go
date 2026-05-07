package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"foldersguard/internal/app"
	"foldersguard/internal/db"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/project"
	"foldersguard/internal/storage"
)

type renameOptions struct {
	projectRef      string
	itemPath        string
	newName         string
	passwordOptions passwordOptions
}

type removeOptions struct {
	projectRef      string
	itemPath        string
	contentRoot     string
	passwordOptions passwordOptions
	force           bool
}

type moveOptions struct {
	projectRef       string
	itemPath         string
	targetFolderPath string
	contentRoot      string
	passwordOptions  passwordOptions
}

type addOptions struct {
	projectRef       string
	sourcePath       string
	targetFolderPath string
	stagingContent   string
	contentRoot      string
	maxPartSize      int64
	passwordOptions  passwordOptions
	force            bool
}

func (c cli) renameCommand() *cobra.Command {
	options := renameOptions{}
	command := &cobra.Command{
		Use:           "rename <project-id> <item-path> <new-name>",
		Short:         "Rename a file or folder in FG metadata.",
		Example:       c.name + " rename <project-id> Root/old.txt new.txt --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectRef = args[0]
			options.itemPath = args[1]
			options.newName = args[2]
			return c.runRename(options)
		},
	}
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) removeCommand() *cobra.Command {
	options := removeOptions{}
	command := &cobra.Command{
		Use:           "remove <project-id> <item-path>",
		Short:         "Remove a file or folder from FG metadata.",
		Example:       c.name + " remove <project-id> Root/old.txt --force --content ./encrypted --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectRef = args[0]
			options.itemPath = args[1]
			return c.runRemove(options)
		},
	}
	command.Flags().StringVar(&options.contentRoot, "content", "", "encrypted content folder")
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.Flags().BoolVar(&options.force, "force", false, "accept metadata and content deletion")
	mustMarkRequired(command, "force")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) moveCommand() *cobra.Command {
	options := moveOptions{}
	command := &cobra.Command{
		Use:           "move <project-id> <item-path> <target-folder-path>",
		Short:         "Move a file or folder in FG metadata.",
		Example:       c.name + " move <project-id> Root/old.txt Root/docs --content ./encrypted --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectRef = args[0]
			options.itemPath = args[1]
			options.targetFolderPath = args[2]
			return c.runMove(options)
		},
	}
	command.Flags().StringVar(&options.contentRoot, "content", "", "encrypted content folder")
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) addCommand() *cobra.Command {
	options := addOptions{}
	command := &cobra.Command{
		Use:           "add <project-id> <source-path> <target-folder-path>",
		Short:         "Add cleartext content to an existing FG project.",
		Example:       c.name + " add <project-id> ./new Root/docs --staging-content ./staging --max-part-size 1073741824 --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectRef = args[0]
			options.sourcePath = args[1]
			options.targetFolderPath = args[2]
			if options.maxPartSize <= 0 {
				return fmt.Errorf("max part size must be positive")
			}
			return c.runAdd(options)
		},
	}
	command.Flags().StringVar(&options.stagingContent, "staging-content", "", "staged encrypted content folder")
	command.Flags().StringVar(&options.contentRoot, "content", "", "encrypted content folder")
	command.Flags().Int64Var(&options.maxPartSize, "max-part-size", 0, "maximum part size in bytes")
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.Flags().BoolVar(&options.force, "force", false, "replace existing staging content output")
	mustMarkRequired(command, "staging-content")
	mustMarkRequired(command, "max-part-size")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) runRename(options renameOptions) error {
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}
	databasePath, err := activeProjectDatabasePathFromID(options.projectRef)
	if err != nil {
		return err
	}
	if err := validateExistingFile(databasePath, "database"); err != nil {
		return err
	}

	ctx := context.Background()
	database, err := db.OpenProject(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return err
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		return err
	}
	result, err := store.RenameItem(ctx, options.itemPath, options.newName, time.Now())
	if err != nil {
		return err
	}

	fmt.Fprintf(c.out, "project_id=%s\n", result.ProjectID)
	fmt.Fprintf(c.out, "item_id=%s\n", result.ItemID)
	fmt.Fprintf(c.out, "old_name=%s\n", result.OldName)
	fmt.Fprintf(c.out, "new_name=%s\n", result.NewName)
	fmt.Fprintln(c.out, "content_operations=0")
	return nil
}

func (c cli) runRemove(options removeOptions) error {
	if !options.force {
		return fmt.Errorf("remove requires --force")
	}
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}
	databasePath, err := activeProjectDatabasePathFromID(options.projectRef)
	if err != nil {
		return err
	}
	if err := validateExistingFile(databasePath, "database"); err != nil {
		return err
	}
	if options.contentRoot != "" {
		if err := validateExistingDirectory(options.contentRoot, "encrypted content"); err != nil {
			return err
		}
	}

	ctx := context.Background()
	database, err := db.OpenProject(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return err
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		return err
	}
	result, err := store.RemoveItem(ctx, options.itemPath, time.Now())
	if err != nil {
		return err
	}
	if options.contentRoot != "" {
		if _, err := app.ApplyStorageContentOperations(result.Operations, app.ContentOperationApplyOptions{
			ContentRoot: options.contentRoot,
		}); err != nil {
			return err
		}
	}

	fmt.Fprintf(c.out, "project_id=%s\n", result.ProjectID)
	fmt.Fprintf(c.out, "operation_plan_id=%s\n", result.OperationPlanID)
	fmt.Fprintf(c.out, "operations=%d\n", len(result.Operations))
	for _, operation := range result.Operations {
		fmt.Fprintf(c.out, "operation=%s target=%s\n", operation.Type, operation.TargetPath)
	}
	return nil
}

func (c cli) runMove(options moveOptions) error {
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}
	databasePath, err := activeProjectDatabasePathFromID(options.projectRef)
	if err != nil {
		return err
	}
	if err := validateExistingFile(databasePath, "database"); err != nil {
		return err
	}
	if options.contentRoot != "" {
		if err := validateExistingDirectory(options.contentRoot, "encrypted content"); err != nil {
			return err
		}
	}

	ctx := context.Background()
	database, err := db.OpenProject(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return err
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		return err
	}
	result, err := store.MoveItem(ctx, options.itemPath, options.targetFolderPath, time.Now())
	if err != nil {
		return err
	}
	if options.contentRoot != "" {
		if _, err := app.ApplyStorageContentOperations(result.Operations, app.ContentOperationApplyOptions{
			ContentRoot: options.contentRoot,
		}); err != nil {
			return err
		}
	}

	fmt.Fprintf(c.out, "project_id=%s\n", result.ProjectID)
	fmt.Fprintf(c.out, "operation_plan_id=%s\n", result.OperationPlanID)
	fmt.Fprintf(c.out, "operations=%d\n", len(result.Operations))
	for _, operation := range result.Operations {
		fmt.Fprintf(c.out, "operation=%s source=%s target=%s\n", operation.Type, operation.SourcePath, operation.TargetPath)
	}
	return nil
}

func (c cli) runAdd(options addOptions) error {
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}
	databasePath, err := activeProjectDatabasePathFromID(options.projectRef)
	if err != nil {
		return err
	}
	if err := validateExistingFile(databasePath, "database"); err != nil {
		return err
	}
	if err := validateOutputOutsideSource(options.sourcePath, options.stagingContent); err != nil {
		return err
	}
	if options.contentRoot != "" {
		if err := validateExistingDirectory(options.contentRoot, "encrypted content"); err != nil {
			return err
		}
		if err := validateDistinctPaths(options.stagingContent, options.contentRoot); err != nil {
			return err
		}
	}
	if err := prepareDirectoryOutput(options.stagingContent, options.force, "staging content"); err != nil {
		return err
	}

	ctx := context.Background()
	database, err := db.OpenProject(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return err
	}
	defer database.Close()

	scan, err := fswalk.ScanPath(options.sourcePath)
	if err != nil {
		return err
	}
	addition, err := project.AddPlanner{MaxPartSize: options.maxPartSize}.Plan(scan)
	if err != nil {
		return err
	}
	store, err := storage.NewStore(database)
	if err != nil {
		return err
	}
	addition, operations, err := store.PrepareAdd(ctx, options.targetFolderPath, addition)
	if err != nil {
		return err
	}
	if err := (project.Executor{OutputRoot: options.stagingContent}).EncryptContent(ctx, addition); err != nil {
		return err
	}
	result, err := store.CommitAdd(ctx, options.targetFolderPath, addition, operations, time.Now())
	if err != nil {
		return err
	}
	if options.contentRoot != "" {
		if _, err := app.ApplyStorageContentOperations(result.Operations, app.ContentOperationApplyOptions{
			ContentRoot: options.contentRoot,
			StagingRoot: options.stagingContent,
		}); err != nil {
			return err
		}
	}

	fmt.Fprintf(c.out, "project_id=%s\n", result.ProjectID)
	fmt.Fprintf(c.out, "operation_plan_id=%s\n", result.OperationPlanID)
	fmt.Fprintf(c.out, "staging_content=%s\n", options.stagingContent)
	fmt.Fprintf(c.out, "operations=%d\n", len(result.Operations))
	for _, operation := range result.Operations {
		fmt.Fprintf(c.out, "operation=%s source=%s target=%s\n", operation.Type, operation.SourcePath, operation.TargetPath)
	}
	return nil
}

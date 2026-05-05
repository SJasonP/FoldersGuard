package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"foldersguard/internal/content"
	"foldersguard/internal/db"
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

func (c cli) renameCommand() *cobra.Command {
	options := renameOptions{}
	command := &cobra.Command{
		Use:           "rename <project-ref> <item-path> <new-name>",
		Short:         "Rename a file or folder in FG metadata.",
		Example:       c.name + " rename ./project.fg Root/old.txt new.txt --password-env FG_PASSWORD",
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
		Use:           "remove <project-ref> <item-path>",
		Short:         "Remove a file or folder from FG metadata.",
		Example:       c.name + " remove ./project.fg Root/old.txt --force --content ./encrypted --password-env FG_PASSWORD",
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

func (c cli) runRename(options renameOptions) error {
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}
	databasePath, err := databasePathFromProjectRef(options.projectRef)
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
	databasePath, err := databasePathFromProjectRef(options.projectRef)
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
		for _, operation := range result.Operations {
			if operation.Type != "delete" {
				return fmt.Errorf("unsupported content operation %q", operation.Type)
			}
			target, err := content.SafeJoin(options.contentRoot, operation.TargetPath)
			if err != nil {
				return fmt.Errorf("resolve delete target: %w", err)
			}
			if err := os.RemoveAll(target); err != nil {
				return fmt.Errorf("delete encrypted content %s: %w", operation.TargetPath, err)
			}
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

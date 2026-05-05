package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"foldersguard/internal/db"
	"foldersguard/internal/storage"
)

type renameOptions struct {
	projectRef      string
	itemPath        string
	newName         string
	passwordOptions passwordOptions
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

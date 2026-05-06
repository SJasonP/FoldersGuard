package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/storage"
)

type shareOptions struct {
	projectRef           string
	itemPaths            []string
	outputDatabase       string
	passwordOptions      passwordOptions
	sharePasswordOptions sharePasswordOptions
	force                bool
}

func (c cli) shareCommand() *cobra.Command {
	options := shareOptions{}
	command := &cobra.Command{
		Use:           "share <project-id> <item-path>...",
		Short:         "Create a share database for selected files and folders.",
		Example:       c.name + " share <project-id> Root/docs Root/readme.txt --out-database ./share.fgs --share-password-env FG_SHARE_PASSWORD --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectRef = args[0]
			options.itemPaths = args[1:]
			return c.runShare(options)
		},
	}
	command.Flags().StringVar(&options.outputDatabase, "out-database", "", "share database output path")
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read project password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read project password from an environment variable")
	command.Flags().BoolVar(&options.sharePasswordOptions.passwordStdin, "share-password-stdin", false, "read share password from stdin")
	command.Flags().StringVar(&options.sharePasswordOptions.passwordEnv, "share-password-env", "", "read share password from an environment variable")
	command.Flags().BoolVar(&options.sharePasswordOptions.noPassword, "no-share-password", false, "create an unprotected bearer share database")
	command.Flags().BoolVar(&options.force, "force", false, "replace existing outputs")
	mustMarkRequired(command, "out-database")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	command.MarkFlagsMutuallyExclusive("share-password-stdin", "share-password-env", "no-share-password")
	command.MarkFlagsMutuallyExclusive("password-stdin", "share-password-stdin")
	return command
}

func (c cli) runShare(options shareOptions) error {
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}
	if !format.IsSetExtension(options.outputDatabase) {
		return fmt.Errorf("share database output must use %s extension", format.SetExtension)
	}
	databasePath, err := activeProjectDatabasePathFromID(options.projectRef)
	if err != nil {
		return err
	}
	if err := prepareFileOutput(options.outputDatabase, options.force); err != nil {
		return err
	}
	sharePassword, passwordProtected, err := c.readSharePassword(options.sharePasswordOptions)
	if err != nil {
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
	selection, err := store.SelectShare(ctx, options.itemPaths, time.Now())
	if err != nil {
		return err
	}
	if err := writeShareDatabase(ctx, db.Config{
		Path:       options.outputDatabase,
		DriverName: db.SQLCipherDriver,
		Password:   sharePassword,
	}, selection.Plan); err != nil {
		return err
	}

	fmt.Fprintf(c.out, "project_id=%s\n", selection.SourceProjectID)
	fmt.Fprintf(c.out, "share_id=%s\n", selection.ShareID)
	fmt.Fprintf(c.out, "share_database=%s\n", options.outputDatabase)
	fmt.Fprintf(c.out, "items=%d\n", len(selection.Plan.Items))
	fmt.Fprintf(c.out, "files=%d\n", len(selection.Plan.Files))
	fmt.Fprintf(c.out, "folders=%d\n", len(selection.Plan.Folders))
	fmt.Fprintf(c.out, "parts=%d\n", len(selection.Plan.Parts))
	fmt.Fprintf(c.out, "password_protected=%t\n", passwordProtected)
	for _, location := range selection.ContentLocations {
		fmt.Fprintf(c.out, "content_location source=%s target=%s\n", location.SourcePath, location.TargetPath)
	}
	return nil
}

package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/project"
)

type encryptOptions struct {
	sourceFolder    string
	contentOutput   string
	databaseExport  string
	maxPartSize     int64
	passwordOptions passwordOptions
	force           bool
}

func (c cli) encryptCommand() *cobra.Command {
	options := encryptOptions{}
	command := &cobra.Command{
		Use:           "encrypt <source-folder>",
		Short:         "Encrypt one cleartext top-level folder and create one active FG project.",
		Example:       c.name + " encrypt ./clear --content-out ./encrypted --max-part-size 1073741824 --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.sourceFolder = args[0]
			if options.maxPartSize <= 0 {
				return fmt.Errorf("max part size must be positive")
			}
			return c.runEncrypt(options)
		},
	}
	command.Flags().StringVar(&options.contentOutput, "content-out", "", "encrypted content output folder")
	command.Flags().StringVar(&options.databaseExport, "export", "", "exported project database path")
	command.Flags().Int64Var(&options.maxPartSize, "max-part-size", 0, "maximum part size in bytes")
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.Flags().BoolVar(&options.force, "force", false, "replace existing outputs")
	mustMarkRequired(command, "content-out")
	mustMarkRequired(command, "max-part-size")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) runEncrypt(options encryptOptions) error {
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}

	sourceFolder := options.sourceFolder
	contentOutput := options.contentOutput
	if err := validateOutputOutsideSource(sourceFolder, contentOutput); err != nil {
		return err
	}
	if options.databaseExport != "" {
		if err := validateOutputOutsideSource(sourceFolder, options.databaseExport); err != nil {
			return err
		}
		if !format.IsProjectExtension(options.databaseExport) {
			return fmt.Errorf("database export must use %s extension", format.ProjectExtension)
		}
	}
	if err := prepareContentOutput(contentOutput, options.force); err != nil {
		return err
	}
	if options.databaseExport != "" {
		if err := prepareFileOutput(options.databaseExport, options.force); err != nil {
			return err
		}
	}

	ctx := context.Background()
	scan, err := fswalk.ScanTopFolder(sourceFolder)
	if err != nil {
		return err
	}
	plan, err := project.Planner{MaxPartSize: options.maxPartSize}.Plan(scan)
	if err != nil {
		return err
	}

	activeDatabase, err := activeProjectDatabasePath(plan.Project.ID.String())
	if err != nil {
		return err
	}
	if err := validateOutputOutsideSource(sourceFolder, activeDatabase); err != nil {
		return err
	}

	if err := writeProjectDatabase(ctx, db.Config{
		Path:       activeDatabase,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	}, plan); err != nil {
		return err
	}
	if options.databaseExport != "" {
		if err := writeProjectDatabase(ctx, db.Config{
			Path:       options.databaseExport,
			DriverName: db.SQLCipherDriver,
			Password:   password,
		}, plan); err != nil {
			return err
		}
	}
	if err := (project.Executor{OutputRoot: contentOutput}).EncryptContent(ctx, plan); err != nil {
		return err
	}

	fmt.Fprintf(c.out, "project_id=%s\n", plan.Project.ID)
	fmt.Fprintf(c.out, "root_folder_id=%s\n", plan.Project.RootFolderID)
	fmt.Fprintf(c.out, "content_output=%s\n", contentOutput)
	if options.databaseExport != "" {
		fmt.Fprintf(c.out, "database_export=%s\n", options.databaseExport)
	}
	fmt.Fprintf(c.out, "items=%d\n", len(plan.Items)+1)
	fmt.Fprintf(c.out, "folders=%d\n", len(plan.Folders)+1)
	fmt.Fprintf(c.out, "files=%d\n", len(plan.Files))
	fmt.Fprintf(c.out, "parts=%d\n", len(plan.Parts))
	fmt.Fprintf(c.out, "storage_objects=%d\n", len(plan.StorageObjects))
	fmt.Fprintf(c.out, "skipped=%d\n", len(scan.Skipped))
	for _, skipped := range scan.Skipped {
		fmt.Fprintf(c.out, "skipped_entry=%s reason=%s\n", skipped.RootRelativePath, skipped.Reason)
	}
	return nil
}

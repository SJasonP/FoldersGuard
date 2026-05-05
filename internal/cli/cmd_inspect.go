package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/db"
)

type inspectOptions struct {
	projectRef      string
	passwordOptions passwordOptions
}

func (c cli) inspectCommand() *cobra.Command {
	options := inspectOptions{}
	command := &cobra.Command{
		Use:           "inspect <project-ref>",
		Short:         "Display FG metadata without decrypting file content.",
		Example:       c.name + " inspect ./project.fg --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectRef = args[0]
			return c.runInspect(options)
		},
	}
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) runInspect(options inspectOptions) error {
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}
	databasePath, err := databasePathFromProjectRef(options.projectRef)
	if err != nil {
		return err
	}

	ctx := context.Background()
	plan, meta, err := readProjectDatabaseWithMeta(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(c.out, "project_id=%s\n", plan.Project.ID)
	fmt.Fprintf(c.out, "database_type=%s\n", meta["database_type"])
	fmt.Fprintf(c.out, "root_folder_id=%s\n", plan.Project.RootFolderID)
	fmt.Fprintf(c.out, "root_name=%s\n", plan.RootItem.RealName)
	fmt.Fprintf(c.out, "format_version=%s\n", meta["format_version"])
	fmt.Fprintf(c.out, "schema_version=%s\n", meta["schema_version"])
	fmt.Fprintf(c.out, "items=%d\n", len(plan.Items)+1)
	fmt.Fprintf(c.out, "folders=%d\n", len(plan.Folders)+1)
	fmt.Fprintf(c.out, "files=%d\n", len(plan.Files))
	fmt.Fprintf(c.out, "parts=%d\n", len(plan.Parts))
	fmt.Fprintf(c.out, "storage_objects=%d\n", len(plan.StorageObjects))
	return nil
}

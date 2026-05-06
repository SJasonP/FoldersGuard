package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/app"
	"foldersguard/internal/db"
	"foldersguard/internal/format"
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
		Example:       c.name + " inspect <project-id> --password-env FG_PASSWORD",
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
	ctx := context.Background()
	service, err := app.NewService("")
	if err != nil {
		return err
	}
	if format.IsSetExtension(options.projectRef) && !hasPasswordInput(options.passwordOptions) {
		result, err := service.Inspect(ctx, app.DatabaseOpen{
			ProjectRef: options.projectRef,
			Password:   db.UnprotectedSharePassword,
		})
		if err == nil {
			writeInspectResult(c.out, result)
			return nil
		}
	}
	password, err := c.readDatabasePassword(options.projectRef, options.passwordOptions)
	if err != nil {
		return err
	}
	result, err := service.Inspect(ctx, app.DatabaseOpen{
		ProjectRef: options.projectRef,
		Password:   password,
	})
	if err != nil {
		return err
	}
	writeInspectResult(c.out, result)
	return nil
}

func writeInspectResult(out interface {
	Write([]byte) (int, error)
}, result app.InspectResult) {
	fmt.Fprintf(out, "project_id=%s\n", result.ProjectID)
	fmt.Fprintf(out, "database_type=%s\n", result.DatabaseType)
	fmt.Fprintf(out, "root_folder_id=%s\n", result.RootFolderID)
	fmt.Fprintf(out, "root_name=%s\n", result.RootName)
	fmt.Fprintf(out, "format_version=%s\n", result.FormatVersion)
	fmt.Fprintf(out, "schema_version=%s\n", result.SchemaVersion)
	fmt.Fprintf(out, "items=%d\n", result.Items)
	fmt.Fprintf(out, "folders=%d\n", result.Folders)
	fmt.Fprintf(out, "files=%d\n", result.Files)
	fmt.Fprintf(out, "parts=%d\n", result.Parts)
	fmt.Fprintf(out, "storage_objects=%d\n", result.StorageObjects)
}

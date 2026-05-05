package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
)

type exportOptions struct {
	projectID       string
	outputPath      string
	passwordOptions passwordOptions
	force           bool
}

type importOptions struct {
	inputPath       string
	passwordOptions passwordOptions
	force           bool
}

func (c cli) exportCommand() *cobra.Command {
	options := exportOptions{}
	command := &cobra.Command{
		Use:           "export <project-id>",
		Short:         "Export an active project database from FG's data directory.",
		Example:       c.name + " export <project-id> --out ./project.fg --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectID = args[0]
			return c.runExport(options)
		},
	}
	command.Flags().StringVar(&options.outputPath, "out", "", "exported project database path")
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.Flags().BoolVar(&options.force, "force", false, "replace existing outputs")
	mustMarkRequired(command, "out")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) importCommand() *cobra.Command {
	options := importOptions{}
	command := &cobra.Command{
		Use:           "import <project.fg>",
		Short:         "Import an exported project database into FG's data directory.",
		Example:       c.name + " import ./project.fg --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.inputPath = args[0]
			return c.runImport(options)
		},
	}
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.Flags().BoolVar(&options.force, "force", false, "replace existing active project")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) runExport(options exportOptions) error {
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}
	if !format.IsProjectExtension(options.outputPath) {
		return fmt.Errorf("database output must use %s extension", format.ProjectExtension)
	}
	sourcePath, err := activeProjectDatabasePath(options.projectID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	plan, meta, err := readProjectDatabaseWithMeta(ctx, db.Config{
		Path:       sourcePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return err
	}
	if meta["database_type"] != "project" {
		return fmt.Errorf("database type = %q, want project", meta["database_type"])
	}
	if err := validateDistinctPaths(sourcePath, options.outputPath); err != nil {
		return err
	}
	if err := prepareFileOutput(options.outputPath, options.force); err != nil {
		return err
	}
	if err := copyFile(sourcePath, options.outputPath); err != nil {
		return err
	}

	fmt.Fprintf(c.out, "project_id=%s\n", plan.Project.ID)
	fmt.Fprintf(c.out, "database_output=%s\n", options.outputPath)
	return nil
}

func (c cli) runImport(options importOptions) error {
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}
	if !format.IsProjectExtension(options.inputPath) {
		return fmt.Errorf("input must use %s extension", format.ProjectExtension)
	}

	ctx := context.Background()
	plan, meta, err := readProjectDatabaseWithMeta(ctx, db.Config{
		Path:       options.inputPath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return err
	}
	if meta["database_type"] != "project" {
		return fmt.Errorf("database type = %q, want project", meta["database_type"])
	}
	activePath, err := activeProjectDatabasePath(plan.Project.ID.String())
	if err != nil {
		return err
	}
	if err := validateDistinctPaths(options.inputPath, activePath); err != nil {
		return err
	}
	if err := prepareFileOutput(activePath, options.force); err != nil {
		return err
	}
	if err := copyFile(options.inputPath, activePath); err != nil {
		return err
	}

	fmt.Fprintf(c.out, "project_id=%s\n", plan.Project.ID)
	fmt.Fprintln(c.out, "imported=true")
	return nil
}

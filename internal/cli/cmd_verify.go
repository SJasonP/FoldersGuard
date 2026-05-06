package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/db"
	"foldersguard/internal/project"
)

type verifyOptions struct {
	projectRef      string
	contentRoot     string
	passwordOptions passwordOptions
}

func (c cli) verifyCommand() *cobra.Command {
	options := verifyOptions{}
	command := &cobra.Command{
		Use:           "verify <project-ref>",
		Short:         "Verify database and encrypted content consistency.",
		Example:       c.name + " verify ./project.fg --content ./encrypted --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectRef = args[0]
			return c.runVerify(options)
		},
	}
	command.Flags().StringVar(&options.contentRoot, "content", "", "encrypted content folder")
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	mustMarkRequired(command, "content")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) runVerify(options verifyOptions) error {
	password, err := c.readDatabasePassword(options.projectRef, options.passwordOptions)
	if err != nil {
		return err
	}
	databasePath, err := databasePathFromProjectRef(options.projectRef)
	if err != nil {
		return err
	}
	if err := validateExistingDirectory(options.contentRoot, "content"); err != nil {
		return err
	}

	ctx := context.Background()
	plan, err := readProjectDatabase(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return err
	}
	report, err := (project.Verifier{EncryptedRoot: options.contentRoot}).VerifyContent(ctx, plan)
	if err != nil {
		return err
	}

	status := "ok"
	if !report.OK() {
		status = "failed"
	}
	fmt.Fprintf(c.out, "project_id=%s\n", plan.Project.ID)
	fmt.Fprintf(c.out, "checked_objects=%d\n", report.CheckedObjects)
	fmt.Fprintf(c.out, "missing_objects=%d\n", report.MissingObjects)
	fmt.Fprintf(c.out, "tampered_objects=%d\n", report.TamperedObjects)
	fmt.Fprintf(c.out, "extra_objects=%d\n", report.ExtraObjects)
	fmt.Fprintf(c.out, "status=%s\n", status)
	return nil
}

package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/app"
	"foldersguard/internal/db"
	"foldersguard/internal/format"
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
		Example:       c.name + " verify <project-id> --content ./encrypted --password-env FG_PASSWORD",
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
	ctx := context.Background()
	service, err := app.NewService("")
	if err != nil {
		return err
	}
	if format.IsSetExtension(options.projectRef) && !hasPasswordInput(options.passwordOptions) {
		result, err := service.Verify(ctx, app.DatabaseOpen{
			ProjectRef: options.projectRef,
			Password:   db.UnprotectedSharePassword,
		}, options.contentRoot)
		if err == nil {
			writeVerifyResult(c.out, result)
			return nil
		}
	}
	password, err := c.readDatabasePassword(options.projectRef, options.passwordOptions)
	if err != nil {
		return err
	}
	result, err := service.Verify(ctx, app.DatabaseOpen{
		ProjectRef: options.projectRef,
		Password:   password,
	}, options.contentRoot)
	if err != nil {
		return err
	}

	writeVerifyResult(c.out, result)
	return nil
}

func writeVerifyResult(out interface {
	Write([]byte) (int, error)
}, result app.VerifyResult) {
	fmt.Fprintf(out, "project_id=%s\n", result.ProjectID)
	fmt.Fprintf(out, "checked_objects=%d\n", result.CheckedObjects)
	fmt.Fprintf(out, "missing_objects=%d\n", result.MissingObjects)
	fmt.Fprintf(out, "tampered_objects=%d\n", result.TamperedObjects)
	fmt.Fprintf(out, "extra_objects=%d\n", result.ExtraObjects)
	fmt.Fprintf(out, "status=%s\n", result.Status)
}

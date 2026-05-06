package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/db"
	"foldersguard/internal/project"
)

type decryptOptions struct {
	projectRef      string
	contentRoot     string
	outputRoot      string
	passwordOptions passwordOptions
	force           bool
}

func (c cli) decryptCommand() *cobra.Command {
	options := decryptOptions{}
	command := &cobra.Command{
		Use:           "decrypt <project-ref>",
		Short:         "Decrypt encrypted content using a project or share database.",
		Example:       c.name + " decrypt ./project.fg --content ./encrypted --out ./restored --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectRef = args[0]
			return c.runDecrypt(options)
		},
	}
	command.Flags().StringVar(&options.contentRoot, "content", "", "encrypted content folder")
	command.Flags().StringVar(&options.outputRoot, "out", "", "restored plaintext output folder")
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.Flags().BoolVar(&options.force, "force", false, "replace existing outputs")
	mustMarkRequired(command, "content")
	mustMarkRequired(command, "out")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) runDecrypt(options decryptOptions) error {
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
	if err := validateOutputOutsideSource(options.contentRoot, options.outputRoot); err != nil {
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
	if err := prepareDirectoryOutput(options.outputRoot, options.force, "output"); err != nil {
		return err
	}
	if err := (project.Restorer{EncryptedRoot: options.contentRoot, OutputRoot: options.outputRoot}).RestoreContent(ctx, plan); err != nil {
		return err
	}

	fmt.Fprintf(c.out, "project_id=%s\n", plan.Project.ID)
	fmt.Fprintf(c.out, "output=%s\n", options.outputRoot)
	fmt.Fprintf(c.out, "folders=%d\n", countFolders(plan))
	fmt.Fprintf(c.out, "files=%d\n", len(plan.Files))
	fmt.Fprintf(c.out, "parts=%d\n", len(plan.Parts))
	fmt.Fprintf(c.out, "restored_files=%d\n", len(plan.Files))
	return nil
}

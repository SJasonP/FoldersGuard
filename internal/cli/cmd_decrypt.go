package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"foldersguard/internal/model"
	"foldersguard/internal/project"
)

type decryptOptions struct {
	projectRef      string
	contentRoot     string
	outputRoot      string
	passwordOptions passwordOptions
	force           bool
	resume          bool
	continueOnError bool
}

func (c cli) decryptCommand() *cobra.Command {
	options := decryptOptions{}
	command := &cobra.Command{
		Use:           "decrypt <project-ref>",
		Short:         "Decrypt encrypted content using a project or share database.",
		Example:       c.name + " decrypt <project-id> --content ./encrypted --out ./restored --password-env FG_PASSWORD",
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
	command.Flags().BoolVar(&options.resume, "resume", false, "continue an interrupted decryption, skipping outputs that already exist at the expected size")
	command.Flags().BoolVar(&options.continueOnError, "continue-on-error", false, "record item-level failures and restore the remaining files instead of aborting on the first error")
	command.MarkFlagsMutuallyExclusive("force", "resume")
	mustMarkRequired(command, "content")
	mustMarkRequired(command, "out")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) runDecrypt(options decryptOptions) error {
	if err := validateExistingDirectory(options.contentRoot, "content"); err != nil {
		return err
	}
	if err := validateOutputOutsideSource(options.contentRoot, options.outputRoot); err != nil {
		return err
	}

	ctx := context.Background()
	plan, err := c.readDatabaseFromProjectRef(ctx, options.projectRef, options.passwordOptions)
	if err != nil {
		return err
	}
	if options.resume {
		// Resuming keeps the existing partial output, so the non-empty output
		// must not be rejected or wiped.
		if err := os.MkdirAll(options.outputRoot, 0o755); err != nil {
			return fmt.Errorf("create output folder: %w", err)
		}
	} else if err := prepareDirectoryOutput(options.outputRoot, options.force, "output"); err != nil {
		return err
	}
	noiseMode, err := readNoiseFileHandling()
	if err != nil {
		return err
	}
	var failures []model.File
	onFileError := func(file model.File, _ error) {
		failures = append(failures, file)
	}
	report, err := (project.Restorer{
		EncryptedRoot:   options.contentRoot,
		OutputRoot:      options.outputRoot,
		NoiseMode:       noiseMode,
		Resume:          options.resume,
		ContinueOnError: options.continueOnError,
		OnFileError:     onFileError,
	}).RestoreContentReport(ctx, plan)
	if err != nil {
		return err
	}

	fmt.Fprintf(c.out, "project_id=%s\n", plan.Project.ID)
	fmt.Fprintf(c.out, "output=%s\n", options.outputRoot)
	fmt.Fprintf(c.out, "folders=%d\n", countFolders(plan))
	fmt.Fprintf(c.out, "files=%d\n", len(plan.Files))
	fmt.Fprintf(c.out, "parts=%d\n", len(plan.Parts))
	fmt.Fprintf(c.out, "restored_files=%d\n", report.DecryptedFiles)
	fmt.Fprintf(c.out, "failed_files=%d\n", len(failures))
	if len(failures) > 0 {
		for _, file := range failures {
			fmt.Fprintf(c.err, "failed_file=%s\n", file.ID)
		}
		return fmt.Errorf("%d file(s) failed to decrypt", len(failures))
	}
	return nil
}

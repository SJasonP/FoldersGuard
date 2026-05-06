package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/fswalk"
	"foldersguard/internal/project"
)

func (c cli) planCommand() *cobra.Command {
	plan := &cobra.Command{
		Use:           "plan",
		Short:         "Preview FG operations without writing encrypted content or databases.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	plan.AddCommand(c.planEncryptCommand())
	return plan
}

func (c cli) planEncryptCommand() *cobra.Command {
	var maxPartSize int64
	command := &cobra.Command{
		Use:           "encrypt <source-folder>",
		Short:         "Print an encryption plan without writing encrypted content or FG databases.",
		Example:       c.name + " plan encrypt ./clear --max-part-size 1073741824",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if maxPartSize <= 0 {
				return fmt.Errorf("max part size must be positive")
			}
			return c.runPlanEncrypt(args[0], maxPartSize)
		},
	}
	command.Flags().Int64Var(&maxPartSize, "max-part-size", 0, "maximum part size in bytes")
	mustMarkRequired(command, "max-part-size")
	return command
}

func (c cli) runPlanEncrypt(sourceFolder string, maxPartSize int64) error {
	scan, err := fswalk.ScanTopFolder(sourceFolder)
	if err != nil {
		return err
	}
	plan, err := project.Planner{MaxPartSize: maxPartSize}.Plan(scan)
	if err != nil {
		return err
	}

	fmt.Fprintf(c.out, "items=%d\n", len(plan.Items)+1)
	fmt.Fprintf(c.out, "folders=%d\n", countFolders(plan))
	fmt.Fprintf(c.out, "files=%d\n", len(plan.Files))
	fmt.Fprintf(c.out, "parts=%d\n", len(plan.Parts))
	fmt.Fprintf(c.out, "storage_objects=%d\n", len(plan.StorageObjects))
	fmt.Fprintf(c.out, "skipped=%d\n", len(scan.Skipped))
	for _, skipped := range scan.Skipped {
		fmt.Fprintf(c.out, "skipped_entry=%s reason=%s\n", skipped.RootRelativePath, skipped.Reason)
	}
	return nil
}

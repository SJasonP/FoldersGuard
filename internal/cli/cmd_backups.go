package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/app"
)

func (c cli) backupsCommand() *cobra.Command {
	command := &cobra.Command{
		Use:           "backups",
		Short:         "Manage project database backups.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	command.AddCommand(c.backupsListCommand())
	command.AddCommand(c.backupsRestoreCommand())
	return command
}

func (c cli) backupsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "list <project-id>",
		Short:         "List retained database backups for a project, newest first.",
		Example:       c.name + " backups list <project-id>",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runBackupsList(args[0])
		},
	}
}

func (c cli) runBackupsList(projectID string) error {
	service, err := app.NewService("")
	if err != nil {
		return err
	}
	backups, err := service.ListProjectBackups(projectID)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.out, "project_id=%s\n", projectID)
	for _, backup := range backups {
		fmt.Fprintf(c.out, "backup_id=%s reason=%s created=%s size=%d\n",
			backup.ID,
			backup.Reason,
			backup.CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
			backup.Size,
		)
	}
	return nil
}

func (c cli) backupsRestoreCommand() *cobra.Command {
	var force bool
	command := &cobra.Command{
		Use:           "restore <project-id> <backup-id>",
		Short:         "Restore a project database from a retained backup.",
		Example:       c.name + " backups restore <project-id> <backup-id> --force",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runBackupsRestore(args[0], args[1], force)
		},
	}
	command.Flags().BoolVar(&force, "force", false, "overwrite the existing active project database")
	return command
}

func (c cli) runBackupsRestore(projectID, backupID string, force bool) error {
	service, err := app.NewService("")
	if err != nil {
		return err
	}
	if _, err := service.RestoreProjectBackup(projectID, backupID, force); err != nil {
		return err
	}
	fmt.Fprintf(c.out, "project_id=%s\n", projectID)
	fmt.Fprintf(c.out, "restored_from=%s\n", backupID)
	return nil
}

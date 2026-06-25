package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/app"
)

type passwdOptions struct {
	projectRef  string
	sharePath   string
	oldPassword passwordOptions
	newPassword passwordOptions
}

func (c cli) passwdCommand() *cobra.Command {
	options := passwdOptions{}
	command := &cobra.Command{
		Use:           "passwd [project-ref]",
		Short:         "Change a project or share database password without re-encrypting content.",
		Example:       c.name + " passwd <project-id>\n  " + c.name + " passwd --share path/to/share.fgs",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				options.projectRef = args[0]
			}
			return c.runPasswd(options)
		},
	}
	command.Flags().StringVar(&options.sharePath, "share", "", "change the password of a share database instead of a project")
	command.Flags().BoolVar(&options.oldPassword.passwordStdin, "password-stdin", false, "read the current password from stdin")
	command.Flags().StringVar(&options.oldPassword.passwordEnv, "password-env", "", "read the current password from an environment variable")
	command.Flags().BoolVar(&options.newPassword.passwordStdin, "new-password-stdin", false, "read the new password from stdin")
	command.Flags().StringVar(&options.newPassword.passwordEnv, "new-password-env", "", "read the new password from an environment variable")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	command.MarkFlagsMutuallyExclusive("new-password-stdin", "new-password-env")
	command.MarkFlagsMutuallyExclusive("password-stdin", "new-password-stdin")
	return command
}

func (c cli) runPasswd(options passwdOptions) error {
	if options.sharePath == "" && options.projectRef == "" {
		return fmt.Errorf("a project id or --share path is required")
	}
	if options.sharePath != "" && options.projectRef != "" {
		return fmt.Errorf("specify either a project id or --share, not both")
	}

	ctx := context.Background()
	service, err := app.NewService("")
	if err != nil {
		return err
	}

	oldPassword, err := c.readPasswordFor(options.oldPassword, passwordPrompt{label: "Current password"})
	if err != nil {
		return err
	}
	newPassword, err := c.readPasswordFor(options.newPassword, passwordPrompt{
		label:        "New password",
		confirm:      true,
		confirmLabel: "Confirm new password",
	})
	if err != nil {
		return err
	}

	if options.sharePath != "" {
		if err := service.ChangeSharePassword(ctx, options.sharePath, oldPassword, newPassword); err != nil {
			return err
		}
		fmt.Fprintf(c.out, "share=%s\n", options.sharePath)
		fmt.Fprintln(c.out, "rekeyed=true")
		fmt.Fprintln(c.out, "content_operations=0")
		return nil
	}

	if err := service.ChangeProjectPassword(ctx, options.projectRef, oldPassword, newPassword); err != nil {
		return err
	}
	fmt.Fprintf(c.out, "project_id=%s\n", options.projectRef)
	fmt.Fprintln(c.out, "rekeyed=true")
	fmt.Fprintln(c.out, "content_operations=0")
	return nil
}

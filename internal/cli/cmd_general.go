package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/format"
)

func (c cli) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "version",
		Short:         "Print the application id and native format version.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(c.out, "app_id=%s\n", format.AppID)
			fmt.Fprintf(c.out, "format_version=%s\n", format.NativeFormatVersion)
			return nil
		},
	}
}

func (c cli) schemaCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "schema",
		Short:         "Print the FG database schema version.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(c.out, "schema_version=%d\n", format.SchemaVersion)
			return nil
		},
	}
}

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/format"
)

func (c cli) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "version",
		Short:         "Print product and format version information.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(c.out, "app_id=%s\n", format.AppID)
			fmt.Fprintf(c.out, "product_version=%s\n", format.ProductVersion)
			fmt.Fprintf(c.out, "format_version=%s\n", format.FormatVersion)
			return nil
		},
	}
}

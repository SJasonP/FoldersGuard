package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func Run(args []string) error {
	return newCLI("foldersguard", os.Stdin, os.Stdout).run(args)
}

func RunWithIO(name string, args []string, in io.Reader, out io.Writer) error {
	return newCLI(name, in, out).run(args)
}

type cli struct {
	name string
	in   io.Reader
	out  io.Writer
}

func newCLI(name string, in io.Reader, out io.Writer) cli {
	if name == "" {
		name = "foldersguard"
	}
	if in == nil {
		in = strings.NewReader("")
	}
	if out == nil {
		out = io.Discard
	}
	return cli{name: name, in: in, out: out}
}

func (c cli) run(args []string) error {
	root := c.rootCommand()
	root.SetArgs(args)
	return root.Execute()
}

func (c cli) rootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           c.name,
		Short:         "FoldersGuard protects folder contents with encrypted metadata and content objects.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unknown command %q", args[0])
			}
			return cmd.Help()
		},
	}
	root.SetIn(c.in)
	root.SetOut(c.out)
	root.SetErr(io.Discard)
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(c.versionCommand())
	root.AddCommand(c.schemaCommand())
	root.AddCommand(c.planCommand())
	root.AddCommand(c.encryptCommand())
	root.AddCommand(c.decryptCommand())
	root.AddCommand(c.inspectCommand())
	root.AddCommand(c.verifyCommand())
	root.AddCommand(c.exportCommand())
	root.AddCommand(c.importCommand())
	root.AddCommand(c.renameCommand())
	return root
}

func mustMarkRequired(command *cobra.Command, name string) {
	if err := command.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

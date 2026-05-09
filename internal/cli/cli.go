package cli

import (
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func Run(args []string) error {
	return newCLIWithErr("foldersguard", os.Stdin, os.Stdout, os.Stderr).run(args)
}

func RunWithIO(name string, args []string, in io.Reader, out io.Writer) error {
	return newCLI(name, in, out).run(args)
}

func RunWithIOErr(name string, args []string, in io.Reader, out io.Writer, errOut io.Writer) error {
	return newCLIWithErr(name, in, out, errOut).run(args)
}

type cli struct {
	name string
	in   io.Reader
	out  io.Writer
	err  io.Writer
}

func newCLI(name string, in io.Reader, out io.Writer) cli {
	return newCLIWithErr(name, in, out, io.Discard)
}

func newCLIWithErr(name string, in io.Reader, out io.Writer, errOut io.Writer) cli {
	if name == "" {
		name = "foldersguard"
	}
	if in == nil {
		in = strings.NewReader("")
	}
	if out == nil {
		out = io.Discard
	}
	if errOut == nil {
		errOut = io.Discard
	}
	return cli{name: name, in: in, out: out, err: errOut}
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
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	root.SetIn(c.in)
	root.SetOut(c.out)
	root.SetErr(io.Discard)
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(c.versionCommand())
	root.AddCommand(c.planCommand())
	root.AddCommand(c.encryptCommand())
	root.AddCommand(c.decryptCommand())
	root.AddCommand(c.inspectCommand())
	root.AddCommand(c.verifyCommand())
	root.AddCommand(c.exportCommand())
	root.AddCommand(c.importCommand())
	root.AddCommand(c.shareCommand())
	root.AddCommand(c.renameCommand())
	root.AddCommand(c.addCommand())
	root.AddCommand(c.moveCommand())
	root.AddCommand(c.removeCommand())
	return root
}

func mustMarkRequired(command *cobra.Command, name string) {
	if err := command.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

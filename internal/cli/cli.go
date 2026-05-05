package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
	"foldersguard/internal/project"
	"foldersguard/internal/storage"
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

type passwordOptions struct {
	passwordStdin bool
	passwordEnv   string
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
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("unknown command %q", args[0])
			}
			c.printUsage()
			return nil
		},
	}
	root.SetIn(c.in)
	root.SetOut(c.out)
	root.SetErr(io.Discard)
	root.SetHelpFunc(func(command *cobra.Command, args []string) {
		c.printUsage()
	})
	root.AddCommand(c.versionCommand())
	root.AddCommand(c.schemaCommand())
	root.AddCommand(c.planCommand())
	root.AddCommand(c.encryptCommand())
	return root
}

func (c cli) printUsage() {
	fmt.Fprintln(c.out, "FoldersGuard")
	fmt.Fprintln(c.out)
	fmt.Fprintln(c.out, "Usage:")
	fmt.Fprintf(c.out, "  %s help\n", c.name)
	fmt.Fprintf(c.out, "  %s version\n", c.name)
	fmt.Fprintf(c.out, "  %s schema\n", c.name)
	fmt.Fprintf(c.out, "  %s plan encrypt <source-folder> --max-part-size <bytes>\n", c.name)
	fmt.Fprintf(c.out, "  %s encrypt <source-folder> --content-out <encrypted-content-folder> --max-part-size <bytes> [--export <project.fg>] [--password-stdin | --password-env <name>]\n", c.name)
}

func (c cli) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "version",
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
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(c.out, "schema_version=%d\n", format.SchemaVersion)
			return nil
		},
	}
}

func (c cli) planCommand() *cobra.Command {
	plan := &cobra.Command{
		Use:           "plan",
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
	return command
}

func (c cli) encryptCommand() *cobra.Command {
	options := encryptOptions{}
	command := &cobra.Command{
		Use:           "encrypt <source-folder>",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.sourceFolder = args[0]
			if options.contentOutput == "" {
				return fmt.Errorf("--content-out is required")
			}
			if options.maxPartSize <= 0 {
				return fmt.Errorf("max part size must be positive")
			}
			if options.passwordOptions.passwordStdin && options.passwordOptions.passwordEnv != "" {
				return fmt.Errorf("choose only one password input mode")
			}
			return c.runEncrypt(options)
		},
	}
	command.Flags().StringVar(&options.contentOutput, "content-out", "", "encrypted content output folder")
	command.Flags().StringVar(&options.databaseExport, "export", "", "exported project database path")
	command.Flags().Int64Var(&options.maxPartSize, "max-part-size", 0, "maximum part size in bytes")
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.Flags().BoolVar(&options.force, "force", false, "replace existing outputs")
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
	fmt.Fprintf(c.out, "folders=%d\n", len(plan.Folders)+1)
	fmt.Fprintf(c.out, "files=%d\n", len(plan.Files))
	fmt.Fprintf(c.out, "parts=%d\n", len(plan.Parts))
	fmt.Fprintf(c.out, "storage_objects=%d\n", len(plan.StorageObjects))
	fmt.Fprintf(c.out, "skipped=%d\n", len(scan.Skipped))
	for _, skipped := range scan.Skipped {
		fmt.Fprintf(c.out, "skipped_entry=%s reason=%s\n", skipped.RootRelativePath, skipped.Reason)
	}
	return nil
}

func (c cli) runEncrypt(options encryptOptions) error {
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}

	sourceFolder := options.sourceFolder
	contentOutput := options.contentOutput
	if err := validateOutputOutsideSource(sourceFolder, contentOutput); err != nil {
		return err
	}
	if options.databaseExport != "" {
		if err := validateOutputOutsideSource(sourceFolder, options.databaseExport); err != nil {
			return err
		}
		if !format.IsProjectExtension(options.databaseExport) {
			return fmt.Errorf("database export must use %s extension", format.ProjectExtension)
		}
	}
	if err := prepareContentOutput(contentOutput, options.force); err != nil {
		return err
	}
	if options.databaseExport != "" {
		if err := prepareFileOutput(options.databaseExport, options.force); err != nil {
			return err
		}
	}

	ctx := context.Background()
	scan, err := fswalk.ScanTopFolder(sourceFolder)
	if err != nil {
		return err
	}
	plan, err := project.Planner{MaxPartSize: options.maxPartSize}.Plan(scan)
	if err != nil {
		return err
	}

	activeDatabase, err := activeProjectDatabasePath(plan.Project.ID.String())
	if err != nil {
		return err
	}
	if err := validateOutputOutsideSource(sourceFolder, activeDatabase); err != nil {
		return err
	}

	if err := writeProjectDatabase(ctx, db.Config{
		Path:       activeDatabase,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	}, plan); err != nil {
		return err
	}
	if options.databaseExport != "" {
		if err := writeProjectDatabase(ctx, db.Config{
			Path:       options.databaseExport,
			DriverName: db.SQLCipherDriver,
			Password:   password,
		}, plan); err != nil {
			return err
		}
	}
	if err := (project.Executor{OutputRoot: contentOutput}).EncryptContent(ctx, plan); err != nil {
		return err
	}

	fmt.Fprintf(c.out, "project_id=%s\n", plan.Project.ID)
	fmt.Fprintf(c.out, "root_folder_id=%s\n", plan.Project.RootFolderID)
	fmt.Fprintf(c.out, "content_output=%s\n", contentOutput)
	if options.databaseExport != "" {
		fmt.Fprintf(c.out, "database_export=%s\n", options.databaseExport)
	}
	fmt.Fprintf(c.out, "items=%d\n", len(plan.Items)+1)
	fmt.Fprintf(c.out, "folders=%d\n", len(plan.Folders)+1)
	fmt.Fprintf(c.out, "files=%d\n", len(plan.Files))
	fmt.Fprintf(c.out, "parts=%d\n", len(plan.Parts))
	fmt.Fprintf(c.out, "storage_objects=%d\n", len(plan.StorageObjects))
	fmt.Fprintf(c.out, "skipped=%d\n", len(scan.Skipped))
	for _, skipped := range scan.Skipped {
		fmt.Fprintf(c.out, "skipped_entry=%s reason=%s\n", skipped.RootRelativePath, skipped.Reason)
	}
	return nil
}

type encryptOptions struct {
	sourceFolder    string
	contentOutput   string
	databaseExport  string
	maxPartSize     int64
	passwordOptions passwordOptions
	force           bool
}

func prepareContentOutput(path string, force bool) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return fmt.Errorf("create content output folder: %w", err)
			}
			return nil
		}
		return fmt.Errorf("stat content output: %w", err)
	}
	if !info.IsDir() {
		if !force {
			return fmt.Errorf("content output exists and is not a directory; use --force to replace it")
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove existing content output file: %w", err)
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create content output folder: %w", err)
		}
		return nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read content output folder: %w", err)
	}
	if len(entries) > 0 {
		if !force {
			return fmt.Errorf("content output folder is not empty; use --force to replace it")
		}
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove existing content output folder: %w", err)
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create content output folder: %w", err)
		}
	}
	return nil
}

func prepareFileOutput(path string, force bool) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return fmt.Errorf("create output folder: %w", err)
			}
			return nil
		}
		return fmt.Errorf("stat output file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("output file path is a directory")
	}
	if !force {
		return fmt.Errorf("output file exists; use --force to replace it")
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove existing output file: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output folder: %w", err)
	}
	return nil
}

func (c cli) readPassword(options passwordOptions) (string, error) {
	if options.passwordStdin {
		data, err := io.ReadAll(c.in)
		if err != nil {
			return "", fmt.Errorf("read password from stdin: %w", err)
		}
		password := strings.TrimRight(string(data), "\r\n")
		if password == "" {
			return "", fmt.Errorf("password must not be empty")
		}
		return password, nil
	}
	if options.passwordEnv != "" {
		password := os.Getenv(options.passwordEnv)
		if password == "" {
			return "", fmt.Errorf("password environment variable %s is empty or unset", options.passwordEnv)
		}
		return password, nil
	}
	return "", fmt.Errorf("password input is required")
}

func activeProjectDatabasePath(projectID string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(configDir, format.AppID, "projects", projectID+format.ProjectExtension), nil
}

func validateOutputOutsideSource(source, output string) error {
	sourceAbs, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("resolve source path: %w", err)
	}
	outputAbs, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}
	relative, err := filepath.Rel(sourceAbs, outputAbs)
	if err != nil {
		return fmt.Errorf("compare source and output paths: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return nil
	}
	return fmt.Errorf("output path must be outside the source folder")
}

func writeProjectDatabase(ctx context.Context, config db.Config, plan model.PlannedProject) error {
	database, err := db.OpenProject(ctx, config)
	if err != nil {
		return err
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		return err
	}
	if err := store.InitProject(ctx, storage.ProjectSpec{
		ProjectID:       plan.Project.ID,
		RootFolderID:    plan.Project.RootFolderID,
		RootVisibleName: plan.RootItem.VisibleName,
		RootRealName:    plan.RootItem.RealName,
		RootFolderKey:   plan.RootFolder.Key,
		DatabaseType:    "project",
		CreatedAt:       plan.Project.CreatedAt,
	}); err != nil {
		return err
	}
	if err := store.WritePlannedProject(ctx, plan); err != nil {
		return err
	}
	return nil
}

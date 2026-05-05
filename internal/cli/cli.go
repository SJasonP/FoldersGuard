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
	return root
}

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

func (c cli) encryptCommand() *cobra.Command {
	options := encryptOptions{}
	command := &cobra.Command{
		Use:           "encrypt <source-folder>",
		Short:         "Encrypt one cleartext top-level folder and create one active FG project.",
		Example:       c.name + " encrypt ./clear --content-out ./encrypted --max-part-size 1073741824 --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.sourceFolder = args[0]
			if options.maxPartSize <= 0 {
				return fmt.Errorf("max part size must be positive")
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
	mustMarkRequired(command, "content-out")
	mustMarkRequired(command, "max-part-size")
	mustMarkMutuallyExclusive(command, "password-stdin", "password-env")
	return command
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
	mustMarkMutuallyExclusive(command, "password-stdin", "password-env")
	return command
}

func mustMarkRequired(command *cobra.Command, name string) {
	if err := command.MarkFlagRequired(name); err != nil {
		panic(err)
	}
}

func mustMarkMutuallyExclusive(command *cobra.Command, names ...string) {
	command.MarkFlagsMutuallyExclusive(names...)
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

type decryptOptions struct {
	projectRef      string
	contentRoot     string
	outputRoot      string
	passwordOptions passwordOptions
	force           bool
}

func (c cli) runDecrypt(options decryptOptions) error {
	password, err := c.readPassword(options.passwordOptions)
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
	fmt.Fprintf(c.out, "folders=%d\n", len(plan.Folders)+1)
	fmt.Fprintf(c.out, "files=%d\n", len(plan.Files))
	fmt.Fprintf(c.out, "parts=%d\n", len(plan.Parts))
	fmt.Fprintf(c.out, "restored_files=%d\n", len(plan.Files))
	return nil
}

func prepareContentOutput(path string, force bool) error {
	return prepareDirectoryOutput(path, force, "content output")
}

func prepareDirectoryOutput(path string, force bool, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return fmt.Errorf("create %s folder: %w", label, err)
			}
			return nil
		}
		return fmt.Errorf("stat %s: %w", label, err)
	}
	if !info.IsDir() {
		if !force {
			return fmt.Errorf("%s exists and is not a directory; use --force to replace it", label)
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove existing %s file: %w", label, err)
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create %s folder: %w", label, err)
		}
		return nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read %s folder: %w", label, err)
	}
	if len(entries) > 0 {
		if !force {
			return fmt.Errorf("%s folder is not empty; use --force to replace it", label)
		}
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove existing %s folder: %w", label, err)
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create %s folder: %w", label, err)
		}
	}
	return nil
}

func validateExistingDirectory(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s folder: %w", label, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s must be a directory", label)
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

func databasePathFromProjectRef(projectRef string) (string, error) {
	if projectRef == "" {
		return "", fmt.Errorf("project reference is required")
	}
	if format.IsProjectExtension(projectRef) || format.IsSetExtension(projectRef) {
		return projectRef, nil
	}
	return activeProjectDatabasePath(projectRef)
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

func readProjectDatabase(ctx context.Context, config db.Config) (model.PlannedProject, error) {
	database, err := db.OpenProject(ctx, config)
	if err != nil {
		return model.PlannedProject{}, err
	}
	defer database.Close()

	store, err := storage.NewStore(database)
	if err != nil {
		return model.PlannedProject{}, err
	}
	return store.ReadPlannedProject(ctx)
}

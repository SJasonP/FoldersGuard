package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"foldersguard/internal/db"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/project"
	"foldersguard/internal/storage"
)

func (c cli) planCommand() *cobra.Command {
	plan := &cobra.Command{
		Use:           "plan",
		Short:         "Preview FG operations without writing encrypted content or databases.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	plan.AddCommand(c.planEncryptCommand())
	plan.AddCommand(c.planAddCommand())
	plan.AddCommand(c.planMoveCommand())
	plan.AddCommand(c.planRemoveCommand())
	return plan
}

type planAddOptions struct {
	projectRef       string
	sourcePath       string
	targetFolderPath string
	stagingContent   string
	maxPartSize      int64
	passwordOptions  passwordOptions
}

type planMoveOptions struct {
	projectRef       string
	itemPath         string
	targetFolderPath string
	passwordOptions  passwordOptions
}

type planRemoveOptions struct {
	projectRef      string
	itemPath        string
	passwordOptions passwordOptions
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

func (c cli) planAddCommand() *cobra.Command {
	options := planAddOptions{}
	command := &cobra.Command{
		Use:           "add <project-id> <source-path> <target-folder-path>",
		Short:         "Print add operations without writing content or metadata.",
		Example:       c.name + " plan add <project-id> ./new Root/docs --staging-content ./staging --max-part-size 1073741824 --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectRef = args[0]
			options.sourcePath = args[1]
			options.targetFolderPath = args[2]
			if options.maxPartSize <= 0 {
				return fmt.Errorf("max part size must be positive")
			}
			return c.runPlanAdd(options)
		},
	}
	command.Flags().StringVar(&options.stagingContent, "staging-content", "", "staged encrypted content folder")
	command.Flags().Int64Var(&options.maxPartSize, "max-part-size", 0, "maximum part size in bytes")
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	mustMarkRequired(command, "staging-content")
	mustMarkRequired(command, "max-part-size")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) planMoveCommand() *cobra.Command {
	options := planMoveOptions{}
	command := &cobra.Command{
		Use:           "move <project-id> <item-path> <target-folder-path>",
		Short:         "Print move operations without writing content or metadata.",
		Example:       c.name + " plan move <project-id> Root/old.txt Root/docs --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectRef = args[0]
			options.itemPath = args[1]
			options.targetFolderPath = args[2]
			return c.runPlanMove(options)
		},
	}
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
	return command
}

func (c cli) planRemoveCommand() *cobra.Command {
	options := planRemoveOptions{}
	command := &cobra.Command{
		Use:           "remove <project-id> <item-path>",
		Short:         "Print remove operations without writing content or metadata.",
		Example:       c.name + " plan remove <project-id> Root/old.txt --password-env FG_PASSWORD",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.projectRef = args[0]
			options.itemPath = args[1]
			return c.runPlanRemove(options)
		},
	}
	command.Flags().BoolVar(&options.passwordOptions.passwordStdin, "password-stdin", false, "read password from stdin")
	command.Flags().StringVar(&options.passwordOptions.passwordEnv, "password-env", "", "read password from an environment variable")
	command.MarkFlagsMutuallyExclusive("password-stdin", "password-env")
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
	return nil
}

func (c cli) runPlanAdd(options planAddOptions) error {
	password, err := c.readPassword(options.passwordOptions)
	if err != nil {
		return err
	}
	databasePath, err := activeProjectDatabasePathFromID(options.projectRef)
	if err != nil {
		return err
	}
	if err := validateExistingFile(databasePath, "database"); err != nil {
		return err
	}
	if err := validateOutputOutsideSource(options.sourcePath, options.stagingContent); err != nil {
		return err
	}

	ctx := context.Background()
	database, err := db.OpenProject(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return err
	}
	defer database.Close()

	scan, err := fswalk.ScanPath(options.sourcePath)
	if err != nil {
		return err
	}
	addition, err := project.AddPlanner{MaxPartSize: options.maxPartSize}.Plan(scan)
	if err != nil {
		return err
	}
	store, err := storage.NewStore(database)
	if err != nil {
		return err
	}
	_, operations, err := store.PrepareAdd(ctx, options.targetFolderPath, addition)
	if err != nil {
		return err
	}
	plan, err := store.ReadPlannedProject(ctx)
	if err != nil {
		return err
	}
	printOperations(c, plan.Project.ID.String(), operations)
	return nil
}

func (c cli) runPlanMove(options planMoveOptions) error {
	projectID, operations, err := c.planMetadataOperation(options.projectRef, options.passwordOptions, func(ctx context.Context, store *storage.Store) (string, []storage.ContentOperation, error) {
		projectID, operations, err := store.PlanMove(ctx, options.itemPath, options.targetFolderPath)
		return projectID.String(), operations, err
	})
	if err != nil {
		return err
	}
	printOperations(c, projectID, operations)
	return nil
}

func (c cli) runPlanRemove(options planRemoveOptions) error {
	projectID, operations, err := c.planMetadataOperation(options.projectRef, options.passwordOptions, func(ctx context.Context, store *storage.Store) (string, []storage.ContentOperation, error) {
		projectID, operations, err := store.PlanRemove(ctx, options.itemPath)
		return projectID.String(), operations, err
	})
	if err != nil {
		return err
	}
	printOperations(c, projectID, operations)
	return nil
}

func (c cli) planMetadataOperation(projectRef string, passwordOptions passwordOptions, run func(context.Context, *storage.Store) (string, []storage.ContentOperation, error)) (string, []storage.ContentOperation, error) {
	password, err := c.readPassword(passwordOptions)
	if err != nil {
		return "", nil, err
	}
	databasePath, err := activeProjectDatabasePathFromID(projectRef)
	if err != nil {
		return "", nil, err
	}
	if err := validateExistingFile(databasePath, "database"); err != nil {
		return "", nil, err
	}

	ctx := context.Background()
	database, err := db.OpenProject(ctx, db.Config{
		Path:       databasePath,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	})
	if err != nil {
		return "", nil, err
	}
	defer database.Close()
	store, err := storage.NewStore(database)
	if err != nil {
		return "", nil, err
	}
	return run(ctx, store)
}

func printOperations(c cli, projectID string, operations []storage.ContentOperation) {
	fmt.Fprintf(c.out, "project_id=%s\n", projectID)
	fmt.Fprintf(c.out, "operations=%d\n", len(operations))
	for _, operation := range operations {
		switch operation.Type {
		case "delete":
			fmt.Fprintf(c.out, "operation=%s target=%s\n", operation.Type, operation.TargetPath)
		default:
			fmt.Fprintf(c.out, "operation=%s source=%s target=%s\n", operation.Type, operation.SourcePath, operation.TargetPath)
		}
	}
}

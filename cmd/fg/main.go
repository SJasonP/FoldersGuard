package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"foldersguard/internal/db"
	"foldersguard/internal/format"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/model"
	"foldersguard/internal/project"
	"foldersguard/internal/storage"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "fg:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "version":
		fmt.Println(format.AppID, format.NativeFormatVersion)
		return nil
	case "schema":
		fmt.Println(format.SchemaVersion)
		return nil
	case "plan":
		return runPlan(args[1:])
	case "encrypt":
		return runEncrypt(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage() {
	fmt.Println("fg - FoldersGuard")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  fg version")
	fmt.Println("  fg schema")
	fmt.Println("  fg plan <folder> <max-part-size-bytes>")
	fmt.Println("  FG_PASSWORD=<password> fg encrypt <folder> <content-output-folder> <database-output.fg> <max-part-size-bytes>")
}

func runPlan(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: fg plan <folder> <max-part-size-bytes>")
	}
	maxPartSize, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("parse max part size: %w", err)
	}

	scan, err := fswalk.ScanTopFolder(args[0])
	if err != nil {
		return err
	}
	plan, err := project.Planner{MaxPartSize: maxPartSize}.Plan(scan)
	if err != nil {
		return err
	}

	fmt.Printf("project_id=%s\n", plan.Project.ID)
	fmt.Printf("root_folder_id=%s\n", plan.Project.RootFolderID)
	fmt.Printf("items=%d\n", len(plan.Items)+1)
	fmt.Printf("folders=%d\n", len(plan.Folders)+1)
	fmt.Printf("files=%d\n", len(plan.Files))
	fmt.Printf("parts=%d\n", len(plan.Parts))
	fmt.Printf("storage_objects=%d\n", len(plan.StorageObjects))
	fmt.Printf("skipped=%d\n", len(scan.Skipped))
	for _, skipped := range scan.Skipped {
		fmt.Printf("skipped_entry=%s reason=%s\n", skipped.RootRelativePath, skipped.Reason)
	}
	return nil
}

func runEncrypt(args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("usage: FG_PASSWORD=<password> fg encrypt <folder> <content-output-folder> <database-output.fg> <max-part-size-bytes>")
	}
	password := os.Getenv("FG_PASSWORD")
	if password == "" {
		return fmt.Errorf("FG_PASSWORD is required")
	}

	sourceFolder := args[0]
	contentOutput := args[1]
	databaseOutput := args[2]
	if err := validateOutputOutsideSource(sourceFolder, contentOutput); err != nil {
		return err
	}
	if err := validateOutputOutsideSource(sourceFolder, databaseOutput); err != nil {
		return err
	}
	if !format.IsProjectExtension(databaseOutput) {
		return fmt.Errorf("database output must use %s extension", format.ProjectExtension)
	}
	maxPartSize, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return fmt.Errorf("parse max part size: %w", err)
	}
	if maxPartSize <= 0 {
		return fmt.Errorf("max part size must be positive")
	}

	ctx := context.Background()
	scan, err := fswalk.ScanTopFolder(sourceFolder)
	if err != nil {
		return err
	}
	plan, err := project.Planner{MaxPartSize: maxPartSize}.Plan(scan)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(contentOutput, 0o755); err != nil {
		return fmt.Errorf("create content output folder: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(databaseOutput), 0o755); err != nil {
		return fmt.Errorf("create database output folder: %w", err)
	}

	if err := writeProjectDatabase(ctx, db.Config{
		Path:       databaseOutput,
		DriverName: db.SQLCipherDriver,
		Password:   password,
	}, plan); err != nil {
		return err
	}
	if err := (project.Executor{OutputRoot: contentOutput}).EncryptContent(ctx, plan); err != nil {
		return err
	}

	fmt.Printf("project_id=%s\n", plan.Project.ID)
	fmt.Printf("content_output=%s\n", contentOutput)
	fmt.Printf("database_output=%s\n", databaseOutput)
	fmt.Printf("items=%d\n", len(plan.Items)+1)
	fmt.Printf("folders=%d\n", len(plan.Folders)+1)
	fmt.Printf("files=%d\n", len(plan.Files))
	fmt.Printf("parts=%d\n", len(plan.Parts))
	fmt.Printf("storage_objects=%d\n", len(plan.StorageObjects))
	fmt.Printf("skipped=%d\n", len(scan.Skipped))
	for _, skipped := range scan.Skipped {
		fmt.Printf("skipped_entry=%s reason=%s\n", skipped.RootRelativePath, skipped.Reason)
	}
	return nil
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

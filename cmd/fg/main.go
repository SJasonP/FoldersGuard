package main

import (
	"fmt"
	"os"
	"strconv"

	"foldersguard/internal/format"
	"foldersguard/internal/fswalk"
	"foldersguard/internal/project"
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

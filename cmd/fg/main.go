package main

import (
	"fmt"
	"os"

	"foldersguard/internal/format"
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
}

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"foldersguard/internal/cli"
)

func main() {
	name := filepath.Base(os.Args[0])
	if err := cli.RunWithIO(name, os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
		os.Exit(1)
	}
}

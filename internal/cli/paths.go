package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

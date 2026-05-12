package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ValidateDatabasePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat database: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("database path is a directory")
	}
	return nil
}

func ValidateExistingDirectory(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s folder: %w", label, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s must be a directory", label)
	}
	return nil
}

func ValidateExistingFile(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s file: %w", label, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s must be a file", label)
	}
	return nil
}

func PrepareContentOutput(path string, force bool) error {
	return PrepareDirectoryOutput(path, force, "content output")
}

func PrepareDirectoryOutput(path string, force bool, label string) error {
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

func PrepareFileOutput(path string, force bool) error {
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

func ValidateOutputOutsideSource(source, output string) error {
	sourceAbs, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("resolve source path: %w", err)
	}
	outputAbs, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}
	outputInsideSource, err := pathContainsOrEquals(sourceAbs, outputAbs)
	if err != nil {
		return err
	}
	if outputInsideSource {
		return fmt.Errorf("output path must be outside the source folder")
	}
	sourceInsideOutput, err := pathContainsOrEquals(outputAbs, sourceAbs)
	if err != nil {
		return err
	}
	if sourceInsideOutput {
		return fmt.Errorf("output path must not contain the source folder")
	}
	return nil
}

func ValidateDistinctPaths(left, right string) error {
	leftAbs, err := filepath.Abs(left)
	if err != nil {
		return fmt.Errorf("resolve source path: %w", err)
	}
	rightAbs, err := filepath.Abs(right)
	if err != nil {
		return fmt.Errorf("resolve target path: %w", err)
	}
	if leftAbs == rightAbs {
		return fmt.Errorf("source and target paths must be different")
	}
	return nil
}

func pathContainsOrEquals(parent, child string) (bool, error) {
	relative, err := filepath.Rel(parent, child)
	if err != nil {
		return false, fmt.Errorf("compare paths: %w", err)
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))), nil
}

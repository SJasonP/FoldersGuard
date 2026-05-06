package fswalk

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"foldersguard/internal/fsmeta"
)

type EntryType string

const (
	EntryTypeFile   EntryType = "file"
	EntryTypeFolder EntryType = "folder"
)

type Entry struct {
	RootRelativePath string
	AbsolutePath     string
	Type             EntryType
	Size             int64
	Metadata         fsmeta.Metadata
}

type ScanResult struct {
	Root    Entry
	Entries []Entry
}

func ScanTopFolder(root string) (ScanResult, error) {
	if root == "" {
		return ScanResult{}, errors.New("root path is required")
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return ScanResult{}, fmt.Errorf("resolve root: %w", err)
	}

	info, err := os.Lstat(absRoot)
	if err != nil {
		return ScanResult{}, fmt.Errorf("stat root: %w", err)
	}
	if !info.IsDir() {
		return ScanResult{}, fmt.Errorf("root must be a directory")
	}
	if !isRegularDir(info) {
		return ScanResult{}, fmt.Errorf("root must be a regular directory")
	}

	result := ScanResult{
		Root: Entry{
			RootRelativePath: ".",
			AbsolutePath:     absRoot,
			Type:             EntryTypeFolder,
		},
	}
	rootMetadata, err := fsmeta.Capture(absRoot, info)
	if err != nil {
		return ScanResult{}, err
	}
	result.Root.Metadata = rootMetadata

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path == absRoot {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		mode := info.Mode()
		metadata, err := fsmeta.Capture(path, info)
		if err != nil {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		switch {
		case mode.Type() == 0:
			result.Entries = append(result.Entries, Entry{
				RootRelativePath: rel(absRoot, path),
				AbsolutePath:     path,
				Type:             EntryTypeFile,
				Size:             info.Size(),
				Metadata:         metadata,
			})
		case isRegularDir(info):
			result.Entries = append(result.Entries, Entry{
				RootRelativePath: rel(absRoot, path),
				AbsolutePath:     path,
				Type:             EntryTypeFolder,
				Metadata:         metadata,
			})
		default:
			if d.IsDir() {
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return ScanResult{}, fmt.Errorf("scan root: %w", err)
	}

	sort.Slice(result.Entries, func(i, j int) bool {
		return result.Entries[i].RootRelativePath < result.Entries[j].RootRelativePath
	})

	return result, nil
}

func ScanPath(path string) (ScanResult, error) {
	if path == "" {
		return ScanResult{}, errors.New("path is required")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return ScanResult{}, fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Lstat(absPath)
	if err != nil {
		return ScanResult{}, fmt.Errorf("stat path: %w", err)
	}
	if info.Mode().Type() == 0 {
		metadata, err := fsmeta.Capture(absPath, info)
		if err != nil {
			return ScanResult{}, err
		}
		return ScanResult{
			Root: Entry{
				RootRelativePath: ".",
				AbsolutePath:     absPath,
				Type:             EntryTypeFile,
				Size:             info.Size(),
				Metadata:         metadata,
			},
		}, nil
	}
	if info.IsDir() {
		return ScanTopFolder(absPath)
	}
	return ScanResult{}, fmt.Errorf("path is unsupported: %s", unsupportedReason(info.Mode()))
}

func isRegularDir(info fs.FileInfo) bool {
	return info.IsDir() && info.Mode().Type() == os.ModeDir
}

func rel(root, path string) string {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(relative)
}

func unsupportedReason(mode fs.FileMode) string {
	switch {
	case mode&os.ModeSymlink != 0:
		return "symlink unsupported"
	case mode&os.ModeSocket != 0:
		return "socket unsupported"
	case mode&os.ModeDevice != 0:
		return "device unsupported"
	case mode&os.ModeNamedPipe != 0:
		return "fifo unsupported"
	case mode&os.ModeIrregular != 0:
		return "irregular file unsupported"
	default:
		return "special file unsupported"
	}
}

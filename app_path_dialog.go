package main

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type PathDialogFilter struct {
	DisplayName string `json:"displayName"`
	Pattern     string `json:"pattern"`
}

type SelectPathRequest struct {
	Kind             string             `json:"kind"`
	Title            string             `json:"title"`
	DefaultDirectory string             `json:"defaultDirectory"`
	DefaultFilename  string             `json:"defaultFilename"`
	Filters          []PathDialogFilter `json:"filters"`
}

func (a *App) SelectPath(request SelectPathRequest) (string, error) {
	filters := make([]runtime.FileFilter, 0, len(request.Filters))
	for _, filter := range request.Filters {
		filters = append(filters, runtime.FileFilter{
			DisplayName: filter.DisplayName,
			Pattern:     filter.Pattern,
		})
	}

	switch request.Kind {
	case "open-directory":
		return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
			Title:            request.Title,
			DefaultDirectory: request.DefaultDirectory,
		})
	case "open-file":
		return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
			Title:            request.Title,
			DefaultDirectory: request.DefaultDirectory,
			DefaultFilename:  request.DefaultFilename,
			Filters:          filters,
		})
	case "save-file":
		return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
			Title:                request.Title,
			DefaultDirectory:     request.DefaultDirectory,
			DefaultFilename:      request.DefaultFilename,
			Filters:              filters,
			CanCreateDirectories: true,
		})
	default:
		return "", fmt.Errorf("unsupported path dialog kind %q", request.Kind)
	}
}

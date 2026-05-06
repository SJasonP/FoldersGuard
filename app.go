package main

import (
	"context"
	"time"

	"foldersguard/internal/app"
	"foldersguard/internal/format"
)

type App struct {
	ctx     context.Context
	service app.Service
}

type AppInfo struct {
	ProductName         string `json:"productName"`
	AppID               string `json:"appId"`
	NativeFormatVersion string `json:"nativeFormatVersion"`
	SchemaVersion       int    `json:"schemaVersion"`
	DataDir             string `json:"dataDir"`
	CLIExecutableName   string `json:"cliExecutableName"`
	CLIShortAlias       string `json:"cliShortAlias"`
}

type LocalProjectSummary struct {
	ProjectID          string `json:"projectId"`
	FileName           string `json:"fileName"`
	ModifiedAt         string `json:"modifiedAt"`
	AvailabilityStatus string `json:"availabilityStatus"`
}

func NewApp() (*App, error) {
	service, err := app.NewService("")
	if err != nil {
		return nil, err
	}
	if err := service.EnsureDataDir(); err != nil {
		return nil, err
	}
	return &App{service: service}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) AppInfo() AppInfo {
	return AppInfo{
		ProductName:         "FoldersGuard",
		AppID:               format.AppID,
		NativeFormatVersion: format.NativeFormatVersion,
		SchemaVersion:       format.SchemaVersion,
		DataDir:             a.service.DataDir,
		CLIExecutableName:   "foldersguard",
		CLIShortAlias:       "fg",
	}
}

func (a *App) ListLocalProjects() ([]LocalProjectSummary, error) {
	projects, err := a.service.ListActiveProjects()
	if err != nil {
		return nil, err
	}

	result := make([]LocalProjectSummary, 0, len(projects))
	for _, project := range projects {
		modifiedAt := ""
		if !project.ModifiedAt.IsZero() {
			modifiedAt = project.ModifiedAt.Format(time.RFC3339)
		}
		result = append(result, LocalProjectSummary{
			ProjectID:          project.ProjectID,
			FileName:           project.FileName,
			ModifiedAt:         modifiedAt,
			AvailabilityStatus: project.Availability,
		})
	}
	return result, nil
}

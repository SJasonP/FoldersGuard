package main

import (
	"context"

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

func NewApp() (*App, error) {
	service, err := app.NewService("")
	if err != nil {
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

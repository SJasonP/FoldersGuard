package main

import "foldersguard/internal/format"

func (a *App) AppInfo() AppInfo {
	startupError := ""
	if a.startupError != nil {
		startupError = a.startupError.Error()
	}
	return AppInfo{
		ProductName:         "FoldersGuard",
		AppID:               format.AppID,
		NativeFormatVersion: format.NativeFormatVersion,
		SchemaVersion:       format.SchemaVersion,
		DataDir:             a.service.DataDir,
		StartupError:        startupError,
		CopyrightNotice:     "Copyright (c) 2026 SJasonP",
		ProjectLink:         "https://github.com/SJasonP/FoldersGuard",
	}
}

package main

import "foldersguard/internal/format"

func (a *App) AppInfo() AppInfo {
	return AppInfo{
		ProductName:         "FoldersGuard",
		AppID:               format.AppID,
		NativeFormatVersion: format.NativeFormatVersion,
		SchemaVersion:       format.SchemaVersion,
		DataDir:             a.service.DataDir,
		CopyrightNotice:     "Copyright (c) 2026 SJasonP",
		ProjectLink:         "https://github.com/SJasonP/FoldersGuard",
	}
}

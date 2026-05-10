package main

import "foldersguard/internal/format"

func (a *App) AppInfo() AppInfo {
	startupError := ""
	if a.startupError != nil {
		startupError = a.startupError.Error()
	}
	return AppInfo{
		ProductName:     "FoldersGuard",
		ProductVersion:  format.ProductVersion,
		AppID:           format.AppID,
		FormatVersion:   format.FormatVersion,
		DataDir:         a.service.DataDir,
		StartupError:    startupError,
		CopyrightNotice: "Copyright (C) 2026 SJasonP",
		ProjectLink:     "https://github.com/SJasonP/FoldersGuard",
		ThirdPartyLink:  "https://github.com/SJasonP/FoldersGuard/blob/main/THIRD-PARTY-NOTICES.md",
	}
}

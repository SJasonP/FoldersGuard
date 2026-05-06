package main

import "foldersguard/internal/app"

func (a *App) ReadSettings() (Settings, error) {
	settings, err := a.service.ReadSettings()
	if err != nil {
		return Settings{}, err
	}
	return mapSettings(settings), nil
}

func (a *App) SaveSettings(settings Settings) (Settings, error) {
	saved, err := a.service.SaveSettings(app.Settings{
		OperationGuideFormat:   settings.OperationGuideFormat,
		DefaultMaxPartSize:     settings.DefaultMaxPartSize,
		SourceCleanupMode:      settings.SourceCleanupMode,
		RememberRecentPaths:    settings.RememberRecentPaths,
		RecentPaths:            append([]string(nil), settings.RecentPaths...),
		WindowStatePersistence: settings.WindowStatePersistence,
		Theme:                  settings.Theme,
		Language:               settings.Language,
	})
	if err != nil {
		return Settings{}, err
	}
	return mapSettings(saved), nil
}

func (a *App) ClearRecentPaths() (Settings, error) {
	settings, err := a.service.ClearRecentPaths()
	if err != nil {
		return Settings{}, err
	}
	return mapSettings(settings), nil
}

func mapSettings(settings app.Settings) Settings {
	return Settings{
		OperationGuideFormat:   settings.OperationGuideFormat,
		DefaultMaxPartSize:     settings.DefaultMaxPartSize,
		SourceCleanupMode:      settings.SourceCleanupMode,
		RememberRecentPaths:    settings.RememberRecentPaths,
		RecentPaths:            append([]string(nil), settings.RecentPaths...),
		WindowStatePersistence: settings.WindowStatePersistence,
		Theme:                  settings.Theme,
		Language:               settings.Language,
	}
}

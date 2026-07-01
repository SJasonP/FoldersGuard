package main

import "foldersguard/internal/app"

func (a *App) ReadSettings() (Settings, error) {
	settings, err := a.service.ReadSettings()
	if err != nil {
		return Settings{}, frontendError(err)
	}
	return mapSettings(settings), nil
}

func (a *App) SaveSettings(settings Settings) (Settings, error) {
	saved, err := a.service.SaveSettings(app.Settings{
		DefaultMaxPartSize: settings.DefaultMaxPartSize,
		SourceCleanupMode:  settings.SourceCleanupMode,
		NoiseFileHandling:  settings.NoiseFileHandling,
		Theme:              settings.Theme,
		Language:           settings.Language,
		BackupRetention:    settings.BackupRetention,
		FailureHandling:    settings.FailureHandling,
	})
	if err != nil {
		return Settings{}, frontendError(err)
	}
	return mapSettings(saved), nil
}

func mapSettings(settings app.Settings) Settings {
	return Settings{
		DefaultMaxPartSize: settings.DefaultMaxPartSize,
		SourceCleanupMode:  settings.SourceCleanupMode,
		NoiseFileHandling:  settings.NoiseFileHandling,
		Theme:              settings.Theme,
		Language:           settings.Language,
		BackupRetention:    settings.BackupRetention,
		FailureHandling:    settings.FailureHandling,
	}
}

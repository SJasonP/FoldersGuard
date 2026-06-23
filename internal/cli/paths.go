package cli

import (
	"foldersguard/internal/app"
)

func prepareContentOutput(path string, force bool) error {
	noiseMode, err := readNoiseFileHandling()
	if err != nil {
		return err
	}
	return app.PrepareDirectoryOutputWithNoiseMode(path, force, "content output", noiseMode)
}

func prepareDirectoryOutput(path string, force bool, label string) error {
	noiseMode, err := readNoiseFileHandling()
	if err != nil {
		return err
	}
	return app.PrepareDirectoryOutputWithNoiseMode(path, force, label, noiseMode)
}

func validateExistingDirectory(path, label string) error {
	return app.ValidateExistingDirectory(path, label)
}

func validateExistingFile(path, label string) error {
	return app.ValidateExistingFile(path, label)
}

func prepareFileOutput(path string, force bool) error {
	return app.PrepareFileOutput(path, force)
}

func validateOutputOutsideSource(source, output string) error {
	return app.ValidateOutputOutsideSource(source, output)
}

func validateDistinctPaths(left, right string) error {
	return app.ValidateDistinctPaths(left, right)
}

package main

import (
	"errors"
	"fmt"

	"foldersguard/internal/app"
	fgdb "foldersguard/internal/db"
)

const (
	errorCodeInvalidPassword      = "FG_INVALID_PASSWORD"
	errorCodeOutputFolderNotEmpty = "FG_OUTPUT_FOLDER_NOT_EMPTY"
	errorCodeOutputInsideSource   = "FG_OUTPUT_INSIDE_SOURCE"
	errorCodeOutputContainsSource = "FG_OUTPUT_CONTAINS_SOURCE"
	errorCodeSourceTargetSame     = "FG_SOURCE_TARGET_SAME"
)

func frontendError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, fgdb.ErrInvalidDatabasePassword) {
		return codedError(errorCodeInvalidPassword, err)
	}
	if errors.Is(err, app.ErrOutputFolderNotEmpty) {
		return codedError(errorCodeOutputFolderNotEmpty, err)
	}
	if errors.Is(err, app.ErrOutputInsideSource) {
		return codedError(errorCodeOutputInsideSource, err)
	}
	if errors.Is(err, app.ErrOutputContainsSource) {
		return codedError(errorCodeOutputContainsSource, err)
	}
	if errors.Is(err, app.ErrSourceTargetSame) {
		return codedError(errorCodeSourceTargetSame, err)
	}
	return err
}

func codedError(code string, err error) error {
	return fmt.Errorf("%s: %w", code, err)
}

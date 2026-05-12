package app

import "errors"

var (
	ErrOutputFolderNotEmpty = errors.New("output folder is not empty")
	ErrOutputInsideSource   = errors.New("output path is inside the source folder")
	ErrOutputContainsSource = errors.New("output path contains the source folder")
	ErrSourceTargetSame     = errors.New("source and target paths are the same")
)

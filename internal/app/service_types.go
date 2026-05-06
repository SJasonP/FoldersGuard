package app

import (
	"time"
)

type DatabaseOpen struct {
	ProjectRef string
	Password   string
}

type ShareOpen struct {
	DatabasePath string
	Password     string
}

type InspectResult struct {
	ProjectID      string
	DatabaseType   string
	RootFolderID   string
	RootName       string
	FormatVersion  string
	SchemaVersion  string
	Items          int
	Folders        int
	Files          int
	Parts          int
	StorageObjects int
}

type VerifyResult struct {
	ProjectID       string
	CheckedObjects  int
	MissingObjects  int
	TamperedObjects int
	ExtraObjects    int
	Status          string
}

type ActiveProjectSummary struct {
	ProjectID    string
	FileName     string
	ModifiedAt   time.Time
	Availability string
}

type ExportProjectInput struct {
	ProjectID  string
	Password   string
	OutputPath string
	Force      bool
}

type ExportProjectResult struct {
	ProjectID  string
	OutputPath string
}

type DeleteProjectInput struct {
	ProjectID string
	Password  string
}

type DeleteProjectResult struct {
	ProjectID string
}

type ImportProjectInput struct {
	InputPath string
	Password  string
	Force     bool
}

type ImportProjectResult struct {
	ProjectID string
}

type CreateProjectInput struct {
	SourcePath     string
	ContentOutput  string
	Password       string
	MaxPartSize    int64
	Force          bool
	SourceCleanup  string
	DatabaseExport string
}

type CreateProjectResult struct {
	ProjectID               string
	ProjectName             string
	ContentOutput           string
	DatabaseExport          string
	EncryptedFiles          int
	EncryptedFolders        int
	EncryptedParts          int
	DeletedCleartextFiles   int
	DeletedCleartextFolders int
	FailedFiles             int
}

type ShareSummary struct {
	ShareID           string
	DatabaseType      string
	FormatVersion     string
	SchemaVersion     string
	TopLevelItems     int
	Files             int
	Folders           int
	Parts             int
	StorageObjects    int
	PasswordProtected bool
}

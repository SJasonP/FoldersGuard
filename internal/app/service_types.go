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

type DecryptShareInput struct {
	DatabasePath  string
	Password      string
	EncryptedRoot string
	OutputRoot    string
	Force         bool
	SourceCleanup string
}

type DecryptShareResult struct {
	ShareID               string
	OutputRoot            string
	DecryptedFiles        int
	RestoredFolders       int
	SkippedFolders        int
	DeletedEncryptedFiles int
	FailedEncryptedFiles  int
}

type DecryptProjectInput struct {
	ProjectID     string
	Password      string
	EncryptedRoot string
	OutputRoot    string
	Force         bool
	SourceCleanup string
}

type DecryptProjectResult struct {
	ProjectID             string
	OutputRoot            string
	DecryptedFiles        int
	RestoredFolders       int
	SkippedFolders        int
	DeletedEncryptedFiles int
	FailedEncryptedFiles  int
}

type InspectResult struct {
	ProjectID      string
	DatabaseType   string
	ProjectName    string
	RootFolderID   string
	RootName       string
	FormatVersion  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
	MissingPaths    []string
	TamperedPaths   []string
	ExtraPaths      []string
	Status          string
}

type ActiveProjectSummary struct {
	ProjectID    string
	ProjectName  string
	FileName     string
	ModifiedAt   time.Time
	Availability string
}

type SaveLocalProjectNameInput struct {
	ProjectID   string
	ProjectName string
}

type SaveLocalProjectNameResult struct {
	ProjectID   string
	ProjectName string
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
	TopLevelItems     int
	Files             int
	Folders           int
	Parts             int
	StorageObjects    int
	PasswordProtected bool
}

type ShareableItem struct {
	ID         string
	ParentID   string
	Path       string
	ParentPath string
	Name       string
	Type       string
	Size       int64
	ChildCount int
	ModifiedAt time.Time
}

type CreateShareInput struct {
	ProjectID         string
	ProjectPassword   string
	ItemPaths         []string
	OutputPath        string
	Force             bool
	PasswordProtected bool
	SharePassword     string
}

type ShareContentLocation struct {
	SourcePath string
	TargetPath string
}

type CreateShareResult struct {
	ProjectID         string
	ShareID           string
	OutputPath        string
	TopLevelItems     int
	Files             int
	Folders           int
	Parts             int
	PasswordProtected bool
	ContentLocations  []ShareContentLocation
}

type OpenProjectBrowserInput struct {
	ProjectID     string
	Password      string
	EncryptedRoot string
}

type ProjectBrowserItem struct {
	ID               string
	ParentID         string
	Path             string
	ParentPath       string
	Name             string
	Type             string
	Size             int64
	ChildCount       int
	ModifiedAt       time.Time
	MetadataCaptured bool
	ContentAvailable bool
}

type ProjectBrowserState struct {
	ProjectID        string
	ProjectName      string
	RootFolderID     string
	RootFolderName   string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Files            int
	Folders          int
	Parts            int
	ContentConnected bool
	EncryptedRoot    string
	Items            []ProjectBrowserItem
}

type ProjectRenameChange struct {
	ItemPath string
	NewName  string
}

type ProjectMoveChange struct {
	ItemPath         string
	TargetFolderPath string
}

type ProjectRemoveChange struct {
	ItemPath string
}

type ProjectAddChange struct {
	SourcePath       string
	TargetFolderPath string
	MaxPartSize      int64
}

type ProjectCreateFolderChange struct {
	TargetFolderPath string
	Name             string
}

type ProjectContentOperation struct {
	Type       string
	SourcePath string
	TargetPath string
}

type ApplyProjectChangesInput struct {
	ProjectID           string
	Password            string
	EncryptedRoot       string
	RenameChanges       []ProjectRenameChange
	MoveChanges         []ProjectMoveChange
	RemoveChanges       []ProjectRemoveChange
	AddChanges          []ProjectAddChange
	CreateFolderChanges []ProjectCreateFolderChange
}

type ApplyProjectChangesResult struct {
	ProjectID              string
	AppliedRenames         int
	AppliedMoves           int
	AppliedRemoves         int
	AppliedAdds            int
	AppliedCreatedFolders  int
	ManualContentGuide     bool
	StagedContentPath      string
	StagedContentName      string
	StagedContentOnDesktop bool
	ContentOperations      []ProjectContentOperation
	AppliedContentChanges  []ProjectContentOperation
	BrowserState           ProjectBrowserState
}

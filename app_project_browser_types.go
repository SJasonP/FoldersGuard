package main

type OpenProjectBrowserRequest struct {
	ProjectID     string `json:"projectId"`
	Password      string `json:"password"`
	EncryptedPath string `json:"encryptedPath"`
}

type ProjectBrowserItem struct {
	ID               string `json:"id"`
	ParentID         string `json:"parentId"`
	Path             string `json:"path"`
	ParentPath       string `json:"parentPath"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	Size             int64  `json:"size"`
	ChildCount       int    `json:"childCount"`
	ModifiedAt       string `json:"modifiedAt"`
	MetadataCaptured bool   `json:"metadataCaptured"`
	ContentAvailable bool   `json:"contentAvailable"`
}

type ProjectBrowserState struct {
	ProjectID        string               `json:"projectId"`
	ProjectName      string               `json:"projectName"`
	RootFolderID     string               `json:"rootFolderId"`
	RootFolderName   string               `json:"rootFolderName"`
	CreatedAt        string               `json:"createdAt"`
	UpdatedAt        string               `json:"updatedAt"`
	Files            int                  `json:"files"`
	Folders          int                  `json:"folders"`
	Parts            int                  `json:"parts"`
	ContentConnected bool                 `json:"contentConnected"`
	EncryptedPath    string               `json:"encryptedPath"`
	Items            []ProjectBrowserItem `json:"items"`
}

type ProjectRenameChange struct {
	ItemPath string `json:"itemPath"`
	NewName  string `json:"newName"`
}

type ProjectMoveChange struct {
	ItemPath         string `json:"itemPath"`
	TargetFolderPath string `json:"targetFolderPath"`
}

type ProjectRemoveChange struct {
	ItemPath string `json:"itemPath"`
}

type ProjectAddChange struct {
	SourcePath       string `json:"sourcePath"`
	TargetFolderPath string `json:"targetFolderPath"`
	MaxPartSize      int64  `json:"maxPartSize"`
}

type ProjectContentOperation struct {
	Type       string `json:"type"`
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
}

type ApplyProjectChangesRequest struct {
	ProjectID     string                `json:"projectId"`
	Password      string                `json:"password"`
	EncryptedPath string                `json:"encryptedPath"`
	RenameChanges []ProjectRenameChange `json:"renameChanges"`
	MoveChanges   []ProjectMoveChange   `json:"moveChanges"`
	RemoveChanges []ProjectRemoveChange `json:"removeChanges"`
	AddChanges    []ProjectAddChange    `json:"addChanges"`
}

type ApplyProjectChangesResult struct {
	ProjectID             string                    `json:"projectId"`
	AppliedRenames        int                       `json:"appliedRenames"`
	AppliedMoves          int                       `json:"appliedMoves"`
	AppliedRemoves        int                       `json:"appliedRemoves"`
	AppliedAdds           int                       `json:"appliedAdds"`
	OperationGuidePath    string                    `json:"operationGuidePath"`
	StagedContentPath     string                    `json:"stagedContentPath"`
	ContentOperations     []ProjectContentOperation `json:"contentOperations"`
	AppliedContentChanges []ProjectContentOperation `json:"appliedContentChanges"`
	BrowserState          ProjectBrowserState       `json:"browserState"`
}

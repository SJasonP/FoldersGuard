package main

type ShareableItem struct {
	ID         string `json:"id"`
	ParentID   string `json:"parentId"`
	Path       string `json:"path"`
	ParentPath string `json:"parentPath"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Size       int64  `json:"size"`
	ChildCount int    `json:"childCount"`
	ModifiedAt string `json:"modifiedAt"`
}

type ListShareableItemsRequest struct {
	ProjectID string `json:"projectId"`
	Password  string `json:"password"`
}

type CreateShareRequest struct {
	ProjectID         string   `json:"projectId"`
	ProjectPassword   string   `json:"projectPassword"`
	ItemPaths         []string `json:"itemPaths"`
	OutputPath        string   `json:"outputPath"`
	Force             bool     `json:"force"`
	PasswordProtected bool     `json:"passwordProtected"`
	SharePassword     string   `json:"sharePassword"`
}

type ShareContentLocation struct {
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
}

type CreateShareResult struct {
	ProjectID         string                 `json:"projectId"`
	ShareID           string                 `json:"shareId"`
	OutputPath        string                 `json:"outputPath"`
	TopLevelItems     int                    `json:"topLevelItems"`
	Files             int                    `json:"files"`
	Folders           int                    `json:"folders"`
	Parts             int                    `json:"parts"`
	PasswordProtected bool                   `json:"passwordProtected"`
	ContentLocations  []ShareContentLocation `json:"contentLocations"`
}

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

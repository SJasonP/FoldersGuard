package main

type AppInfo struct {
	ProductName         string `json:"productName"`
	AppID               string `json:"appId"`
	NativeFormatVersion string `json:"nativeFormatVersion"`
	SchemaVersion       int    `json:"schemaVersion"`
	DataDir             string `json:"dataDir"`
	CLIExecutableName   string `json:"cliExecutableName"`
	CLIShortAlias       string `json:"cliShortAlias"`
}

type LocalProjectSummary struct {
	ProjectID          string `json:"projectId"`
	FileName           string `json:"fileName"`
	ModifiedAt         string `json:"modifiedAt"`
	AvailabilityStatus string `json:"availabilityStatus"`
}

type Settings struct {
	OperationGuideFormat   string   `json:"operationGuideFormat"`
	DefaultMaxPartSize     int64    `json:"defaultMaxPartSize"`
	SourceCleanupMode      string   `json:"sourceCleanupMode"`
	RememberRecentPaths    bool     `json:"rememberRecentPaths"`
	RecentPaths            []string `json:"recentPaths"`
	WindowStatePersistence bool     `json:"windowStatePersistence"`
	Theme                  string   `json:"theme"`
	Language               string   `json:"language"`
}

type InspectProjectRequest struct {
	ProjectID string `json:"projectId"`
	Password  string `json:"password"`
}

type InspectProjectResult struct {
	ProjectID      string `json:"projectId"`
	DatabaseType   string `json:"databaseType"`
	RootFolderID   string `json:"rootFolderId"`
	RootName       string `json:"rootName"`
	FormatVersion  string `json:"formatVersion"`
	SchemaVersion  string `json:"schemaVersion"`
	Items          int    `json:"items"`
	Folders        int    `json:"folders"`
	Files          int    `json:"files"`
	Parts          int    `json:"parts"`
	StorageObjects int    `json:"storageObjects"`
}

type ExportProjectRequest struct {
	ProjectID  string `json:"projectId"`
	Password   string `json:"password"`
	OutputPath string `json:"outputPath"`
	Force      bool   `json:"force"`
}

type ExportProjectResult struct {
	ProjectID  string `json:"projectId"`
	OutputPath string `json:"outputPath"`
}

type DeleteProjectRequest struct {
	ProjectID string `json:"projectId"`
	Password  string `json:"password"`
}

type DeleteProjectResult struct {
	ProjectID string `json:"projectId"`
}

type ImportProjectRequest struct {
	InputPath string `json:"inputPath"`
	Password  string `json:"password"`
	Force     bool   `json:"force"`
}

type ImportProjectResult struct {
	ProjectID string `json:"projectId"`
}

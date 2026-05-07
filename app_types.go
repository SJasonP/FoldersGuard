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

type ShareSummary struct {
	ShareID           string `json:"shareId"`
	DatabaseType      string `json:"databaseType"`
	FormatVersion     string `json:"formatVersion"`
	SchemaVersion     string `json:"schemaVersion"`
	TopLevelItems     int    `json:"topLevelItems"`
	Files             int    `json:"files"`
	Folders           int    `json:"folders"`
	Parts             int    `json:"parts"`
	StorageObjects    int    `json:"storageObjects"`
	PasswordProtected bool   `json:"passwordProtected"`
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

type VerifyProjectRequest struct {
	ProjectID     string `json:"projectId"`
	Password      string `json:"password"`
	EncryptedPath string `json:"encryptedPath"`
}

type VerifyProjectResult struct {
	ProjectID       string `json:"projectId"`
	CheckedObjects  int    `json:"checkedObjects"`
	MissingObjects  int    `json:"missingObjects"`
	TamperedObjects int    `json:"tamperedObjects"`
	ExtraObjects    int    `json:"extraObjects"`
	Status          string `json:"status"`
}

type LoadShareRequest struct {
	DatabasePath string `json:"databasePath"`
	Password     string `json:"password"`
}

type VerifyShareRequest struct {
	DatabasePath  string `json:"databasePath"`
	Password      string `json:"password"`
	EncryptedPath string `json:"encryptedPath"`
}

type DecryptShareRequest struct {
	DatabasePath  string `json:"databasePath"`
	Password      string `json:"password"`
	EncryptedPath string `json:"encryptedPath"`
	OutputPath    string `json:"outputPath"`
	Force         bool   `json:"force"`
	SourceCleanup string `json:"sourceCleanup"`
}

type DecryptShareResult struct {
	ShareID               string `json:"shareId"`
	OutputPath            string `json:"outputPath"`
	DecryptedFiles        int    `json:"decryptedFiles"`
	RestoredFolders       int    `json:"restoredFolders"`
	DeletedEncryptedFiles int    `json:"deletedEncryptedFiles"`
	FailedEncryptedFiles  int    `json:"failedEncryptedFiles"`
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

type CreateProjectRequest struct {
	SourcePath     string `json:"sourcePath"`
	ContentOutput  string `json:"contentOutput"`
	Password       string `json:"password"`
	MaxPartSize    int64  `json:"maxPartSize"`
	Force          bool   `json:"force"`
	SourceCleanup  string `json:"sourceCleanup"`
	DatabaseExport string `json:"databaseExport"`
}

type CreateProjectResult struct {
	ProjectID               string `json:"projectId"`
	ProjectName             string `json:"projectName"`
	ContentOutput           string `json:"contentOutput"`
	DatabaseExport          string `json:"databaseExport"`
	EncryptedFiles          int    `json:"encryptedFiles"`
	EncryptedFolders        int    `json:"encryptedFolders"`
	EncryptedParts          int    `json:"encryptedParts"`
	DeletedCleartextFiles   int    `json:"deletedCleartextFiles"`
	DeletedCleartextFolders int    `json:"deletedCleartextFolders"`
	FailedFiles             int    `json:"failedFiles"`
}

type ImportProjectRequest struct {
	InputPath string `json:"inputPath"`
	Password  string `json:"password"`
	Force     bool   `json:"force"`
}

type ImportProjectResult struct {
	ProjectID string `json:"projectId"`
}

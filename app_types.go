package main

type AppInfo struct {
	ProductName     string `json:"productName"`
	ProductVersion  string `json:"productVersion"`
	AppID           string `json:"appId"`
	FormatVersion   string `json:"formatVersion"`
	DataDir         string `json:"dataDir"`
	StartupError    string `json:"startupError"`
	CopyrightNotice string `json:"copyrightNotice"`
	ProjectLink     string `json:"projectLink"`
	ThirdPartyLink  string `json:"thirdPartyLink"`
}

type LocalProjectSummary struct {
	ProjectID          string `json:"projectId"`
	ProjectName        string `json:"projectName"`
	FileName           string `json:"fileName"`
	ModifiedAt         string `json:"modifiedAt"`
	AvailabilityStatus string `json:"availabilityStatus"`
}

type SaveLocalProjectNameRequest struct {
	ProjectID   string `json:"projectId"`
	ProjectName string `json:"projectName"`
}

type SaveLocalProjectNameResult struct {
	ProjectID   string `json:"projectId"`
	ProjectName string `json:"projectName"`
}

type ShareSummary struct {
	ShareID           string `json:"shareId"`
	DatabaseType      string `json:"databaseType"`
	FormatVersion     string `json:"formatVersion"`
	TopLevelItems     int    `json:"topLevelItems"`
	Files             int    `json:"files"`
	Folders           int    `json:"folders"`
	Parts             int    `json:"parts"`
	StorageObjects    int    `json:"storageObjects"`
	PasswordProtected bool   `json:"passwordProtected"`
}

type Settings struct {
	DefaultMaxPartSize int64  `json:"defaultMaxPartSize"`
	SourceCleanupMode  string `json:"sourceCleanupMode"`
	NoiseFileHandling  string `json:"noiseFileHandling"`
	Theme              string `json:"theme"`
	Language           string `json:"language"`
	BackupRetention    int    `json:"backupRetention"`
}

type ProjectBackupInfo struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Reason    string `json:"reason"`
	CreatedAt string `json:"createdAt"`
	Size      int64  `json:"size"`
}

type RestoreProjectBackupRequest struct {
	ProjectID string `json:"projectId"`
	BackupID  string `json:"backupId"`
	Force     bool   `json:"force"`
}

type RestoreProjectBackupResult struct {
	ProjectID string `json:"projectId"`
}

type InspectProjectRequest struct {
	ProjectID string `json:"projectId"`
	Password  string `json:"password"`
}

type InspectProjectResult struct {
	ProjectID      string `json:"projectId"`
	DatabaseType   string `json:"databaseType"`
	ProjectName    string `json:"projectName"`
	RootFolderID   string `json:"rootFolderId"`
	RootName       string `json:"rootName"`
	FormatVersion  string `json:"formatVersion"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
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
	ProjectID       string   `json:"projectId"`
	CheckedObjects  int      `json:"checkedObjects"`
	MissingObjects  int      `json:"missingObjects"`
	TamperedObjects int      `json:"tamperedObjects"`
	ExtraObjects    int      `json:"extraObjects"`
	MissingPaths    []string `json:"missingPaths"`
	TamperedPaths   []string `json:"tamperedPaths"`
	ExtraPaths      []string `json:"extraPaths"`
	Status          string   `json:"status"`
}

type DecryptProjectRequest struct {
	ProjectID     string `json:"projectId"`
	Password      string `json:"password"`
	EncryptedPath string `json:"encryptedPath"`
	OutputPath    string `json:"outputPath"`
	Force         bool   `json:"force"`
	SourceCleanup string `json:"sourceCleanup"`
}

type DecryptProjectResult struct {
	ProjectID             string `json:"projectId"`
	OutputPath            string `json:"outputPath"`
	DecryptedFiles        int    `json:"decryptedFiles"`
	RestoredFolders       int    `json:"restoredFolders"`
	SkippedFolders        int    `json:"skippedFolders"`
	DeletedEncryptedFiles int    `json:"deletedEncryptedFiles"`
	FailedEncryptedFiles  int    `json:"failedEncryptedFiles"`
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
	SkippedFolders        int    `json:"skippedFolders"`
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

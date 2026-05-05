package model

import (
	"time"

	"github.com/google/uuid"
)

type ItemType string

const (
	ItemTypeFile   ItemType = "file"
	ItemTypeFolder ItemType = "folder"
)

type StorageKind string

const (
	StorageKindSingle StorageKind = "single"
	StorageKindSplit  StorageKind = "split"
)

type Project struct {
	ID           uuid.UUID
	RootFolderID uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Item struct {
	ID          uuid.UUID
	ParentID    *uuid.UUID
	Type        ItemType
	VisibleName uuid.UUID
	RealName    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type File struct {
	ID               uuid.UUID
	Key              []byte
	SourcePath       string
	OriginalSize     int64
	ContentAlgorithm string
	StorageKind      StorageKind
}

type Folder struct {
	ID  uuid.UUID
	Key []byte
}

type Part struct {
	ID          uuid.UUID
	FileID      uuid.UUID
	Index       int
	VisibleName uuid.UUID
	Offset      int64
	Size        int64
	Integrity   []byte
}

type StorageObjectType string

const (
	StorageObjectTypeFile   StorageObjectType = "file"
	StorageObjectTypeFolder StorageObjectType = "folder"
	StorageObjectTypePart   StorageObjectType = "part"
)

type StorageObject struct {
	ID          uuid.UUID
	ItemID      uuid.UUID
	Type        StorageObjectType
	VisiblePath string
	Size        *int64
	Integrity   []byte
}

type PlannedProject struct {
	Project        Project
	RootItem       Item
	RootFolder     Folder
	Items          []Item
	Folders        []Folder
	Files          []File
	Parts          []Part
	StorageObjects []StorageObject
}

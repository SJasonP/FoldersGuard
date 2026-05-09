package storage

import "foldersguard/internal/format"

const Schema = `
CREATE TABLE IF NOT EXISTS meta (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS items (
	item_id TEXT PRIMARY KEY,
	parent_id TEXT,
	item_type TEXT NOT NULL CHECK (item_type IN ('file', 'folder')),
	visible_name TEXT NOT NULL,
	real_name TEXT NOT NULL,
	sort_name TEXT NOT NULL,
	original_mode INTEGER NOT NULL,
	original_mod_time TEXT NOT NULL,
	original_access_time TEXT,
	original_birth_time TEXT,
	windows_attributes INTEGER,
	metadata_capabilities TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	deleted_at TEXT,
	UNIQUE(parent_id, visible_name),
	UNIQUE(parent_id, real_name)
);

CREATE TABLE IF NOT EXISTS folders (
	folder_id TEXT PRIMARY KEY,
	folder_key BLOB NOT NULL,
	FOREIGN KEY(folder_id) REFERENCES items(item_id)
);

CREATE TABLE IF NOT EXISTS files (
	file_id TEXT PRIMARY KEY,
	file_key BLOB NOT NULL,
	original_size INTEGER NOT NULL,
	content_algorithm TEXT NOT NULL,
	storage_kind TEXT NOT NULL CHECK (storage_kind IN ('single', 'split')),
	FOREIGN KEY(file_id) REFERENCES items(item_id)
);

CREATE TABLE IF NOT EXISTS parts (
	part_id TEXT PRIMARY KEY,
	file_id TEXT NOT NULL,
	part_index INTEGER NOT NULL,
	visible_name TEXT NOT NULL,
	offset INTEGER NOT NULL,
	size INTEGER NOT NULL,
	integrity BLOB,
	UNIQUE(file_id, part_index),
	UNIQUE(file_id, visible_name),
	FOREIGN KEY(file_id) REFERENCES files(file_id)
);

CREATE TABLE IF NOT EXISTS storage_objects (
	object_id TEXT PRIMARY KEY,
	item_id TEXT NOT NULL,
	object_type TEXT NOT NULL CHECK (object_type IN ('file', 'folder', 'part')),
	visible_path TEXT NOT NULL,
	size INTEGER,
	integrity BLOB,
	UNIQUE(visible_path),
	FOREIGN KEY(item_id) REFERENCES items(item_id)
);

CREATE TABLE IF NOT EXISTS operation_plans (
	plan_id TEXT PRIMARY KEY,
	status TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS operation_steps (
	step_id TEXT PRIMARY KEY,
	plan_id TEXT NOT NULL,
	step_index INTEGER NOT NULL,
	operation_type TEXT NOT NULL,
	source_visible_path TEXT,
	target_visible_path TEXT,
	expected_integrity BLOB,
	UNIQUE(plan_id, step_index),
	FOREIGN KEY(plan_id) REFERENCES operation_plans(plan_id)
);

CREATE INDEX IF NOT EXISTS idx_items_parent ON items(parent_id);
CREATE INDEX IF NOT EXISTS idx_items_sort ON items(parent_id, sort_name);
CREATE INDEX IF NOT EXISTS idx_parts_file ON parts(file_id, part_index);
`

var RequiredMeta = map[string]string{
	"app_id":                format.AppID,
	"format_version":        format.FormatVersion,
	"database_type":         "",
	"project_id":            "",
	"root_folder_id":        "",
	"created_at":            "",
	"updated_at":            "",
	"crypto_suite":          format.CryptoSuite,
	"content_crypto_suite":  format.ContentAlgorithm,
	"database_crypto_suite": format.DatabaseAlgorithm,
}

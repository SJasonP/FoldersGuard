package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"foldersguard/internal/format"
)

type Store struct {
	db *sql.DB
}

type ProjectSpec struct {
	ProjectID       uuid.UUID
	RootFolderID    uuid.UUID
	RootVisibleName uuid.UUID
	RootRealName    string
	RootFolderKey   []byte
	DatabaseType    string
	CreatedAt       time.Time
}

func NewStore(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("db is required")
	}
	return &Store{db: db}, nil
}

func (s *Store) ApplySchema(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, Schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}

func (s *Store) InitProject(ctx context.Context, spec ProjectSpec) error {
	if err := validateProjectSpec(spec); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin init project: %w", err)
	}
	defer rollback(tx)

	if _, err := tx.ExecContext(ctx, Schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	createdAt := spec.CreatedAt.UTC().Format(time.RFC3339Nano)
	meta := map[string]string{
		"app_id":                format.AppID,
		"format_version":        format.NativeFormatVersion,
		"schema_version":        strconv.Itoa(format.SchemaVersion),
		"database_type":         spec.DatabaseType,
		"project_id":            spec.ProjectID.String(),
		"root_folder_id":        spec.RootFolderID.String(),
		"created_at":            createdAt,
		"updated_at":            createdAt,
		"crypto_suite":          format.CryptoSuite,
		"content_crypto_suite":  format.ContentAlgorithm,
		"database_crypto_suite": format.DatabaseAlgorithm,
	}
	if err := insertMeta(ctx, tx, meta); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit init project: %w", err)
	}
	return nil
}

func (s *Store) Meta(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM meta`)
	if err != nil {
		return nil, fmt.Errorf("query meta: %w", err)
	}
	defer rows.Close()

	meta := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan meta: %w", err)
		}
		meta[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meta: %w", err)
	}
	return meta, nil
}

func validateProjectSpec(spec ProjectSpec) error {
	if spec.ProjectID == uuid.Nil {
		return errors.New("project id is required")
	}
	if spec.RootFolderID == uuid.Nil {
		return errors.New("root folder id is required")
	}
	if spec.RootVisibleName == uuid.Nil {
		return errors.New("root visible name is required")
	}
	if strings.TrimSpace(spec.RootRealName) == "" {
		return errors.New("root real name is required")
	}
	if len(spec.RootFolderKey) != 32 {
		return errors.New("root folder key must be 32 bytes")
	}
	if spec.DatabaseType == "" {
		return errors.New("database type is required")
	}
	if spec.CreatedAt.IsZero() {
		return errors.New("created at is required")
	}
	return nil
}

func insertMeta(ctx context.Context, tx *sql.Tx, meta map[string]string) error {
	keys := make([]string, 0, len(meta))
	for key := range meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO meta (key, value) VALUES (?, ?)`,
			key,
			meta[key],
		); err != nil {
			return fmt.Errorf("insert meta %s: %w", key, err)
		}
	}
	return nil
}

func sortName(name string) string {
	return strings.ToLower(name)
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}

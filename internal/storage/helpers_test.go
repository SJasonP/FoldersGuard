package storage

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func openMemoryDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func bytesOf(value byte, size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = value
	}
	return data
}

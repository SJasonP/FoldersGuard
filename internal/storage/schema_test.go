package storage

import (
	"strings"
	"testing"
)

func TestSchemaDoesNotContainRemovedAccessRecords(t *testing.T) {
	if strings.Contains(Schema, "access_records") {
		t.Fatal("schema must not contain access_records")
	}
}

func TestRequiredMeta(t *testing.T) {
	for _, key := range []string{
		"app_id",
		"format_version",
		"schema_version",
		"crypto_suite",
		"content_crypto_suite",
		"database_crypto_suite",
	} {
		if RequiredMeta[key] == "" {
			t.Fatalf("RequiredMeta[%q] is empty", key)
		}
	}
}

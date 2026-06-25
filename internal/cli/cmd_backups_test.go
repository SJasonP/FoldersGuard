package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestBackupsListEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	const projectID = "11111111-1111-1111-1111-111111111111"
	var out bytes.Buffer
	if err := RunWithIO("fg", []string{"backups", "list", projectID}, nil, &out); err != nil {
		t.Fatalf("backups list: %v", err)
	}
	if !strings.Contains(out.String(), "project_id="+projectID+"\n") {
		t.Fatalf("output = %q", out.String())
	}
	if strings.Contains(out.String(), "backup_id=") {
		t.Fatalf("expected no backups, got %q", out.String())
	}
}

func TestBackupsRestoreUnknownFails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var out bytes.Buffer
	err := RunWithIO("fg", []string{"backups", "restore", "11111111-1111-1111-1111-111111111111", "nope", "--force"}, nil, &out)
	if err == nil {
		t.Fatal("expected restore of an unknown backup to fail")
	}
}

func TestBackupsRequiresSubcommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	var out bytes.Buffer
	if err := RunWithIO("fg", []string{"backups"}, nil, &out); err != nil {
		t.Fatalf("backups: %v", err)
	}
	if !strings.Contains(out.String(), "list") || !strings.Contains(out.String(), "restore") {
		t.Fatalf("backups help = %q, want list and restore subcommands", out.String())
	}
}

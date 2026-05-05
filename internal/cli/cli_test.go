package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpUsesInvokedName(t *testing.T) {
	var out bytes.Buffer
	if err := RunWithIO("foldersguard", []string{"help"}, nil, &out); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if !strings.Contains(output, "Usage:") {
		t.Fatalf("help output = %q, want Cobra usage", output)
	}
	if !strings.Contains(output, "foldersguard [flags]") {
		t.Fatalf("help output = %q, want foldersguard command examples", output)
	}
	if !strings.Contains(output, "decrypt") {
		t.Fatalf("help output = %q, want decrypt command", output)
	}
	if strings.Contains(output, "completion") {
		t.Fatalf("help output = %q, want no undocumented completion command", output)
	}
}

func TestVersionAndSchemaOutput(t *testing.T) {
	var version bytes.Buffer
	if err := RunWithIO("fg", []string{"version"}, nil, &version); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(version.String(), "app_id=com.SJasonP.FoldersGuard\n") {
		t.Fatalf("version output = %q", version.String())
	}
	if !strings.Contains(version.String(), "format_version=fg-native-v1\n") {
		t.Fatalf("version output = %q", version.String())
	}

	var schema bytes.Buffer
	if err := RunWithIO("fg", []string{"schema"}, nil, &schema); err != nil {
		t.Fatal(err)
	}
	if schema.String() != "schema_version=1\n" {
		t.Fatalf("schema output = %q", schema.String())
	}
}

func TestPlanEncryptDoesNotEmitDurableProjectIDs(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := RunWithIO("fg", []string{"plan", "encrypt", source, "--max-part-size", "1024"}, nil, &out); err != nil {
		t.Fatal(err)
	}
	output := out.String()
	if strings.Contains(output, "project_id=") || strings.Contains(output, "root_folder_id=") {
		t.Fatalf("plan output contains durable ids: %q", output)
	}
	if !strings.Contains(output, "files=1\n") {
		t.Fatalf("plan output = %q, want files count", output)
	}
}

func TestEncryptRejectsExistingExportWithoutForce(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	contentOutput := filepath.Join(root, "content")
	databaseOutput := filepath.Join(root, "project.fg")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "note.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(databaseOutput, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", root)
	t.Setenv("FG_TEST_PASSWORD", "test-password")
	err := RunWithIO("fg", []string{
		"encrypt",
		source,
		"--content-out", contentOutput,
		"--export", databaseOutput,
		"--max-part-size", "1024",
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, nil)
	if err == nil {
		t.Fatal("expected existing export to be rejected")
	}
}

func TestExportMissingProjectDoesNotCreateFiles(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "missing.fg")

	t.Setenv("HOME", root)
	t.Setenv("FG_TEST_PASSWORD", "test-password")
	err := RunWithIO("fg", []string{
		"export",
		"missing-project",
		"--out", output,
		"--password-env", "FG_TEST_PASSWORD",
	}, nil, nil)
	if err == nil {
		t.Fatal("expected missing project export to fail")
	}
	if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
		t.Fatalf("output stat error = %v, want not exist", statErr)
	}
	active, err := activeProjectDatabasePath("missing-project")
	if err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(active); !os.IsNotExist(statErr) {
		t.Fatalf("active database stat error = %v, want not exist", statErr)
	}
}

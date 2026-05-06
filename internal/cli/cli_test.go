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
	if !strings.Contains(active, string(filepath.Separator)+"FoldersGuard"+string(filepath.Separator)) {
		t.Fatalf("active database path = %q, want FoldersGuard data directory", active)
	}
	if _, statErr := os.Stat(active); !os.IsNotExist(statErr) {
		t.Fatalf("active database stat error = %v, want not exist", statErr)
	}
}

func TestProjectCommandsRejectDatabasePaths(t *testing.T) {
	root := t.TempDir()
	t.Setenv("FG_TEST_PASSWORD", "test-password")
	shareDatabase := filepath.Join(root, "share.fgs")
	projectDatabase := filepath.Join(root, "project.fg")
	if err := os.WriteFile(shareDatabase, []byte("not a database"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectDatabase, []byte("not a database"), 0o600); err != nil {
		t.Fatal(err)
	}

	commands := [][]string{
		{"rename", shareDatabase, "Root/old.txt", "new.txt", "--password-env", "FG_TEST_PASSWORD"},
		{"rename", projectDatabase, "Root/old.txt", "new.txt", "--password-env", "FG_TEST_PASSWORD"},
		{"add", shareDatabase, filepath.Join(root, "new.txt"), "Root", "--staging-content", filepath.Join(root, "staging"), "--max-part-size", "1024", "--password-env", "FG_TEST_PASSWORD"},
		{"add", projectDatabase, filepath.Join(root, "new.txt"), "Root", "--staging-content", filepath.Join(root, "staging"), "--max-part-size", "1024", "--password-env", "FG_TEST_PASSWORD"},
		{"move", shareDatabase, "Root/old.txt", "Root/docs", "--password-env", "FG_TEST_PASSWORD"},
		{"move", projectDatabase, "Root/old.txt", "Root/docs", "--password-env", "FG_TEST_PASSWORD"},
		{"remove", shareDatabase, "Root/old.txt", "--force", "--password-env", "FG_TEST_PASSWORD"},
		{"remove", projectDatabase, "Root/old.txt", "--force", "--password-env", "FG_TEST_PASSWORD"},
		{"plan", "add", shareDatabase, filepath.Join(root, "new.txt"), "Root", "--staging-content", filepath.Join(root, "staging-plan"), "--max-part-size", "1024", "--password-env", "FG_TEST_PASSWORD"},
		{"plan", "add", projectDatabase, filepath.Join(root, "new.txt"), "Root", "--staging-content", filepath.Join(root, "staging-plan"), "--max-part-size", "1024", "--password-env", "FG_TEST_PASSWORD"},
		{"share", shareDatabase, "Root", "--content", root, "--out-content", filepath.Join(root, "out-content"), "--out-database", filepath.Join(root, "out.fgs"), "--password-env", "FG_TEST_PASSWORD", "--no-share-password"},
		{"share", projectDatabase, "Root", "--content", root, "--out-content", filepath.Join(root, "out-content"), "--out-database", filepath.Join(root, "out.fgs"), "--password-env", "FG_TEST_PASSWORD", "--no-share-password"},
	}
	for _, args := range commands {
		err := RunWithIO("foldersguard", args, nil, nil)
		if err == nil || !strings.Contains(err.Error(), "project id must reference an active project") {
			t.Fatalf("command %v error = %v, want active project id rejection", args, err)
		}
	}
}

func TestReadCommandsRejectExportedProjectDatabasePath(t *testing.T) {
	root := t.TempDir()
	contentRoot := filepath.Join(root, "content")
	outputRoot := filepath.Join(root, "out")
	if err := os.MkdirAll(contentRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FG_TEST_PASSWORD", "test-password")
	projectDatabase := filepath.Join(root, "project.fg")
	if err := os.WriteFile(projectDatabase, []byte("not a database"), 0o600); err != nil {
		t.Fatal(err)
	}

	commands := [][]string{
		{"inspect", projectDatabase, "--password-env", "FG_TEST_PASSWORD"},
		{"verify", projectDatabase, "--content", contentRoot, "--password-env", "FG_TEST_PASSWORD"},
		{"decrypt", projectDatabase, "--content", contentRoot, "--out", outputRoot, "--password-env", "FG_TEST_PASSWORD"},
	}
	for _, args := range commands {
		err := RunWithIO("foldersguard", args, nil, nil)
		if err == nil || !strings.Contains(err.Error(), "must be imported before use") {
			t.Fatalf("command %v error = %v, want import requirement", args, err)
		}
	}
}

func TestPasswordRequiredInNonInteractiveMode(t *testing.T) {
	root := t.TempDir()
	database := filepath.Join(root, "project.fg")
	if err := os.WriteFile(database, []byte("not a database"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := RunWithIO("fg", []string{
		"import",
		database,
	}, strings.NewReader(""), nil)
	if err == nil {
		t.Fatal("expected password requirement error")
	}
	if !strings.Contains(err.Error(), "password input is required") {
		t.Fatalf("error = %v, want password requirement", err)
	}
}

func TestShareRequiresActiveProjectBeforeDefaultSharePassword(t *testing.T) {
	root := t.TempDir()
	contentRoot := filepath.Join(root, "content")
	if err := os.MkdirAll(contentRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	database := filepath.Join(root, "project.fg")
	if err := os.WriteFile(database, []byte("not a database"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FG_TEST_PASSWORD", "test-password")
	err := RunWithIO("fg", []string{
		"share",
		database,
		"Root",
		"--content", contentRoot,
		"--out-content", filepath.Join(root, "out-content"),
		"--out-database", filepath.Join(root, "share.fgs"),
		"--password-env", "FG_TEST_PASSWORD",
	}, strings.NewReader(""), nil)
	if err == nil {
		t.Fatal("expected active project id requirement")
	}
	if !strings.Contains(err.Error(), "project id must reference an active project") {
		t.Fatalf("error = %v, want active project id requirement", err)
	}
}

func TestShareRejectsTwoStdinPasswordModes(t *testing.T) {
	root := t.TempDir()
	err := RunWithIO("fg", []string{
		"share",
		filepath.Join(root, "missing.fg"),
		"Root",
		"--content", root,
		"--out-content", filepath.Join(root, "out-content"),
		"--out-database", filepath.Join(root, "share.fgs"),
		"--password-stdin",
		"--share-password-stdin",
	}, strings.NewReader("test-password"), nil)
	if err == nil {
		t.Fatal("expected mutually exclusive stdin flags")
	}
	if !strings.Contains(err.Error(), "if any flags in the group") {
		t.Fatalf("error = %v, want Cobra mutual exclusion error", err)
	}
}

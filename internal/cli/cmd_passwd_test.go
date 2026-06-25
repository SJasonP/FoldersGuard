package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPasswdChangesProjectPasswordEndToEnd(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	contentOutput := filepath.Join(root, "encrypted")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "file.txt"), []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", root)
	t.Setenv("FG_OLD_PASSWORD", "old-password")
	t.Setenv("FG_NEW_PASSWORD", "new-password")

	var encryptOut bytes.Buffer
	if err := RunWithIO("fg", []string{
		"encrypt", source,
		"--content-out", contentOutput,
		"--max-part-size", "1024",
		"--password-env", "FG_OLD_PASSWORD",
	}, nil, &encryptOut); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	projectID := parseProjectID(t, encryptOut.String())

	if err := RunWithIO("fg", []string{
		"passwd", projectID,
		"--password-env", "FG_OLD_PASSWORD",
		"--new-password-env", "FG_NEW_PASSWORD",
	}, nil, nil); err != nil {
		t.Fatalf("passwd: %v", err)
	}

	// The old password no longer opens the project.
	if err := RunWithIO("fg", []string{
		"inspect", projectID,
		"--password-env", "FG_OLD_PASSWORD",
	}, nil, nil); err == nil {
		t.Fatal("expected the old password to fail after passwd")
	}
	// The new password does.
	if err := RunWithIO("fg", []string{
		"inspect", projectID,
		"--password-env", "FG_NEW_PASSWORD",
	}, nil, nil); err != nil {
		t.Fatalf("inspect with new password: %v", err)
	}
}

func TestPasswdRequiresTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("FG_OLD_PASSWORD", "old")
	t.Setenv("FG_NEW_PASSWORD", "new")
	err := RunWithIO("fg", []string{
		"passwd",
		"--password-env", "FG_OLD_PASSWORD",
		"--new-password-env", "FG_NEW_PASSWORD",
	}, nil, nil)
	if err == nil {
		t.Fatal("expected passwd with no project or --share to fail")
	}
}

func TestPasswdRejectsBothStdinPasswords(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	err := RunWithIO("fg", []string{
		"passwd", "some-project",
		"--password-stdin",
		"--new-password-stdin",
	}, strings.NewReader("x"), nil)
	if err == nil {
		t.Fatal("expected both stdin password flags to be rejected")
	}
}

func parseProjectID(t *testing.T, output string) string {
	t.Helper()
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "project_id=") {
			return strings.TrimPrefix(line, "project_id=")
		}
	}
	t.Fatalf("no project_id in output %q", output)
	return ""
}

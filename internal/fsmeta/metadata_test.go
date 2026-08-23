package fsmeta

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestApplyDoesNotRestoreSpecialPermissionBits(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix special permission bits")
	}

	path := filepath.Join(t.TempDir(), "restored.txt")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	mode := fs.FileMode(0o640) | fs.ModeSetuid | fs.ModeSetgid | fs.ModeSticky
	if err := Apply(path, Metadata{
		Mode:         uint32(mode),
		Capabilities: []string{CapabilityMode},
	}); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode() & (fs.ModeSetuid | fs.ModeSetgid | fs.ModeSticky); got != 0 {
		t.Fatalf("restored special permission bits: %v", got)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("restored permissions = %o, want 640", got)
	}
}

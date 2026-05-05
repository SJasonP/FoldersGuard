package content

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	fgcrypto "foldersguard/internal/crypto"
	"foldersguard/internal/model"
)

func TestEncryptSingleFile(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "input.txt")
	plaintext := []byte("hello foldersguard")
	mustWrite(t, input, plaintext)

	output := filepath.Join(root, "out")
	key, err := fgcrypto.GenerateKey256()
	if err != nil {
		t.Fatal(err)
	}

	fileID := uuid.New().String()
	visiblePath := uuid.New().String()
	err = Encryptor{OutputRoot: output}.EncryptFile(context.Background(), FileSource{
		FileID:       fileID,
		AbsolutePath: input,
		Key:          key,
		StorageKind:  model.StorageKindSingle,
		VisiblePath:  visiblePath,
	})
	if err != nil {
		t.Fatal(err)
	}

	encrypted, err := os.ReadFile(filepath.Join(output, visiblePath))
	if err != nil {
		t.Fatal(err)
	}
	opened, err := OpenObject(key, encrypted, []byte("fg-content-v1:file:"+fileID))
	if err != nil {
		t.Fatal(err)
	}
	if string(opened) != string(plaintext) {
		t.Fatalf("opened = %q, want %q", opened, plaintext)
	}

	encrypted[len(encrypted)-1] ^= 0xff
	if _, err := OpenObject(key, encrypted, []byte("fg-content-v1:file:"+fileID)); err == nil {
		t.Fatal("expected authentication failure after ciphertext tampering")
	}
}

func TestEncryptSplitFile(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "input.txt")
	plaintext := []byte("abcdefghijkl")
	mustWrite(t, input, plaintext)

	output := filepath.Join(root, "out")
	key, err := fgcrypto.GenerateKey256()
	if err != nil {
		t.Fatal(err)
	}

	fileID := uuid.New()
	partA := model.Part{ID: uuid.New(), FileID: fileID, Index: 0, VisibleName: uuid.New(), Offset: 0, Size: 6}
	partB := model.Part{ID: uuid.New(), FileID: fileID, Index: 1, VisibleName: uuid.New(), Offset: 6, Size: 6}
	visiblePath := uuid.New().String()
	err = Encryptor{OutputRoot: output}.EncryptFile(context.Background(), FileSource{
		FileID:       fileID.String(),
		AbsolutePath: input,
		Key:          key,
		StorageKind:  model.StorageKindSplit,
		VisiblePath:  visiblePath,
		Parts:        []model.Part{partA, partB},
	})
	if err != nil {
		t.Fatal(err)
	}

	opened := make([]byte, 0, len(plaintext))
	for _, part := range []model.Part{partA, partB} {
		encrypted, err := os.ReadFile(filepath.Join(output, visiblePath, part.VisibleName.String()))
		if err != nil {
			t.Fatal(err)
		}
		ad := []byte(fmt.Sprintf("fg-content-v1:part:%s:%d:%d:%d", fileID.String(), part.Index, part.Offset, part.Size))
		partPlaintext, err := OpenObject(key, encrypted, ad)
		if err != nil {
			t.Fatal(err)
		}
		opened = append(opened, partPlaintext...)
	}
	if string(opened) != string(plaintext) {
		t.Fatalf("opened = %q, want %q", opened, plaintext)
	}
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

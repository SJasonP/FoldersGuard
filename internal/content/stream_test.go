package content

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	fgcrypto "foldersguard/internal/crypto"
	"foldersguard/internal/model"
)

func TestStreamRoundTripWithProgress(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "input.bin")
	// Larger than one default chunk so progress is reported in several steps.
	plaintext := bytes.Repeat([]byte("foldersguard-streaming-"), 600_000)
	mustWrite(t, input, plaintext)

	key, err := fgcrypto.GenerateKey256()
	if err != nil {
		t.Fatal(err)
	}
	fileID := uuid.New().String()
	visiblePath := uuid.New().String()
	output := filepath.Join(root, "out")

	var encryptBytes int64
	err = Encryptor{OutputRoot: output, OnBytes: func(n int64) { encryptBytes += n }}.EncryptFile(
		context.Background(), FileSource{
			FileID:       fileID,
			AbsolutePath: input,
			Key:          key,
			StorageKind:  model.StorageKindSingle,
			VisiblePath:  visiblePath,
		})
	if err != nil {
		t.Fatal(err)
	}
	if encryptBytes != int64(len(plaintext)) {
		t.Fatalf("encrypt progress = %d bytes, want %d", encryptBytes, len(plaintext))
	}

	encryptedPath := filepath.Join(output, visiblePath)
	ad := []byte("fg-content-v1:file:" + fileID)

	// Verify-stream authenticates and reports the full plaintext size.
	var verifyBytes int64
	if err := VerifyObjectFileStream(context.Background(), key, encryptedPath, ad, func(n int64) { verifyBytes += n }); err != nil {
		t.Fatal(err)
	}
	if verifyBytes != int64(len(plaintext)) {
		t.Fatalf("verify progress = %d bytes, want %d", verifyBytes, len(plaintext))
	}

	// Decrypt-stream reproduces the original plaintext.
	restored := filepath.Join(root, "restored.bin")
	var decryptBytes int64
	if err := OpenObjectFileStream(context.Background(), key, encryptedPath, restored, ad, func(n int64) { decryptBytes += n }); err != nil {
		t.Fatal(err)
	}
	if decryptBytes != int64(len(plaintext)) {
		t.Fatalf("decrypt progress = %d bytes, want %d", decryptBytes, len(plaintext))
	}
	got, err := os.ReadFile(restored)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatal("streamed decrypt did not reproduce the original plaintext")
	}

	// Tampering must fail authentication during verify.
	encrypted, err := os.ReadFile(encryptedPath)
	if err != nil {
		t.Fatal(err)
	}
	encrypted[len(encrypted)-1] ^= 0xff
	tampered := filepath.Join(root, "tampered.bin")
	mustWrite(t, tampered, encrypted)
	if err := VerifyObjectFileStream(context.Background(), key, tampered, ad, nil); err == nil {
		t.Fatal("expected verify failure after ciphertext tampering")
	}
}

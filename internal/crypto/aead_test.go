package crypto

import (
	"bytes"
	"testing"
)

func TestSealOpenAES256GCM(t *testing.T) {
	key, err := GenerateKey256()
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("foldersguard")
	ad := []byte("item-id")

	sealed, err := SealAES256GCM(key, plaintext, ad)
	if err != nil {
		t.Fatal(err)
	}

	opened, err := OpenAES256GCM(key, sealed, ad)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(opened, plaintext) {
		t.Fatalf("opened = %q, want %q", opened, plaintext)
	}

	if _, err := OpenAES256GCM(key, sealed, []byte("wrong")); err == nil {
		t.Fatal("expected authentication failure with wrong associated data")
	}
}

func TestDeriveKey(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}

	params := Argon2idParams{Time: 1, MemoryKiB: 64, Parallelism: 1, KeyLen: KeySize256}
	key, err := DeriveKey("password", salt, params)
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != KeySize256 {
		t.Fatalf("key length = %d, want %d", len(key), KeySize256)
	}
}

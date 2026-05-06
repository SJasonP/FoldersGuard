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

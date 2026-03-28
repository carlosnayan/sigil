package crypto

import (
	"bytes"
	"errors"
	"testing"
)

func TestWrapUnwrapSecret(t *testing.T) {
	pass := "correct-horse-battery-staple"
	secret := "TeamSharedSymmetricKey123"
	blob, err := WrapSecret(pass, secret)
	if err != nil {
		t.Fatal(err)
	}
	out, err := UnwrapSecret(pass, blob)
	if err != nil {
		t.Fatal(err)
	}
	if out != secret {
		t.Fatalf("got %q want %q", out, secret)
	}
}

func TestUnwrapWrongPassphrase(t *testing.T) {
	blob, err := WrapSecret("good", "secret-value")
	if err != nil {
		t.Fatal(err)
	}
	_, err = UnwrapSecret("bad", blob)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrDecrypt) {
		t.Fatalf("want ErrDecrypt, got %v", err)
	}
}

func TestEncryptDecryptVault(t *testing.T) {
	key := "shared-team-secret-abc"
	plain := []byte("API_KEY=xyz\nDB_URL=postgres://\n")
	blob, err := EncryptVault(key, plain)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecryptVault(key, blob)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("got %q want %q", out, plain)
	}
}

func TestDecryptVaultWrongSecret(t *testing.T) {
	blob, err := EncryptVault("key-a", []byte("data"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = DecryptVault("key-b", blob)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrDecrypt) {
		t.Fatalf("want ErrDecrypt, got %v", err)
	}
}

func TestEncryptVaultEmptyPlaintext(t *testing.T) {
	key := "k"
	blob, err := EncryptVault(key, nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecryptVault(key, blob)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("want empty, got %q", out)
	}
}

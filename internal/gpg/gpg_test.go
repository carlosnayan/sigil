package gpg

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func requireGPG(t *testing.T) {
	t.Helper()
	if _, err := gpgBin(); err != nil {
		t.Skip("gpg not in PATH:", err)
	}
}

// tempKeyringDir returns a short path for GNUPGHOME (macOS socket path limit).
func tempKeyringDir(t *testing.T) string {
	t.Helper()
	base := "/tmp"
	if runtime.GOOS == "windows" {
		base = t.TempDir()
	}
	d, err := os.MkdirTemp(base, "vg")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(d) })
	return d
}

func TestEnsureHomeDir(t *testing.T) {
	requireGPG(t)
	home := filepath.Join(tempKeyringDir(t), "gnupg")
	g := New(home)
	if err := g.EnsureHomeDir(); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(home)
	if err != nil {
		t.Fatal(err)
	}
	if !st.IsDir() {
		t.Fatal("expected directory")
	}
	data, err := os.ReadFile(filepath.Join(home, "gpg-agent.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "allow-loopback-pinentry\n" {
		t.Fatalf("gpg-agent.conf: got %q", data)
	}
	if runtime.GOOS != "windows" {
		if got := st.Mode().Perm() & 0o777; got != 0o700 {
			t.Fatalf("dir mode: want 0700, got %04o", got)
		}
	}
}

func TestGenerateKey(t *testing.T) {
	requireGPG(t)
	home := filepath.Join(tempKeyringDir(t), "gnupg")
	g := New(home)
	if err := g.GenerateKey("test-secret-passphrase"); err != nil {
		t.Fatal(err)
	}
	if !g.HasKey() {
		t.Fatal("expected HasKey true after GenerateKey")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	requireGPG(t)
	home := filepath.Join(tempKeyringDir(t), "gnupg")
	g := New(home)
	pass := "round-trip-pass-xyz"
	if err := g.GenerateKey(pass); err != nil {
		t.Fatal(err)
	}
	plain := []byte("hello vault secrets\n")
	cipher, err := g.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(cipher, []byte("BEGIN PGP MESSAGE")) {
		t.Fatalf("expected armored message, got prefix %q", cipher[:min(40, len(cipher))])
	}
	out, err := g.Decrypt(cipher, pass)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("plaintext mismatch: got %q want %q", out, plain)
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	requireGPG(t)
	home := filepath.Join(tempKeyringDir(t), "gnupg")
	g := New(home)
	if err := g.GenerateKey("correct-horse"); err != nil {
		t.Fatal(err)
	}
	cipher, err := g.Encrypt([]byte("data"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.Decrypt(cipher, "wrong-pass")
	if err == nil {
		t.Fatal("expected error for wrong passphrase")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

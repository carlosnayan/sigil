package vaultstore

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		in    string
		valid bool
	}{
		{"dev", true},
		{"my-vault_1", true},
		{"", false},
		{"..", false},
		{"a/b", false},
		{"a\\b", false},
		{"bad space", false},
		{"dot.not", false},
	}
	for _, tt := range tests {
		err := ValidateSlug(tt.in)
		if tt.valid && err != nil {
			t.Errorf("ValidateSlug(%q): want nil, got %v", tt.in, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("ValidateSlug(%q): want error, got nil", tt.in)
		}
	}
}

func TestListEncBasenames_sortedAndFiltered(t *testing.T) {
	d := t.TempDir()
	_ = os.WriteFile(filepath.Join(d, "b.enc"), []byte("1"), 0o600)
	_ = os.WriteFile(filepath.Join(d, "a.enc"), []byte("1"), 0o600)
	_ = os.WriteFile(filepath.Join(d, "skip.txt"), []byte("1"), 0o600)
	_ = os.Mkdir(filepath.Join(d, "nested.enc"), 0o700)

	got, err := ListEncBasenames(d)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.enc", "b.enc"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestListEncBasenames_missingDir(t *testing.T) {
	got, err := ListEncBasenames(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want empty, got %v", got)
	}
}

func TestStem(t *testing.T) {
	if got, want := Stem("dev.enc"), "dev"; got != want {
		t.Fatalf("Stem(dev.enc) = %q, want %q", got, want)
	}
	if got, want := Stem("default.dev.env.ENC"), "default.dev.env"; got != want {
		t.Fatalf("Stem(default.dev.env.ENC) = %q, want %q", got, want)
	}
}

func TestEncPath(t *testing.T) {
	home := t.TempDir()
	want := filepath.Join(home, "vaults", "prod.enc")
	got := EncPath(home, "prod")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

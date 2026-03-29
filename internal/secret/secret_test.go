package secret

import (
	"strings"
	"testing"
)

func TestGenerate_length(t *testing.T) {
	s, err := Generate(32)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 32 {
		t.Fatalf("len %d", len(s))
	}
}

func TestGenerate_charset(t *testing.T) {
	const charsetSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	s, err := Generate(200)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range s {
		if !strings.ContainsRune(charsetSet, r) {
			t.Fatalf("invalid rune %q in %q", r, s)
		}
	}
}

func TestGenerate_zeroLength(t *testing.T) {
	_, err := Generate(0)
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = Generate(-1)
	if err == nil {
		t.Fatal("expected error")
	}
}

package link

import (
	"os"
	"path/filepath"
	"testing"
)

func minimalVaultYAML(secret string) []byte {
	return []byte(`project: test
env: dev
secret: ` + secret + `
vaults: []
inject: {}
`)
}

func TestGet_missingVaultYAML(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "vault.yaml")
	_, _, err := Get(cfgPath, "/tmp/foo")
	if err == nil {
		t.Fatal("expected error when vault.yaml missing")
	}
}

func TestSet_Get_Remove(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "vault.yaml")
	if err := os.WriteFile(cfgPath, minimalVaultYAML("s"), 0o600); err != nil {
		t.Fatal(err)
	}

	dirA := "/tmp/proj-a"
	dirB := "/tmp/proj-b"

	if err := Set(cfgPath, dirA, "dev"); err != nil {
		t.Fatal(err)
	}
	if err := Set(cfgPath, dirB, "prod"); err != nil {
		t.Fatal(err)
	}

	s, ok, err := Get(cfgPath, dirA)
	if err != nil || !ok || s != "dev" {
		t.Fatalf("Get dirA: %q %v err=%v", s, ok, err)
	}

	if err := Set(cfgPath, dirA, "staging"); err != nil {
		t.Fatal(err)
	}
	s, ok, err = Get(cfgPath, dirA)
	if err != nil || !ok || s != "staging" {
		t.Fatalf("after update: %q %v", s, ok)
	}

	if err := Remove(cfgPath, dirA); err != nil {
		t.Fatal(err)
	}
	_, ok, err = Get(cfgPath, dirA)
	if err != nil || ok {
		t.Fatalf("expected dirA removed, ok=%v err=%v", ok, err)
	}
	s, ok, err = Get(cfgPath, dirB)
	if err != nil || !ok || s != "prod" {
		t.Fatalf("dirB: %q %v err=%v", s, ok, err)
	}
}

func TestGet_emptyLinks(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "vault.yaml")
	if err := os.WriteFile(cfgPath, minimalVaultYAML("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, ok, err := Get(cfgPath, "/any/path")
	if err != nil || ok {
		t.Fatalf("want no link, ok=%v err=%v", ok, err)
	}
}

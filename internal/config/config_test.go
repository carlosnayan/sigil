package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_emptyPath(t *testing.T) {
	c, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if c.Secret != "" || len(c.Links) != 0 {
		t.Fatalf("want empty config, got %+v", c)
	}
}

func TestLoad_roundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "vault.yaml")
	orig := &Config{
		Project: "p",
		Env:     "dev",
		Secret:  "sekret",
		Vaults:  []string{"a"},
		Inject:  map[string]string{"K": "v"},
		Links:   map[string]string{"/proj": "dev"},
	}
	if err := Save(p, orig); err != nil {
		t.Fatal(err)
	}
	got, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.Project != "p" || got.Env != "dev" || got.Secret != "sekret" {
		t.Fatalf("fields: %+v", got)
	}
	if len(got.Vaults) != 1 || got.Vaults[0] != "a" {
		t.Fatalf("vaults: %v", got.Vaults)
	}
	if got.Inject["K"] != "v" {
		t.Fatalf("inject: %v", got.Inject)
	}
	if got.Links["/proj"] != "dev" {
		t.Fatalf("links: %v", got.Links)
	}
}

func TestLoad_missingFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nope.yaml")
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "config: read") {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestLoad_invalidYAML(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(p, []byte("not: yaml: [[["), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestSave_nilConfig(t *testing.T) {
	p := filepath.Join(t.TempDir(), "v.yaml")
	if err := Save(p, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected yaml output")
	}
}

package link

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_missingFile(t *testing.T) {
	dir := t.TempDir()
	m, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 0 {
		t.Fatalf("want empty map, got %v", m)
	}
}

func TestSet_Get_Remove(t *testing.T) {
	home := t.TempDir()
	dirA := "/tmp/proj-a"
	dirB := "/tmp/proj-b"

	if err := Set(home, dirA, "dev"); err != nil {
		t.Fatal(err)
	}
	if err := Set(home, dirB, "prod"); err != nil {
		t.Fatal(err)
	}

	m, err := Load(home)
	if err != nil {
		t.Fatal(err)
	}
	if m[dirA] != "dev" || m[dirB] != "prod" {
		t.Fatalf("Load: %v", m)
	}
	s, ok := Get(home, dirA)
	if !ok || s != "dev" {
		t.Fatalf("Get dirA: %q %v", s, ok)
	}

	if err := Set(home, dirA, "staging"); err != nil {
		t.Fatal(err)
	}
	s, _ = Get(home, dirA)
	if s != "staging" {
		t.Fatalf("after update want staging, got %q", s)
	}

	if err := Remove(home, dirA); err != nil {
		t.Fatal(err)
	}
	_, ok = Get(home, dirA)
	if ok {
		t.Fatal("expected dirA removed")
	}
	s, ok = Get(home, dirB)
	if !ok || s != "prod" {
		t.Fatalf("dirB: %q %v", s, ok)
	}
}

func TestLinksFile(t *testing.T) {
	got := LinksFile("/home/x/.sigil")
	want := filepath.Join("/home/x/.sigil", "links.yaml")
	if got != want {
		t.Fatalf("LinksFile: %q want %q", got, want)
	}
}

func TestSave_nilMap(t *testing.T) {
	home := t.TempDir()
	if err := Save(home, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(LinksFile(home))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty yaml for empty map")
	}
}

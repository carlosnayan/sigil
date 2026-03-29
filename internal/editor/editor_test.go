package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolve_prefersVisualOverEditor(t *testing.T) {
	t.Setenv("VISUAL", "vim")
	t.Setenv("EDITOR", "nano")
	if got := Resolve(); got != "vim" {
		t.Fatalf("got %q", got)
	}
}

func TestResolve_editorFallback(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "nano")
	if got := Resolve(); got != "nano" {
		t.Fatalf("got %q", got)
	}
}

func TestSplitCmd_quotedPath(t *testing.T) {
	got := splitCmd(`"/path/with spaces/bin" --wait`)
	want := []string{"/path/with spaces/bin", "--wait"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestSplitCmd_fieldsLikeBefore(t *testing.T) {
	got := splitCmd("code --wait")
	want := []string{"code", "--wait"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestResolve_defaultNano(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")
	if got := Resolve(); got != "nano" {
		t.Fatalf("got %q", got)
	}
}

func TestOpen_appendsAndExitsZero(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires sh helper script")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "ed.sh")
	body := "#!/bin/sh\necho 'APPENDED=1' >> \"$1\"\nexit 0\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VISUAL", script)
	t.Setenv("EDITOR", "")

	target := filepath.Join(dir, "work.txt")
	if err := os.WriteFile(target, []byte("start\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	code, err := Open(target)
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}
	b, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "APPENDED=1") {
		t.Fatalf("file content %q missing APPENDED=1", string(b))
	}
}

func TestOpen_nonZeroExitDoesNotRequireError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires sh helper script")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "ed.sh")
	body := "#!/bin/sh\nexit 7\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VISUAL", script)
	t.Setenv("EDITOR", "")

	target := filepath.Join(dir, "work.txt")
	if err := os.WriteFile(target, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	code, err := Open(target)
	if err != nil {
		t.Fatal(err)
	}
	if code != 7 {
		t.Fatalf("want exit 7, got %d", code)
	}
}

func TestOpen_withEditorArgs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires sh helper script")
	}
	dir := t.TempDir()
	script := filepath.Join(dir, "ed.sh")
	body := "#!/bin/sh\nprintf 'ok' >> \"$3\"\nexit 0\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VISUAL", fmt.Sprintf("%s -ignored -ignored", script))
	t.Setenv("EDITOR", "")

	target := filepath.Join(dir, "work.txt")
	if err := os.WriteFile(target, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	code, err := Open(target)
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	b, _ := os.ReadFile(target)
	if string(b) != "ok" {
		t.Fatalf("got %q", string(b))
	}
}

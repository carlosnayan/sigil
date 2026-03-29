package runner

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

func echoHelloCmd() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/c", "echo", "hello"}
	}
	return "echo", []string{"hello"}
}

func exit42Cmd() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/c", "exit", "42"}
	}
	return "sh", []string{"-c", "exit 42"}
}

func TestRun_echoExit0(t *testing.T) {
	name, args := echoHelloCmd()
	code, err := Run(name, args, os.Environ())
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}
}

func TestRun_nonZeroExit(t *testing.T) {
	name, args := exit42Cmd()
	code, err := Run(name, args, os.Environ())
	if err != nil {
		t.Fatal(err)
	}
	if code != 42 {
		t.Fatalf("want exit 42, got %d", code)
	}
}

func TestRun_passesEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell env check is unix-oriented")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not found")
	}
	base := os.Environ()
	custom := append(append([]string{}, base...), "SIGIL_RUNNER_TEST=ok")
	code, err := Run("sh", []string{"-c", `[ "$SIGIL_RUNNER_TEST" = "ok" ]`}, custom)
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("script failed, exit %d", code)
	}
}

func TestRun_startError(t *testing.T) {
	code, err := Run("/nonexistent/sigil-runner-binary-xyz", nil, os.Environ())
	if err == nil {
		t.Fatal("expected error")
	}
	if code != 1 {
		t.Fatalf("want code 1, got %d", code)
	}
}

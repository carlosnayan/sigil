package editor

import (
	"os"
	"os/exec"
	"strings"
)

func Resolve() string {
	if v := strings.TrimSpace(os.Getenv("VISUAL")); v != "" {
		return v
	}
	if e := strings.TrimSpace(os.Getenv("EDITOR")); e != "" {
		return e
	}
	return "nano"
}

func Open(path string) (exitCode int, err error) {
	cmdline := Resolve()
	parts := splitCmd(cmdline)
	if len(parts) == 0 {
		parts = []string{"nano"}
	}
	bin := parts[0]
	args := append(append([]string{}, parts[1:]...), path)
	c := exec.Command(bin, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err == nil {
		return 0, nil
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode(), nil
	}
	return -1, err
}

func splitCmd(s string) []string {
	return strings.Fields(strings.TrimSpace(s))
}

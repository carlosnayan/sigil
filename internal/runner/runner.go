package runner

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// Run starts a child process with the given environment (full KEY=value slice),
// wires stdin/stdout/stderr to the parent, forwards SIGINT and SIGTERM to the child,
// and returns the child's exit code (0 on success).
func Run(name string, args []string, environ []string) (exitCode int, err error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = environ

	if err := cmd.Start(); err != nil {
		return 1, err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	for {
		select {
		case sig := <-sigCh:
			if cmd.Process != nil {
				_ = cmd.Process.Signal(sig)
			}
		case waitErr := <-waitDone:
			if waitErr == nil {
				return 0, nil
			}
			var exitErr *exec.ExitError
			if errors.As(waitErr, &exitErr) {
				return exitErr.ExitCode(), nil
			}
			return 1, waitErr
		}
	}
}

package gpg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const recipientEmail = "vault@local"

// GPG runs GnuPG with an isolated keyring in HomeDir.
type GPG struct {
	HomeDir string
}

// New returns a GPG helper for the given homedir (e.g. ~/.vault/gnupg).
func New(homeDir string) *GPG {
	return &GPG{HomeDir: homeDir}
}

func gpgBin() (string, error) {
	for _, name := range []string{"gpg2", "gpg"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", errors.New("gpg: gpg or gpg2 not found in PATH")
}

func gpgconfBin() (string, error) {
	if p, err := exec.LookPath("gpgconf"); err == nil {
		return p, nil
	}
	g, err := gpgBin()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(g)
	for _, name := range []string{"gpgconf", "gpgconf.exe"} {
		candidate := filepath.Join(dir, name)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate, nil
		}
	}
	return "", errors.New("gpg: gpgconf not found")
}

// launchAgent starts gpg-agent for this homedir (required for key generation on GnuPG 2.x).
func (g *GPG) launchAgent() error {
	if g.HomeDir == "" {
		return errors.New("gpg: HomeDir is empty")
	}
	conf, err := gpgconfBin()
	if err != nil {
		return err
	}
	cmd := exec.Command(conf, "--homedir", g.HomeDir, "--launch", "gpg-agent")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gpgconf --launch gpg-agent: %w\n%s", err, stderr.String())
	}
	return nil
}

func (g *GPG) run(args []string, stdin []byte) ([]byte, error) {
	if g.HomeDir == "" {
		return nil, errors.New("gpg: HomeDir is empty")
	}
	bin, err := gpgBin()
	if err != nil {
		return nil, err
	}
	base := []string{"--homedir", g.HomeDir, "--batch", "--yes"}
	cmd := exec.Command(bin, append(base, args...)...)
	if stdin == nil {
		cmd.Stdin = bytes.NewReader(nil)
	} else {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gpg %v: %w\n%s", args, err, stderr.String())
	}
	return stdout.Bytes(), nil
}

func (g *GPG) runWithPassphrase(args []string, passphrase string) ([]byte, error) {
	if g.HomeDir == "" {
		return nil, errors.New("gpg: HomeDir is empty")
	}
	bin, err := gpgBin()
	if err != nil {
		return nil, err
	}
	base := []string{
		"--homedir", g.HomeDir,
		"--batch", "--yes",
		"--pinentry-mode", "loopback",
		"--passphrase-fd", "0",
	}
	cmd := exec.Command(bin, append(base, args...)...)
	cmd.Stdin = bytes.NewReader([]byte(passphrase + "\n"))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gpg %v: %w\n%s", args, err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// EnsureHomeDir creates the isolated GNUPGHOME directory and gpg-agent.conf.
func (g *GPG) EnsureHomeDir() error {
	if g.HomeDir == "" {
		return errors.New("gpg: HomeDir is empty")
	}
	if err := os.MkdirAll(g.HomeDir, 0o700); err != nil {
		return err
	}
	agentConf := filepath.Join(g.HomeDir, "gpg-agent.conf")
	return os.WriteFile(agentConf, []byte("allow-loopback-pinentry\n"), 0o600)
}

// GenerateKey creates the Vault Local / vault@local keypair in the isolated keyring.
func (g *GPG) GenerateKey(passphrase string) error {
	if err := g.EnsureHomeDir(); err != nil {
		return err
	}
	if err := g.launchAgent(); err != nil {
		return err
	}
	// RSA for broad compatibility (default/ed25519 curves vary across GnuPG builds).
	var b strings.Builder
	b.WriteString("Key-Type: RSA\nKey-Length: 4096\nSubkey-Type: RSA\nSubkey-Length: 4096\nName-Real: Vault Local\nName-Email: ")
	b.WriteString(recipientEmail)
	b.WriteString("\nExpire-Date: 0\nPassphrase: ")
	b.WriteString(passphrase)
	b.WriteString("\n%commit\n")
	batch := b.String()
	_, err := g.run([]string{"--pinentry-mode", "loopback", "--generate-key"}, []byte(batch))
	return err
}

// HasKey reports whether a key for vault@local exists in the keyring.
func (g *GPG) HasKey() bool {
	if g.HomeDir == "" {
		return false
	}
	_, err := g.run([]string{"--list-keys", recipientEmail}, nil)
	return err == nil
}

// Encrypt encrypts plaintext for vault@local (ASCII-armored).
func (g *GPG) Encrypt(plaintext []byte) ([]byte, error) {
	return g.run([]string{
		"--pinentry-mode", "loopback",
		"--trust-model", "always",
		"--recipient", recipientEmail,
		"--encrypt",
		"--armor",
	}, plaintext)
}

// Decrypt decrypts ciphertext using the vault private key and passphrase (cross-platform: temp file + stdin passphrase).
func (g *GPG) Decrypt(ciphertext []byte, passphrase string) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, errors.New("gpg: empty ciphertext")
	}
	f, err := os.CreateTemp("", "vault-gpg-*.asc")
	if err != nil {
		return nil, err
	}
	path := f.Name()
	defer func() { _ = os.Remove(path) }()

	if _, err := f.Write(ciphertext); err != nil {
		_ = f.Close()
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	return g.runWithPassphrase([]string{"--decrypt", path}, passphrase)
}

// IsUnlocked reports whether the keyring is accessible (same check as HasKey for an isolated homedir).
func (g *GPG) IsUnlocked() bool {
	return g.HasKey()
}

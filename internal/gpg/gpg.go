package gpg

import "errors"

// GPG encapsula operações com keyring isolado em HomeDir.
type GPG struct {
	HomeDir string
}

// GenerateKey gera par de chaves no keyring. Stub até Fase 2.
func (g *GPG) GenerateKey(passphrase string) error {
	_ = passphrase
	if g.HomeDir == "" {
		return errors.New("gpg: HomeDir vazio")
	}
	return nil
}

// Encrypt criptografa plaintext. Stub até Fase 2.
func (g *GPG) Encrypt(plaintext []byte, recipients ...string) ([]byte, error) {
	_ = plaintext
	_ = recipients
	return nil, nil
}

// Decrypt descriptografa ciphertext. Stub até Fase 2.
func (g *GPG) Decrypt(ciphertext []byte) ([]byte, error) {
	_ = ciphertext
	return nil, nil
}

// IsUnlocked indica se o agente está desbloqueado. Stub até Fase 2.
func (g *GPG) IsUnlocked() bool {
	return false
}

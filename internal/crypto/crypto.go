// Package crypto provides symmetric encryption for Sigil: AES-256-GCM with keys derived via scrypt.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/scrypt"
)

const (
	// Scrypt parameters (memory-hard key derivation).
	scryptN  = 32768
	scryptR  = 8
	scryptP  = 1
	saltLen  = 16
	nonceLen = 12
	keyLen   = 32
	wireV1   = byte(1)
)

// ErrDecrypt is returned when decryption fails (wrong key or corrupt data).
var ErrDecrypt = errors.New("crypto: decryption failed")

func deriveKey(password []byte, salt []byte) ([]byte, error) {
	return scrypt.Key(password, salt, scryptN, scryptR, scryptP, keyLen)
}

func seal(password []byte, plaintext []byte) ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	dk, err := deriveKey(password, salt)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceLen)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ct := gcm.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 1+saltLen+nonceLen+len(ct))
	out[0] = wireV1
	copy(out[1:], salt)
	copy(out[1+saltLen:], nonce)
	copy(out[1+saltLen+nonceLen:], ct)
	return out, nil
}

func open(password []byte, blob []byte) ([]byte, error) {
	if len(blob) < 1+saltLen+nonceLen+aes.BlockSize {
		return nil, ErrDecrypt
	}
	if blob[0] != wireV1 {
		return nil, ErrDecrypt
	}
	salt := blob[1 : 1+saltLen]
	nonce := blob[1+saltLen : 1+saltLen+nonceLen]
	ct := blob[1+saltLen+nonceLen:]
	dk, err := deriveKey(password, salt)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecrypt, err)
	}
	return pt, nil
}

// WrapSecret encrypts the vault symmetric secret (UTF-8 string) with the user's passphrase.
// Stored at ~/.vault/secret.enc.
func WrapSecret(passphrase string, secretPlain string) ([]byte, error) {
	if passphrase == "" {
		return nil, errors.New("crypto: passphrase is empty")
	}
	return seal([]byte(passphrase), []byte(secretPlain))
}

// UnwrapSecret decrypts secret.enc using the passphrase.
func UnwrapSecret(passphrase string, blob []byte) (string, error) {
	pt, err := open([]byte(passphrase), blob)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

// EncryptVault encrypts vault plaintext (e.g. .env body) with the team/shared secret string.
func EncryptVault(secret string, plaintext []byte) ([]byte, error) {
	if secret == "" {
		return nil, errors.New("crypto: secret is empty")
	}
	return seal([]byte(secret), plaintext)
}

// DecryptVault decrypts a vault blob produced by EncryptVault.
func DecryptVault(secret string, blob []byte) ([]byte, error) {
	if secret == "" {
		return nil, errors.New("crypto: secret is empty")
	}
	return open([]byte(secret), blob)
}

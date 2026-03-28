package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carlos/sigil/internal/crypto"
	"github.com/carlos/sigil/internal/editor"
	"github.com/carlos/sigil/internal/ui"
	"github.com/carlos/sigil/internal/vaultstore"
	"github.com/manifoldco/promptui"
)

const vaultStarterTemplate = "# Add KEY=value lines below.\n"

func runManageSecrets(sigilHome, encryptionKey string) error {
	vaultsDir := vaultstore.Dir(sigilHome)
	if err := os.MkdirAll(vaultsDir, 0o700); err != nil {
		return fmt.Errorf("sigil: vaults directory: %w", err)
	}

	for {
		basenames, err := vaultstore.ListEncBasenames(vaultsDir)
		if err != nil {
			return fmt.Errorf("sigil: list vaults: %w", err)
		}

		items := []string{"Add new config"}
		for _, b := range basenames {
			items = append(items, vaultstore.Stem(b))
		}
		items = append(items, "Back")

		idx, _, err := ui.SelectList("Manage configs", items)
		if err != nil {
			if errors.Is(err, promptui.ErrInterrupt) || errors.Is(err, promptui.ErrEOF) {
				return nil
			}
			return err
		}

		backIdx := len(items) - 1
		switch {
		case idx == backIdx:
			return nil
		case idx == 0:
			if err := addNewVault(vaultsDir, encryptionKey); err != nil {
				if !isPromptAbort(err) {
					ui.Error(err.Error())
				}
			}
		default:
			basename := basenames[idx-1]
			if err := vaultSubmenu(vaultsDir, basename, encryptionKey); err != nil {
				if errors.Is(err, promptui.ErrInterrupt) || errors.Is(err, promptui.ErrEOF) {
					return nil
				}
				if !isPromptAbort(err) {
					ui.Error(err.Error())
				}
			}
		}
	}
}

func isPromptAbort(err error) bool {
	return errors.Is(err, promptui.ErrInterrupt) || errors.Is(err, promptui.ErrEOF)
}

func addNewVault(vaultsDir, encryptionKey string) error {
	slug, err := ui.PromptText("Environment name (slug)")
	if err != nil {
		return err
	}
	if err := vaultstore.ValidateSlug(slug); err != nil {
		return err
	}

	dest := filepath.Join(vaultsDir, vaultstore.EncFileName(slug))
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("environment %q already exists", slug)
	} else if !os.IsNotExist(err) {
		return err
	}

	tmp, err := os.CreateTemp("", "sigil-new-vault-*.env")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmp.WriteString(vaultStarterTemplate); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	code, err := editor.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("launch editor: %w", err)
	}
	if code != 0 {
		ui.Warn("Editor exited without saving; environment not created.")
		return nil
	}

	plaintext, err := os.ReadFile(tmpPath)
	if err != nil {
		return err
	}
	enc, err := crypto.EncryptVault(encryptionKey, plaintext)
	if err != nil {
		return fmt.Errorf("encrypt environment: %w", err)
	}
	if err := os.WriteFile(dest, enc, 0o600); err != nil {
		return err
	}
	ui.Success(fmt.Sprintf("Environment %q created.", slug))
	return nil
}

func vaultSubmenu(vaultsDir, basename, encryptionKey string) error {
	path := vaultstore.Path(vaultsDir, basename)
	title := vaultstore.Stem(basename)

	for {
		items := []string{"Edit", "Delete", "Back"}
		idx, _, err := ui.SelectList(title, items)
		if err != nil {
			return err
		}
		switch idx {
		case 2:
			return nil
		case 0:
			if err := editVaultFile(path, encryptionKey); err != nil {
				return err
			}
		case 1:
			ok, err := ui.Confirm("Delete environment "+title+"? This cannot be undone.", false)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("remove vault: %w", err)
			}
			ui.Success("Environment deleted.")
			return nil
		}
	}
}

func editVaultFile(encPath, encryptionKey string) error {
	blob, err := os.ReadFile(encPath)
	if err != nil {
		return fmt.Errorf("read environment file: %w", err)
	}
	plaintext, err := crypto.DecryptVault(encryptionKey, blob)
	if err != nil {
		if errors.Is(err, crypto.ErrDecrypt) {
			return fmt.Errorf("cannot decrypt (wrong encryption key or corrupt file)")
		}
		return fmt.Errorf("decrypt environment: %w", err)
	}

	tmp, err := os.CreateTemp("", "sigil-edit-vault-*.env")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if err := os.WriteFile(tmpPath, []byte(plaintext), 0o600); err != nil {
		return err
	}

	code, err := editor.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("launch editor: %w", err)
	}
	if code != 0 {
		ui.Warn("Editor exited without saving; file unchanged.")
		return nil
	}

	newPlain, err := os.ReadFile(tmpPath)
	if err != nil {
		return err
	}
	out, err := crypto.EncryptVault(encryptionKey, newPlain)
	if err != nil {
		return fmt.Errorf("encrypt environment: %w", err)
	}
	if err := os.WriteFile(encPath, out, 0o600); err != nil {
		return err
	}
	ui.Success("Environment saved.")
	return nil
}

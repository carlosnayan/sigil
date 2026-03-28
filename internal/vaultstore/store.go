package vaultstore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

var ErrInvalidSlug = errors.New("vaultstore: invalid vault slug")

func Dir(sigilHome string) string {
	return filepath.Join(sigilHome, "vaults")
}

func ValidateSlug(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("%w: empty", ErrInvalidSlug)
	}
	if strings.Contains(s, "..") {
		return fmt.Errorf("%w: .. not allowed", ErrInvalidSlug)
	}
	if strings.ContainsAny(s, `/\`) {
		return fmt.Errorf("%w: path separators not allowed", ErrInvalidSlug)
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			continue
		}
		return fmt.Errorf("%w: use only letters, digits, underscore, hyphen", ErrInvalidSlug)
	}
	return nil
}

func EncFileName(slug string) string {
	return slug + ".enc"
}

func EncPath(sigilHome, slug string) string {
	return filepath.Join(Dir(sigilHome), EncFileName(slug))
}

func ListEncBasenames(vaultsDir string) ([]string, error) {
	entries, err := os.ReadDir(vaultsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.EqualFold(filepath.Ext(name), ".enc") {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out, nil
}

func Stem(basename string) string {
	const suf = ".enc"
	if len(basename) > len(suf) && strings.EqualFold(basename[len(basename)-len(suf):], suf) {
		return basename[:len(basename)-len(suf)]
	}
	return basename
}

func Path(vaultsDir, basename string) string {
	return filepath.Join(vaultsDir, basename)
}

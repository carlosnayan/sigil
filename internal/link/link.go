package link

import (
	"github.com/carlos/sigil/internal/config"
)

// Get returns the vault config slug linked to dir (absolute path) from vault.yaml's links map.
func Get(cfgPath, dir string) (slug string, ok bool, err error) {
	c, err := config.Load(cfgPath)
	if err != nil {
		return "", false, err
	}
	if c.Links == nil {
		return "", false, nil
	}
	s, ok := c.Links[dir]
	return s, ok, nil
}

// Set records dir -> slug in vault.yaml (merges with existing config).
func Set(cfgPath, dir, slug string) error {
	c, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	if c.Links == nil {
		c.Links = make(map[string]string)
	}
	c.Links[dir] = slug
	return config.Save(cfgPath, c)
}

// Remove deletes the link for dir from vault.yaml.
func Remove(cfgPath, dir string) error {
	c, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	if c.Links == nil {
		return nil
	}
	delete(c.Links, dir)
	return config.Save(cfgPath, c)
}

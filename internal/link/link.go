package link

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LinksFile(sigilHome string) string {
	return filepath.Join(sigilHome, "links.yaml")
}

func Load(sigilHome string) (map[string]string, error) {
	p := LinksFile(sigilHome)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	var m map[string]string
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m == nil {
		return map[string]string{}, nil
	}
	return m, nil
}

func Save(sigilHome string, links map[string]string) error {
	if links == nil {
		links = map[string]string{}
	}
	data, err := yaml.Marshal(links)
	if err != nil {
		return err
	}
	return os.WriteFile(LinksFile(sigilHome), data, 0o600)
}

func Set(sigilHome, dir, slug string) error {
	m, err := Load(sigilHome)
	if err != nil {
		return err
	}
	m[dir] = slug
	return Save(sigilHome, m)
}

func Get(sigilHome, dir string) (slug string, ok bool) {
	m, err := Load(sigilHome)
	if err != nil {
		return "", false
	}
	s, ok := m[dir]
	return s, ok
}

func Remove(sigilHome, dir string) error {
	m, err := Load(sigilHome)
	if err != nil {
		return err
	}
	delete(m, dir)
	return Save(sigilHome, m)
}

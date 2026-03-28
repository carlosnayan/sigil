package vault

import "errors"

type Vault struct {
	Name string
	Env  string
}

func (v *Vault) Create() error {
	if v.Name == "" {
		return errors.New("vault: Name vazio")
	}
	return nil
}

func (v *Vault) Read() ([]byte, error) {
	return nil, nil
}

func (v *Vault) Write(data []byte) error {
	_ = data
	return nil
}

func (v *Vault) List() ([]string, error) {
	_ = v
	return nil, nil
}

package vault

import "errors"

// Vault representa um conjunto de secrets criptografados por nome/ambiente.
type Vault struct {
	Name string
	Env  string
}

// Create garante arquivo vazio criptografado. Stub até Fase 5.
func (v *Vault) Create() error {
	if v.Name == "" {
		return errors.New("vault: Name vazio")
	}
	return nil
}

// Read retorna conteúdo descriptografado. Stub.
func (v *Vault) Read() ([]byte, error) {
	return nil, nil
}

// Write persiste conteúdo criptografado. Stub.
func (v *Vault) Write(data []byte) error {
	_ = data
	return nil
}

// List retorna chaves ou entradas do vault. Stub.
func (v *Vault) List() ([]string, error) {
	_ = v
	return nil, nil
}

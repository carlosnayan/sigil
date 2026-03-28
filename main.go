package main

import (
	"github.com/carlos/sigil/cmd"

	// Garante que os pacotes internal compilam com `go build .` (Fase 1).
	_ "github.com/carlos/sigil/internal/config"
	_ "github.com/carlos/sigil/internal/crypto"
	_ "github.com/carlos/sigil/internal/env"
	_ "github.com/carlos/sigil/internal/runner"
	_ "github.com/carlos/sigil/internal/secret"
	_ "github.com/carlos/sigil/internal/ui"
	_ "github.com/carlos/sigil/internal/vault"
)

func main() {
	cmd.Execute()
}

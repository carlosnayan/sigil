package runner

import "errors"

// Run executa comando com ambiente extra. Stub até Fase 8.
func Run(name string, args []string, env map[string]string) error {
	_ = name
	_ = args
	_ = env
	return errors.New("runner: não implementado (Fase 8)")
}

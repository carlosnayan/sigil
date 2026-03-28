package env

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/joho/godotenv"
)

// Parse interpreta conteúdo estilo .env.
func Parse(data []byte) (map[string]string, error) {
	m, err := godotenv.Unmarshal(string(data))
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Merge combina mapas; entradas posteriores sobrescrevem anteriores.
func Merge(envs ...map[string]string) map[string]string {
	out := make(map[string]string)
	for _, e := range envs {
		for k, v := range e {
			out[k] = v
		}
	}
	return out
}

// Serialize gera texto KEY=VALUE ordenado por chave.
func Serialize(env map[string]string) []byte {
	if len(env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b bytes.Buffer
	for _, k := range keys {
		_, _ = fmt.Fprintf(&b, "%s=%s\n", k, env[k])
	}
	return b.Bytes()
}

// SerializeDotEnv alias com quebras de linha Windows-friendly desabilitado (LF apenas).
func SerializeDotEnv(env map[string]string) string {
	return strings.TrimSuffix(string(Serialize(env)), "\n")
}

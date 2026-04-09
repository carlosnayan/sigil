# Sigil

> Environment files, sealed with a Sigil.

Sigil is a secure, offline-first CLI vault manager for encrypting and managing environment variables. It encrypts `.env` files with AES-256-GCM, stores them locally, and injects secrets into child processes at runtime — no cloud, no daemon, no plaintext on disk.

## Features

- **AES-256-GCM encryption** with scrypt key derivation — secrets encrypted at rest
- **Offline-first** — no network, no server, no third-party dependency
- **Environment injection** — decrypt and inject vars into any command via `sigil run`
- **Multiple environments** — manage dev, staging, prod configs as separate vaults
- **Project linking** — bind directories to vault configs for automatic resolution
- **Interactive menus** — add, edit, and delete encrypted configs from the terminal
- **Editor integration** — uses `$VISUAL` / `$EDITOR` for editing vault contents

## Installation

### From source

```bash
git clone https://github.com/carlos/sigil.git
cd sigil
make build
```

The binary will be at `bin/sigil`. Move it to your `$PATH`:

```bash
cp bin/sigil /usr/local/bin/
```

### With `go install`

```bash
go install github.com/carlos/sigil@latest
```

Requires Go 1.26+.

## Quick Start

```bash
# 1. Initialize Sigil (creates ~/.sigil/ with encryption key)
sigil init

# 2. Create an encrypted vault config via interactive menu
sigil config

# 3. Link current project to a vault config
sigil setup

# 4. Run a command with vault secrets injected
sigil run -- docker compose up
```

## Commands

### `sigil init`

Initializes Sigil by creating `~/.sigil/vault.yaml` and the `~/.sigil/vaults/` directory. Prompts for an encryption secret or generates one automatically.

```bash
sigil init
```

### `sigil config`

Opens an interactive menu to manage encrypted vault configurations:

- **Add new config** — create a new encrypted `.env` file
- **Manage configs** — edit, delete, or view existing vault configs

```bash
sigil config
```

Editing opens your `$VISUAL` or `$EDITOR` (falls back to `nano`). Write environment variables in standard `.env` format:

```
DATABASE_URL=postgres://localhost:5432/mydb
API_KEY=sk-secret-value
REDIS_HOST=127.0.0.1
```

### `sigil setup`

Links the current project directory to a vault config. Reads `sigil.yaml` from the project root and records the mapping in `vault.yaml`.

```bash
cd /path/to/my-project
sigil setup
```

Requires a `sigil.yaml` in the project directory:

```yaml
setup:
  config: dev  # references ~/.sigil/vaults/dev.enc
```

### `sigil run`

Executes a command with decrypted vault variables injected into the environment.

```bash
sigil run -- <command> [args...]
```

Examples:

```bash
sigil run -- docker compose up -d
sigil run -- npm start
sigil run -- env | grep DATABASE
```

**Environment priority** (highest to lowest):

| Priority | Source | Description |
|----------|--------|-------------|
| 1 | `inject` (vault.yaml) | Static vars from config |
| 2 | Vault (`.enc` file) | Decrypted secrets |
| 3 | OS environment | Inherited from shell |

### Global Flags

```
-v, --verbose        verbose output
    --config <path>  path to vault.yaml (default: ~/.sigil/vault.yaml)
    --version        show version
```

## Configuration

### `~/.sigil/vault.yaml`

Main configuration file created by `sigil init`:

```yaml
project: my-project
env: dev
secret: <encryption-key>
inject:
  NODE_ENV: development
  LOG_LEVEL: debug
links:
  /home/user/projects/api: api-prod
  /home/user/projects/web: web-dev
```

| Field | Description |
|-------|-------------|
| `project` | Project identifier |
| `env` | Default environment name |
| `secret` | Master encryption/decryption key |
| `inject` | Static key-value pairs injected into every `sigil run` |
| `links` | Maps absolute directory paths to vault config slugs |

### `sigil.yaml` (per-project)

Place in the project root to declare which vault config to use:

```yaml
setup:
  config: dev
```

The `config` value references a file at `~/.sigil/vaults/<config>.enc`.

### Encrypted Vault Files

Stored at `~/.sigil/vaults/<slug>.enc` with `0600` permissions. Binary format:

```
[1B version][16B salt][12B nonce][ciphertext + GCM auth tag]
```

## Security

- **Encryption**: AES-256-GCM (authenticated encryption)
- **Key derivation**: scrypt (N=32768, r=8, p=1) with random 16-byte salt per encryption
- **File permissions**: vault files are `0600`, vaults directory is `0700`
- **No plaintext on disk**: decryption happens in-memory during `sigil run` only
- **Slug validation**: prevents directory traversal (no `/`, `\`, or `..`)

## Development

```bash
# Build
make build

# Run tests (includes gofmt, go vet, go test)
make test

# Run directly
make run ARGS='config --help'
```

## License

[GNU AGPL v3](LICENSE)

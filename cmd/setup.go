package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/carlos/sigil/internal/link"
	"github.com/carlos/sigil/internal/ui"
	"github.com/carlos/sigil/internal/vaultstore"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Link a config from sigil.yaml to the current directory",
	Long: `Reads sigil.yaml (or sigil.yml) in the current directory for setup.config,
verifies ~/.sigil/vaults/<config>.enc exists, and records the mapping in vault.yaml (links)
for use by sigil run.`,
	Args: cobra.NoArgs,
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

type sigilYAML struct {
	Setup struct {
		Config string `yaml:"config"`
	} `yaml:"setup"`
}

var (
	setupGetwd = os.Getwd
)

func findSigilManifest(dir string) (path string, err error) {
	for _, name := range []string{"sigil.yaml", "sigil.yml"} {
		p := filepath.Join(dir, name)
		st, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		if st.IsDir() {
			continue
		}
		return p, nil
	}
	return "", errSigilNotFound
}

var errSigilNotFound = errors.New("sigil: no sigil.yaml found in current directory")

func runSetup(cmd *cobra.Command, _ []string) error {
	wd, err := setupGetwd()
	if err != nil {
		return fmt.Errorf("sigil: get working directory: %w", err)
	}
	cwd, err := filepath.Abs(wd)
	if err != nil {
		return fmt.Errorf("sigil: resolve working directory: %w", err)
	}

	manifest, err := findSigilManifest(cwd)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(manifest)
	if err != nil {
		return fmt.Errorf("sigil: read %s: %w", manifest, err)
	}

	var doc sigilYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("sigil: parse %s: %w", filepath.Base(manifest), err)
	}

	config := doc.Setup.Config
	if config == "" {
		return fmt.Errorf("sigil: setup.config is required in sigil.yaml")
	}
	if err := vaultstore.ValidateSlug(config); err != nil {
		return fmt.Errorf("sigil: invalid setup.config: %w", err)
	}

	sigilHome, err := SigilHome()
	if err != nil {
		return err
	}

	encPath := vaultstore.EncPath(sigilHome, config)
	if _, err := os.Stat(encPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("sigil: config %q not found in ~/.sigil/vaults/ — create it first via sigil config > Manage configs", config)
		}
		return fmt.Errorf("sigil: stat vault config: %w", err)
	}

	cfgPath, err := configPathForCmd(cmd)
	if err != nil {
		return err
	}
	if _, err := os.Stat(cfgPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("sigil: no vault.yaml at %s — run `sigil init` first", cfgPath)
		}
		return fmt.Errorf("sigil: vault.yaml: %w", err)
	}

	if err := link.Set(cfgPath, cwd, config); err != nil {
		return fmt.Errorf("sigil: save link: %w", err)
	}

	ui.PrintTable(
		[]string{"NAME", "VALUE", "SCOPE"},
		[][]string{{"config", config, cwd}},
	)
	return nil
}

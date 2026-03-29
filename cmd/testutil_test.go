package cmd

import (
	"sync"
	"testing"

	"github.com/mitchellh/go-homedir"
)

var cmdExecuteMu sync.Mutex

func resetCmdState(t *testing.T) {
	t.Helper()
	homedir.Reset()
	_ = rootCmd.PersistentFlags().Set("config", "")
	_ = rootCmd.PersistentFlags().Set("verbose", "false")
}

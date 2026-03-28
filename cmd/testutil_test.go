package cmd

import (
	"sync"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cmdExecuteMu sync.Mutex

func resetCmdState(t *testing.T) {
	t.Helper()
	homedir.Reset()
	viper.Reset()
	_ = rootCmd.PersistentFlags().Set("config", "")
	_ = rootCmd.PersistentFlags().Set("verbose", "false")
}

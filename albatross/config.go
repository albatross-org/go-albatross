package albatross

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var log *logrus.Entry = logrus.New().WithField("prefix", "albatross")

// getConfigDir gets the user's configuration directory.
// TODO: At the moment, this uses $XDG_CONFIG_HOME and falls back to
// $HOME/.config which isn't cross platform.
func getConfigDir() string {
	config := os.Getenv("XDG_CONFIG_HOME")
	if config != "" {
		return config
	}

	home, err := homedir.Dir()
	if err != nil {
		panic(err) // This really shouldn't happen.
	}

	return filepath.Join(home, ".config")
}

// parseConfigFile parses a config file into a Viper configuration.
// This function sets the defaults.
func parseConfigFile(path string) (*viper.Viper, error) {
	v := viper.New()

	v.SetDefault("dates.format", "2006-01-02 15:04")
	v.SetDefault("tags.prefix-builtin", "@!")
	v.SetDefault("tags.prefix-custom", "@?")

	defaultPublicKeyPath := filepath.Join(getConfigDir(), "albatross", "keys", "public.key")
	defaultPrivateKeyPath := filepath.Join(getConfigDir(), "albatross", "keys", "private.key")

	v.SetDefault("encryption.public-key", defaultPublicKeyPath)
	v.SetDefault("encryption.private-key", defaultPrivateKeyPath)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	err = v.ReadConfig(f)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// SetLogger sets the logger used by the package.
func SetLogger(logger *logrus.Entry) {
	log = logger
}

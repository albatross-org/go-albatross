package albatross

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var log *logrus.Entry = logrus.New().WithField("prefix", "albatross")

// Config holds configuration options for an Albatross store.
type Config struct {
	// Path contains the path to the Albatross store.
	Path string `yaml:"path"`

	// DateFormat is a Go-formatted date that is used when parsing the front matter of entries.
	// By default, it is "2006-01-02 15:04".
	DateFormat string `yaml:"date-format"`

	// TagPrefix is the prefix used to identify a tag in an entry. By default, this is "@?".
	TagPrefix string `yaml:"tag-prefix"`

	// UseGit specifies whether the store should use Git.
	// By default, it is true.
	UseGit bool `yaml:"use-git"`

	// Encryption contains the configuration for how the store is encrypted.
	Encryption *EncryptionConfig `yaml:"encryption"`
}

// EncryptionConfig holds configuration info for the encryption functionality.
// Most of the time, this is part of an albatross.Config.
type EncryptionConfig struct {
	PublicKey  string `yaml:"public-key"`
	PrivateKey string `yaml:"private-key"`
}

// DefaultConfig contains all the default configuration options for a store.
var DefaultConfig = &Config{
	DateFormat: "2006-01-02 15:04",
	TagPrefix:  "@?",
	Encryption: &EncryptionConfig{
		PublicKey:  filepath.Join(getConfigDir(), "albatross", "keys", "public.key"),
		PrivateKey: filepath.Join(getConfigDir(), "albatross", "keys", "private.key"),
	},
}

// NewConfig returns a new configuration struct with default values supplied.
func NewConfig() *Config {
	// We can't just return the DefaultConfig since it's a pointer -- we don't want to run
	// the risk of actually mutating the DefaultConfig rather than just a copy of it.
	return &Config{
		Path: "", // Can't set a default path.

		DateFormat: DefaultConfig.DateFormat,
		TagPrefix:  DefaultConfig.TagPrefix,
		UseGit:     true,
		Encryption: &EncryptionConfig{
			PublicKey:  DefaultConfig.Encryption.PublicKey,
			PrivateKey: DefaultConfig.Encryption.PrivateKey,
		},
	}
}

// setDefaults sets all the default values in a config.
func (c *Config) setDefaults() {
	if c.DateFormat == "" {
		c.DateFormat = DefaultConfig.DateFormat
	}

	if c.TagPrefix == "" {
		c.TagPrefix = DefaultConfig.TagPrefix
	}

	// TODO: need to flip this around to c.DisableGit because it's impossible to distinguish between c.UseGit being uninitialised or
	// purposefully being set to false.
	if c.UseGit == false {
		c.UseGit = true
	}

	if c.Encryption == nil {
		c.Encryption = &EncryptionConfig{}
	}

	if c.Encryption.PublicKey == "" {
		c.Encryption.PublicKey = DefaultConfig.Encryption.PublicKey
	}

	if c.Encryption.PrivateKey == "" {
		c.Encryption.PrivateKey = DefaultConfig.Encryption.PrivateKey
	}
}

// parseTopLevelConfig takes the path to a config file and returns a map of store names to their configurations.
func parseTopLevelConfig(path string) (map[string]*Config, error) {
	configs := make(map[string]*Config)

	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read the config file %s: %w", path, err)
	}

	err = yaml.Unmarshal(fileBytes, &configs)
	if err != nil {
		return nil, err
	}

	for _, c := range configs {
		c.setDefaults()
	}

	return configs, nil
}

// parseConfig takes the path to a config file which contains the configuration information for only one store.
func parseConfig(path string) (*Config, error) {
	config := Config{}

	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read the config file %s: %w", path, err)
	}

	err = yaml.Unmarshal(fileBytes, &config)
	if err != nil {
		return nil, err
	}

	config.setDefaults()

	return &config, nil
}

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

// SetLogger sets the logger used by the package.
func SetLogger(logger *logrus.Entry) {
	log = logger
}

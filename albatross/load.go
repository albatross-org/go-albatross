package albatross

import (
	"path/filepath"
)

// Load returns a new Albatross store representation given the name of a store and the path
// to its configuration.
// If the name is set, the path to the config is treated as a top-level config which is a map of store names,
// such as "default", "phd" and "testing" to individual configurations. The name given will be used to pick
// the corresponding configuration.
// If the name is not set, the path to the config is treated as a standalone config for one particular store.
// For example:
//   albatross.Load("default", "/home/david/.config/albatross/config.yaml")
// If "default" is a store located at "/home/david/documents/albatross", you could instead load the store like this,
// assuming a config file is present in the root of the store:
//   albatross.Load("", "/home/david/documents/albatross/config.yaml")
// If configPath is unset, it defaults to "~/.config/albatross/config.yaml"
func Load(name, configPath string) (*Store, error) {
	if configPath == "" {
		configPath = filepath.Join(getConfigDir(), "albatross", "config.yaml")
	}

	var config *Config
	var err error

	if name == "" {
		config, err = parseConfig(configPath)
		if err != nil {
			return nil, err
		}

		// If the path isn't specified in this config file, use the folder that
		// the store is in as the path.
		if config.Path == "" {
			config.Path = filepath.Dir(configPath)
		}
	} else {
		configs, err := parseTopLevelConfig(configPath)
		if err != nil {
			return nil, err
		}

		config = configs[name]

		if config == nil {
			return nil, ErrLoadStoreInvalid{name, configPath}
		}
	}

	return FromConfig(config)
}

// FromConfig returns a new Albatross store representation from a Config struct.
func FromConfig(config *Config) (*Store, error) {
	s := &Store{
		Path:   config.Path,
		Config: config,

		entriesPath:     filepath.Join(config.Path, "entries"),
		configPath:      filepath.Join(config.Path, "config.yaml"),
		attachmentsPath: filepath.Join(config.Path, "attachments"),
		gitPath:         filepath.Join(config.Path, "entries", ".git"),

		disableGit: !config.UseGit, // TODO: flip this around
	}

	encrypted, err := s.Encrypted()
	if err != nil {
		return nil, err
	}

	if !encrypted {
		err = s.load()
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// StoreOptions returns all the options of stores in the top level config file
func StoreOptions(configPath string) []string {
	configs, err := parseTopLevelConfig(configPath)
	if err != nil {
		return []string{}
	}

	options := []string{}
	for name := range configs {
		options = append(options, name)
	}

	return options
}

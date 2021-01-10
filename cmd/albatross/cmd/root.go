package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	albatross "github.com/albatross-org/go-albatross/pkg/core"

	_ "net/http/pprof"
)

var cfgFile string
var logLvl string
var leaveDecrypted bool
var disableGit bool

var storeName string
var storePath string

var store *albatross.Store
var globalLog *logrus.Logger
var log *logrus.Entry

var pprof bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "albatross",
	Short: "Albatross is a note-taking and journalling application.",
	Long: `Albatross is a distributed note-taking and journalling application, optimised for usage by a single
individual as a secure place for networked thoughts, ideas and information. You could think of it as a Zettelkasten
or a memex.

Based on a simple format that is plain text first and not reliant on any proprietary third party apps or software,
Albatross makes the guarantee that your personal data will be safe and accessible with time.

This program is a command line tool for interfacing with Albatross stores in a composable and succinct way.

	$ albatross decrypt
	$ albatross create food/pizza
	$ albatross get --path food/pizza --update

Setup
-----

See the README, https://github.com/albatross-org/go-albatross/albatross as a guide on how to setup an Albatross store.

Basic Usage
-----------

	# Create an entry
	$ albatross create food/pizza

	# Update the entry
	$ albatross get --path food/pizza update

	# Encrypt the store using your public and private key
	$ albatross encrypt

	# Decrypt the store
	$ albatross decrypt

	# Search for all entries with a given tag
	$ albatross get --tag "@!public"

Entries
-------

Entries are Markdown files with a YAML frontmatter.

	---
	title: "My Entry"
	date: "2020-10-26 16:03"
	---

	This is an example of an entry.

If the title isn't specified, the first sentence is used as the title. If the date isn't specified, the modification
time of the file is used.

Links
-----

You can link to other entries using two different syntaxes:

	- {{path/to/entry}}
	- [[My Entry Title]]

More Help
---------

See the README: https://github.com/albatross-org/go-albatross/tree/master/cmd/albatross
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Store '%s':\n", storeName)
		fmt.Println("  Path:", storePath)

		encrypted, err := store.Encrypted()
		if err != nil {
			log.Fatal(err)
		}

		if encrypted {
			fmt.Println("  Encrypted: yes")
			os.Exit(0)
		} else {
			fmt.Println("  Encrypted: no")

			collection, err := store.Collection()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("  Entries:", collection.Len())
			fmt.Println("  Using Git:", store.UsingGit())
			fmt.Println("")
		}

		err = cmd.Usage()
		if err != nil {
			log.Fatal(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if pprof {
		log.Println("Starting profilling server on localhost:6060...")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}
}

func init() {
	cobra.OnInitialize(initLogging, initConfig, initStore)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/albatross/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLvl, "level", "info", "logging level (trace, debug, info, warning, error, fatal, panic)")
	rootCmd.PersistentFlags().StringVar(&storeName, "store", "default", "store to use, as defined in config file (e.g. default, thesis)")
	rootCmd.PersistentFlags().BoolVarP(&leaveDecrypted, "leave-decrypted", "l", false, "whether to leave the store decrypted or encrypt it again after decrypting it")
	rootCmd.PersistentFlags().BoolVarP(&disableGit, "disable-git", "d", false, "don't use git for version control (mainly used when you want to make commits by hand)")
	rootCmd.PersistentFlags().BoolVar(&pprof, "pprof", false, "after the command has executed, start a pprof server on port 6060")
}

// getConfigDirectory gets the configuration directory that should be used for the program.
// It uses $XDG_CONFIG_HOME/albatross if set and defaults to $HOME/.config/albatross otherwise.
// TODO: make this cross-platform
func getConfigDirectory() string {
	var dir string

	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		dir = filepath.Join(home, ".config", "albatross")
	} else {
		dir = filepath.Join(xdg, "albatross")
	}

	return dir
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		configDir := getConfigDirectory()

		// Search config in home directory with name ".albatross" (without extension).
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
	}

	viper.SetEnvPrefix("ALBATROSS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Using config file:", viper.ConfigFileUsed())
	}
}

// initStore sets the store using the configuration the program has.
func initStore() {
	storePath = viper.GetString(fmt.Sprintf("%s.path", storeName))
	if storePath == "" {
		fmt.Printf("Couldn't find path for store '%s'.\n", storeName)
		fmt.Printf("Make sure you have an path in your config file for that store, something like:\n\n")

		fmt.Printf("%s:\n", storeName)
		fmt.Printf("\tpath: /path/to/the/store\n\n")

		os.Exit(1)
	}

	log.Debugf(
		"Using store named '%s', located at: %s",
		storeName,
		storePath, // This really doesn't seem ideal.
	)

	var err error

	start := time.Now()
	store, err = albatross.Load(storePath)
	if err != nil {
		log.Fatal(err)
	}
	end := time.Now()

	if globalLog.IsLevelEnabled(logrus.DebugLevel) {
		log.Debugf("Parsing store took %s.", end.Sub(start))
	}

	if disableGit {
		store.DisableGit()
	}
}

// initLogging initialises the logger.
func initLogging() {
	lvl, err := logrus.ParseLevel(logLvl)
	if err != nil {
		log.Fatalf("Invalid log level '%s'\nPlease choose from: trace, debug, info, warning, error, fatal, panic", logLvl)
	}

	globalLog = logrus.New()
	globalLog.SetLevel(lvl)
	globalLog.Formatter = new(prefixed.TextFormatter)

	log = globalLog.WithField("prefix", "albatross")
	albatross.SetLogger(log)
}

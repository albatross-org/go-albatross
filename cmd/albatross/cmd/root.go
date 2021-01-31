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

	"github.com/albatross-org/go-albatross/albatross"

	// Used for profiling purposes.
	_ "net/http/pprof"
)

var cfgFile string
var logLvl string
var leaveDecrypted bool
var disableGit bool

var storeLocation string
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
		fmt.Printf("Store '%s':\n", storeLocation)
		fmt.Println("  Path:", store.Path)

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
	cobra.OnInitialize(initLogging, initStore)

	// Global flags.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", getDefaultConfigPath(), "path to top-level config file")
	rootCmd.PersistentFlags().StringVar(&logLvl, "level", "info", "logging level (trace, debug, info, warning, error, fatal, panic)")
	rootCmd.PersistentFlags().StringVar(&storeLocation, "store", "default", "store to use, either as defined in config file (e.g. default, thesis) or the path to a store")
	rootCmd.PersistentFlags().BoolVarP(&leaveDecrypted, "leave-decrypted", "l", false, "whether to leave the store decrypted or encrypt it again after decrypting it")
	rootCmd.PersistentFlags().BoolVarP(&disableGit, "disable-git", "d", false, "don't use git for version control (mainly used when you want to make commits by hand)")

	// Special hidden flags for development purposes.
	rootCmd.PersistentFlags().BoolVar(&pprof, "pprof", false, "after the command has executed, start a pprof server on port 6060")
	rootCmd.PersistentFlags().MarkHidden("pprof")
}

// getDefaultConfigPath gets the default configuration path that should be used for the program.
// It uses $XDG_CONFIG_HOME/albatross/config.yaml if set and defaults to $HOME/.config/albatross/config.yaml otherwise.
// TODO: make this cross-platform
func getDefaultConfigPath() string {
	var dir string

	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		dir = filepath.Join(home, ".config", "albatross", "config.yaml")
	} else {
		dir = filepath.Join(xdg, "albatross", "config.yaml")
	}

	return dir
}

// initStore sets the store using the configuration the program has.
// There's three cases here:
//  1. storeLocation is the name of a store, like "default" or "phd".
//     In this case we need to "resolve" the name of the store by using the top-level config as defined in cfgFile.
//     This always takes precedence over it being a path -- e.g. default could be the name of the store or the relative
//     path to a store in the current directory.
//
//  2. storeLocation is the path to a config file.
//     In this case, we need to load the store from that config file.
//
//  3. storeLocation is the path to a store.
//     In this case, if the store contains a config.yaml file, we need to use that. Otherwise we need to just use
//     the default config.
func initStore() {
	log.Debugf(
		"Store location specified as '%s'",
		storeLocation,
	)

	start := time.Now()
	var err error

	// First attempt to load the store as if it definitely is in the top-level config. If this errors, we can
	// test for the other two cases.
	store, err = albatross.Load(storeLocation, cfgFile)
	if err == nil {
		if disableGit {
			store.DisableGit()
		}

		return
	} else if _, ok := err.(albatross.ErrLoadStoreInvalid); !ok {
		log.Fatal(err)
	}

	// If we're here in the code, it's either Case 2 or Case 3.

	if !exists(storeLocation) {
		fmt.Printf("Can't figure out what is meant by store '%s'. It is not:\n\n", storeLocation)
		fmt.Printf("- The name of a store defined in the top-level config %s.\n", cfgFile)
		fmt.Printf("- The path to a config file.\n")
		fmt.Printf("- The path to a store itself.\n")

		options := albatross.StoreOptions(cfgFile)
		if len(options) > 0 {
			fmt.Printf("\nAvailable options are: %s\n", strings.Join(options, ", "))
		}

		os.Exit(1)
	}

	// Test if the path ends in config.yaml to see if it's a path to
	// a config.
	if strings.HasSuffix(storeLocation, "config.yaml") {
		store, err = albatross.Load("", storeLocation)
		if err != nil {
			log.Fatal(err)
		}

		if disableGit {
			store.DisableGit()
		}

		return
	}

	// Check if it's a store -- this is a very simple check, just seeing if
	// the folder given contains an "entries" folder.
	if !exists(filepath.Join(storeLocation, "entries")) {
		fmt.Printf("Can't figure out what is meant by store '%s':\n\n", storeLocation)
		fmt.Printf("- The name of a store defined in the top-level config %s.\n", cfgFile)
		fmt.Printf("- The path to a config file.\n")
		fmt.Printf("- The path to a store itself.\n")

		options := albatross.StoreOptions(cfgFile)
		if len(options) > 0 {
			fmt.Printf("\nAvailable options are: %s\n", strings.Join(options, ", "))
		}

		os.Exit(1)
	}

	// If a config file is present in the current directory, use that one.
	if exists(filepath.Join(storeLocation, "config.yaml")) {
		store, err = albatross.Load("", filepath.Join(storeLocation, "config.yaml"))
		if err != nil {
			log.Fatal(err)
		}

		if disableGit {
			store.DisableGit()
		}

		return
	}

	// Otherwise use a default config instead.
	defaultConfig := albatross.NewConfig()
	defaultConfig.Path = storeLocation

	store, err = albatross.FromConfig(defaultConfig)
	if err != nil {
		log.Fatal(err)
	}

	if disableGit {
		store.DisableGit()
	}

	end := time.Now()

	log.Debugf("Parsing store took %s.", end.Sub(start))

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

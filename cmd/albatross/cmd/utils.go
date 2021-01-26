package cmd

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// getEditor gets the $EDITOR environment variable, defaulting to the argument specified if none has been set.
func getEditor(def string) string {
	env := os.Getenv("EDITOR")
	if env == "" {
		return def
	}

	return env
}

// checkArg checks an error returned by a call to cmd.Flags().Get and prints an error if it fails.
func checkArg(err error) {
	if err != nil {
		fmt.Println("Can't get argument:")
		fmt.Println(err)
		os.Exit(1)
	}
}

// checkArgVerbose checks an error returned by a call to cmd.Flags().Get and prints a detailed error if it fails.
func checkArgVerbose(cmd *cobra.Command, flag string, err error) {
	if err != nil {
		fmt.Printf("Can't get argument %s for command %s:\n", flag, cmd.Name())
		fmt.Println(err)
		os.Exit(1)
	}
}

// tempFile returns the path to a temporary file for editing, initialised with the content specified.
func tempFile(content string) (path string, cleanup func(), err error) {
	f, err := ioutil.TempFile("", "albatross*.md")
	if err != nil {
		return "", func() {}, err
	}

	_, err = f.Write([]byte(content))
	if err != nil {
		return "", func() {}, err
	}

	return f.Name(), func() {
		os.Remove(f.Name())
	}, nil
}

// edit will open an editor and let them edit the content specified. It will return the new content.
func edit(editor string, content string) (string, error) {
	path, cleanup, err := tempFile(content)
	if err != nil {
		return "", err
	}
	defer cleanup()

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	newContent, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(newContent), nil
}

// hashString is a shorthand for a doing the SHA-1 hash of a string.
func hashString(path string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(path))
	return fmt.Sprintf("%x.xhtml", h.Sum(nil))
}

// confirmPrompt displays a prompt `s` to the user and returns a bool indicating yes / no
// If the lowercased, trimmed input begins with anything other than 'y', it returns false
// Courtesy https://gist.github.com/r0l1/3dcbb0c8f6cfe9c66ab8008f55f8f28b
func confirmPrompt(s string) bool {
	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		res, err := r.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		// Empty input (i.e. "\n")
		if len(res) < 2 {
			continue
		}

		switch strings.ToLower(strings.TrimSpace(res))[0] {
		case 'y':
			return true
		case 'n':
			return false
		default:
			fmt.Println("Please enter [y/n].")
		}
	}
}

// commandExists checks if a command exists.
// Courtesy: https://gist.github.com/miguelmota/ed4ec562b8cd1781e7b20151b37de8a0
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// letterBytes are the letters used to generate a random string.
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// randomString generates a string consisting of characters from letterBytes that is n characters long.
// Courtesy: https://stackoverflow.com/a/31832326
func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func init() {
	// Seed the random number generator.
	rand.Seed(time.Now().UnixNano())
}

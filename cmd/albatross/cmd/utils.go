package cmd

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	h.Write([]byte(path))
	return fmt.Sprintf("%x.xhtml", h.Sum(nil))
}

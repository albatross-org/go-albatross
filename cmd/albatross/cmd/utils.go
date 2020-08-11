package cmd

import (
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

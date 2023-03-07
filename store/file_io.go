package store

import (
	"os"
	"path/filepath"
)

// Replaces the contents of file `filename` with `contents`.
//
// Creates file `filename` if it did not exist before.
func overwriteFile(filename string, contents string) {
	// Make the file's parent directory (no-op if directory already exists).
	dir := filepath.Dir(filename)
	os.MkdirAll(dir, os.ModePerm)

	// Delete the file if it already exists.
	_, err := os.Open(filename)
	if err == nil {
		os.Remove(filename)
	}

	newFile, _ := os.Create(filename)
	newFile.WriteString(contents)
	newFile.Close()
}

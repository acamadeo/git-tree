package store

import (
	"os"
	"path/filepath"
	"strings"
)

// Replaces the contents of file `filename` with `contents`.
//
// Creates file `filename` if it did not exist before.
func OverwriteFile(filename string, contents string) {
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

// Return contents of a file as a string.
func ReadFile(filename string) string {
	bytes, _ := os.ReadFile(filename)
	return string(bytes)
}

// Add a newline and append `contents` to the file.
func AppendToFile(filename string, contents string) {
	existingContents := ReadFile(filename)
	OverwriteFile(filename, existingContents+"\n"+contents)
}

// Returns true if the file exists.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// Returns true if the file is empty.
func FileEmpty(filename string) bool {
	return ReadFile(filename) == ""
}

// Returns true if the directory exists.
func DirExists(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil
}

// Returns true if the file contains the given line exactly.
func FileContainsLine(filename string, line string) bool {
	contents := ReadFile(filename)
	for _, l := range strings.Split(contents, "\n") {
		if l == line {
			return true
		}
	}
	return false
}

// Removes any lines matching `line` in the file.
func DeleteLineInFile(filename string, line string) {
	lines := []string{}

	contents := ReadFile(filename)
	for _, l := range strings.Split(contents, "\n") {
		if l == line {
			continue
		}
		lines = append(lines, l)
	}

	OverwriteFile(filename, strings.Join(lines, "\n"))
}

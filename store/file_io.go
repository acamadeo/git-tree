package store

import "os"

// Replaces the contents of file `filepath` with `contents`.
//
// Creates file `filepath` if it did not exist before.
func overwriteFile(filepath string, contents string) {
	_, err := os.Open(filepath)

	// Delete the file if it already exists
	if err == nil {
		os.Remove(filepath)
	}

	newFile, _ := os.Create(filepath)
	newFile.WriteString(contents)
	newFile.Close()
}

// TODO: We'll probably want a read util that we can use in
//  - ReadBranchMap()
//  - ReadObsolescenceMap()

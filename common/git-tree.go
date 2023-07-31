package common

import "github.com/abaresk/git-tree/store"

// Returns true if `git-tree init` has been run.
func GitTreeInited(gitPath string) bool {
	// A branch map file should exist if git-tree has been initialized.
	return store.FileExists(BranchMapPath(gitPath))
}

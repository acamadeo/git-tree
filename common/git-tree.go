package common

import "os"

// Returns true if `git-tree init` has been run.
func GitTreeInited(gitPath string) bool {
	// A branch map file should exist if git-tree has been initialized.
	if _, err := os.Stat(BranchMapPath(gitPath)); err == nil {
		return true
	}
	return false
}

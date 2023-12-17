package common

import (
	"github.com/acamadeo/git-tree/models"
	"github.com/acamadeo/git-tree/utils"
	git "github.com/libgit2/git2go/v34"
)

// Returns true if `git-tree init` has been run.
func GitTreeInited(gitPath string) bool {
	// A branch map file should exist if git-tree has been initialized.
	return utils.FileExists(BranchMapPath(gitPath))
}

// Returns true if you are on the a tip commit (i.e., a commit referenced by a
// branch).
func OnTipCommit(repo *git.Repository, branchMap *models.BranchMap) bool {
	// Check whether HEAD points to any of the commits pointed to by any of the
	// branches we are tracking.
	//
	// We do this instead of checking whether HEAD is detached because HEAD is
	// often detached from the tree navigation commands.
	head, _ := repo.Head()

	for branch := range branchMap.Children {
		if branch.Cmp(head) == 0 {
			return true
		}
	}
	return false
}

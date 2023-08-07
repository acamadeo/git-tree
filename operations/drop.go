package operations

import (
	"fmt"
	"os"

	"github.com/acamadeo/git-tree/common"
	"github.com/acamadeo/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

// Remove git-tree tracking for the given repository.
func Drop(repo *git.Repository) error {
	// Read the branch map file.
	branchMapPath := common.BranchMapPath(repo.Path())
	branchMap := store.ReadBranchMap(repo, branchMapPath)

	// Delete the root branch created by `git-tree init`.
	if err := branchMap.Root.Delete(); err != nil {
		return fmt.Errorf("Could not delete root branch: %s.", err.Error())
	}

	// Delete local git-tree storage (i.e. the branch map and obsolescence map
	// files).
	gitTreePath := common.GitTreeSubdirPath(repo.Path())
	if err := os.RemoveAll(gitTreePath); err != nil {
		return fmt.Errorf("Could not delete git-tree files: %s.", err.Error())
	}

	return nil
}

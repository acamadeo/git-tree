package operations

import (
	"fmt"

	"github.com/acamadeo/git-tree/common"
	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/models"
	"github.com/acamadeo/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

// Initialize git-tree for the given repository.
//
// Track the specified branches, or all branches if none were specified.
func Init(repo *git.Repository, branches ...*git.Branch) error {
	// If there were no branches passed in, initialize with all local branches.
	if len(branches) == 0 {
		branches = gitutil.AllLocalBranches(repo)
	}

	// Create the root branch as the most-common ancestor of the provided
	// branches.
	rootBranch, err := createRootBranch(repo, branches)
	if err != nil {
		return fmt.Errorf("Could not create temporary root branch: %s.", err.Error())
	}

	// Construct a branch map from the branches and store the branch map in our
	// file.
	branchMap := models.BranchMapFromRepo(repo, rootBranch, branches)
	store.WriteBranchMap(branchMap, common.BranchMapPath(repo.Path()))

	return nil
}

func createRootBranch(repo *git.Repository, branches []*git.Branch) (*git.Branch, error) {
	var rootOid *git.Oid
	if len(branches) == 1 {
		rootOid = branches[0].Target()
	} else {
		// Find the commit that will serve as the root of the git-tree. Create a new
		// branch pointed to this commit.
		rootOid, _ = repo.MergeBaseMany(branchOids(branches))
	}

	rootCommit, _ := repo.LookupCommit(rootOid)
	return repo.CreateBranch(common.GitTreeRootBranch, rootCommit, false)
}

func branchOids(branches []*git.Branch) []*git.Oid {
	oids := []*git.Oid{}
	for _, branch := range branches {
		oids = append(oids, branch.Target())
	}
	return oids
}

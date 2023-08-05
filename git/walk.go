package gitutil

import git "github.com/libgit2/git2go/v34"

// Create a walk that will include the tip commit of every local branch.
func InitWalkWithAllBranches(repo *git.Repository) *git.RevWalk {
	revWalk, _ := repo.Walk()
	revWalk.Sorting(git.SortTopological)
	for _, branch := range AllLocalBranches(repo) {
		tipCommitOid := branch.Target()
		revWalk.Push(tipCommitOid)
	}
	return revWalk
}

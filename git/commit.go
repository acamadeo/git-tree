package gitutil

import (
	"encoding/hex"

	git "github.com/libgit2/git2go/v34"
)

func CommitByReference(repo *git.Repository, ref *git.Reference) *git.Commit {
	commit, _ := repo.LookupCommit(ref.Target())
	return commit
}

func AllLocalCommits(repo *git.Repository) []*git.Commit {
	// Create a walk that will include the tip commit of every local branch.
	revWalk := InitWalkWithAllBranches(repo)

	// Perform the walk, creating a set of every commit oid.
	commitOidsSet := map[git.Oid]bool{}
	revWalk.Iterate(func(commit *git.Commit) bool {
		commitOidsSet[*commit.Id()] = true
		return true
	})

	// Lookup each commit in the set.
	commits := []*git.Commit{}
	for oid := range commitOidsSet {
		commit, _ := repo.LookupCommit(&oid)
		commits = append(commits, commit)
	}
	return commits
}

func CommitShortHash(commit *git.Commit) string {
	oid := commit.Id()
	oidString := hex.EncodeToString(oid[:])
	return oidString[:shortHashLength]
}

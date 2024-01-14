package gitutil

import (
	"encoding/hex"
	"fmt"

	git "github.com/libgit2/git2go/v34"
)

func CommitByOid(repo *git.Repository, oid git.Oid) *git.Commit {
	commit, _ := repo.LookupCommit(&oid)
	return commit
}

func CommitByReference(repo *git.Repository, ref *git.Reference) *git.Commit {
	commit, _ := repo.LookupCommit(ref.Target())
	return commit
}

// Returns a list of all local commits up to and including the `root`.
//
// If `root` is nil, it returns a list of all local commits in the entire
// history.
func AllLocalCommits(repo *git.Repository, root *git.Branch) []*git.Commit {
	return LocalCommitsFromBranches(repo, root, AllLocalBranches(repo)...)
}

// Returns a list of all commits up to and including the `root` that are ancestors
// of the provided `branches`.
//
// If `root` is nil, the result will include commits from the entire history.
func LocalCommitsFromBranches(repo *git.Repository, root *git.Branch, branches ...*git.Branch) []*git.Commit {
	// Create a walk that will include the tip commit of each provided branch.
	revWalk := InitWalkWithBranches(repo, branches...)

	// Perform the walk, creating a set of every commit oid.
	commitOidsSet := map[git.Oid]bool{}
	revWalk.Iterate(func(commit *git.Commit) bool {
		commitOidsSet[*commit.Id()] = true

		// Stop adding commits once we hit `root` (if specified).
		return root == nil || *root.Target() != *commit.Id()
	})

	// Lookup each commit in the set.
	commits := []*git.Commit{}
	for oid := range commitOidsSet {
		commit, _ := repo.LookupCommit(&oid)
		commits = append(commits, commit)
	}
	return commits
}

// Returns true if commit `a` is an ancestor of commit `b`.
func IsAncestor(repo *git.Repository, a *git.Commit, b *git.Commit) bool {
	ancestorOid, _ := repo.MergeBase(a.Id(), b.Id())
	return a.Id().Equal(ancestorOid)
}

func CreateReferenceAtCommit(repo *git.Repository, commit *git.Commit, name string) *git.Reference {
	logMessage := fmt.Sprintf(
		"Created reference %q pointing to commit %s",
		name, CommitShortHash(commit))
	ref, _ := repo.References.Create("refs/heads/"+name, commit.Id(), false, logMessage)
	return ref
}

func CommitShortHash(commit *git.Commit) string {
	oid := commit.Id()
	oidString := hex.EncodeToString(oid[:])
	return oidString[:shortHashLength]
}

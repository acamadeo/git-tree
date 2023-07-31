package gitutil

import (
	"encoding/hex"

	git "github.com/libgit2/git2go/v34"
)

func CommitByReference(repo *git.Repository, ref *git.Reference) *git.Commit {
	commit, _ := repo.LookupCommit(ref.Target())
	return commit
}

func CommitShortHash(commit *git.Commit) string {
	oid := commit.Id()
	oidString := hex.EncodeToString(oid[:])
	return oidString[:shortHashLength]
}

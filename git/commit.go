package gitutil

import (
	"encoding/hex"

	"github.com/abaresk/git-tree/models"
	git "github.com/libgit2/git2go/v34"
)

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

func CommitByReference(repo *git.Repository, ref *git.Reference) *git.Commit {
	commit, _ := repo.LookupCommit(ref.Target())
	return commit
}

func CommitShortHash(commit *git.Commit) string {
	oid := commit.Id()
	oidString := hex.EncodeToString(oid[:])
	return oidString[:shortHashLength]
}

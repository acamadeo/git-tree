package operations

import (
	"os"
	"strings"

	"github.com/acamadeo/git-tree/common"
	"github.com/acamadeo/git-tree/models"
	"github.com/acamadeo/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

// KNOWN ISSUE:
//   When you amend a commit during a split rebase (via `edit`), the rebase
//   entries appears under `action amend` instead of `action rebase`.
//
//   Perhaps we should check whether there is a rebase in progress before
//   creating a new action in the `pre-commit` hook?

// NOTE TO SELF: If there's extraneous entries under certain actions (e.g.
// post-commit entries in rebase action), don't include them.

// -------------------------------------------------------------------------- \
// ObsoletePreRebase                                                          |
// -------------------------------------------------------------------------- /

func ObsoletePreRebase(repo *git.Repository) error {
	// Add a new Obsolescence Action with the Rebase ActionType.
	obsmapFile := common.ObsoleteMapPath(repo.Path())
	store.AppendObsolescenceAction(repo, obsmapFile, models.ActionTypeRebase)
	return nil
}

// -------------------------------------------------------------------------- \
// ObsoletePostRewriteAmend                                                   |
// -------------------------------------------------------------------------- /

func ObsoletePostRewriteAmend(repo *git.Repository, lines []string) error {
	if err := validateObsoleteLines(repo, lines); err != nil {
		return err
	}

	// Commit actions and Amend actions both start with a `pre-commit` hook. If we
	// receive `post-rewrite.amend` after `pre-commit`, we know this is an Amend
	// action, not a Commit action.
	obsmapFile := common.ObsoleteMapPath(repo.Path())
	if store.LastObsolescenceActionType(repo, obsmapFile) == models.ActionTypeCommit {
		store.SetLastObsolescenceActionType(repo, obsmapFile, models.ActionTypeAmend)
	}

	return appendEntriesToObsoleteMap(repo, lines, models.PostRewriteAmend)
}

// -------------------------------------------------------------------------- \
// ObsoletePostRewriteRebase                                                  |
// -------------------------------------------------------------------------- /

func ObsoletePostRewriteRebase(repo *git.Repository, lines []string) error {
	if err := validateObsoleteLines(repo, lines); err != nil {
		return err
	}
	return appendEntriesToObsoleteMap(repo, lines, models.PostRewriteRebase)
}

// -------------------------------------------------------------------------- \
// ObsoletePreCommit                                                          |
// -------------------------------------------------------------------------- /

func ObsoletePreCommit(repo *git.Repository) error {
	// Store the parent of the commit at HEAD, or "null" if HEAD is at the
	// initial commit.
	headRef, _ := repo.Head()
	headCommit, _ := repo.LookupCommit(headRef.Target())

	headParent := "null"
	if headCommit.ParentCount() > 0 {
		headParent = headCommit.ParentId(0).String()
	}
	store.OverwriteFile(common.PreCommitParentPath(repo.Path()), headParent)

	// Add a new Obsolescence Action. We assume it's a Commit action by default.
	// If it was an Amend action, we'll modify the type of this action when the
	// `post-rewrite.amend` hook fires.
	obsmapFile := common.ObsoleteMapPath(repo.Path())
	store.AppendObsolescenceAction(repo, obsmapFile, models.ActionTypeCommit)
	return nil
}

// -------------------------------------------------------------------------- \
// ObsoletePostCommit                                                         |
// -------------------------------------------------------------------------- /

// TODO: Only obsolete commits if the commit has descendants. Come back to this
// after creating the RepoTree struct.
func ObsoletePostCommit(repo *git.Repository) error {
	headRef, _ := repo.Head()
	headCommit, _ := repo.LookupCommit(headRef.Target())

	preCommitParentPath := common.PreCommitParentPath(repo.Path())
	preCommitParent := store.ReadFile(preCommitParentPath)
	defer os.Remove(preCommitParentPath)

	// If the parent of HEAD are the same pre- and post-commit, this was
	// `git commit --amend`. Ignore.
	headParent := "null"
	if headCommit.ParentCount() > 0 {
		headParent = headCommit.ParentId(0).String()
	}
	if headParent == preCommitParent {
		return nil
	}

	// Mark the commit at HEAD as obsoleting its parent.
	entry := models.ObsolescenceEntry{
		Commit:    headCommit.Parent(0),
		Obsoleter: headCommit,
		HookType:  models.PostCommit,
	}

	obsmapFile := common.ObsoleteMapPath(repo.Path())
	return store.AppendEntriesToLastObsolescenceAction(repo, obsmapFile, entry)
}

func validateObsoleteLines(repo *git.Repository, lines []string) error {
	for _, line := range lines {
		lineParts := strings.Fields(line)
		oldOid, _ := git.NewOid(lineParts[0])
		newOid, _ := git.NewOid(lineParts[1])

		if _, err := repo.LookupCommit(oldOid); err != nil {
			return err
		}
		if _, err := repo.LookupCommit(newOid); err != nil {
			return err
		}
	}
	return nil
}

func appendEntriesToObsoleteMap(repo *git.Repository, lines []string, hookType models.HookType) error {
	obsmapEntries := []models.ObsolescenceEntry{}

	for _, line := range lines {
		lineParts := strings.Fields(line)
		oldHash := lineParts[0]
		newHash := lineParts[1]

		commitOid, _ := git.NewOid(oldHash)
		commit, _ := repo.LookupCommit(commitOid)

		obsoleterOid, _ := git.NewOid(newHash)
		obsoleter, _ := repo.LookupCommit(obsoleterOid)

		obsmapEntries = append(obsmapEntries, models.ObsolescenceEntry{
			Commit:    commit,
			Obsoleter: obsoleter,
			HookType:  hookType,
		})
	}

	obsmapFile := common.ObsoleteMapPath(repo.Path())
	return store.AppendEntriesToLastObsolescenceAction(repo, obsmapFile, obsmapEntries...)
}

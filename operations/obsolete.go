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
//   entries appears under `event amend` instead of `event rebase`.
//
//   Perhaps we should check whether there is a rebase in progress before
//   creating a new event in the `pre-commit` hook?

// NOTE TO SELF: If there's extraneous entries under certain events (e.g.
// post-commit entries in rebase event), don't include them.

// -------------------------------------------------------------------------- \
// ObsoletePreRebase                                                          |
// -------------------------------------------------------------------------- /

func ObsoletePreRebase(repo *git.Repository) error {
	// Add a new Obsolescence Event with the Rebase EventType.
	obsmapFile := common.ObsoleteMapPath(repo.Path())
	store.AppendObsolescenceEvent(repo, obsmapFile, models.EventTypeRebase)
	return nil
}

// -------------------------------------------------------------------------- \
// ObsoleteAmend                                                              |
// -------------------------------------------------------------------------- /

func ObsoleteAmend(repo *git.Repository, lines []string) error {
	if err := validateObsoleteLines(repo, lines); err != nil {
		return err
	}

	// Commit events and Amend events both start with a `pre-commit` hook. If we
	// receive `post-rewrite.amend` after `pre-commit`, we know this is an Amend
	// event, not a Commit event.
	obsmapFile := common.ObsoleteMapPath(repo.Path())
	if store.LastObsolescenceEventType(repo, obsmapFile) == models.EventTypeCommit {
		store.SetLastObsolescenceEventType(repo, obsmapFile, models.EventTypeAmend)
	}

	return appendEntriesToObsoleteMap(repo, lines, models.PostRewriteAmend)
}

// -------------------------------------------------------------------------- \
// ObsoleteRebase                                                             |
// -------------------------------------------------------------------------- /

func ObsoleteRebase(repo *git.Repository, lines []string) error {
	if err := validateObsoleteLines(repo, lines); err != nil {
		return err
	}
	return appendEntriesToObsoleteMap(repo, lines, models.PostRewriteRebase)
}

// -------------------------------------------------------------------------- \
// ObsoletePreCommit                                                          |
// -------------------------------------------------------------------------- /

func ObsoletePreCommit(repo *git.Repository) error {
	// Store the parent of the commit at HEAD.
	headRef, _ := repo.Head()
	headCommit, _ := repo.LookupCommit(headRef.Target())

	contents := headCommit.ParentId(0).String()
	store.OverwriteFile(common.PreCommitParentPath(repo.Path()), contents)

	// Add a new Obsolescence Event. We assume it's a Commit event by default.
	// If it was an Amend event, we'll modify the type of this event when the
	// `post-rewrite.amend` hook fires.
	obsmapFile := common.ObsoleteMapPath(repo.Path())
	store.AppendObsolescenceEvent(repo, obsmapFile, models.EventTypeCommit)
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
	if headCommit.ParentId(0).String() == preCommitParent {
		return nil
	}

	// Mark the commit at HEAD as obsoleting its parent.
	entry := models.ObsolescenceEntry{
		Commit:    headCommit.Parent(0),
		Obsoleter: headCommit,
		HookType:  models.PostCommit,
	}

	obsmapFile := common.ObsoleteMapPath(repo.Path())
	return store.AppendEntriesToLastObsolescenceEvent(repo, obsmapFile, entry)
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
	return store.AppendEntriesToLastObsolescenceEvent(repo, obsmapFile, obsmapEntries...)
}

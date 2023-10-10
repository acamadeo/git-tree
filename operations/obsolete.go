package operations

import (
	"os"
	"strings"

	"github.com/acamadeo/git-tree/common"
	"github.com/acamadeo/git-tree/models"
	"github.com/acamadeo/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

// -------------------------------------------------------------------------- \
// ObsoleteAmend                                                              |
// -------------------------------------------------------------------------- /

func ObsoleteAmend(repo *git.Repository, lines []string) error {
	if err := validateObsoleteLines(repo, lines); err != nil {
		return err
	}
	return writeToObsoleteMap(repo, lines, models.PostRewriteAmend)
}

// -------------------------------------------------------------------------- \
// ObsoleteRebase                                                             |
// -------------------------------------------------------------------------- /

func ObsoleteRebase(repo *git.Repository, lines []string) error {
	if err := validateObsoleteLines(repo, lines); err != nil {
		return err
	}
	return writeToObsoleteMap(repo, lines, models.PostRewriteRebase)
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
	entry := models.ObsolescenceMapEntry{
		Commit:    headCommit.ParentId(0).String(),
		Obsoleter: headCommit.Id().String(),
		HookType:  models.PostCommit,
	}

	obsmapFile := common.ObsoleteMapPath(repo.Path())
	store.AppendToObsolescenceMap(obsmapFile, entry)
	return nil
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

func writeToObsoleteMap(repo *git.Repository, lines []string, hookType models.HookType) error {
	obsmapEntries := []models.ObsolescenceMapEntry{}

	for _, line := range lines {
		lineParts := strings.Fields(line)
		oldHash := lineParts[0]
		newHash := lineParts[1]

		obsmapEntries = append(obsmapEntries, models.ObsolescenceMapEntry{
			Commit:    oldHash,
			Obsoleter: newHash,
			HookType:  hookType,
		})
	}

	obsmapFile := common.ObsoleteMapPath(repo.Path())
	store.AppendToObsolescenceMap(obsmapFile, obsmapEntries...)
	return nil
}

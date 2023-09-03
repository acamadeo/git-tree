package operations

import (
	"strings"

	"github.com/acamadeo/git-tree/common"
	"github.com/acamadeo/git-tree/models"
	"github.com/acamadeo/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

// -------------------------------------------------------------------------- \
// ObsoleteAmend                                                             |
// -------------------------------------------------------------------------- /

func ObsoleteAmend(repo *git.Repository, lines []string) error {
	if err := validateObsoleteLines(repo, lines); err != nil {
		return err
	}
	return writeToObsoleteMap(repo, lines)
}

// -------------------------------------------------------------------------- \
// ObsoleteRebase                                                             |
// -------------------------------------------------------------------------- /

func ObsoleteRebase(repo *git.Repository, lines []string) error {
	if err := validateObsoleteLines(repo, lines); err != nil {
		return err
	}
	return writeToObsoleteMap(repo, lines)
}

// -------------------------------------------------------------------------- \
// ObsoleteCommit                                                             |
// -------------------------------------------------------------------------- /

// TODO: Only obsolete commits if the commit has descendants. Come back to this
// after creating the RepoTree struct.
func ObsoleteCommit(repo *git.Repository) error {
	// Mark the commit at HEAD as obsoleting its parent.
	headRef, _ := repo.Head()
	headCommit, _ := repo.LookupCommit(headRef.Target())

	entry := models.ObsolescenceMapEntry{
		Commit:    headCommit.ParentId(0).String(),
		Obsoleter: headCommit.Id().String(),
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

func writeToObsoleteMap(repo *git.Repository, lines []string) error {
	obsmapEntries := []models.ObsolescenceMapEntry{}

	for _, line := range lines {
		lineParts := strings.Fields(line)
		oldHash := lineParts[0]
		newHash := lineParts[1]

		obsmapEntries = append(obsmapEntries, models.ObsolescenceMapEntry{
			Commit:    oldHash,
			Obsoleter: newHash,
		})
	}

	obsmapFile := common.ObsoleteMapPath(repo.Path())
	store.AppendToObsolescenceMap(obsmapFile, obsmapEntries...)
	return nil
}

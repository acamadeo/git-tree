package operations

import (
	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/models"
)

// Reconcile any troubled commits within the repository.
func Evolve(repo *gitutil.RepoTree, obsmap *models.ObsolescenceMap) error {
	return nil
}

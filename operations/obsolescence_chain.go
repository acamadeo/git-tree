package operations

import (
	"sort"

	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/models"
	git "github.com/libgit2/git2go/v34"
)

// TODO: Consider moving into a dedicated `evolve/` directory!

// Represents the commits that obsoleted another set of commits in a Git action.
type obsolescenceChain struct {
	// Commits on the obsoleted side of the chain (from oldest to newest).
	obsoleted []*git.Commit
	// Commits on the obsoleter side of the chain (from oldest to newest).
	obsoleter []*git.Commit
}

func (oc *obsolescenceChain) HasObsoleteCommit(commit *git.Commit) bool {
	for _, c := range oc.obsoleted {
		if *c.Id() == *commit.Id() {
			return true
		}
	}
	return false
}

type obsolescenceChains []obsolescenceChain

// Find the obsolescence chain where the given commit got obsoleted.
//
// Returns nil if the given commit was not obsoleted.
func (ocs *obsolescenceChains) FindChainWithObsoleteCommit(commit *git.Commit) *obsolescenceChain {
	for _, chain := range *ocs {
		if chain.HasObsoleteCommit(commit) {
			return &chain
		}
	}
	return nil
}

func buildObsolescenceChains(repoTree *gitutil.RepoTree, obsmap *models.ObsolescenceMap) obsolescenceChains {
	chains := []obsolescenceChain{}
	for _, action := range obsmap.Actions {
		chains = append(chains, buildObsolescenceChain(repoTree, action))
	}
	return chains
}

func buildObsolescenceChain(repoTree *gitutil.RepoTree, action models.ObsolescenceAction) obsolescenceChain {
	obsoleted, obsoleter := []*git.Commit{}, []*git.Commit{}
	for _, entry := range action.Entries {
		obsoleted = append(obsoleted, entry.Commit)
		obsoleter = append(obsoleter, entry.Obsoleter)
	}

	// De-duplicate commits.
	obsoleted, obsoleter = uniqueCommits(repoTree.Repo, obsoleted), uniqueCommits(repoTree.Repo, obsoleter)

	return obsolescenceChain{
		obsoleted: sortCommitsInGraph(repoTree, obsoleted),
		obsoleter: sortCommitsInGraph(repoTree, obsoleter),
	}
}

func uniqueCommits(repo *git.Repository, commits []*git.Commit) []*git.Commit {
	oidSet := map[git.Oid]bool{}
	for _, c := range commits {
		oidSet[*c.Id()] = true
	}

	unique := []*git.Commit{}
	for oid := range oidSet {
		unique = append(unique, gitutil.CommitByOid(repo, oid))
	}
	return unique
}

// Re-orders the commits from oldest to newest, i.e. by proximity to the root of
// the commit graph.
func sortCommitsInGraph(repoTree *gitutil.RepoTree, commits []*git.Commit) []*git.Commit {
	sort.Slice(commits, func(i, j int) bool {
		return gitutil.IsAncestor(repoTree.Repo, commits[i], commits[j])
	})
	return commits
}

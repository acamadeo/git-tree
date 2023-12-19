package operations

import (
	"fmt"

	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/models"
	git "github.com/libgit2/git2go/v34"
)

// NOTE TO SELF: Make sure there are enough comments to describe how the
// algorithm is working in English!!

type evolveRunner struct {
	repoTree  *gitutil.RepoTree
	obsChains obsolescenceChains
}

// Reconcile any troubled commits within the repository.
func Evolve(repoTree *gitutil.RepoTree, obsmap *models.ObsolescenceMap) error {
	runner := evolveRunner{
		repoTree:  repoTree,
		obsChains: buildObsolescenceChains(repoTree, obsmap),
	}
	return runner.Execute(repoTree.Root, nil)
}

func (r *evolveRunner) Execute(commitOid git.Oid, rebaseHead *git.Oid) error {
	return r.executeRecurse(commitOid, rebaseHead)
}

// Recursive evolve function, which is run on each commit in the `RepoTree`.
func (r *evolveRunner) executeRecurse(commitOid git.Oid, rebaseHead *git.Oid) error {
	// Find the obsolescence chain, if any, where this commit got obsoleted.
	commit := gitutil.CommitByOid(r.repoTree.Repo, commitOid)
	obsChain := r.obsChains.FindChainWithObsoleteCommit(commit)

	commitsToRebase := []*git.Commit{commit}
	newRebaseHead := rebaseHead
	if obsChain != nil {
		// The current commit is obsolete.

		// Resolve any rebases in the obsolescence chain. We receive a list of
		// commits in the resolved version of the chain. The last of these is
		// the commit that descendant commits should rebase onto.
		commitsToRebase = r.resolveObsolescences(*obsChain)
		newRebaseHead = commitsToRebase[len(commitsToRebase)-1].Id()

		// We've addressed all the obsolescences in this chain. Skip ahead to
		// the final commit in the obsolescence chain.
		commitOid = *obsChain.obsoleted[len(obsChain.obsoleted)-1].Id()
	}

	if rebaseHead != nil {
		// Rebase `commitsToRebase` onto the old `rebaseHead`.
		rebaseHeadCommit := gitutil.CommitByOid(r.repoTree.Repo, *rebaseHead)
		// TODO: Handle errors!
		rebaseCommits(r.repoTree.Repo, commitsToRebase[0], commitsToRebase[len(commitsToRebase)-1], rebaseHeadCommit)
	}

	for _, child := range r.repoTree.FindChildren(commitOid) {
		// Abort early if evolve failed for any children.
		//
		// TODO: Is this the error behavior we want??
		if err := r.executeRecurse(child, newRebaseHead); err != nil {
			return err
		}
	}
	return nil
}

// Resolve any obsolescences within the given chain.
//
// Returns a list of commits in the resolved version of the chain.
//
// TODO: Consider whether this function should return an error!
func (r *evolveRunner) resolveObsolescences(thisChain obsolescenceChain) []*git.Commit {
	var rebaseHead *git.Commit
	resolvedCommits := []*git.Commit{}

	// Go through each commit on the obsoleter side of the chain, checking if it
	// has been obsoleted itself.
	for i := 0; i < len(thisChain.obsoleter); i++ {
		obsoleter := thisChain.obsoleter[i]
		obsChain := r.obsChains.FindChainWithObsoleteCommit(obsoleter)

		// The first time we see an obsolete commit in the chain...
		if rebaseHead == nil && obsChain != nil {
			// Resolve the obsolescences. Fast-forward past any obsolete commits
			// in this chain and continue iterating.
			//
			// TODO: Maybe should have a better name than `commitsToRebase`!!
			commitsToRebase := r.resolveObsolescences(*obsChain)
			i = lastObsoleteIdx(i, thisChain, *obsChain)
			rebaseHead = commitsToRebase[len(commitsToRebase)-1]
			resolvedCommits = append(resolvedCommits, commitsToRebase...)
			continue
		}

		// If an obsolete commit has been found, any commit afterwards must be
		// rebased onto the ultimate sucessor (`rebaseHead`).
		if rebaseHead != nil {
			commitsToRebase := []*git.Commit{obsoleter}
			obsChain = r.obsChains.FindChainWithObsoleteCommit(obsoleter)
			if obsChain != nil {
				commitsToRebase = r.resolveObsolescences(*obsChain)
				i = lastObsoleteIdx(i, thisChain, *obsChain)
			}
			rebaseHead = commitsToRebase[len(commitsToRebase)-1]

			rebasedCommits := rebaseCommits(r.repoTree.Repo, commitsToRebase[0], commitsToRebase[len(commitsToRebase)-1], rebaseHead)
			resolvedCommits = append(resolvedCommits, rebasedCommits...)
			continue
		}

		// There hasn't been an obsolete commit, and this commit isn't obsolete either.
		resolvedCommits = append(resolvedCommits, obsoleter)
	}

	return resolvedCommits
}

// Return the index of the last commit in `thisChain` that was obsoleted by
// `obsChain`.
//
// It is assumed that `index` points to an obsolete commit in `thisChain`.
func lastObsoleteIdx(index int, thisChain, obsChain obsolescenceChain) int {
	for index < len(thisChain.obsoleter) && obsChain.HasObsoleteCommit(thisChain.obsoleter[index]) {
		index++
	}
	return index - 1
}

// Rebase the sequence of commits from `start` to `end` onto commit `onto`.
//
// NOTE: `start` and `end` are inclusive.
//
// TODO: Move this into `rebase_tree`, or refactor rebase_tree to do something like this.
//
// TODO: This should probably return an error...
func rebaseCommits(repo *git.Repository, start *git.Commit, end *git.Commit, onto *git.Commit) []*git.Commit {
	// DEBUG
	fmt.Printf("Rebasing commits [%s..%s] onto %s\n",
		gitutil.CommitShortHash(start),
		gitutil.CommitShortHash(end),
		gitutil.CommitShortHash(onto))
	return []*git.Commit{}
}

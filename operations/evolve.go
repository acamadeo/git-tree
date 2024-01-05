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
	repoTree        *gitutil.RepoTree
	obsChains       obsolescenceChains
	tempBranchNames []string
}

// Reconcile any troubled commits within the repository.
func Evolve(repoTree *gitutil.RepoTree, obsmap *models.ObsolescenceMap) error {
	runner := evolveRunner{
		repoTree:  repoTree,
		obsChains: buildObsolescenceChains(repoTree, obsmap),
	}
	return runner.Execute(gitutil.CommitByOid(repoTree.Repo, repoTree.Root))
}

func (r *evolveRunner) Execute(root *git.Commit) error {
	// TODO: We may only need one temp branch. If the tree splits at some point,
	// we need to be able to point the branch to the current commit. If we can't
	// do that, we'll need to create more temp branches for every fork in the
	// tree.
	branchName := gitutil.UniqueBranchName(r.repoTree.Repo, "git-tree-evolve-head")
	head, _ := r.repoTree.Repo.CreateBranch(branchName, root, false)
	r.tempBranchNames = append(r.tempBranchNames, branchName)
	defer r.cleanupTempBranches()

	return r.executeRecurse(root, head)
}

// Recursive evolve function, which is run on each commit in the `RepoTree`.
func (r *evolveRunner) executeRecurse(commit *git.Commit, evolveHead *git.Branch) error {
	// Find the obsolescence chain, if any, where this commit got obsoleted.
	obsChain := r.obsChains.FindChainWithObsoleteCommit(commit)

	if obsChain != nil {
		// The current commit is obsolete.

		// Resolve any rebases in the obsolescence chain. `evolveHead` points to
		//  the last commit resolved commit of the chain.
		r.resolveObsolescences(*obsChain, &evolveHead)

		// The obsolescence chain is resolved. Skip to final commit in the chain.
		commit = obsChain.obsoleted[len(obsChain.obsoleted)-1]
	} else {
		// Rebase the current commit onto `evolveHead`. `evolveHead` points to
		// the last commit that was rebased.
		// TODO: Handle errors!
		rebaseCommits(r.repoTree.Repo, commit, commit, &evolveHead)
	}

	for _, childOid := range r.repoTree.FindChildren(*commit.Id()) {
		// Abort early if evolve failed for any children.
		//
		// TODO: Is this the error behavior we want??
		child := gitutil.CommitByOid(r.repoTree.Repo, childOid)
		if err := r.executeRecurse(child, evolveHead); err != nil {
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
//
// NEXT: Consider whether this function should accept the `evolveHead` branch (and just add onto it as it resolves the obsChain)!
//   - Make sure to save my work as a new commit! Don't amend the previous implementation in case we need it!
func (r *evolveRunner) resolveObsolescences(thisChain obsolescenceChain, evolveHead **git.Branch) {
	// QUESTION: Is this and the closing statement in the function needed??
	head := *evolveHead
	// Go through each commit on the obsoleter side of the chain, checking if it
	// has been obsoleted itself.
	for i := 0; i < len(thisChain.obsoleter); i++ {
		obsoleter := thisChain.obsoleter[i]
		if obsChain := r.obsChains.FindChainWithObsoleteCommit(obsoleter); obsChain != nil {
			// Resolve the obsolescences. Fast-forward past any obsolete commits
			// in this chain and continue iterating.
			r.resolveObsolescences(*obsChain, &head)
			i = lastObsoleteIdx(i, thisChain, *obsChain)
		} else {
			// This commit is not obsolete; add it to the head branch.
			rebaseCommits(r.repoTree.Repo, obsoleter, obsoleter, &head)
		}
	}

	*evolveHead = head
}

func (r *evolveRunner) cleanupTempBranches() {
	for _, branchName := range r.tempBranchNames {
		branch, _ := r.repoTree.Repo.LookupBranch(branchName, git.BranchLocal)
		branch.Delete()
	}
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

// Rebase the sequence of commits from `start` to `end` onto branch `onto`.
//
// NOTE: `start` and `end` are inclusive.
//
// NOTE TO SELF: Afterwards, branch `onto` should point to the rebased commit.

// NOTE TO SELF: We definitely need to rebase in case moving the commit over
// results in merge conflicts!
//
// TODO: Move this into `rebase_tree`, or refactor rebase_tree to do something like this.
//
// TODO: This should probably return an error...
func rebaseCommits(repo *git.Repository, start *git.Commit, end *git.Commit, onto **git.Branch) []*git.Commit {
	// rebase := gitutil.InitRebase(repo)

	// DEBUG
	ontoOid := (*onto).Target()
	ontoCommit, _ := repo.LookupCommit(ontoOid)
	fmt.Printf("Rebasing commits [%s..%s] onto %s\n",
		gitutil.CommitShortHash(start),
		gitutil.CommitShortHash(end),
		gitutil.CommitShortHash(ontoCommit))
	return []*git.Commit{}
}

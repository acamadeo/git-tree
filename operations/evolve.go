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
	headBranch      string
	tempBranchNames []string
}

// Reconcile any troubled commits within the repository.
func Evolve(repoTree *gitutil.RepoTree, obsmap *models.ObsolescenceMap) error {
	runner := evolveRunner{
		repoTree:   repoTree,
		obsChains:  buildObsolescenceChains(repoTree.Repo, obsmap),
		headBranch: gitutil.BranchName(gitutil.HeadBranch(repoTree.Repo)),
	}
	return runner.Execute(gitutil.CommitByOid(repoTree.Repo, repoTree.Root))
}

func (r *evolveRunner) Execute(root *git.Commit) error {
	// TODO: We may only need one temp branch. If the tree splits at some point,
	// we need to be able to point the branch to the current commit. If we can't
	// do that, we'll need to create more temp branches for every fork in the
	// tree.
	branchName := gitutil.UniqueBranchName(r.repoTree.Repo, "git-tree-evolve-head")
	head, _ := r.repoTree.Repo.CreateBranch(branchName, root.Parent(0), false)
	r.tempBranchNames = append(r.tempBranchNames, branchName)
	defer r.cleanup()

	return r.executeRecurse(root, &head)
}

// Recursive evolve function, which is run on each commit in the `RepoTree`.
func (r *evolveRunner) executeRecurse(commit *git.Commit, evolveHead **git.Branch) error {
	var obsoletedBranches []*git.Branch
	// Find the obsolescence chain, if any, where this commit got obsoleted.
	obsChain := r.obsChains.FindChainWithObsoleteCommit(commit)

	evolveOid := (*evolveHead).Target()
	evolveCommit, _ := r.repoTree.Repo.LookupCommit(evolveOid)
	fmt.Printf("executeRecurse(%s, %s)\n", gitutil.CommitShortHash(commit), gitutil.CommitShortHash(evolveCommit))
	if obsChain != nil {
		// The current commit is obsolete.

		// Resolve any obsolete commits in the obsolescence chain. `evolveHead`
		// points to the last resolved commit of the chain.
		r.resolveObsolescences(*obsChain, evolveHead)
		obsoletedBranches = r.findBranchesInObsChain(*obsChain)

		// The obsolescence chain is resolved. Skip to final commit in the chain.
		commit = obsChain.obsoleted[len(obsChain.obsoleted)-1]
	} else {
		// Rebase the current commit onto `evolveHead`. `evolveHead` points to
		// the last commit that was rebased.
		// TODO: Handle errors!
		rebaseCommit(r.repoTree.Repo, commit, evolveHead)
		obsoletedBranches = r.findBranchesAtCommit(commit)
	}

	// Update any branches that pointed to the current commit (or its ultimate successor).
	for _, branch := range obsoletedBranches {
		gitutil.UpdateBranchTarget(r.repoTree.Repo, &branch, (*evolveHead).Target())
	}

	headTarget := (*evolveHead).Target()
	for _, childOid := range r.repoTree.FindChildren(*commit.Id()) {
		// Abort early if evolve failed for any children.
		//
		// TODO: Is this the error behavior we want??
		child := gitutil.CommitByOid(r.repoTree.Repo, childOid)
		if err := r.executeRecurse(child, evolveHead); err != nil {
			return err
		}
		// Restore `evolveHead` to current commit before proceeding.
		gitutil.UpdateBranchTarget(r.repoTree.Repo, evolveHead, headTarget)
	}
	return nil
}

// Resolve any obsolescences within the given chain.
//
// Returns a list of commits in the resolved version of the chain.
//
// TODO: Consider whether this function should return an error!
func (r *evolveRunner) resolveObsolescences(thisChain obsolescenceChain, evolveHead **git.Branch) {
	evolveOid := (*evolveHead).Target()
	evolveCommit, _ := r.repoTree.Repo.LookupCommit(evolveOid)
	fmt.Printf("resolveObsolescences(<chain>, %s), <chain>:\n", gitutil.CommitShortHash(evolveCommit))
	fmt.Println(thisChain.String())
	// Go through each commit on the obsoleter side of the chain, checking if it
	// has been obsoleted itself.
	for i := 0; i < len(thisChain.obsoleter); i++ {
		obsoleter := thisChain.obsoleter[i]
		if obsChain := r.obsChains.FindChainWithObsoleteCommit(obsoleter); obsChain != nil {
			// Resolve the obsolescences. Fast-forward past any obsolete commits
			// in this chain and continue iterating.
			r.resolveObsolescences(*obsChain, evolveHead)
			i = lastObsoleteIdx(i, thisChain, *obsChain)
		} else {
			// This commit is not obsolete; add it to the head branch.
			rebaseCommit(r.repoTree.Repo, obsoleter, evolveHead)
		}
	}
}

func (r *evolveRunner) findBranchesInObsChain(thisChain obsolescenceChain) []*git.Branch {
	// We assume branch pointers can only be on the final commit in the
	// obsolescence chain.
	lastCommit := thisChain.obsoleter[len(thisChain.obsoleter)-1]
	for obsChain := r.obsChains.FindChainWithObsoleteCommit(lastCommit); obsChain != nil; {
		lastCommit = obsChain.obsoleter[len(obsChain.obsoleter)-1]
		obsChain = r.obsChains.FindChainWithObsoleteCommit(lastCommit)
	}
	return r.findBranchesAtCommit(lastCommit)
}

func (r *evolveRunner) findBranchesAtCommit(commit *git.Commit) []*git.Branch {
	branches := []*git.Branch{}
	for _, branchName := range r.repoTree.FindBranches(*commit.Id()) {
		branch, _ := r.repoTree.Repo.LookupBranch(branchName, git.BranchLocal)
		branches = append(branches, branch)
	}
	return branches
}

func (r *evolveRunner) cleanup() {
	// Switch back to the original HEAD branch.
	gitutil.CheckoutBranchByName(r.repoTree.Repo, r.headBranch)

	// Remove temporary branches.
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
// TODO: This should probably return an error...
func rebaseCommit(repo *git.Repository, commit *git.Commit, onto **git.Branch) {
	// DEBUG
	ontoOid := (*onto).Target()
	ontoCommit, _ := repo.LookupCommit(ontoOid)
	fmt.Printf("Rebasing commit [%s] onto %s\n",
		gitutil.CommitShortHash(commit),
		gitutil.CommitShortHash(ontoCommit))

	startParent := commit.Parent(0)
	startParentBranch := gitutil.CreateBranchAtCommit(
		repo, startParent, "git-tree-evolve-commit-start")
	endBranch := gitutil.CreateBranchAtCommit(
		repo, commit, "git-tree-evolve-commit-end")
	defer func() {
		gitutil.CheckoutBranch(repo, *onto)
		startParentBranch.Delete()
		endBranch.Delete()
	}()

	// Rebase commit `commit` onto branch `onto`.
	result := gitutil.Rebase_UpdateOnto(
		repo, startParentBranch, onto, &endBranch)

	// TODO: Actually handle the error!
	if result.Type == gitutil.RebaseError {
		fmt.Printf("Error during rebase: %v\n", result.Error)
	}
}

package common

import (
	"errors"

	"github.com/abaresk/git-tree/models"
	"github.com/abaresk/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

type RebaseTreeResultType int

const (
	RebaseTreeError RebaseTreeResultType = iota
	RebaseTreeMergeConflict
	RebaseTreeSuccess
)

// The result of a RebaseTree operation.
type RebaseTreeResult struct {
	// The type of result that occurred.
	Type RebaseTreeResultType
	// The error returned by the operation, if any.
	Error error
}

type rebaseTreeRunner struct {
	repo      *git.Repository
	source    *git.Branch
	dest      *git.Branch
	branchMap *models.BranchMap
	// A map from the temporary branch to the branch it replaced.
	tempBranches map[*git.Branch]*git.Branch
}

// QUESTIONS:
//  1. How does git natively store that a repo is in `rebasing` mode?
//      - We may want to mirror this for git-tree.
//      - We'll need to persist the temporary branches if `git-tree` enters
//        `rebasing` mode.

// Rebase a branch and all its descendants onto another branch.
//
// Under the hood, this is performed as a sequence of git rebase operations.
func RebaseTree(repo *git.Repository, source *git.Branch, dest *git.Branch, branchMap *models.BranchMap) RebaseTreeResult {

	// Steps:
	//  1. Validation:
	//      - source and destination cannot be the same
	//      - source cannot be an ancestor of destination
	//      * Might not specifically need validation but if `source` is already
	//        on `dest`, it's not an error but it should be a no-op.
	//  1. Perform the rebases in the sequence.
	//      - If it's a clean rebase (no merge conflicts). Perform the rebases
	//        as listed and then clean up (delete temporary branches).
	//      - If there's a merge conflict, set the repository into a git-tree
	//        `rebasing` state and return RebaseTreeMergeConflict.

	if err := validateRebaseTree(source, dest, branchMap); err != nil {
		return RebaseTreeResult{Type: RebaseTreeError, Error: err}
	}

	runner := newRebaseTreeRunner(repo, source, dest, branchMap)
	return runner.Execute()
}

func RebaseTreeContinue(repo *git.Repository) error {
	return nil
}

func RebaseTreeAbort(repo *git.Repository) error {
	return nil
}

func validateRebaseTree(source *git.Branch, dest *git.Branch, branchMap *models.BranchMap) error {
	// Source and destination cannot be the same.
	if source.Cmp(dest.Reference) == 0 {
		return errors.New("Source and destination cannot be the same")
	}

	// Source cannot be an ancestor of destination.
	sourceName, _ := source.Name()
	destName, _ := dest.Name()
	if branchMap.IsBranchAncestor(sourceName, destName) {
		return errors.New("Source cannot be an ancestor of destination")
	}

	// Source should not be a child of dest.
	if branchMap.IsBranchParent(destName, sourceName) {
		return errors.New("Source is already a child of destination")
	}

	return nil
}

func newRebaseTreeRunner(repo *git.Repository, source *git.Branch, dest *git.Branch, branchMap *models.BranchMap) *rebaseTreeRunner {
	return &rebaseTreeRunner{
		repo:         repo,
		source:       source,
		dest:         dest,
		branchMap:    branchMap,
		tempBranches: map[*git.Branch]*git.Branch{},
	}
}

func (r *rebaseTreeRunner) Execute() RebaseTreeResult {
	// Algorithm:
	//  1. Starting with the source branch, create a temporary branch pointing
	//     to the same commit.
	//  1. Rebase the source branch onto the destination.
	//  1. Recursively go through each of source's children and rebase it onto
	//     source's new location.
	//
	//  - I'm leaning toward not updating BranchMap until the operation
	//    succeeds. It should be easy to modify this data structure.
	//
	// Maybe pass in recursively rebaseTreeImpl(parentBranch, branchToMove, branchToMoveOnto).

	destName, _ := r.dest.Name()
	sourceName, _ := r.source.Name()

	sourceParent := r.branchMap.FindParent(sourceName)
	destBranch := r.branchMap.FindBranch(destName)
	sourceBranch := r.branchMap.FindBranch(sourceName)

	err := r.executeRecurse(sourceParent, destBranch, sourceBranch)
	if err != nil {
		return RebaseTreeResult{Type: RebaseTreeError, Error: err}
	}

	// Delete temporary branches.
	for tempBranch := range r.tempBranches {
		if err := tempBranch.Delete(); err != nil {
			return RebaseTreeResult{Type: RebaseTreeError, Error: err}
		}
	}

	if err := r.updateAndWriteBranchMap(); err != nil {
		return RebaseTreeResult{Type: RebaseTreeError, Error: err}
	}

	return RebaseTreeResult{Type: RebaseTreeSuccess}
}

func (r *rebaseTreeRunner) executeRecurse(parent, onto, toMove *git.Branch) error {
	// Create a temporary branch pointing to the same commit as `toMove`.
	toMoveName, _ := toMove.Name()
	tempName := UniqueBranchName(r.repo, "rebase-"+toMoveName)
	toMoveCommit := CommitByReference(r.repo, toMove.Reference)
	tempBranch, _ := r.repo.CreateBranch(tempName, toMoveCommit, false)

	// Keep track of the temporary branch and which branch it is replacing.
	r.tempBranches[tempBranch] = toMove

	// Rebase branch `toMove` onto branch `onto`.
	err := Rebase(r.repo, parent, onto, &toMove)
	// Pause the rebase if we encountered a merge conflict.
	if err != nil {
		return err
	}

	// Otherwise, recurse into each child of `toMove` and move it onto the new
	// location of `toMove`.
	children := r.branchMap.FindChildren(toMoveName)
	for _, child := range children {
		err := r.executeRecurse(tempBranch, toMove, child)
		// Abort early if the rebase failed for any of the children.
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *rebaseTreeRunner) updateBranchMap() {
	// Look up `source`, `dest`, and `parent` before making any changes.
	sourceName, _ := r.source.Name()
	source := r.branchMap.FindBranch(sourceName)

	destName, _ := r.dest.Name()
	dest := r.branchMap.FindBranch(destName)

	parent := r.branchMap.FindParent(sourceName)
	parentName, _ := parent.Name()

	// Move `source` under `dest`.
	childrenMap := r.branchMap.Children
	childrenMap[dest] = append(childrenMap[dest], source)

	// Remove `source` as a child of its parent.
	r.branchMap.RemoveChildren(parentName, []string{sourceName})
}

func (r *rebaseTreeRunner) updateAndWriteBranchMap() error {
	r.updateBranchMap()

	// Rewrite the branch map file to disk.
	branchFile := BranchMapPath(r.repo.Path())
	store.WriteBranchMap(r.branchMap, branchFile)

	return nil
}

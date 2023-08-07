package operations

import (
	"errors"
	"fmt"
	"os"

	"github.com/acamadeo/git-tree/common"
	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/models"
	"github.com/acamadeo/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

type RebaseTreeResultType int

const (
	RebaseTreeError RebaseTreeResultType = iota
	RebaseTreeMergeConflict
	RebaseTreeUnstagedChanges
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
	tempBranches models.TempBranchMap
}

// -------------------------------------------------------------------------- \
// RebaseTree                                                                 |
// -------------------------------------------------------------------------- /

// Rebase a branch and all its descendants onto another branch.
//
// Under the hood, this is performed as a sequence of git rebase operations.
func RebaseTree(repo *git.Repository, source *git.Branch, dest *git.Branch) RebaseTreeResult {
	// Read the branch map file.
	branchMap := store.ReadBranchMap(repo, common.BranchMapPath(repo.Path()))

	if err := validateRebaseTree(repo, source, dest, branchMap); err != nil {
		return RebaseTreeResult{Type: RebaseTreeError, Error: err}
	}

	runner := newRebaseTreeRunner(repo, source, dest, branchMap)
	return runner.Execute()
}

// -------------------------------------------------------------------------- \
// RebaseTreeContinue                                                         |
// -------------------------------------------------------------------------- /

func RebaseTreeContinue(repo *git.Repository) RebaseTreeResult {
	// Try finishing the in-progress rebase.
	rebaseResult := continueExistingRebase(repo)
	if rebaseResult.Type != RebaseTreeSuccess {
		return rebaseResult
	}

	// Read the branch map file.
	branchMap := store.ReadBranchMap(repo, common.BranchMapPath(repo.Path()))

	// Look up source and dest branches.
	sourceName := store.ReadFile(common.RebasingSourcePath(repo.Path()))
	source := branchMap.FindBranch(sourceName)

	destName := store.ReadFile(common.RebasingDestPath(repo.Path()))
	dest := branchMap.FindBranch(destName)

	runner := newRebaseTreeRunner(repo, source, dest, branchMap)

	// Populate the runner with temporary branches from previous runs.
	path := common.RebasingTempsPath(repo.Path())
	runner.tempBranches = store.ReadTemporaryBranches(repo, path)

	return runner.Execute()
}

func continueExistingRebase(repo *git.Repository) RebaseTreeResult {
	// Open the existing rebase.
	rebase, err := gitutil.OpenRebase(repo)
	if err != nil {
		err := fmt.Errorf("Error opening rebase: %v", err)
		return RebaseTreeResult{Type: RebaseTreeError, Error: err}
	}

	// Continue the existing rebase.
	rebaseResult := gitutil.ContinueRebase(repo, rebase)

	if rebaseResult.Type == gitutil.RebaseError {
		return RebaseTreeResult{Type: RebaseTreeError, Error: rebaseResult.Error}
	} else if rebaseResult.Type == gitutil.RebaseMergeConflict {
		return RebaseTreeResult{Type: RebaseTreeMergeConflict}
	} else if rebaseResult.Type == gitutil.RebaseUnstagedChanges {
		return RebaseTreeResult{Type: RebaseTreeUnstagedChanges}
	} else {
		return RebaseTreeResult{Type: RebaseTreeSuccess}
	}
}

// -------------------------------------------------------------------------- \
// RebaseTreeAbort                                                            |
// -------------------------------------------------------------------------- /

// Abort a RebaseTree operation in progress.
func RebaseTreeAbort(repo *git.Repository) RebaseTreeResult {
	rebase, err := gitutil.OpenRebase(repo)
	if err != nil {
		err := fmt.Errorf("Error opening rebase: %v", err)
		return RebaseTreeResult{Type: RebaseTreeError, Error: err}
	}

	if err := rebase.Abort(); err != nil {
		err := fmt.Errorf("Error aborting rebase: %v", err)
		return RebaseTreeResult{Type: RebaseTreeError, Error: err}
	}

	// Read temporary branches.
	path := common.RebasingTempsPath(repo.Path())
	tempBranchMap := store.ReadTemporaryBranches(repo, path)

	// Move rebased branches back to their original positions.
	restoreRebasedBranches(tempBranchMap)

	// Delete temporary branches and storage files.
	deleteTemporaryBranches(tempBranchMap)
	deleteStorage(repo)

	return RebaseTreeResult{Type: RebaseTreeSuccess}
}

// validateRebaseTree checks whether the RebaseTree operation is valid,
// returning an error if it is not.
func validateRebaseTree(repo *git.Repository, source *git.Branch, dest *git.Branch, branchMap *models.BranchMap) error {
	// Cannot run `git-tree rebase` if another rebase is in progress.
	if store.FileExists(common.RebasingPath(repo.Path())) {
		return errors.New("Cannot rebase while another rebase is in progress. Abort or continue the existing rebase")
	}

	// Source and destination cannot be the same.
	if source.Cmp(dest.Reference) == 0 {
		return errors.New("Source and destination cannot be the same")
	}

	// Source cannot be an ancestor of destination.
	sourceName := gitutil.BranchName(source)
	destName := gitutil.BranchName(dest)
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
		tempBranches: models.TempBranchMap{},
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

	destName := gitutil.BranchName(r.dest)
	sourceName := gitutil.BranchName(r.source)

	sourceParent := r.branchMap.FindParent(sourceName)
	destBranch := r.branchMap.FindBranch(destName)
	sourceBranch := r.branchMap.FindBranch(sourceName)

	result := r.executeRecurse(sourceParent, destBranch, sourceBranch)
	if result.Type == RebaseTreeMergeConflict {
		r.handleMergeConflict()
		return result
	} else if result.Type == RebaseTreeError {
		return result
	}

	r.handleSuccess()
	return RebaseTreeResult{Type: RebaseTreeSuccess}
}

func (r *rebaseTreeRunner) executeRecurse(parent, onto, toMove *git.Branch) RebaseTreeResult {
	// Check if the branch was already rebased. If it was already rebased, it
	// should have a persisted temporary branch.
	tempBranch := r.persistedTempBranch(toMove)
	if tempBranch == nil {
		// The rebase has not happened yet. Rebase branch `toMove` onto branch `onto`.

		// Create a temporary branch pointing to the same commit as `toMove`.
		// Now we won't try to rebase this branch again if `git-tree rebase`
		// gets interrupted (here or in a downstream branch).
		tempBranch = r.createTempBranch(toMove)

		rebaseResult := gitutil.InitAndRunRebase(r.repo, parent, onto, &toMove)

		// Pause the rebase if we encountered an error.
		if rebaseResult.Type == gitutil.RebaseError {
			return RebaseTreeResult{Type: RebaseTreeError, Error: rebaseResult.Error}
		}

		// Bubble out of the rebase if we encountered a merge conflict.
		if rebaseResult.Type == gitutil.RebaseMergeConflict {
			return RebaseTreeResult{Type: RebaseTreeMergeConflict}
		}
	}

	// Otherwise, recurse into each child of `toMove` and move it onto the new
	// location of `toMove`.
	children := r.branchMap.FindChildren(gitutil.BranchName(toMove))
	for _, child := range children {
		result := r.executeRecurse(tempBranch, toMove, child)
		// Abort early if the rebase failed for any of the children.
		if result.Type != RebaseTreeSuccess {
			return result
		}
	}

	return RebaseTreeResult{Type: RebaseTreeSuccess}
}

// Returns the persisted temporary branch that replaced the given branch, or nil
// if no temporary branch exists.
//
// Persisted temporary branches are created when `git-tree rebase` has rebased
// some branches but encountered a merge conflict.
func (r *rebaseTreeRunner) persistedTempBranch(branch *git.Branch) *git.Branch {
	// Branch map that was persisted in an interrupted `git-tree rebase` run.
	path := common.RebasingTempsPath(r.repo.Path())
	tempBranchMap := store.ReadTemporaryBranches(r.repo, path)
	for tempBranch, origBranch := range tempBranchMap {
		if gitutil.BranchName(origBranch) == gitutil.BranchName(branch) {
			return tempBranch
		}
	}
	return nil
}

// Create a temporary branch pointing to the same commit as `branch`.
func (r *rebaseTreeRunner) createTempBranch(branch *git.Branch) *git.Branch {
	branchName := gitutil.BranchName(branch)

	tempName := gitutil.UniqueBranchName(r.repo, "rebase-"+branchName)
	toMoveCommit := gitutil.CommitByReference(r.repo, branch.Reference)
	tempBranch, _ := r.repo.CreateBranch(tempName, toMoveCommit, false)

	// Keep track of the temporary branch and which branch it is replacing.
	r.tempBranches[tempBranch] = branch
	return tempBranch
}

func (r *rebaseTreeRunner) handleMergeConflict() {
	// Create a file indicating a rebase is in progress.
	path := common.RebasingPath(r.repo.Path())
	os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)

	// Store the `source` and `dest` branches.
	path = common.RebasingSourcePath(r.repo.Path())
	store.OverwriteFile(path, gitutil.BranchName(r.source))

	path = common.RebasingDestPath(r.repo.Path())
	store.OverwriteFile(path, gitutil.BranchName(r.dest))

	// Store the temporary branches with pointers to each one's original branch.
	path = common.RebasingTempsPath(r.repo.Path())
	store.WriteTemporaryBranches(r.tempBranches, path)
}

func (r *rebaseTreeRunner) handleSuccess() {
	deleteTemporaryBranches(r.tempBranches)
	deleteStorage(r.repo)
	r.updateAndWriteBranchMap()
}

func (r *rebaseTreeRunner) updateBranchMap() {
	// Look up `source`, `dest`, and `parent` before making any changes.
	sourceName := gitutil.BranchName(r.source)
	source := r.branchMap.FindBranch(sourceName)

	destName := gitutil.BranchName(r.dest)
	dest := r.branchMap.FindBranch(destName)

	parent := r.branchMap.FindParent(sourceName)
	parentName := gitutil.BranchName(parent)

	// Move `source` under `dest`.
	childrenMap := r.branchMap.Children
	childrenMap[dest] = append(childrenMap[dest], source)

	// Remove `source` as a child of its parent.
	r.branchMap.RemoveChildren(parentName, []string{sourceName})
}

func (r *rebaseTreeRunner) updateAndWriteBranchMap() error {
	r.updateBranchMap()

	// Rewrite the branch map file to disk.
	branchFile := common.BranchMapPath(r.repo.Path())
	store.WriteBranchMap(r.branchMap, branchFile)

	return nil
}

// Restore branches that have already been rebased back to their original
// positions, using the temporary branches map.
func restoreRebasedBranches(tempBranches models.TempBranchMap) {
	for tempBranch, origBranch := range tempBranches {
		// Move `origBranch` to point back to `tempBranch`.
		tempBranchOid := tempBranch.Reference.Target()
		origBranch.Branch().SetTarget(tempBranchOid, "aborting RebaseTree")
	}
}

func deleteTemporaryBranches(tempBranches models.TempBranchMap) {
	for tempBranch := range tempBranches {
		tempBranch.Delete()
	}
}

func deleteStorage(repo *git.Repository) {
	// Delete the file that indicates a rebase is in progress.
	rebasingPath := common.RebasingPath(repo.Path())
	os.Remove(rebasingPath)

	// Delete the file with the RebaseTree source.
	rebasingSourcePath := common.RebasingSourcePath(repo.Path())
	os.Remove(rebasingSourcePath)

	// Delete the file with the RebaseTree dest.
	rebasingDestPath := common.RebasingDestPath(repo.Path())
	os.Remove(rebasingDestPath)

	// Delete the file with the RebaseTree temporary branches.
	rebasingTempsPath := common.RebasingTempsPath(repo.Path())
	os.Remove(rebasingTempsPath)
}

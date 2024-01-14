package gitutil

import (
	"fmt"
	"path/filepath"

	"github.com/acamadeo/git-tree/utils"
	git "github.com/libgit2/git2go/v34"
)

type RebaseResultType int

const (
	RebaseError RebaseResultType = iota
	RebaseMergeConflict
	RebaseUnstagedChanges
	RebaseSuccess
)

// The result of a Rebase operation.
type RebaseResult struct {
	// The type of result that occurred.
	Type RebaseResultType
	// The error returned by the operation, if any.
	Error error
}

const successfulRebaseError = "IterOver"
const unstagedChangesError = "unstaged changes exist in workdir"

func InteractiveRebaseInProgress(repo *git.Repository) bool {
	filename := filepath.Join(repo.Path(), "rebase-merge", "interactive")
	return utils.FileExists(filename)
}

func initRebase(repo *git.Repository, upstream, onto *git.Branch, toMove **git.Branch) (*git.Rebase, error) {
	toMoveAC := AnnotatedCommitFromBranch(repo, *toMove)
	upstreamAC := AnnotatedCommitFromBranch(repo, upstream)
	ontoAC := AnnotatedCommitFromBranch(repo, onto)

	return repo.InitRebase(toMoveAC, upstreamAC, ontoAC, rebaseOptions())
}

// Rebase commits in branch `toMove` that aren't in branch `upstream` onto branch `onto`.
func Rebase(repo *git.Repository, upstream, onto *git.Branch, toMove **git.Branch) RebaseResult {
	rebase, err := initRebase(repo, upstream, onto, toMove)
	if err != nil {
		err = fmt.Errorf("Error initializing rebase: %s\n", err)
		return RebaseResult{Type: RebaseError, Error: err}
	}

	rebaseResult := doRebase(repo, rebase)

	if rebaseResult.Type == RebaseSuccess {
		// libgit2 does not update the target of `toMove` branch after rebase. Look
		// up the branch again to get the updated target.
		toMoveName := BranchName(*toMove)
		toMoveNew, _ := repo.LookupBranch(toMoveName, git.BranchLocal)
		*toMove = toMoveNew
	}
	return rebaseResult
}

// Rebase commits in branch `toMove` that aren't in branch `upstream` onto branch `onto`.
//
// Updates `onto` to point to the same commit as the newly rebased `toMove` branch.
func Rebase_UpdateOnto(repo *git.Repository, upstream *git.Branch, onto, toMove **git.Branch) RebaseResult {
	rebaseResult := Rebase(repo, upstream, *onto, toMove)
	if rebaseResult.Type == RebaseSuccess {
		UpdateBranchTarget(onto, (*toMove).Target())
	}
	return rebaseResult
}

// Returns the result of the rebase.
func doRebase(repo *git.Repository, rebase *git.Rebase) RebaseResult {
	// Perform each operation in the rebase. Breaks with an error when there
	// are no more operations in the rebase.
	var rebaseError error
	for {
		rebaseOp, err := rebase.Next()
		rebaseError = err
		if err != nil {
			fmt.Println(err)
			break
		}
		if err := commitPatch(repo, rebase, rebaseOp); err != nil {
			rebaseError = err
			break
		}
	}

	// This is the expected error when all the operations have successfully
	// completed.
	if rebaseError.Error() == successfulRebaseError {
		rebase.Finish()
	}
	return processRebaseError(rebaseError, false)
}

func ContinueRebase(repo *git.Repository, rebase *git.Rebase) RebaseResult {
	// Get the current operation in the rebase.
	curOpIdx, _ := rebase.CurrentOperationIndex()
	curOp := rebase.OperationAt(curOpIdx)

	// Commit the resolved files.
	err := commitPatch(repo, rebase, curOp)
	if result := processRebaseError(err, true); result.Type != RebaseSuccess {
		return result
	}

	return doRebase(repo, rebase)
}

// Process the rebaseError into a RebaseResult. `continuing` is true if we are
// continuing an existing rebase.
func processRebaseError(rebaseError error, continuing bool) RebaseResult {
	if rebaseError == nil || rebaseError.Error() == successfulRebaseError {
		return RebaseResult{Type: RebaseSuccess}
	}
	if rebaseError.Error() == unstagedChangesError {
		if continuing {
			return RebaseResult{Type: RebaseUnstagedChanges}
		}
		return RebaseResult{Type: RebaseMergeConflict}
	}
	return RebaseResult{Type: RebaseError, Error: rebaseError}
}

func OpenRebase(repo *git.Repository) (*git.Rebase, error) {
	return repo.OpenRebase(rebaseOptions())
}

func rebaseOptions() *git.RebaseOptions {
	return &git.RebaseOptions{
		Quiet:    0,
		InMemory: 0,
		// QUESTION: Is it still possible to stop if there are merge conflicts
		// with this setting??
		//  - seems not to be an issue?
		CheckoutOptions: git.CheckoutOptions{
			Strategy: git.CheckoutForce,
		},
		MergeOptions: git.MergeOptions{
			TreeFlags: git.MergeTreeFindRenames,
		},
	}
}

func commitPatch(repo *git.Repository, rebase *git.Rebase, rebaseOp *git.RebaseOperation) error {
	originalCommit, _ := repo.LookupCommit(rebaseOp.Id)
	return rebase.Commit(rebaseOp.Id, originalCommit.Author(), originalCommit.Committer(), originalCommit.Message())
}

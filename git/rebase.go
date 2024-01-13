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

func InitRebase(repo *git.Repository, parent, onto *git.Branch, toMove **git.Branch) (*git.Rebase, error) {
	toMoveAC := AnnotatedCommitFromBranch(repo, *toMove)
	parentAC := AnnotatedCommitFromBranch(repo, parent)
	ontoAC := AnnotatedCommitFromBranch(repo, onto)

	return repo.InitRebase(toMoveAC, parentAC, ontoAC, rebaseOptions())
}

func InitAndRunRebase(repo *git.Repository, parent, onto *git.Branch, toMove **git.Branch) RebaseResult {
	rebase, err := InitRebase(repo, parent, onto, toMove)
	if err != nil {
		err = fmt.Errorf("Error initializing rebase: %s\n", err)
		return RebaseResult{Type: RebaseError, Error: err}
	}

	rebaseResult := Rebase(repo, rebase)

	if rebaseResult.Type == RebaseSuccess {
		// libgit2 does not update the target of `toMove` branch after rebase. Look
		// up the branch again to get the updated target.
		toMoveName := BranchName(*toMove)
		toMoveNew, _ := repo.LookupBranch(toMoveName, git.BranchLocal)
		*toMove = toMoveNew
	}
	return rebaseResult
}

// TODO: Consider having this accept two Commit's and manage the references and AnnotatedCommits in here!
func InitRebase_CommitsOntoBranch(repo *git.Repository, start, end *git.AnnotatedCommit, onto **git.Branch) (*git.Rebase, error) {
	ontoAC := AnnotatedCommitFromBranch(repo, *onto)
	return repo.InitRebase(end, start, ontoAC, rebaseOptions())
}

// TODO: Consider consolidating some of the machinery here with
//   - InitAndRunRebase
func InitAndRunRebase_CommitsOntoBranch(repo *git.Repository, parent *git.Branch, toMove, onto **git.Branch) RebaseResult {
	rebaseResult := InitAndRunRebase(repo, parent, *onto, toMove)
	if rebaseResult.Type == RebaseSuccess {
		// NEXT: Update `onto` to point to the new location of `toMove`!
		UpdateBranchTarget(onto, (*toMove).Target())
	}
	return rebaseResult

	// rebase, err := InitRebase_CommitsOntoBranch(repo, parent, toMove, onto)
	// if err != nil {
	// 	err = fmt.Errorf("Error initializing rebase: %s\n", err)
	// 	return RebaseResult{Type: RebaseError, Error: err}
	// }

	// rebaseResult := Rebase(repo, rebase)

	// if rebaseResult.Type == RebaseSuccess {
	// 	// NEXT: Update `onto` to point to the new location of `toMove`!
	// 	(*onto).SetTarget(toMove.Id(), "Update `onto` branch after rebase")

	// 	// // libgit2 does not update the target of `onto` branch after rebase. Look
	// 	// // up the branch again to get the updated target.
	// 	// ontoName := BranchName(*onto)
	// 	// ontoNew, _ := repo.LookupBranch(ontoName, git.BranchLocal)
	// 	// *onto = ontoNew
	// }
	// return rebaseResult
}

// Returns the result of the rebase.
func Rebase(repo *git.Repository, rebase *git.Rebase) RebaseResult {
	// Perform each operation in the rebase. Breaks with an error when there
	// are no more operations in the rebase.
	var rebaseError error
	for true {
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

	return Rebase(repo, rebase)
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

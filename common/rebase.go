package common

import (
	"fmt"

	git "github.com/libgit2/git2go/v34"
)

const successfulRebaseError = "IterOver"

func Rebase(repo *git.Repository, parent, onto *git.Branch, toMove **git.Branch) error {
	toMoveAC := annotatedCommit(repo, *toMove)
	parentAC := annotatedCommit(repo, parent)
	ontoAC := annotatedCommit(repo, onto)

	rebase, err := repo.InitRebase(toMoveAC, parentAC, ontoAC, rebaseOptions())
	if err != nil {
		return fmt.Errorf("Error initializing rebase: %s\n", err)
	}

	// Perform each operation in the rebase. Breaks with an error when there
	// are no more operations in the rebase.
	var rebaseError error
	for true {
		rebaseOp, err := rebase.Next()
		rebaseError = err
		if err != nil {
			break
		}
		if err := commitPatch(repo, rebase, rebaseOp); err != nil {
			rebaseError = err
			break
		}
	}

	rebase.Finish()
	rebase.Free()

	// This is the expected error when all the operations have successfully
	// completed.
	if rebaseError.Error() == successfulRebaseError {
		// libgit2 does not update the target of `toMove` branch after rebase. Look
		// up the branch again to get the updated target.
		toMoveName, _ := (*toMove).Name()
		toMoveNew, _ := repo.LookupBranch(toMoveName, git.BranchLocal)
		*toMove = toMoveNew

		return nil
	}
	return rebaseError
}

func annotatedCommit(repo *git.Repository, branch *git.Branch) *git.AnnotatedCommit {
	annotatedCommit, _ := repo.AnnotatedCommitFromRef(branch.Reference)
	return annotatedCommit
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

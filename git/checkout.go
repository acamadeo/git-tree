package gitutil

import (
	"fmt"

	git "github.com/libgit2/git2go/v34"
)

func CheckoutBranch(repo *git.Repository, branch *git.Branch) error {
	commit, _ := repo.LookupCommit(branch.Target())
	commitTree, _ := commit.Tree()

	// Check out the working tree at the given branch.
	if err := repo.CheckoutTree(commitTree, checkoutOpts()); err != nil {
		return fmt.Errorf("Could not checkout tree: %s", err)
	}

	// Set HEAD to the given branch.
	repo.SetHead(branch.Reference.Name())

	return nil
}

func CheckoutBranchByName(repo *git.Repository, branchName string) error {
	branch, err := repo.LookupBranch(branchName, git.BranchLocal)
	if err != nil {
		return fmt.Errorf("Branch %q does not exist", branchName)
	}

	return CheckoutBranch(repo, branch)
}

func checkoutOpts() *git.CheckoutOptions {
	return &git.CheckoutOptions{
		Strategy: git.CheckoutSafe,
	}
}

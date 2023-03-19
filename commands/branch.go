package commands

import (
	"errors"
	"fmt"

	"github.com/abaresk/git-tree/common"
	"github.com/abaresk/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

func newCmdBranch() *Command {
	return &Command{
		Name:         "branch",
		Run:          runBranch,
		ValidateArgs: validateBranchArgs,
	}
}

// Add a new branch pointing to the current commit and checkout that branch.
func runBranch(context *Context, args []string) error {
	// Create the new branch.
	newBranchName := args[0]
	newBranch, err := context.Repo.CreateBranch(newBranchName, headCommit(context.Repo), false)
	if err != nil {
		return fmt.Errorf("Could not create branch: %s.", err.Error())
	}

	branchMap := store.ReadBranchMap(context.Repo, common.BranchMapPath(context.Repo.Path()))

	// Add the new branch as a child of the head branch in the branch map.
	headRef, _ := context.Repo.Head()
	headName, _ := headRef.Branch().Name()

	headBranch := branchMap.FindBranch(headName)
	branchMap.Children[headBranch] = append(branchMap.Children[headBranch], newBranch)

	// Rewrite the branch map file to disk.
	branchFile := common.BranchMapPath(context.Repo.Path())
	store.WriteBranchMap(branchMap, branchFile)

	// Checkout the new branch.
	if err := context.Repo.SetHead("refs/heads/" + newBranchName); err != nil {
		return fmt.Errorf("Could not checkout new branch: %s.", err.Error())
	}

	return nil
}

func headCommit(repo *git.Repository) *git.Commit {
	headRef, _ := repo.Head()
	return common.CommitByReference(repo, headRef)
}

func validateBranchArgs(context *Context, args []string) error {
	if !common.GitTreeInited(context.Repo.Path()) {
		return errors.New("git-tree is not initialized. Run `git-tree init` to initialize.")
	}

	if len(args) != 1 {
		return errors.New("Command should be followed by a single branch name.")
	}

	if branch, _ := context.Repo.LookupBranch(args[0], git.BranchLocal); branch != nil {
		return fmt.Errorf("Branch %q already exists in the git repository.", args[0])
	}

	// Check if you are on a tip commit.
	branchMap := store.ReadBranchMap(context.Repo, common.BranchMapPath(context.Repo.Path()))
	if !common.OnTipCommit(context.Repo, branchMap) {
		headCommit, _ := context.Repo.Head()
		return fmt.Errorf("HEAD commit %q is not pointed to by any tracked branches.", common.ReferenceShortHash(headCommit))
	}

	return nil
}

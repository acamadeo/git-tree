package branch

import (
	"errors"
	"fmt"

	"github.com/acamadeo/git-tree/commands"
	"github.com/acamadeo/git-tree/common"
	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/store"
	git "github.com/libgit2/git2go/v34"
	"github.com/spf13/cobra"
)

func NewBranchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch",
		Short: "Add a new branch at current commit",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			context, err := commands.CreateContext()
			if err != nil {
				return err
			}

			return validateBranchArgs(context, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := commands.CreateContext()
			if err != nil {
				return err
			}

			return runBranch(context, args)
		},
	}

	return cmd
}

// Add a new branch pointing to the current commit and checkout that branch.
func runBranch(context *commands.Context, args []string) error {
	// Create the new branch.
	newBranchName := args[0]
	newBranch, err := context.Repo.CreateBranch(newBranchName, headCommit(context.Repo), false)
	if err != nil {
		return fmt.Errorf("Could not create branch: %s.", err.Error())
	}

	branchMap := store.ReadBranchMap(context.Repo, common.BranchMapPath(context.Repo.Path()))

	// Add the new branch as a child of the head branch in the branch map.
	headRef, _ := context.Repo.Head()
	headName := gitutil.BranchName(headRef.Branch())

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
	return gitutil.CommitByReference(repo, headRef)
}

func validateBranchArgs(context *commands.Context, args []string) error {
	if !common.GitTreeInited(context.Repo.Path()) {
		return errors.New("git-tree is not initialized. Run `git-tree init` to initialize.")
	}

	if branch, _ := context.Repo.LookupBranch(args[0], git.BranchLocal); branch != nil {
		return fmt.Errorf("Branch %q already exists in the git repository.", args[0])
	}

	// Check if you are on a tip commit.
	branchMap := store.ReadBranchMap(context.Repo, common.BranchMapPath(context.Repo.Path()))
	if !common.OnTipCommit(context.Repo, branchMap) {
		headCommit, _ := context.Repo.Head()
		return fmt.Errorf("HEAD commit %q is not pointed to by any tracked branches.", gitutil.ReferenceShortHash(headCommit))
	}

	return nil
}

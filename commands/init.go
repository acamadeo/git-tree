package commands

import (
	"errors"
	"fmt"

	"github.com/acamadeo/git-tree/common"
	"github.com/acamadeo/git-tree/operations"
	git "github.com/libgit2/git2go/v34"
	"github.com/spf13/cobra"
)

// TODO: Handle special cases like:
//   - Detached HEAD
//   - HEAD is not at the tip of a git branch
//   - Some of the branches split at a commit instead of a branch.
//      * An invariant of git-tree is that branches only split at other
//        branches.

type initOptions struct {
	branches []string
}

func NewInitCommand() *cobra.Command {
	var opts initOptions

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initializes git-tree for a repository",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			context, err := CreateContext()
			if err != nil {
				return err
			}

			return validateInitArgs(context, &opts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := CreateContext()
			if err != nil {
				return err
			}

			return runInit(context, &opts)
		},
	}

	flags := cmd.Flags()

	flags.StringArrayVarP(&opts.branches, "branches", "b", []string{}, "The branches to track with git-tree")

	return cmd
}

func runInit(context *Context, opts *initOptions) error {
	return operations.Init(context.Repo, branchesFromNames(context, opts.branches)...)
}

func validateInitArgs(context *Context, opts *initOptions) error {
	// If the branch map file already exists, then `git tree init` has already
	// been run.
	if common.GitTreeInited(context.Repo.Path()) {
		return errors.New("`git-tree init` has already been run on this respository.")
	}

	if len(opts.branches) == 0 {
		return validateInitArgless(context)
	}

	for _, branch := range opts.branches {
		if _, err := context.Repo.LookupBranch(branch, git.BranchLocal); err != nil {
			return fmt.Errorf("Branch %q does not exist in the git repository.", branch)
		}
	}

	return nil
}

func validateInitArgless(context *Context) error {
	head, err := context.Repo.Head()
	if err != nil {
		return fmt.Errorf("Cannot find HEAD reference.")
	}

	if !head.IsBranch() {
		return fmt.Errorf("HEAD is not a branch.")
	}

	return nil
}

func branchesFromNames(context *Context, branchNames []string) []*git.Branch {
	branches := []*git.Branch{}
	for _, arg := range branchNames {
		branch, _ := context.Repo.LookupBranch(arg, git.BranchLocal)
		branches = append(branches, branch)
	}

	return branches
}

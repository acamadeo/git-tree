package commands

import (
	"errors"
	"os"

	"github.com/abaresk/git-tree/common"
	"github.com/abaresk/git-tree/store"
	git "github.com/libgit2/git2go/v34"
	"github.com/spf13/cobra"
)

type rebaseArgs struct {
	source *git.Branch
	dest   *git.Branch
}

type rebaseOptions struct {
	sourceName string
	destName   string
	toContinue bool
	toAbort    bool
}

func NewRebaseCommand() *cobra.Command {
	var opts rebaseOptions

	cmd := &cobra.Command{
		Use:   "rebase",
		Short: "Rebase one branch onto another branch",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			context, err := createContext()
			if err != nil {
				return err
			}

			return validateRebaseArgs(context, &opts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := createContext()
			if err != nil {
				return err
			}

			return runRebase(context, &opts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.sourceName, "source", "s", "", "Source branch to rebase")
	flags.StringVarP(&opts.destName, "dest", "d", "", "Branch to rebase onto")
	flags.BoolVar(&opts.toContinue, "continue", false, "Continue an in-progress git-tree rebase")
	flags.BoolVar(&opts.toAbort, "abort", false, "Abort an in-progress git-tree rebase")

	return cmd
}

// CONSIDERATIONS:
//  - Merge conflicts
//     - Each commit being rebased could have merge conflicts (yuck)

// Rebases a branch and all its descendants onto another branch.
func runRebase(context *Context, opts *rebaseOptions) error {
	rebaseArgs := parseRebaseArgs(context.Repo, opts)

	// Read the branch map file.
	branchMapPath := common.BranchMapPath(context.Repo.Path())
	branchMap := store.ReadBranchMap(context.Repo, branchMapPath)

	result := common.RebaseTree(context.Repo, rebaseArgs.source, rebaseArgs.dest, branchMap)

	rebasingPath := common.RebasingPath(context.Repo.Path())

	// If the repository is still in a Rebasing state, create a file to mark
	// that a git-tree rebase is in progress.
	if result.Type != common.RebaseTreeSuccess {
		os.OpenFile(rebasingPath, os.O_RDONLY|os.O_CREATE, 0666)
	} else {
		// Otherwise, delete the Rebasing state file, if it exists.
		os.Remove(rebasingPath)
	}

	return result.Error
}

func validateRebaseArgs(context *Context, opts *rebaseOptions) error {
	// QUESTION: Maybe consider not requiring `git-tree` to be init'ed. This
	// would be a useful util even without the rest of `git-tree`.
	//  - If we don't pursue that approach, we should enforce that the branches
	//    being rebased are tracked by git-tree!
	if !common.GitTreeInited(context.Repo.Path()) {
		return errors.New("git-tree is not initialized. Run `git-tree init` to initialize.")
	}

	rebaseArgs := parseRebaseArgs(context.Repo, opts)
	if rebaseArgs.source == nil && rebaseArgs.dest == nil {
		return errors.New("Command should be followed by `-s <source-branch> -d <dest-branch>`.")
	}
	if rebaseArgs.source == nil {
		return errors.New("Command must be followed by a valid `-s <source-branch>`.")
	}
	if rebaseArgs.dest == nil {
		return errors.New("Command must be followed by a valid `-d <dest-branch>`.")
	}
	return nil
}

func parseRebaseArgs(repo *git.Repository, opts *rebaseOptions) rebaseArgs {
	sourceBranch, _ := repo.LookupBranch(opts.sourceName, git.BranchLocal)
	destBranch, _ := repo.LookupBranch(opts.destName, git.BranchLocal)
	return rebaseArgs{source: sourceBranch, dest: destBranch}
}

package commands

import (
	"errors"

	"github.com/abaresk/git-tree/common"
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

	var result common.RebaseTreeResult
	if opts.toAbort {
		common.RebaseTreeAbort(context.Repo)
	} else if opts.toContinue {
		result = common.RebaseTreeContinue(context.Repo)
	} else {
		result = common.RebaseTree(context.Repo, rebaseArgs.source, rebaseArgs.dest)
	}

	if result.Type == common.RebaseTreeMergeConflict {
		return errors.New("merge conflict encountered")
	} else if result.Type == common.RebaseTreeUnstagedChanges {
		return errors.New("resolved files must be staged")
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
	return nil
}

func parseRebaseArgs(repo *git.Repository, opts *rebaseOptions) rebaseArgs {
	sourceBranch, _ := repo.LookupBranch(opts.sourceName, git.BranchLocal)
	destBranch, _ := repo.LookupBranch(opts.destName, git.BranchLocal)
	return rebaseArgs{source: sourceBranch, dest: destBranch}
}

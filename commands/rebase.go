package commands

import (
	"errors"
	"fmt"

	"github.com/abaresk/git-tree/common"
	"github.com/abaresk/git-tree/operations"
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

func validateRebaseArgs(context *Context, opts *rebaseOptions) error {
	// TODO: Check that all branches being rebased are tracked by git-tree.
	if !common.GitTreeInited(context.Repo.Path()) {
		return errors.New("git-tree is not initialized. Run `git-tree init` to initialize.")
	}

	if opts.toAbort || opts.toContinue {
		return validateAbortOrContinue(opts)
	} else {
		return validateRegularRebase(context.Repo, opts)
	}
}

func validateAbortOrContinue(opts *rebaseOptions) error {
	if opts.sourceName != "" || opts.destName != "" {
		return errors.New("Command does not take --source or --dest arguments.")
	}
	return nil
}

func validateRegularRebase(repo *git.Repository, opts *rebaseOptions) error {
	if opts.sourceName == "" || opts.destName == "" {
		return errors.New("Command should be followed by valid `-s <source-branch> -d <dest-branch>`.")
	}

	rebaseArgs := parseRebaseArgs(repo, opts)
	if rebaseArgs.source == nil {
		return fmt.Errorf("Could not find source branch %q.", opts.sourceName)
	}
	if rebaseArgs.dest == nil {
		return fmt.Errorf("Could not find dest branch %q.", opts.destName)
	}
	return nil
}

// Rebases a branch and all its descendants onto another branch.
func runRebase(context *Context, opts *rebaseOptions) error {
	rebaseArgs := parseRebaseArgs(context.Repo, opts)

	var result operations.RebaseTreeResult
	if opts.toAbort {
		result = operations.RebaseTreeAbort(context.Repo)
	} else if opts.toContinue {
		result = operations.RebaseTreeContinue(context.Repo)
	} else {
		result = operations.RebaseTree(context.Repo, rebaseArgs.source, rebaseArgs.dest)
	}

	if result.Type == operations.RebaseTreeMergeConflict {
		return errors.New("merge conflict encountered")
	} else if result.Type == operations.RebaseTreeUnstagedChanges {
		return errors.New("resolved files must be staged")
	}
	return result.Error
}

func parseRebaseArgs(repo *git.Repository, opts *rebaseOptions) rebaseArgs {
	sourceBranch, _ := repo.LookupBranch(opts.sourceName, git.BranchLocal)
	destBranch, _ := repo.LookupBranch(opts.destName, git.BranchLocal)
	return rebaseArgs{source: sourceBranch, dest: destBranch}
}

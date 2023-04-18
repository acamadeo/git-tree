package commands

import (
	"errors"
	"os"

	"github.com/abaresk/git-tree/common"
	"github.com/abaresk/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

func newCmdRebase() *Command {
	return &Command{
		Name:         "rebase",
		Run:          runRebase,
		ValidateArgs: validateRebaseArgs,
	}
}

type rebaseArgs struct {
	source *git.Branch
	dest   *git.Branch
}

// CONSIDERATIONS:
//  - Merge conflicts
//     - Each commit being rebased could have merge conflicts (yuck)

// Rebases a branch and all its descendants onto another branch.
func runRebase(context *Context, args []string) error {
	rebaseArgs := parseRebaseArgs(context.Repo, args)

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

func validateRebaseArgs(context *Context, args []string) error {
	// QUESTION: Maybe consider not requiring `git-tree` to be init'ed. This
	// would be a useful util even without the rest of `git-tree`.
	//  - If we don't pursue that approach, we should enforce that the branches
	//    being rebased are tracked by git-tree!
	if !common.GitTreeInited(context.Repo.Path()) {
		return errors.New("git-tree is not initialized. Run `git-tree init` to initialize.")
	}

	if len(args) != 4 {
		return errors.New("Command should be followed by `-s <source-branch> -d <dest-branch>`.")
	}

	rebaseArgs := parseRebaseArgs(context.Repo, args)
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

func parseRebaseArgs(repo *git.Repository, args []string) rebaseArgs {
	ret := rebaseArgs{}

	first, second := args[:2], args[2:]
	for _, pair := range [][]string{first, second} {
		if pair[0] == "-s" {
			branch, _ := repo.LookupBranch(pair[1], git.BranchLocal)
			ret.source = branch
		} else if pair[0] == "-d" {
			branch, _ := repo.LookupBranch(pair[1], git.BranchLocal)
			ret.dest = branch
		}
	}

	return ret
}

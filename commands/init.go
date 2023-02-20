package commands

import (
	"fmt"

	git "github.com/libgit2/git2go/v34"
)

func newCmdInit() *Command {
	return &Command{
		Name:         "init",
		Run:          runInit,
		ValidateArgs: validateInitArgs,
	}
}

func runInit(context *Context, args []string) error {
	return nil
}

// TODO: Handle special cases like:
//   - Detached HEAD
//   - HEAD is not at the tip of a git branch
func validateInitArgs(context *Context, args []string) error {
	if len(args) == 0 {
		return validateInitArgless(context)
	}

	if args[0] != "-b" {
		return fmt.Errorf("List of branches should be preceded by %q.", "-b")
	}

	branchNames := args[1:]
	if len(branchNames) == 0 {
		return fmt.Errorf("-b should be followed by a list of branches.")
	}

	for _, branch := range branchNames {
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

package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/abaresk/git-tree/models"
	"github.com/abaresk/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

const gitTreeRoot = "git-tree-root"

// TODO: Handle special cases like:
//   - Detached HEAD
//   - HEAD is not at the tip of a git branch
//   - Some of the branches split at a commit instead of a branch.
//      * An invariant of git-tree is that branches only split at other
//        branches.

func newCmdInit() *Command {
	return &Command{
		Name:         "init",
		Run:          runInit,
		ValidateArgs: validateInitArgs,
	}
}

func runInit(context *Context, args []string) error {
	// NOTE: For now, let's make life easier and just store the branch map in
	// our own file (not managed through git).
	gitPath := context.Repo.Path()
	branchFile := filepath.Join(gitPath, "tree/branches")

	// If the branch map file already exists, then `git tree init` has already
	// been run.
	if _, err := os.Stat(branchFile); err == nil {
		return errors.New("`git tree init` has already been run on this respository.")
	}

	// Extract the branches passed in via the arguments.
	branches := branchesFromNames(context, args[1:])

	// Create the root branch as the most-common ancestor of the provided
	// branches.
	rootBranch, err := createRootBranch(context.Repo, branches)
	if err != nil {
		return fmt.Errorf("Could not create temporary root branch: %s.", err.Error())
	}

	// Construct a branch map from the branches and store the branch map in our
	// file.
	branchMap := models.BranchMapFromRepo(context.Repo, rootBranch, branches)
	store.WriteBranchMap(branchMap, branchFile)

	return nil
}

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

func branchesFromNames(context *Context, branchNames []string) []*git.Branch {
	// If there were branches passed in, use the current branch.
	if len(branchNames) == 0 {
		head, _ := context.Repo.Head()
		return []*git.Branch{head.Branch()}
	}

	branches := []*git.Branch{}
	for _, arg := range branchNames {
		branch, _ := context.Repo.LookupBranch(arg, git.BranchLocal)
		branches = append(branches, branch)
	}

	return branches
}

// NOTE TO SELF: This new branch needs to be cleaned up in `git tree drop`!!
func createRootBranch(repo *git.Repository, branches []*git.Branch) (*git.Branch, error) {
	var rootOid *git.Oid
	// TODO: Handle argless case.
	if len(branches) == 1 {
		rootOid = branches[0].Target()
	} else {
		// Find the commit that will serve as the root of the git-tree. Create a new
		// branch pointed to this commit.
		rootOid, _ = repo.MergeBaseMany(branchOids(branches))
	}

	rootCommit, _ := repo.LookupCommit(rootOid)
	return repo.CreateBranch(gitTreeRoot, rootCommit, false)
}

func branchOids(branches []*git.Branch) []*git.Oid {
	oids := []*git.Oid{}
	for _, branch := range branches {
		oids = append(oids, branch.Target())
	}
	return oids
}

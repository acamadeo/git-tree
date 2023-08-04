package init

import (
	"errors"
	"fmt"

	"github.com/abaresk/git-tree/commands"
	"github.com/abaresk/git-tree/common"
	gitutil "github.com/abaresk/git-tree/git"
	"github.com/abaresk/git-tree/models"
	"github.com/abaresk/git-tree/store"
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
			context, err := commands.CreateContext()
			if err != nil {
				return err
			}

			return validateInitArgs(context, &opts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := commands.CreateContext()
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

func runInit(context *commands.Context, opts *initOptions) error {
	// NOTE: For now, let's make life easier and just store the branch map in
	// our own file (not managed through git).
	branchFile := common.BranchMapPath(context.Repo.Path())

	// If the branch map file already exists, then `git tree init` has already
	// been run.
	if common.GitTreeInited(context.Repo.Path()) {
		return errors.New("`git-tree init` has already been run on this respository.")
	}

	// Extract the branches passed in via the arguments.
	branches := branchesFromNames(context, opts.branches)

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

func validateInitArgs(context *commands.Context, opts *initOptions) error {
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

func validateInitArgless(context *commands.Context) error {
	head, err := context.Repo.Head()
	if err != nil {
		return fmt.Errorf("Cannot find HEAD reference.")
	}

	if !head.IsBranch() {
		return fmt.Errorf("HEAD is not a branch.")
	}

	return nil
}

func branchesFromNames(context *commands.Context, branchNames []string) []*git.Branch {
	// If there were no branches passed in, initialize with all local branches.
	if len(branchNames) == 0 {
		return gitutil.AllLocalBranches(context.Repo)
	}

	branches := []*git.Branch{}
	for _, arg := range branchNames {
		branch, _ := context.Repo.LookupBranch(arg, git.BranchLocal)
		branches = append(branches, branch)
	}

	return branches
}

func createRootBranch(repo *git.Repository, branches []*git.Branch) (*git.Branch, error) {
	var rootOid *git.Oid
	if len(branches) == 1 {
		rootOid = branches[0].Target()
	} else {
		// Find the commit that will serve as the root of the git-tree. Create a new
		// branch pointed to this commit.
		rootOid, _ = repo.MergeBaseMany(branchOids(branches))
	}

	rootCommit, _ := repo.LookupCommit(rootOid)
	return repo.CreateBranch(common.GitTreeRootBranch, rootCommit, false)
}

func branchOids(branches []*git.Branch) []*git.Oid {
	oids := []*git.Oid{}
	for _, branch := range branches {
		oids = append(oids, branch.Target())
	}
	return oids
}

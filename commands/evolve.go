package commands

import (
	"errors"
	"fmt"

	"github.com/acamadeo/git-tree/common"
	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/models"
	"github.com/acamadeo/git-tree/operations"
	"github.com/acamadeo/git-tree/store"
	git "github.com/libgit2/git2go/v34"
	"github.com/spf13/cobra"
)

func NewEvolveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evolve",
		Short: "Reconcile troubled commits in your repository",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			context, err := CreateContext()
			if err != nil {
				return err
			}

			return validateEvolve(context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := CreateContext()
			if err != nil {
				return err
			}

			return runEvolve(context)
		},
	}

	return cmd
}

func validateEvolve(context *Context) error {
	if !common.GitTreeInited(context.Repo.Path()) {
		return errors.New("git-tree is not initialized. Run `git-tree init` to initialize.")
	}
	return nil
}

func runEvolve(context *Context) error {
	obsmap := store.ReadObsolescenceMap(context.Repo, store.ObsoleteMapPath(context.Repo.Path()))

	branchMap := store.ReadBranchMap(context.Repo, store.BranchMapPath(context.Repo.Path()))
	branches := gitutil.LookupBranches(context.Repo, branchMap.ListBranchNames()...)
	root := gitutil.MergeBaseOctopus_Branches(context.Repo, branches...)
	commits := gitutil.LocalCommitsFromBranches_RootOid(context.Repo, root, branches...)

	// If there are no obsolete commits in the repository, notify the user that
	// running `git-tree evolve` is a no-op.
	if !anyObsoleteCommits(obsmap, commits) {
		fmt.Println("No troubled commits in repository.")
		return nil
	}

	repoTree := gitutil.CreateRepoTree(context.Repo, root, branches...)
	return operations.Evolve(repoTree)
}

// Returns true if any obsolete commits are found among the `localCommits`.
func anyObsoleteCommits(obsmap *models.ObsolescenceMap, localCommits []*git.Commit) bool {
	localCommitOids := map[git.Oid]bool{}
	for _, commit := range localCommits {
		localCommitOids[*commit.Id()] = true
	}

	for _, action := range obsmap.Actions {
		for _, entry := range action.Entries {
			if _, ok := localCommitOids[*entry.Commit.Id()]; ok {
				return true
			}
		}
	}
	return false
}

package drop

import (
	"fmt"
	"os"

	"github.com/abaresk/git-tree/commands"
	"github.com/abaresk/git-tree/common"
	"github.com/abaresk/git-tree/store"
	"github.com/spf13/cobra"
)

func NewDropCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Stops tracking the repository for git-tree",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := commands.CreateContext()
			if err != nil {
				return err
			}

			return runDrop(context)
		},
	}

	return cmd
}

func runDrop(context *commands.Context) error {
	// If git-tree is not initalized, notify the user that running
	// `git-tree drop` is a no-op.
	if !common.GitTreeInited(context.Repo.Path()) {
		fmt.Println("There was nothing to drop.")
		return nil
	}

	// Read the branch map file.
	branchMapPath := common.BranchMapPath(context.Repo.Path())
	branchMap := store.ReadBranchMap(context.Repo, branchMapPath)

	// Delete the root branch created by `git-tree init`.
	if err := branchMap.Root.Delete(); err != nil {
		return fmt.Errorf("Could not delete root branch: %s.", err.Error())
	}

	// Delete local git-tree storage (i.e. the branch map and obsolescence map
	// files).
	gitTreePath := common.GitTreeSubdirPath(context.Repo.Path())
	if err := os.RemoveAll(gitTreePath); err != nil {
		return fmt.Errorf("Could not delete git-tree files: %s.", err.Error())
	}

	return nil
}

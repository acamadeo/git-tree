package commands

import (
	"fmt"

	"github.com/acamadeo/git-tree/common"
	"github.com/acamadeo/git-tree/operations"
	"github.com/spf13/cobra"
)

func NewDropCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drop",
		Short: "Stops tracking the repository for git-tree",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := CreateContext()
			if err != nil {
				return err
			}

			return runDrop(context)
		},
	}

	return cmd
}

func runDrop(context *Context) error {
	// If git-tree is not initalized, notify the user that running
	// `git-tree drop` is a no-op.
	if !common.GitTreeInited(context.Repo.Path()) {
		fmt.Println("There was nothing to drop.")
		return nil
	}

	return operations.Drop(context.Repo)
}

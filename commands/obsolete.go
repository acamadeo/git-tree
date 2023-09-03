package commands

import (
	"fmt"
	"strings"

	"github.com/acamadeo/git-tree/operations"
	"github.com/spf13/cobra"
)

func NewObsoleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obsolete",
		Short: "Updates obsolescence map in response to a Git command",
		Args:  cobra.RangeArgs(2, 3),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return validateObsoleteArgs(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			context, err := CreateContext()
			if err != nil {
				return err
			}

			return runObsolete(context, args)
		},
	}

	return cmd
}

func validateObsoleteArgs(args []string) error {
	command := args[1]

	switch command {
	case "amend":
	case "rebase":
		if len(args) != 3 {
			return fmt.Errorf("accepts 3 arg(s), received %d", len(args))
		}
	case "commit":
		if len(args) != 2 {
			return fmt.Errorf("accepts 2 arg(s), received %d", len(args))
		}
	}
	return nil
}

func runObsolete(context *Context, args []string) error {
	command := args[1]

	switch command {
	case "amend":
		return operations.ObsoleteAmend(context.Repo, strings.Split(args[2], "\n"))
	case "rebase":
		return operations.ObsoleteRebase(context.Repo, strings.Split(args[2], "\n"))
	case "commit":
		return operations.ObsoleteCommit(context.Repo)
	default:
		return fmt.Errorf("Obsolescence not supported for operation %q.", command)
	}
}

package main

import (
	"os"

	branchCmd "github.com/abaresk/git-tree/commands/branch"
	dropCmd "github.com/abaresk/git-tree/commands/drop"
	initCmd "github.com/abaresk/git-tree/commands/init"
	rebaseCmd "github.com/abaresk/git-tree/commands/rebase"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "git-tree",
	Short: "Manage trees of dependent git branches",
}

var InitCmd = initCmd.NewInitCommand()
var DropCmd = dropCmd.NewDropCommand()
var BranchCmd = branchCmd.NewBranchCommand()
var RebaseCmd = rebaseCmd.NewRebaseCommand()

func init() {
	// Add all the commands.
	RootCmd.AddCommand(InitCmd, DropCmd, BranchCmd, RebaseCmd)
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

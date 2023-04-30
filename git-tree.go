package main

import (
	"os"

	"github.com/abaresk/git-tree/commands"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "git-tree",
	Short: "Manage trees of dependent git branches",
}

var InitCmd = commands.NewInitCommand()
var DropCmd = commands.NewDropCommand()
var BranchCmd = commands.NewBranchCommand()
var RebaseCmd = commands.NewRebaseCommand()

func init() {
	// Add all the commands.
	RootCmd.AddCommand(InitCmd, DropCmd, BranchCmd, RebaseCmd)
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

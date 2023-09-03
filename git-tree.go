package main

import (
	"os"

	"github.com/acamadeo/git-tree/commands"
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

// Triggered using `post-commit` and `post-rewrite` git-hooks.
// https://www.git-scm.com/docs/githooks
var ObsoleteCmd = commands.NewObsoleteCommand()

var hiddenCommands = map[string]*cobra.Command{
	"obsolete": ObsoleteCmd,
}

func init() {
	// Add all the commands.
	RootCmd.AddCommand(InitCmd, DropCmd, BranchCmd, RebaseCmd)
}

func main() {
	// Some commands are hidden as they are only intended to be used by the Git
	// interceptor.
	if len(os.Args) > 1 && hiddenCommands[os.Args[1]] != nil {
		cmd := hiddenCommands[os.Args[1]]
		if err := cmd.Execute(); err != nil {
			os.Exit(1)
		}
		return
	}

	// The main CLI interface.
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

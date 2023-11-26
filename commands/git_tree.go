package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "git-tree",
	Short: "Manage trees of dependent git branches",
}

var InitCmd = NewInitCommand()
var DropCmd = NewDropCommand()
var BranchCmd = NewBranchCommand()
var RebaseCmd = NewRebaseCommand()
var EvolveCmd = NewEvolveCommand()

// Triggered using git-hooks (https://www.git-scm.com/docs/githooks).
var ObsoleteCmd = NewObsoleteCommand()

var hiddenCommands = map[string]*cobra.Command{
	"obsolete": ObsoleteCmd,
}

func init() {
	// Add all the commands.
	RootCmd.AddCommand(InitCmd, DropCmd, BranchCmd, RebaseCmd, EvolveCmd)
}

// Returns the status code for the program.
func Main() int {
	// Some commands are hidden as they are only intended to be used by the Git
	// interceptor.
	if len(os.Args) > 1 && hiddenCommands[os.Args[1]] != nil {
		cmd := hiddenCommands[os.Args[1]]
		if err := cmd.Execute(); err != nil {
			return 1
		}
		return 0
	}

	// The main CLI interface.
	if err := RootCmd.Execute(); err != nil {
		return 1
	}
	return 0
}

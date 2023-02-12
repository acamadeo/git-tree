package commands

import (
	"errors"
	"fmt"
)

func NewCmdInit(context *Context) *Command {
	return &Command{
		Name: "init",
		Run:  RunInit,
	}
}

func RunInit(context *Context, args []string) error {
	return nil
}

func validateRunArgs(args []string) error {
	if len(args) == 0 {
		return nil
	}

	if args[0] != "-b" {
		msg := fmt.Sprintf("List of branches should be preceded by %q.", "-b")
		return errors.New(msg)
	}

	// TODO LATER: Validate that the branch names are all actual branches in the
	// git repo!

	return nil
}

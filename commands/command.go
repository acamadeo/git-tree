package commands

import (
	"fmt"
)

type Command struct {
	Name         string
	Help         string
	Aliases      []string
	ValidateArgs func(context *Context, args []string) error
	Run          func(context *Context, args []string) error

	// TODO: FILL IN MORE!!
}

var gitTreeCommands []*Command = []*Command{
	newCmdInit(),
	newCmdDrop(),
}

func RunCommand(context *Context, name string, args []string) error {
	for _, cmd := range gitTreeCommands {
		if cmd.Name == name || stringIn(cmd.Aliases, name) {
			if err := cmd.ValidateArgs(context, args); err != nil {
				return err
			}
			return cmd.Run(context, args)
		}
	}

	return fmt.Errorf("Command %q does not exist.", name)
}

func stringIn(list []string, value string) bool {
	for _, el := range list {
		if el == value {
			return true
		}
	}
	return false
}

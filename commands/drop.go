package commands

func newCmdDrop() *Command {
	return &Command{
		Name: "drop",
		Run:  runDrop,
	}
}

func runDrop(context *Context, args []string) error {
	return nil
}

package commands

func NewCmdDrop(context *Context) *Command {
	return &Command{
		Name: "drop",
		Run:  RunDrop,
	}
}

func RunDrop(context *Context, args []string) error {
	return nil
}

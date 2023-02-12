package commands

type Command struct {
	Name    string
	Help    string
	Aliases []string
	Run     func(context *Context, args []string) error

	// TODO: FILL IN MORE!!
}

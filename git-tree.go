package main

import (
	"os"

	"github.com/acamadeo/git-tree/commands"
)

func main() {
	status := commands.Main()
	os.Exit(status)
}

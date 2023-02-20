package main

import (
	"fmt"
	"os"

	"github.com/abaresk/git-tree/commands"
	git "github.com/libgit2/git2go/v34"
)

func main() {
	if len(os.Args) < 2 {
		// TODO: We should print usage()
		printFatalf("No command was specified.\n")
	}

	cwd, _ := os.Getwd()
	context := createContext(cwd)

	runError := commands.RunCommand(context, os.Args[1], os.Args[2:])
	if runError != nil {
		printFatalf(runError.Error())
	}
}

func createContext(cwd string) *commands.Context {
	repo, err := git.OpenRepository(cwd)
	if err != nil {
		printFatalf("Current directory %q is not a git repository.", cwd)
	}

	return &commands.Context{
		Repo: repo,
	}
}

func printFatalf(format string, a ...any) {
	fmt.Print("ERROR: ")
	fmt.Printf(format, a...)
	fmt.Print("\n")

	os.Exit(1)
}

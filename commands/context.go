package commands

import (
	"fmt"
	"os"

	git "github.com/libgit2/git2go/v34"
)

type Context struct {
	Repo *git.Repository
}

func createContext() (*Context, error) {
	cwd, _ := os.Getwd()
	repo, err := git.OpenRepository(cwd)
	if err != nil {
		return nil, fmt.Errorf("Current directory %q is not a git repository.", cwd)
	}

	return &Context{
		Repo: repo,
	}, nil
}

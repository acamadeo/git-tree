package commands

import (
	git "github.com/libgit2/git2go/v34"
)

type Context struct {
	Repo *git.Repository
}

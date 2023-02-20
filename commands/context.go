package commands

import (
	"github.com/abaresk/git-tree/git"
)

// TODO: Install git2go using static linking using most recent stable branch
// (likely libgit2 version 1.5).
//
// Instructions: https://github.com/libgit2/git2go#versioned-branch-static-linking.
//
// Right now, I installed `git2go` under the "Versioned branch, dynamic
// linking" model (https://github.com/libgit2/git2go#versioned-branch-dynamic-linking).
// This was the easiest to set up, as I can just install libgit2 with my package
// manager.
//
// But for this to work with more people, ideally I would target the most recent
// buildable version.

type Context struct {
	Repo *git.Repository
}

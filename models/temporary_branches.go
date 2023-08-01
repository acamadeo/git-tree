package models

import git "github.com/libgit2/git2go/v34"

// A map from each temporary branch to the branch it replaced (i.e., the branch
// that got rebased).
//
// Used in the RebaseTree operation.
type TempBranchMap map[*git.Branch]*git.Branch

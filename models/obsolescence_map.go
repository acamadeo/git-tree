package models

import git "github.com/libgit2/git2go/v34"

// Represents each obsolete commit and the commit that obsoleted it.
type ObsolescenceMap struct {
	Entries []ObsolescenceMapEntry
}

type HookType int

const (
	HookTypeUnknown HookType = iota
	PostRewriteAmend
	PostRewriteRebase
	PostCommit
)

type ObsolescenceMapEntry struct {
	// The commit that has been obsoleted.
	Commit *git.Commit
	// The commit that obsoleted this entry's commit.
	Obsoleter *git.Commit
	// Which git-hook added this entry.
	HookType HookType
}

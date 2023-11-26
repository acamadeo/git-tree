package models

import git "github.com/libgit2/git2go/v34"

type ActionType int

const (
	ActionTypeUnknown ActionType = iota
	ActionTypeRebase
	ActionTypeAmend
	ActionTypeCommit
)

// TODO: Delete this if it is not necessary!
type HookType int

const (
	HookTypeUnknown HookType = iota
	PostRewriteAmend
	PostRewriteRebase
	PostCommit
)

// Contains a map of the each obsolete commit and the commit that obsoleted it.
//
// This map is broken down into ObsolescenceActions, which are all the obsolescences
// that occurred from running a git command.
type ObsolescenceMap struct {
	Actions []ObsolescenceAction
}

type ObsolescenceAction struct {
	ActionType ActionType
	// All the obsolescences that occurred in this action.
	Entries []ObsolescenceEntry
}

type ObsolescenceEntry struct {
	// The commit that has been obsoleted.
	Commit *git.Commit
	// The commit that obsoleted this entry's commit.
	Obsoleter *git.Commit
	// Which git-hook added this entry.
	//
	// TODO: Delete this if it is unused!
	HookType HookType
}

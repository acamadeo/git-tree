package models

import git "github.com/libgit2/git2go/v34"

// Contains a map of the each obsolete commit and the commit that obsoleted it.
//
// This map is broken down into ObsolescenceEvents, which are all the obsolescences
// that occurred from running a git command.
type ObsolescenceMap struct {
	Events []ObsolescenceEvent
}

type EventType int

const (
	EventTypeUnknown EventType = iota
	EventTypeRebase
	EventTypeAmend
	EventTypeCommit
)

// TODO: Delete this if it is not necessary!
type HookType int

const (
	HookTypeUnknown HookType = iota
	PostRewriteAmend
	PostRewriteRebase
	PostCommit
)

type ObsolescenceEvent struct {
	EventType EventType
	// All the obsolescences that occurred in this event.
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

package models

// TODO: This and BranchMap should probably get moved under store/
//
// Models should contain the internal representation of the ObsolescenceMap
// and BranchMap.
type ObsolescenceMap struct {
	Entries []ObsolescenceMapEntry
}

type ObsolescenceMapEntry struct {
	// Hash of the commit that has been obsoleted.
	Commit string
	// Hash of the commit that obsoleted this commit.
	Obsoleter string
}

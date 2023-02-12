package models

type BranchMap struct {
	Entries []BranchMapEntry
}

type BranchMapEntry struct {
	// Name of parent branch.
	Parent string
	// Names of child branches.
	Children []string
}

package operations

import (
	"testing"

	"github.com/abaresk/git-tree/store"
)

func TestInit_CreatesBranchMapFile(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	env.repo.BranchWithCommit("treecko")

	Init(env.repo.Repo)

	filename := env.repo.Repo.Path() + "tree/branches"
	if !store.FileExists(filename) {
		t.Errorf("Expected file %q to exist, but it does not", filename)
	}
}

// Branches:
//
//	master ─── eevee ─┬─ espeon
//	                  └─ umbreon
func TestInit_CreatesRootBranchAtMostCommonAncestor(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	env.repo.BranchWithCommit("eevee")
	env.repo.BranchWithCommit("espeon")
	env.repo.SwitchBranch("eevee")
	env.repo.BranchWithCommit("umbreon")

	eevee := env.repo.LookupBranch("eevee")
	espeon := env.repo.LookupBranch("espeon")
	umbreon := env.repo.LookupBranch("umbreon")
	Init(env.repo.Repo, eevee, espeon, umbreon)

	rootBranch := env.repo.LookupBranch("eevee")
	gitTreeRoot := env.repo.LookupBranch("git-tree-root")

	if rootBranch.Reference.Cmp(gitTreeRoot.Reference) != 0 {
		t.Errorf("Expected branch %q to point to branch %q, but it does not", "git-tree-root", "eevee")
	}
}

// Branches:
//
//	master ─── mew ─┬─ burmy ───┬─ wormadam
//	                |           └─ mothim
//	                └─ wurmple ─┬─ silcoon ─── beautifly
//	                            └─ cascoon ─── dustox
func TestInit_BranchMapFile_AllBranches(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("burmy")
	env.repo.BranchWithCommit("wormadam")
	env.repo.SwitchBranch("burmy")
	env.repo.BranchWithCommit("mothim")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("wurmple")
	env.repo.BranchWithCommit("silcoon")
	env.repo.BranchWithCommit("beautifly")
	env.repo.SwitchBranch("wurmple")
	env.repo.BranchWithCommit("cascoon")
	env.repo.BranchWithCommit("dustox")

	Init(env.repo.Repo)

	gotString := env.repo.ReadFile(".git/tree/branches")
	wantString :=
		`git-tree-root
git-tree-root master
master mew
mew burmy wurmple
burmy mothim wormadam
wurmple cascoon silcoon
cascoon dustox
silcoon beautifly`

	if gotString != wantString {
		t.Errorf("Got branch map file: %v, but want file: %v", gotString, wantString)
	}
}

// Branches:
//
//	master ─── mew ─┬─ burmy ───┬─ wormadam
//	                |           └─ mothim
//	                └─ wurmple ─┬─ silcoon ─── beautifly
//	                            └─ cascoon ─── dustox
func TestInit_BranchMapFile_SubsetOfBranches(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("burmy")
	env.repo.BranchWithCommit("wormadam")
	env.repo.SwitchBranch("burmy")
	env.repo.BranchWithCommit("mothim")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("wurmple")
	env.repo.BranchWithCommit("silcoon")
	env.repo.BranchWithCommit("beautifly")
	env.repo.SwitchBranch("wurmple")
	env.repo.BranchWithCommit("cascoon")
	env.repo.BranchWithCommit("dustox")

	mew := env.repo.LookupBranch("mew")
	wormadam := env.repo.LookupBranch("wormadam")
	mothim := env.repo.LookupBranch("mothim")
	silcoon := env.repo.LookupBranch("silcoon")
	dustox := env.repo.LookupBranch("dustox")
	Init(env.repo.Repo, mew, wormadam, mothim, silcoon, dustox)

	gotString := env.repo.ReadFile(".git/tree/branches")
	wantString :=
		`git-tree-root
git-tree-root mew
mew wormadam mothim silcoon dustox`

	if gotString != wantString {
		t.Errorf("Got branch map file: %v, but want file: %v", gotString, wantString)
	}
}

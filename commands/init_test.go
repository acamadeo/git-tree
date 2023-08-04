package commands

import (
	"errors"
	"os"
	"testing"

	"github.com/abaresk/git-tree/store"
	"github.com/abaresk/git-tree/testutil"
)

type testEnv struct {
	repo testutil.TestRepository
	// Directory the test is running in. In setUp(), we `cd` into `repo`'s
	// working directory. In tearDown(), we return to `testDir`.
	testDir string
}

func setUp(t *testing.T) testEnv {
	repo := testutil.CreateTestRepo()
	os.Chdir(repo.Repo.Workdir())
	return testEnv{
		repo: repo,
	}
}

func (env *testEnv) tearDown(t *testing.T) {
	os.Chdir(env.testDir)
	env.repo.Free()
}

func TestInit_BranchDoesNotExist_RepoWithBranches(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	env.repo.BranchWithCommit("treecko")

	cmd := NewInitCommand()
	cmd.SetArgs([]string{"-b", "mudkip"})
	gotError := cmd.Execute()

	wantError := errors.New("Branch \"mudkip\" does not exist in the git repository.")
	if gotError.Error() != wantError.Error() {
		t.Errorf("Init command got error %v, but want error %v", gotError, wantError)
	}
}

func TestInit_BranchDoesNotExist_RepoBranchless(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	cmd := NewInitCommand()
	cmd.SetArgs([]string{"-b", "mudkip"})
	gotError := cmd.Execute()

	wantError := errors.New("Branch \"mudkip\" does not exist in the git repository.")
	if gotError.Error() != wantError.Error() {
		t.Errorf("Init command got error %v, but want error %v", gotError, wantError)
	}
}

func TestInit_CreatesBranchMapFile(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	env.repo.BranchWithCommit("treecko")

	cmd := NewInitCommand()
	cmd.Execute()

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

	cmd := NewInitCommand()
	cmd.SetArgs([]string{"-b", "eevee", "-b", "espeon", "-b", "umbreon"})
	cmd.Execute()

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

	cmd := NewInitCommand()
	cmd.Execute()

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

	cmd := NewInitCommand()
	cmd.SetArgs([]string{"-b", "mew", "-b", "wormadam", "-b", "mothim", "-b", "silcoon", "-b", "dustox"})
	cmd.Execute()

	gotString := env.repo.ReadFile(".git/tree/branches")
	wantString :=
		`git-tree-root
git-tree-root mew
mew wormadam mothim silcoon dustox`

	if gotString != wantString {
		t.Errorf("Got branch map file: %v, but want file: %v", gotString, wantString)
	}
}

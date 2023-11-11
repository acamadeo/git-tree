package operations

import (
	"testing"

	"github.com/acamadeo/git-tree/store"
	"github.com/acamadeo/git-tree/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type InitTestSuite struct {
	suite.Suite
	repo testutil.TestRepository
}

func (suite *InitTestSuite) SetupTest() {
	suite.repo = testutil.CreateTestRepo()
}

func (suite *InitTestSuite) TearDownTest() {
	suite.repo.Free()
}

func (suite *InitTestSuite) TestInit_CreatesBranchMapFile() {
	suite.repo.BranchWithCommit("treecko")

	Init(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "tree/branches"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)
}

func (suite *InitTestSuite) TestInit_ModifiesGitHooks() {
	Init(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "hooks/pre-rebase"
	scriptCall := suite.repo.Repo.Path() + `hooks/git-tree-pre-rebase.sh "$@"`
	assert.True(suite.T(), store.FileContainsLine(filename, scriptCall),
		"Expected file %q to contain line %q, but it does not", filename, scriptCall)

	filename = suite.repo.Repo.Path() + "hooks/post-rewrite"
	scriptCall = suite.repo.Repo.Path() + `hooks/git-tree-post-rewrite.sh "$@"`
	assert.True(suite.T(), store.FileContainsLine(filename, scriptCall),
		"Expected file %q to contain line %q, but it does not", filename, scriptCall)

	filename = suite.repo.Repo.Path() + "hooks/pre-commit"
	scriptCall = suite.repo.Repo.Path() + `hooks/git-tree-pre-commit.sh "$@"`
	assert.True(suite.T(), store.FileContainsLine(filename, scriptCall),
		"Expected file %q to contain line %q, but it does not", filename, scriptCall)

	filename = suite.repo.Repo.Path() + "hooks/post-commit"
	scriptCall = suite.repo.Repo.Path() + `hooks/git-tree-post-commit.sh "$@"`
	assert.True(suite.T(), store.FileContainsLine(filename, scriptCall),
		"Expected file %q to contain line %q, but it does not", filename, scriptCall)
}

func (suite *InitTestSuite) TestInit_CopiesGitHookImplementations() {
	Init(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "hooks/git-tree-pre-rebase.sh"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	filename = suite.repo.Repo.Path() + "hooks/git-tree-post-rewrite.sh"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	filename = suite.repo.Repo.Path() + "hooks/git-tree-pre-commit.sh"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	filename = suite.repo.Repo.Path() + "hooks/git-tree-post-commit.sh"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)
}

// Branches:
//
//	master ─── eevee ─┬─ espeon
//	                  └─ umbreon
func (suite *InitTestSuite) TestInit_CreatesRootBranchAtMostCommonAncestor() {
	suite.repo.BranchWithCommit("eevee")
	suite.repo.BranchWithCommit("espeon")
	suite.repo.SwitchBranch("eevee")
	suite.repo.BranchWithCommit("umbreon")

	eevee := suite.repo.LookupBranch("eevee")
	espeon := suite.repo.LookupBranch("espeon")
	umbreon := suite.repo.LookupBranch("umbreon")
	Init(suite.repo.Repo, eevee, espeon, umbreon)

	rootBranch := suite.repo.LookupBranch("eevee")
	gitTreeRoot := suite.repo.LookupBranch("git-tree-root")

	assert.Zero(suite.T(), rootBranch.Reference.Cmp(gitTreeRoot.Reference),
		"Expected branch %q to point to branch %q, but it does not", "git-tree-root", "eevee")
}

// Branches:
//
//	master ─── mew ─┬─ burmy ───┬─ wormadam
//	                |           └─ mothim
//	                └─ wurmple ─┬─ silcoon ─── beautifly
//	                            └─ cascoon ─── dustox
func (suite *InitTestSuite) TestInit_BranchMapFile_AllBranches() {
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("burmy")
	suite.repo.BranchWithCommit("wormadam")
	suite.repo.SwitchBranch("burmy")
	suite.repo.BranchWithCommit("mothim")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("wurmple")
	suite.repo.BranchWithCommit("silcoon")
	suite.repo.BranchWithCommit("beautifly")
	suite.repo.SwitchBranch("wurmple")
	suite.repo.BranchWithCommit("cascoon")
	suite.repo.BranchWithCommit("dustox")

	Init(suite.repo.Repo)

	gotString := suite.repo.ReadFile(".git/tree/branches")
	wantString :=
		`git-tree-root
git-tree-root master
master mew
mew burmy wurmple
burmy mothim wormadam
wurmple cascoon silcoon
cascoon dustox
silcoon beautifly`

	assert.Equal(suite.T(), gotString, wantString,
		"Got branch map file: %v, but want file: %v", gotString, wantString)
}

// Branches:
//
//	master ─── mew ─┬─ burmy ───┬─ wormadam
//	                |           └─ mothim
//	                └─ wurmple ─┬─ silcoon ─── beautifly
//	                            └─ cascoon ─── dustox
func (suite *InitTestSuite) TestInit_BranchMapFile_SubsetOfBranches() {
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("burmy")
	suite.repo.BranchWithCommit("wormadam")
	suite.repo.SwitchBranch("burmy")
	suite.repo.BranchWithCommit("mothim")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("wurmple")
	suite.repo.BranchWithCommit("silcoon")
	suite.repo.BranchWithCommit("beautifly")
	suite.repo.SwitchBranch("wurmple")
	suite.repo.BranchWithCommit("cascoon")
	suite.repo.BranchWithCommit("dustox")

	mew := suite.repo.LookupBranch("mew")
	wormadam := suite.repo.LookupBranch("wormadam")
	mothim := suite.repo.LookupBranch("mothim")
	silcoon := suite.repo.LookupBranch("silcoon")
	dustox := suite.repo.LookupBranch("dustox")
	Init(suite.repo.Repo, mew, wormadam, mothim, silcoon, dustox)

	gotString := suite.repo.ReadFile(".git/tree/branches")
	wantString :=
		`git-tree-root
git-tree-root mew
mew wormadam mothim silcoon dustox`

	assert.Equal(suite.T(), gotString, wantString,
		"Got branch map file: %v, but want file: %v", gotString, wantString)
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

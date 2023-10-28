package operations

import (
	"errors"
	"testing"

	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RebaseTreeTestSuite struct {
	suite.Suite
	repo testutil.TestRepository
}

func (suite *RebaseTreeTestSuite) SetupTest() {
	suite.repo = testutil.CreateTestRepo()
}

func (suite *RebaseTreeTestSuite) TearDownTest() {
	suite.repo.Free()
}

// -------------------------------------------------------------------------- \
// RebaseTree                                                                 |
// -------------------------------------------------------------------------- /

// Initial:
//
//	master ─── mew
func (suite *RebaseTreeTestSuite) TestRebaseTree_SourceAndDestCannotBeTheSame() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("mew")
	dest := suite.repo.LookupBranch("mew")
	gotResult := RebaseTree(suite.repo.Repo, source, dest)

	wantError := errors.New("Source and destination cannot be the same")

	assert.Equal(suite.T(), gotResult.Type, RebaseTreeError)
	assert.Equal(suite.T(), gotResult.Error.Error(), wantError.Error(),
		"Operation got error %v, but want error %v", gotResult.Error, wantError)
}

// Initial:
//
//	master ─── treecko ─── grovyle
func (suite *RebaseTreeTestSuite) TestRebaseTree_SourceCannotBeAncestorOfDest() {
	// Setup initial
	suite.repo.BranchWithCommit("treecko")
	suite.repo.BranchWithCommit("grovyle")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("grovyle")
	gotResult := RebaseTree(suite.repo.Repo, source, dest)

	wantError := errors.New("Source cannot be an ancestor of destination")

	assert.Equal(suite.T(), gotResult.Type, RebaseTreeError)
	assert.Equal(suite.T(), gotResult.Error.Error(), wantError.Error(),
		"Operation got error %v, but want error %v", gotResult.Error, wantError)
}

// Initial:
//
//	master ─── treecko ─── grovyle
func (suite *RebaseTreeTestSuite) TestRebaseTree_SourceCannotBeDirectChildOfDest() {
	// Setup initial
	suite.repo.BranchWithCommit("treecko")
	suite.repo.BranchWithCommit("grovyle")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("grovyle")
	dest := suite.repo.LookupBranch("treecko")
	gotResult := RebaseTree(suite.repo.Repo, source, dest)

	wantError := errors.New("Source is already a child of destination")

	assert.Equal(suite.T(), gotResult.Type, RebaseTreeError)
	assert.Equal(suite.T(), gotResult.Error.Error(), wantError.Error(),
		"Operation got error %v, but want error %v", gotResult.Error, wantError)
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
//
// Result:
//
//	master ─── mew ─── mudkip ─── treecko
func (suite *RebaseTreeTestSuite) TestRebaseTree_RebaseOneChild() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("mudkip")
	// TODO: Figure out why root is at mew instead of master!
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("mudkip")
	expectedRepo.BranchWithCommit("treecko")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle
//	                └─ mudkip
//
// Result:
//
//	master ─── mew ─── mudkip ─── treecko ─── grovyle
func (suite *RebaseTreeTestSuite) TestRebaseTree_RebaseMultipleChildren() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.BranchWithCommit("grovyle")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("mudkip")
	expectedRepo.BranchWithCommit("treecko")
	expectedRepo.BranchWithCommit("grovyle")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle
//	                └─ mudkip
//
// Result:
//
//	master ─── mew ─── treecko ─── grovyle ─── mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTree_RebaseOntoNestedBranch() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.BranchWithCommit("grovyle")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("mudkip")
	dest := suite.repo.LookupBranch("grovyle")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("treecko")
	expectedRepo.BranchWithCommit("grovyle")
	expectedRepo.BranchWithCommit("mudkip")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─── mudkip ─── treecko ─── grovyle
//
// Result:
//
//	master ─── mew ─┬─ treecko ─── grovyle
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTree_ForkBranchLine() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("mudkip")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.BranchWithCommit("grovyle")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mew")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("treecko")
	expectedRepo.BranchWithCommit("grovyle")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("mudkip")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── eevee ─── flareon ─── jolteon ─── vaporeon
//
// Result:
//
//	master ─── eevee ─┬─ vaporeon
//	                  ├─ jolteon
//	                  └─ flareon
func (suite *RebaseTreeTestSuite) TestRebaseTree_MultipleRebases_Fork() {
	// Setup initial
	suite.repo.BranchWithCommit("eevee")
	suite.repo.BranchWithCommit("flareon")
	suite.repo.BranchWithCommit("jolteon")
	suite.repo.BranchWithCommit("vaporeon")
	Init(suite.repo.Repo)

	// Rebase tree operations
	source := suite.repo.LookupBranch("jolteon")
	dest := suite.repo.LookupBranch("eevee")
	RebaseTree(suite.repo.Repo, source, dest)

	source = suite.repo.LookupBranch("vaporeon")
	dest = suite.repo.LookupBranch("eevee")
	RebaseTree(suite.repo.Repo, source, dest)

	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("eevee")
	expectedRepo.BranchWithCommit("vaporeon")
	expectedRepo.SwitchBranch("eevee")
	expectedRepo.BranchWithCommit("jolteon")
	expectedRepo.SwitchBranch("eevee")
	expectedRepo.BranchWithCommit("flareon")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── eevee ─┬─ vaporeon
//	                  ├─ jolteon
//	                  └─ flareon
//
// Result:
//
//	master ─── eevee ─── flareon ─── jolteon ─── vaporeon
func (suite *RebaseTreeTestSuite) TestRebaseTree_MultipleRebases_Merge() {
	// Setup initial
	suite.repo.BranchWithCommit("eevee")
	suite.repo.BranchWithCommit("vaporeon")
	suite.repo.SwitchBranch("eevee")
	suite.repo.BranchWithCommit("jolteon")
	suite.repo.SwitchBranch("eevee")
	suite.repo.BranchWithCommit("flareon")
	Init(suite.repo.Repo)

	// Rebase tree operations
	source := suite.repo.LookupBranch("jolteon")
	dest := suite.repo.LookupBranch("flareon")
	RebaseTree(suite.repo.Repo, source, dest)

	source = suite.repo.LookupBranch("vaporeon")
	dest = suite.repo.LookupBranch("jolteon")
	RebaseTree(suite.repo.Repo, source, dest)

	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("eevee")
	expectedRepo.BranchWithCommit("flareon")
	expectedRepo.BranchWithCommit("jolteon")
	expectedRepo.BranchWithCommit("vaporeon")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─── treecko ─── grovyle
//
// Result:
//
//	master ─┬─ mew
//	        └─ treecko ───grovyle
func (suite *RebaseTreeTestSuite) TestRebaseTree_RebaseOntoFirstBranch() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.BranchWithCommit("grovyle")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("master")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.SwitchBranch("master")
	expectedRepo.BranchWithCommit("treecko")
	expectedRepo.BranchWithCommit("grovyle")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts
//	                └─ snorunt ─┬─ glalie ───  kirlia ─┬─ gardevoir
//	                            └─ froslass            └─ gallade
func (suite *RebaseTreeTestSuite) TestRebaseTree_KirliaOntoGlalie() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("ralts")
	suite.repo.BranchWithCommit("kirlia")
	suite.repo.BranchWithCommit("gardevoir")
	suite.repo.SwitchBranch("kirlia")
	suite.repo.BranchWithCommit("gallade")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("snorunt")
	suite.repo.BranchWithCommit("glalie")
	suite.repo.SwitchBranch("snorunt")
	suite.repo.BranchWithCommit("froslass")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("kirlia")
	dest := suite.repo.LookupBranch("glalie")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─── ralts ─── kirlia ─┬─ gardevoir ─── snorunt ─┬─ glalie
//	                                     └─ gallade                └─ froslass
func (suite *RebaseTreeTestSuite) TestRebaseTree_SnoruntOntoGardevoir() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("ralts")
	suite.repo.BranchWithCommit("kirlia")
	suite.repo.BranchWithCommit("gardevoir")
	suite.repo.SwitchBranch("kirlia")
	suite.repo.BranchWithCommit("gallade")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("snorunt")
	suite.repo.BranchWithCommit("glalie")
	suite.repo.SwitchBranch("snorunt")
	suite.repo.BranchWithCommit("froslass")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("snorunt")
	dest := suite.repo.LookupBranch("gardevoir")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts
//	                |
//	                └─ snorunt ─┬─ glalie
//	                            ├─ froslass
//	                            └─ kirlia ─┬─ gardevoir
//	                                       └─ gallade
func (suite *RebaseTreeTestSuite) TestRebaseTree_KirliaOntoSnorunt() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("ralts")
	suite.repo.BranchWithCommit("kirlia")
	suite.repo.BranchWithCommit("gardevoir")
	suite.repo.SwitchBranch("kirlia")
	suite.repo.BranchWithCommit("gallade")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("snorunt")
	suite.repo.BranchWithCommit("glalie")
	suite.repo.SwitchBranch("snorunt")
	suite.repo.BranchWithCommit("froslass")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("kirlia")
	dest := suite.repo.LookupBranch("snorunt")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─── ralts ─── kirlia ─┬─ gardevoir
//	                                     ├─ gallade
//	                                     └─ snorunt ─┬─ glalie
//	                                                 └─ froslass
func (suite *RebaseTreeTestSuite) TestRebaseTree_SnoruntOntoKirlia() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("ralts")
	suite.repo.BranchWithCommit("kirlia")
	suite.repo.BranchWithCommit("gardevoir")
	suite.repo.SwitchBranch("kirlia")
	suite.repo.BranchWithCommit("gallade")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("snorunt")
	suite.repo.BranchWithCommit("glalie")
	suite.repo.SwitchBranch("snorunt")
	suite.repo.BranchWithCommit("froslass")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("snorunt")
	dest := suite.repo.LookupBranch("kirlia")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      ├─ gallade
//	                |                      └─ glalie
//	                └─ snorunt ─── froslass
func (suite *RebaseTreeTestSuite) TestRebaseTree_GlalieOntoKirlia() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("ralts")
	suite.repo.BranchWithCommit("kirlia")
	suite.repo.BranchWithCommit("gardevoir")
	suite.repo.SwitchBranch("kirlia")
	suite.repo.BranchWithCommit("gallade")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("snorunt")
	suite.repo.BranchWithCommit("glalie")
	suite.repo.SwitchBranch("snorunt")
	suite.repo.BranchWithCommit("froslass")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("glalie")
	dest := suite.repo.LookupBranch("kirlia")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("froslass")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─── gallade
//	                └─ snorunt ─┬─ glalie
//	                            ├─ froslass
//	                            └─ gardevoir
func (suite *RebaseTreeTestSuite) TestRebaseTree_GardevoirOntoSnorunt() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("ralts")
	suite.repo.BranchWithCommit("kirlia")
	suite.repo.BranchWithCommit("gardevoir")
	suite.repo.SwitchBranch("kirlia")
	suite.repo.BranchWithCommit("gallade")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("snorunt")
	suite.repo.BranchWithCommit("glalie")
	suite.repo.SwitchBranch("snorunt")
	suite.repo.BranchWithCommit("froslass")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("gardevoir")
	dest := suite.repo.LookupBranch("snorunt")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("gardevoir")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir ─── glalie
//	                |                      └─ gallade
//	                └─ snorunt ─── froslass
func (suite *RebaseTreeTestSuite) TestRebaseTree_GlalieOntoGardevoir() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("ralts")
	suite.repo.BranchWithCommit("kirlia")
	suite.repo.BranchWithCommit("gardevoir")
	suite.repo.SwitchBranch("kirlia")
	suite.repo.BranchWithCommit("gallade")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("snorunt")
	suite.repo.BranchWithCommit("glalie")
	suite.repo.SwitchBranch("snorunt")
	suite.repo.BranchWithCommit("froslass")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("glalie")
	dest := suite.repo.LookupBranch("gardevoir")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("froslass")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─── gallade
//	                |
//	                └─ snorunt ─┬─ glalie ─── gardevoir
//	                            └─ froslass
func (suite *RebaseTreeTestSuite) TestRebaseTree_GardevoirOntoGlalie() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("ralts")
	suite.repo.BranchWithCommit("kirlia")
	suite.repo.BranchWithCommit("gardevoir")
	suite.repo.SwitchBranch("kirlia")
	suite.repo.BranchWithCommit("gallade")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("snorunt")
	suite.repo.BranchWithCommit("glalie")
	suite.repo.SwitchBranch("snorunt")
	suite.repo.BranchWithCommit("froslass")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("gardevoir")
	dest := suite.repo.LookupBranch("glalie")
	RebaseTree(suite.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTree_MergeConflict_Result() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.CreateAndSwitchBranch("treecko")
	suite.repo.WriteAndCommitFile("starter", "treecko", "treecko")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("starter", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	gotResult := RebaseTree(suite.repo.Repo, source, dest)

	assert.Equal(suite.T(), gotResult.Type, RebaseTreeMergeConflict,
		"Operation did not yield merge conflict, but merge conflict expected")
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTree_MergeConflict_CannotCallRebaseTreeAgain() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.CreateAndSwitchBranch("treecko")
	suite.repo.WriteAndCommitFile("starter", "treecko", "treecko")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("starter", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	// Try doing Rebase tree again
	gotResult := RebaseTree(suite.repo.Repo, source, dest)

	wantError := errors.New("Cannot rebase while another rebase is in progress. Abort or continue the existing rebase")

	assert.Equal(suite.T(), gotResult.Type, RebaseTreeError)
	assert.Equal(suite.T(), gotResult.Error.Error(), wantError.Error(),
		"Operation got error %v, but want error %v", gotResult.Error, wantError)
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTree_MergeConflict_CreatesFiles() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.CreateAndSwitchBranch("treecko")
	suite.repo.WriteAndCommitFile("starter", "treecko", "treecko")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("starter", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	gotString := suite.repo.ReadFile(".git/tree/rebasing")
	wantString := ""
	assert.Equal(suite.T(), gotString, wantString,
		"Got rebasing file: %v, but want file: %v", gotString, wantString)

	gotString = suite.repo.ReadFile(".git/tree/rebasing-source")
	wantString = "treecko"
	assert.Equal(suite.T(), gotString, wantString,
		"Got rebasing-source file: %v, but want file: %v", gotString, wantString)

	gotString = suite.repo.ReadFile(".git/tree/rebasing-dest")
	wantString = "mudkip"
	assert.Equal(suite.T(), gotString, wantString,
		"Got rebasing-dest file: %v, but want file: %v", gotString, wantString)

	gotString = suite.repo.ReadFile(".git/tree/rebasing-temps")
	wantString = "rebase-treecko treecko"
	assert.Equal(suite.T(), gotString, wantString,
		"Got rebasing-temps file: %v, but want file: %v", gotString, wantString)
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTree_MergeConflict_RebasingTempsContainsProperBranches() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.CreateAndSwitchBranch("grovyle")
	suite.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	suite.repo.BranchWithCommit("sceptile")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	// Should contain only branches that we attempted to rebase (we never reached `sceptile`).
	gotString := suite.repo.ReadFile(".git/tree/rebasing-temps")
	wantString := `rebase-grovyle grovyle
rebase-treecko treecko`
	assert.Equal(suite.T(), gotString, wantString,
		"Got rebasing-temps file: %v, but want file: %v", gotString, wantString)
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTree_MergeConflict_TemporaryBranchesPointToProperCommits() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.CreateAndSwitchBranch("grovyle")
	suite.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	suite.repo.BranchWithCommit("sceptile")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Get the oid's of the commits the branches currently point to.
	treeckoOid := suite.repo.LookupBranch("treecko").Target()
	grovyleOid := suite.repo.LookupBranch("grovyle").Target()

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	tempTreeckoOid := suite.repo.LookupBranch("rebase-treecko").Target()
	tempGrovyleOid := suite.repo.LookupBranch("rebase-grovyle").Target()

	// The temporary branches should point to the commits where the rebased
	// branches used to point to.
	assert.Equal(suite.T(), *treeckoOid, *tempTreeckoOid,
		"Expected temporary branch to point to %v, but it points to %v", *treeckoOid, *tempTreeckoOid)
	assert.Equal(suite.T(), *grovyleOid, *tempGrovyleOid,
		"Expected temporary branch to point to %v, but it points to %v", *grovyleOid, *tempGrovyleOid)
}

// -------------------------------------------------------------------------- \
// RebaseTreeContinue                                                         |
// -------------------------------------------------------------------------- /

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTreeContinue_ForgetToStageChanges() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.CreateAndSwitchBranch("grovyle")
	suite.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	suite.repo.BranchWithCommit("sceptile")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	// Fix merge conflicts (but forget to stage changes)
	suite.repo.WriteFile("favorite", "grovyle")

	// Continue the rebase
	gotResult := RebaseTreeContinue(suite.repo.Repo)

	assert.Equal(suite.T(), gotResult.Type, RebaseTreeUnstagedChanges,
		"Expected operation to result in %q, but it did not", "RebaseTreeUnstagedChanges")
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTreeContinue_SuccessfulRebase_SuccessResult() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.CreateAndSwitchBranch("grovyle")
	suite.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	suite.repo.BranchWithCommit("sceptile")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	// Fix merge conflicts
	suite.repo.WriteFile("favorite", "grovyle")
	suite.repo.StageFiles()

	// Continue the rebase
	gotResult := RebaseTreeContinue(suite.repo.Repo)

	assert.Equal(suite.T(), gotResult.Type, RebaseTreeSuccess,
		"Expected operation successful, but it was not")
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
//
// Result:
//
//	master ─── mew ─── mudkip ─── treecko ─── grovyle ─── sceptile
func (suite *RebaseTreeTestSuite) TestRebaseTreeContinue_SuccessfulRebase_BranchesMoved() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.CreateAndSwitchBranch("grovyle")
	suite.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	suite.repo.BranchWithCommit("sceptile")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	// Fix merge conflicts
	suite.repo.WriteFile("favorite", "grovyle")
	suite.repo.StageFiles()

	// Continue the rebase
	RebaseTreeContinue(suite.repo.Repo)
	// Clean up extra branches from `git-tree init`.
	Drop(suite.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("mudkip")
	expectedRepo.BranchWithCommit("treecko")
	expectedRepo.BranchWithCommit("grovyle")
	expectedRepo.BranchWithCommit("sceptile")

	gotRepoTree := gitutil.CreateRepoTree(suite.repo.Repo, nil)
	expectedRepoTree := gitutil.CreateRepoTree(expectedRepo.Repo, nil)
	assert.True(suite.T(), gitutil.TreesEqual(gotRepoTree, expectedRepoTree),
		"Expected rebased repository to match expected, but it does not")
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
//
// Result:
//
//	master ─── mew ─── mudkip ─── treecko ─── grovyle ─── sceptile
func (suite *RebaseTreeTestSuite) TestRebaseTreeContinue_SuccessfulRebase_DeletesFiles() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.CreateAndSwitchBranch("grovyle")
	suite.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	suite.repo.BranchWithCommit("sceptile")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	// Fix merge conflicts
	suite.repo.WriteFile("favorite", "grovyle")
	suite.repo.StageFiles()

	// Continue the rebase
	RebaseTreeContinue(suite.repo.Repo)

	filename := ".git/tree/rebasing"
	assert.False(suite.T(), suite.repo.FileExists(filename),
		"Expected file %q not to exist, but it does", filename)

	filename = ".git/tree/rebasing-source"
	assert.False(suite.T(), suite.repo.FileExists(filename),
		"Expected file %q not to exist, but it does", filename)

	filename = ".git/tree/rebasing-dest"
	assert.False(suite.T(), suite.repo.FileExists(filename),
		"Expected file %q not to exist, but it does", filename)

	filename = ".git/tree/rebasing-temps"
	assert.False(suite.T(), suite.repo.FileExists(filename),
		"Expected file %q not to exist, but it does", filename)
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
//
// Result:
//
//	master ─── mew ─── mudkip ─── treecko ─── grovyle ─── sceptile
func (suite *RebaseTreeTestSuite) TestRebaseTreeContinue_SuccessfulRebase_DeletesTemporaryBranches() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.CreateAndSwitchBranch("grovyle")
	suite.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	suite.repo.BranchWithCommit("sceptile")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	// Fix merge conflicts
	suite.repo.WriteFile("favorite", "grovyle")
	suite.repo.StageFiles()

	// Continue the rebase
	RebaseTreeContinue(suite.repo.Repo)

	tempTreecko := suite.repo.LookupBranch("rebase-treecko")
	tempGrovyle := suite.repo.LookupBranch("rebase-grovyle")

	assert.Nil(suite.T(), tempTreecko,
		"Expected temporary branch %q to not exist, but it does", "rebase-treecko")
	assert.Nil(suite.T(), tempGrovyle,
		"Expected temporary branch %q to not exist, but it does", "rebase-grovyle")
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTreeContinue_RunIntoNewMergeConflict() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.CreateAndSwitchBranch("grovyle")
	suite.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	suite.repo.CreateAndSwitchBranch("sceptile")
	suite.repo.WriteAndCommitFile("overpowered", "sceptile", "sceptile")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	suite.repo.WriteAndCommitFile("overpowered", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	// Fix merge conflicts
	suite.repo.WriteFile("favorite", "grovyle")
	suite.repo.StageFiles()

	// Continue the rebase (but another conflict is expected for `sceptile` branch)
	gotResult := RebaseTreeContinue(suite.repo.Repo)

	assert.Equal(suite.T(), gotResult.Type, RebaseTreeMergeConflict,
		"Operation did not yield merge conflict, but merge conflict expected")
}

// -------------------------------------------------------------------------- \
// RebaseTreeAbort                                                            |
// -------------------------------------------------------------------------- /

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTreeAbort_MovesRebasedBranchesToOriginalLocation() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.CreateAndSwitchBranch("grovyle")
	suite.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	suite.repo.BranchWithCommit("sceptile")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Get the oid's of the commits the branches currently point to.
	treeckoOid := suite.repo.LookupBranch("treecko").Target()
	grovyleOid := suite.repo.LookupBranch("grovyle").Target()

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	// Abort the rebase
	RebaseTreeAbort(suite.repo.Repo)

	newTreeckoOid := suite.repo.LookupBranch("treecko").Target()
	newGrovyleOid := suite.repo.LookupBranch("grovyle").Target()

	// After aborting, the rebased branches should point back to their original
	// commits.
	assert.Equal(suite.T(), *treeckoOid, *newTreeckoOid,
		"Expected temporary branch to point to %v, but it points to %v", *treeckoOid, *newTreeckoOid)
	assert.Equal(suite.T(), *grovyleOid, *newGrovyleOid,
		"Expected temporary branch to point to %v, but it points to %v", *grovyleOid, *newGrovyleOid)
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTreeAbort_DeletesFiles() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.CreateAndSwitchBranch("grovyle")
	suite.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	suite.repo.BranchWithCommit("sceptile")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	// Abort the rebase
	RebaseTreeAbort(suite.repo.Repo)

	filename := ".git/tree/rebasing"
	assert.False(suite.T(), suite.repo.FileExists(filename),
		"Expected file %q not to exist, but it does", filename)

	filename = ".git/tree/rebasing-source"
	assert.False(suite.T(), suite.repo.FileExists(filename),
		"Expected file %q not to exist, but it does", filename)

	filename = ".git/tree/rebasing-dest"
	assert.False(suite.T(), suite.repo.FileExists(filename),
		"Expected file %q not to exist, but it does", filename)

	filename = ".git/tree/rebasing-temps"
	assert.False(suite.T(), suite.repo.FileExists(filename),
		"Expected file %q not to exist, but it does", filename)
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func (suite *RebaseTreeTestSuite) TestRebaseTreeAbort_DeletesTemporaryBranches() {
	// Setup initial - write conflicting contents to the same file.
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.CreateAndSwitchBranch("grovyle")
	suite.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	suite.repo.BranchWithCommit("sceptile")
	suite.repo.SwitchBranch("mew")
	suite.repo.CreateAndSwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(suite.repo.Repo)

	// Rebase tree
	source := suite.repo.LookupBranch("treecko")
	dest := suite.repo.LookupBranch("mudkip")
	RebaseTree(suite.repo.Repo, source, dest)

	// Abort the rebase
	RebaseTreeAbort(suite.repo.Repo)

	tempTreecko := suite.repo.LookupBranch("rebase-treecko")
	tempGrovyle := suite.repo.LookupBranch("rebase-grovyle")

	assert.Nil(suite.T(), tempTreecko,
		"Expected temporary branch %q to not exist, but it does", "rebase-treecko")
	assert.Nil(suite.T(), tempGrovyle,
		"Expected temporary branch %q to not exist, but it does", "rebase-grovyle")
}

func TestRebaseTreeTestSuite(t *testing.T) {
	suite.Run(t, new(RebaseTreeTestSuite))
}

package operations

import (
	"testing"

	"github.com/acamadeo/git-tree/store"
	"github.com/acamadeo/git-tree/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ObsoleteTestSuite struct {
	suite.Suite
	repo testutil.TestRepository
}

func (suite *ObsoleteTestSuite) SetupTest() {
	suite.repo = testutil.CreateTestRepo()
}

func (suite *ObsoleteTestSuite) TearDownTest() {
	suite.repo.Free()
}

func (suite *ObsoleteTestSuite) TestObsoletePreRebase_AddActionToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	// Simulate output from `pre-rebase` hook after user initiates a rebase.
	ObsoletePreRebase(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	wantString := "action rebase"
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoletePostRewriteAmend_AddActionAndEntryToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	// Simulate user amending a commit, which fires the `pre-commit` and
	// `post-rewrite.amend` hooks.
	ObsoletePreCommit(suite.repo.Repo)
	ObsoletePostRewriteAmend(suite.repo.Repo, []string{"cf59c4bf9d3036b68242d6e9db30c0d7654326b6 3316a58b9dd84c7b1864a3eb4d398ca643ac23c7"})

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	// Obsmap:
	//   `action amend` (the action header)
	//   `Parent of HEAD (obsoleted)` - `HEAD (obsoleter)` - `post-rewrite.rebase`
	wantString := `action amend
cf59c4bf9d3036b68242d6e9db30c0d7654326b6 3316a58b9dd84c7b1864a3eb4d398ca643ac23c7 post-rewrite.amend`
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoletePostRewriteRebase_AddActionAndEntryToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	// Simulate user performing a rebase, which fires the `pre-rebase` and
	// `post-rewrite.rebase` hooks.
	ObsoletePreRebase(suite.repo.Repo)
	ObsoletePostRewriteRebase(suite.repo.Repo, []string{"cf59c4bf9d3036b68242d6e9db30c0d7654326b6 3316a58b9dd84c7b1864a3eb4d398ca643ac23c7"})

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	// Obsmap:
	//   `action rebase` (the action header)
	//   `Parent of HEAD (obsoleted)` - `HEAD (obsoleter)` - `post-rewrite.rebase`
	wantString := `action rebase
cf59c4bf9d3036b68242d6e9db30c0d7654326b6 3316a58b9dd84c7b1864a3eb4d398ca643ac23c7 post-rewrite.rebase`
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoletePreCommit_CreatesPreCommitParentFile() {
	suite.repo.BranchWithCommit("treecko")

	ObsoletePreCommit(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "tree/pre-commit-parent"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	// Hash of HEAD commit's parent.
	wantString := "cf59c4bf9d3036b68242d6e9db30c0d7654326b6"
	gotString := suite.repo.ReadFile(".git/tree/pre-commit-parent")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoletePreCommit_AddActionToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	// Simulate output from `pre-commit` hook.
	ObsoletePreCommit(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	wantString := "action commit"
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoletePostCommit_AddEntryToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	// Simulate user adding a commit, which fires the `pre-commit` and
	// `post-commit` hooks.
	ObsoletePreCommit(suite.repo.Repo)
	suite.repo.WriteAndCommitFile("grovyle", "grovyle", "grovyle")
	ObsoletePostCommit(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	// Obsmap:
	//   `action commit` (the action header)
	//   `Parent of HEAD (obsoleted)` - `HEAD (obsoleter)` - `post-commit`
	wantString := `action commit
3316a58b9dd84c7b1864a3eb4d398ca643ac23c7 d916506e4a229b277e6658504ec0321dabe9d797 post-commit`
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoletePostCommit_DontAddEntryIfCommitWasAmended() {
	suite.repo.BranchWithCommit("treecko")

	// Run pre-commit (persists the parent of HEAD in the `pre-commit-parent`
	// file).
	ObsoletePreCommit(suite.repo.Repo)

	// Modify the HEAD commit's message.
	suite.repo.AmendCommit("mudkip")

	// Then run post-commit.
	ObsoletePostCommit(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	// No entry appears under this action.
	//
	// NOTE: The action is still marked as `commit` because it gets changed in
	// the `post-rewrite.amend` hook.
	wantString := "action commit"
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func TestObsoleteTestSuite(t *testing.T) {
	suite.Run(t, new(ObsoleteTestSuite))
}

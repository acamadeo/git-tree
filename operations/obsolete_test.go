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

func (suite *ObsoleteTestSuite) TestObsoleteAmend_AddEntryToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	// Simulate output from `post-rewrite` hook after amending the HEAD commit.
	ObsoleteAmend(suite.repo.Repo, []string{"1bcfb74c7735e96dd69e1369d80d029b4aacbce8 5b8b675e1a0f883a7f9a608460a1f8097741e7a6"})

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	// Obsmap entry:
	//   `Parent of HEAD (obsoleted)` - `HEAD (obsoleter)` - `post-rewrite.amend`
	wantString := "1bcfb74c7735e96dd69e1369d80d029b4aacbce8 5b8b675e1a0f883a7f9a608460a1f8097741e7a6 post-rewrite.amend"
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoleteRebase_AddEntryToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	// Simulate output from `post-rewrite` hook after rebasing.
	ObsoleteRebase(suite.repo.Repo, []string{"1bcfb74c7735e96dd69e1369d80d029b4aacbce8 5b8b675e1a0f883a7f9a608460a1f8097741e7a6"})

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	// Obsmap entry:
	//   `Parent of HEAD (obsoleted)` - `HEAD (obsoleter)` - `post-rewrite.rebase`
	wantString := "1bcfb74c7735e96dd69e1369d80d029b4aacbce8 5b8b675e1a0f883a7f9a608460a1f8097741e7a6 post-rewrite.rebase"
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoletePreCommit_CreatesPreCommitParentFile() {
	suite.repo.BranchWithCommit("treecko")

	ObsoletePreCommit(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "tree/pre-commit-parent"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	// Hash of HEAD commit's parent.
	wantString := "1bcfb74c7735e96dd69e1369d80d029b4aacbce8"
	gotString := suite.repo.ReadFile(".git/tree/pre-commit-parent")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoletePostCommit_AddEntryToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	ObsoletePostCommit(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	// Obsmap entry:
	//   `Parent of HEAD (obsoleted)` - `HEAD (obsoleter)` - `post-commit`
	wantString := "1bcfb74c7735e96dd69e1369d80d029b4aacbce8 5b8b675e1a0f883a7f9a608460a1f8097741e7a6 post-commit"
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoletePostCommit_DontAddEntryIfHeadParentHasntChangedSincePreCommit() {
	suite.repo.BranchWithCommit("treecko")

	// Run pre-commit (persists the parent of HEAD in the `pre-commit-parent`
	// file).
	ObsoletePreCommit(suite.repo.Repo)

	// Modify the HEAD commit's message.
	suite.repo.AmendCommit("mudkip")

	// Then run post-commit.
	ObsoletePostCommit(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.False(suite.T(), store.FileExists(filename), "Expected file %q not to exist, but it does", filename)
}

func TestObsoleteTestSuite(t *testing.T) {
	suite.Run(t, new(ObsoleteTestSuite))
}

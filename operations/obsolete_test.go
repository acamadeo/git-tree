package operations

import (
	"testing"

	"github.com/acamadeo/git-tree/store"
	"github.com/acamadeo/git-tree/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// TODO: Add more end-to-end tests, where we perform the native git CLI commands
// to trigger these obsolete commands.

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

func (suite *ObsoleteTestSuite) TestObsoletePreRebase_AddEventToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	// Simulate output from `pre-rebase` hook after user initiates a rebase.
	ObsoletePreRebase(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	wantString := "event rebase"
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoleteAmend_AddEventAndEntryToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	// Simulate user amending a commit, which fires the `pre-commit` and
	// `post-rewrite.amend` hooks.
	ObsoletePreCommit(suite.repo.Repo)
	ObsoleteAmend(suite.repo.Repo, []string{"1bcfb74c7735e96dd69e1369d80d029b4aacbce8 5b8b675e1a0f883a7f9a608460a1f8097741e7a6"})

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	// Obsmap:
	//   `event amend` (the event header)
	//   `Parent of HEAD (obsoleted)` - `HEAD (obsoleter)` - `post-rewrite.rebase`
	wantString := `event amend
1bcfb74c7735e96dd69e1369d80d029b4aacbce8 5b8b675e1a0f883a7f9a608460a1f8097741e7a6 post-rewrite.amend`
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func (suite *ObsoleteTestSuite) TestObsoleteRebase_AddEventAndEntryToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	// Simulate user performing a rebase, which fires the `pre-rebase` and
	// `post-rewrite.rebase` hooks.
	ObsoletePreRebase(suite.repo.Repo)
	ObsoleteRebase(suite.repo.Repo, []string{"1bcfb74c7735e96dd69e1369d80d029b4aacbce8 5b8b675e1a0f883a7f9a608460a1f8097741e7a6"})

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	// Obsmap:
	//   `event rebase` (the event header)
	//   `Parent of HEAD (obsoleted)` - `HEAD (obsoleter)` - `post-rewrite.rebase`
	wantString := `event rebase
1bcfb74c7735e96dd69e1369d80d029b4aacbce8 5b8b675e1a0f883a7f9a608460a1f8097741e7a6 post-rewrite.rebase`
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

func (suite *ObsoleteTestSuite) TestObsoletePreCommit_AddEventToObsmap() {
	suite.repo.BranchWithCommit("treecko")

	// Simulate output from `pre-commit` hook.
	ObsoletePreCommit(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "tree/obsmap"
	assert.True(suite.T(), store.FileExists(filename), "Expected file %q to exist, but it does not", filename)

	wantString := "event commit"
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
	//   `event commit` (the event header)
	//   `Parent of HEAD (obsoleted)` - `HEAD (obsoleter)` - `post-commit`
	wantString := `event commit
5b8b675e1a0f883a7f9a608460a1f8097741e7a6 82f6f8dbc22cc410119c7a600015b8396ef12064 post-commit`
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

	// No entry appears under this event.
	//
	// NOTE: The event is still marked as `commit` because it gets changed in
	// the `post-rewrite.amend` hook.
	wantString := "event commit"
	gotString := suite.repo.ReadFile(".git/tree/obsmap")
	assert.Equal(suite.T(), wantString, gotString)
}

func TestObsoleteTestSuite(t *testing.T) {
	suite.Run(t, new(ObsoleteTestSuite))
}

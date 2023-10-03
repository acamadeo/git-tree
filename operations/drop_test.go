package operations

import (
	"testing"

	"github.com/acamadeo/git-tree/store"
	"github.com/acamadeo/git-tree/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DropTestSuite struct {
	suite.Suite
	repo testutil.TestRepository
}

func (suite *DropTestSuite) SetupTest() {
	repo := testutil.CreateTestRepo()

	// Run git-tree init.
	Init(repo.Repo)
	suite.repo = repo
}

func (suite *DropTestSuite) TearDownTest() {
	suite.repo.Free()
}

func (suite *DropTestSuite) TestDrop_DeletesRootBranch() {
	Drop(suite.repo.Repo)

	rootBranch := suite.repo.LookupBranch("git-tree-root")
	assert.Nil(suite.T(), rootBranch)
}

func (suite *DropTestSuite) TestDrop_DeletesGitTreeSubdir() {
	Drop(suite.repo.Repo)

	dirName := suite.repo.Repo.Path() + "/tree"
	assert.False(suite.T(), store.DirExists(dirName), "Expected directory %q not to exist, but it does", dirName)
}

func (suite *DropTestSuite) TestDrop_GitHooksDontCallImplementation() {
	// Set up the git-hooks.
	Init(suite.repo.Repo)

	// Perform Drop operation.
	Drop(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "hooks/post-rewrite"
	scriptCall := suite.repo.Repo.Path() + `hooks/git-tree-post-rewrite.sh "$@"`
	assert.False(suite.T(), store.FileContainsLine(filename, scriptCall),
		"Expected file %q to not contain line %q, but it not", filename, scriptCall)

	filename = suite.repo.Repo.Path() + "hooks/post-commit"
	scriptCall = suite.repo.Repo.Path() + `hooks/git-tree-post-commit.sh "$@"`
	assert.False(suite.T(), store.FileContainsLine(filename, scriptCall),
		"Expected file %q to not contain line %q, but it does", filename, scriptCall)
}

func (suite *DropTestSuite) TestDrop_RemovesGitHookImplementations() {
	// Set up the git-hooks.
	Init(suite.repo.Repo)

	// Perform Drop operation.
	Drop(suite.repo.Repo)

	filename := suite.repo.Repo.Path() + "hooks/git-tree-post-rewrite.sh"
	assert.False(suite.T(), store.FileExists(filename),
		"Expected file %q to not exist, but it does", filename)

	filename = suite.repo.Repo.Path() + "hooks/git-tree-post-commit.sh"
	assert.False(suite.T(), store.FileExists(filename),
		"Expected file %q to not exist, but it does", filename)
}

func TestDropTestSuite(t *testing.T) {
	suite.Run(t, new(DropTestSuite))
}

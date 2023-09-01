package commands

import (
	"os"
	"testing"

	"github.com/acamadeo/git-tree/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RebaseTestSuite struct {
	suite.Suite
	repo testutil.TestRepository
	// Directory the test is running in. In setUp(), we `cd` into `repo`'s
	// working directory. In tearDown(), we return to `testDir`.
	testDir string
}

func (suite *RebaseTestSuite) SetupTest() {
	repo := testutil.CreateTestRepo()
	os.Chdir(repo.Repo.Workdir())

	suite.repo = repo
}

func (suite *RebaseTestSuite) TearDownTest() {
	os.Chdir(suite.testDir)
	suite.repo.Free()
}

func (suite *RebaseTestSuite) TestRebase_ErrorIfGitTreeNotInitialized() {
	gotError := NewRebaseCommand().Execute()

	wantError := "git-tree is not initialized. Run `git-tree init` to initialize."
	assert.EqualError(suite.T(), gotError, wantError)
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
func (suite *RebaseTestSuite) TestRebase_SourceBranchDoesNotExist() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("mudkip")

	// Run git-tree init.
	NewInitCommand().Execute()

	cmd := NewRebaseCommand()
	cmd.SetArgs([]string{"-s", "torchic", "-d", "treecko"})
	gotError := cmd.Execute()

	wantError := "Could not find source branch \"torchic\"."
	assert.EqualError(suite.T(), gotError, wantError)
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
func (suite *RebaseTestSuite) TestRebase_DestBranchDoesNotExist() {
	// Setup initial
	suite.repo.BranchWithCommit("mew")
	suite.repo.BranchWithCommit("treecko")
	suite.repo.SwitchBranch("mew")
	suite.repo.BranchWithCommit("mudkip")

	// Run git-tree init.
	NewInitCommand().Execute()

	cmd := NewRebaseCommand()
	cmd.SetArgs([]string{"-s", "treecko", "-d", "torchic"})
	gotError := cmd.Execute()

	wantError := "Could not find dest branch \"torchic\"."
	assert.EqualError(suite.T(), gotError, wantError)
}

func TestRebaseTestSuite(t *testing.T) {
	suite.Run(t, new(RebaseTestSuite))
}

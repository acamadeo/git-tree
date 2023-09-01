package branch

import (
	"errors"
	"os"
	"strings"
	"testing"

	initCmd "github.com/acamadeo/git-tree/commands/init"
	"github.com/acamadeo/git-tree/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BranchWithGitInitedTestSuite struct {
	suite.Suite
	repo testutil.TestRepository
	// Directory the test is running in. In setUp(), we `cd` into `repo`'s
	// working directory. In tearDown(), we return to `testDir`.
	testDir string
}

func (suite *BranchWithGitInitedTestSuite) SetupTest() {
	repo := testutil.CreateTestRepo()
	os.Chdir(repo.Repo.Workdir())

	// Run git-tree init.
	initCmd.NewInitCommand().Execute()

	suite.repo = repo
}

func (suite *BranchWithGitInitedTestSuite) TearDownTest() {
	os.Chdir(suite.testDir)
	suite.repo.Free()
}

func (suite *BranchWithGitInitedTestSuite) TestBranch_BranchAlreadyExists() {
	suite.repo.BranchWithCommit("treecko")

	cmd := NewBranchCommand()
	cmd.SetArgs([]string{"treecko"})
	gotError := cmd.Execute()

	wantError := errors.New("Branch \"treecko\" already exists in the git repository.")
	assert.Equal(suite.T(), gotError.Error(), wantError.Error(),
		"Command got error %v, but want error %v", gotError, wantError)
}

// Branches:
//
//	                    ┌───── *mudkip
//	                    ▼
//	master ─── mud -> kip
func (suite *BranchWithGitInitedTestSuite) TestBranch_NotOnTipCommit() {
	suite.repo.CreateBranch("mudkip")
	suite.repo.SwitchBranch("mudkip")
	suite.repo.WriteAndCommitFile("mud", "mud", "mud")
	suite.repo.WriteAndCommitFile("kip", "kip", "kip")
	suite.repo.SwitchCommit("mud")

	cmd := NewBranchCommand()
	cmd.SetArgs([]string{"treecko"})
	gotError := cmd.Execute()

	assert.ErrorContains(suite.T(), gotError, "HEAD commit")
	assert.ErrorContains(suite.T(), gotError, "is not pointed to by any tracked branches")
}

type BranchTestSuite struct {
	suite.Suite
	repo testutil.TestRepository
	// Directory the test is running in. In setUp(), we `cd` into `repo`'s
	// working directory. In tearDown(), we return to `testDir`.
	testDir string
}

func (suite *BranchTestSuite) SetupTest() {
	repo := testutil.CreateTestRepo()
	os.Chdir(repo.Repo.Workdir())

	suite.repo = repo
}

func (suite *BranchTestSuite) TearDownTest() {
	os.Chdir(suite.testDir)
	suite.repo.Free()
}

// Branches:
//
//	master ─── treecko
func (suite *BranchTestSuite) TestBranch_NewBranchIsChildOfCurrentBranch() {
	suite.repo.BranchWithCommit("treecko")
	initCmd.NewInitCommand().Execute()

	cmd := NewBranchCommand()
	cmd.SetArgs([]string{"grovyle"})
	cmd.Execute()

	assert.True(suite.T(), suite.repo.IsBranchAncestor("treecko", "grovyle"),
		"Expected branch %q to be an ancestor of %q, but it is not", "treecko", "grovyle")
}

// Branches:
//
//	master ─── treecko
func (suite *BranchTestSuite) TestBranch_HeadIsAtNewBranch() {
	suite.repo.BranchWithCommit("treecko")
	initCmd.NewInitCommand().Execute()

	cmd := NewBranchCommand()
	cmd.SetArgs([]string{"grovyle"})
	cmd.Execute()

	head, _ := suite.repo.Repo.Head()
	newBranch := suite.repo.LookupBranch("grovyle")

	assert.Zero(suite.T(), head.Cmp(newBranch.Reference),
		"Expected HEAD to be at branch %q, but it is not", "grovyle")
}

// Branches:
//
//	master ─── treecko
func (suite *BranchTestSuite) TestBranch_UpdatesBranchMapFile() {
	suite.repo.BranchWithCommit("treecko")
	initCmd.NewInitCommand().Execute()

	cmd := NewBranchCommand()
	cmd.SetArgs([]string{"grovyle"})
	cmd.Execute()

	gotString := suite.repo.ReadFile(".git/tree/branches")
	wantString :=
		`git-tree-root
git-tree-root master
master treecko
treecko grovyle`

	assert.Equal(suite.T(), gotString, wantString,
		"Got branch map file: %v, but want file: %v", gotString, wantString)
}

func containsAll(s string, substr ...string) bool {
	for _, sub := range substr {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
func TestBranchWithGitInitedTestSuite(t *testing.T) {
	suite.Run(t, new(BranchWithGitInitedTestSuite))
}

func TestBranchTestSuite(t *testing.T) {
	suite.Run(t, new(BranchTestSuite))
}

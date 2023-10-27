package commands

import (
	"os"
	"testing"

	"github.com/acamadeo/git-tree/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ObsoleteTestSuite struct {
	suite.Suite
	repo testutil.TestRepository
	// Directory the test is running in. In setUp(), we `cd` into `repo`'s
	// working directory. In tearDown(), we return to `testDir`.
	testDir string
}

func (suite *ObsoleteTestSuite) SetupTest() {
	repo := testutil.CreateTestRepo()
	os.Chdir(repo.Repo.Workdir())

	suite.repo = repo
}

func (suite *ObsoleteTestSuite) TearDownTest() {
	os.Chdir(suite.testDir)
	suite.repo.Free()
}

func (suite *ObsoleteTestSuite) TestObsolete_PreCommit_ValidArgs() {
	suite.repo.BranchWithCommit("treecko")

	cmd := NewObsoleteCommand()
	cmd.SetArgs([]string{"obsolete", "pre-commit"})
	gotError := cmd.Execute()

	var wantError error = nil
	assert.Equal(suite.T(), gotError, wantError,
		"Command got error %v, but want error %v", gotError, wantError)
}

func (suite *ObsoleteTestSuite) TestObsolete_PostCommit_ValidArgs() {
	suite.repo.BranchWithCommit("treecko")

	cmd := NewObsoleteCommand()
	cmd.SetArgs([]string{"obsolete", "post-commit"})
	gotError := cmd.Execute()

	var wantError error = nil
	assert.Equal(suite.T(), gotError, wantError,
		"Command got error %v, but want error %v", gotError, wantError)
}

func TestObsoleteTestSuite(t *testing.T) {
	suite.Run(t, new(ObsoleteTestSuite))
}

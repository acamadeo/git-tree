package init

import (
	"errors"
	"os"
	"testing"

	"github.com/acamadeo/git-tree/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type InitTestSuite struct {
	suite.Suite
	repo testutil.TestRepository
	// Directory the test is running in. In setUp(), we `cd` into `repo`'s
	// working directory. In tearDown(), we return to `testDir`.
	testDir string
}

func (suite *InitTestSuite) SetupTest() {
	repo := testutil.CreateTestRepo()
	os.Chdir(repo.Repo.Workdir())

	suite.repo = repo
}

func (suite *InitTestSuite) TearDownTest() {
	os.Chdir(suite.testDir)
	suite.repo.Free()
}

func (suite *InitTestSuite) TestInit_BranchDoesNotExist_RepoWithBranches() {
	suite.repo.BranchWithCommit("treecko")

	cmd := NewInitCommand()
	cmd.SetArgs([]string{"-b", "mudkip"})
	gotError := cmd.Execute()

	wantError := errors.New("Branch \"mudkip\" does not exist in the git repository.")
	assert.Equal(suite.T(), gotError.Error(), wantError.Error(),
		"Command got error %v, but want error %v", gotError, wantError)
}

func (suite *InitTestSuite) TestInit_BranchDoesNotExist_RepoBranchless() {
	cmd := NewInitCommand()
	cmd.SetArgs([]string{"-b", "mudkip"})
	gotError := cmd.Execute()

	wantError := errors.New("Branch \"mudkip\" does not exist in the git repository.")
	assert.Equal(suite.T(), gotError.Error(), wantError.Error(),
		"Command got error %v, but want error %v", gotError, wantError)
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

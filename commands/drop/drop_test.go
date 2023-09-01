package drop

import (
	"os"
	"testing"

	"github.com/acamadeo/git-tree/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DropTestSuite struct {
	suite.Suite
	repo testutil.TestRepository
	// Directory the test is running in. In setUp(), we `cd` into `repo`'s
	// working directory. In tearDown(), we return to `testDir`.
	testDir string
}

func (suite *DropTestSuite) SetupTest() {
	repo := testutil.CreateTestRepo()
	os.Chdir(repo.Repo.Workdir())

	suite.repo = repo
}

func (suite *DropTestSuite) TearDownTest() {
	os.Chdir(suite.testDir)
	suite.repo.Free()
}

func (suite *DropTestSuite) TestDrop_NoErrorIfGitTreeNotInitialized() {
	gotError := NewDropCommand().Execute()

	assert.Nil(suite.T(), gotError,
		"Command got error %v, but want error %v", gotError, nil)
}

func TestDropTestSuite(t *testing.T) {
	suite.Run(t, new(DropTestSuite))
}

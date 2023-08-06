package drop

import (
	"os"
	"testing"

	"github.com/abaresk/git-tree/testutil"
)

type testEnv struct {
	repo testutil.TestRepository
	// Directory the test is running in. In setUp(), we `cd` into `repo`'s
	// working directory. In tearDown(), we return to `testDir`.
	testDir string
}

func setUp(t *testing.T) testEnv {
	repo := testutil.CreateTestRepo()
	os.Chdir(repo.Repo.Workdir())

	return testEnv{
		repo: repo,
	}
}

func (env *testEnv) tearDown(t *testing.T) {
	os.Chdir(env.testDir)
	env.repo.Free()
}

func TestDrop_NoErrorIfGitTreeNotInitialized(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	gotError := NewDropCommand().Execute()

	if gotError != nil {
		t.Errorf("Command got error %v, but want error %v", gotError, nil)
	}
}

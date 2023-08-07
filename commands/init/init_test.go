package init

import (
	"errors"
	"os"
	"testing"

	"github.com/acamadeo/git-tree/testutil"
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

func TestInit_BranchDoesNotExist_RepoWithBranches(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	env.repo.BranchWithCommit("treecko")

	cmd := NewInitCommand()
	cmd.SetArgs([]string{"-b", "mudkip"})
	gotError := cmd.Execute()

	wantError := errors.New("Branch \"mudkip\" does not exist in the git repository.")
	if gotError.Error() != wantError.Error() {
		t.Errorf("Command got error %v, but want error %v", gotError, wantError)
	}
}

func TestInit_BranchDoesNotExist_RepoBranchless(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	cmd := NewInitCommand()
	cmd.SetArgs([]string{"-b", "mudkip"})
	gotError := cmd.Execute()

	wantError := errors.New("Branch \"mudkip\" does not exist in the git repository.")
	if gotError.Error() != wantError.Error() {
		t.Errorf("Command got error %v, but want error %v", gotError, wantError)
	}
}

package rebase

import (
	"errors"
	"os"
	"testing"

	initCmd "github.com/abaresk/git-tree/commands/init"
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

func TestRebase_ErrorIfGitTreeNotInitialized(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	gotError := NewRebaseCommand().Execute()

	wantError := errors.New("git-tree is not initialized. Run `git-tree init` to initialize.")

	if gotError.Error() != wantError.Error() {
		t.Errorf("Command got error %v, but want error %v", gotError, wantError)
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
func TestRebase_SourceBranchDoesNotExist(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("treecko")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("mudkip")

	// Run git-tree init.
	initCmd.NewInitCommand().Execute()

	cmd := NewRebaseCommand()
	cmd.SetArgs([]string{"-s", "torchic", "-d", "treecko"})
	gotError := cmd.Execute()

	wantError := errors.New("Could not find source branch \"torchic\".")
	if gotError.Error() != wantError.Error() {
		t.Errorf("Command got error %v, but want error %v", gotError, wantError)
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
func TestRebase_DestBranchDoesNotExist(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("treecko")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("mudkip")

	// Run git-tree init.
	initCmd.NewInitCommand().Execute()

	cmd := NewRebaseCommand()
	cmd.SetArgs([]string{"-s", "treecko", "-d", "torchic"})
	gotError := cmd.Execute()

	wantError := errors.New("Could not find dest branch \"torchic\".")
	if gotError.Error() != wantError.Error() {
		t.Errorf("Command got error %v, but want error %v", gotError, wantError)
	}
}

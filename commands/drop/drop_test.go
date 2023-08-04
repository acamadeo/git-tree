package drop

import (
	"os"
	"testing"

	initCmd "github.com/abaresk/git-tree/commands/init"
	"github.com/abaresk/git-tree/store"
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

	// Run git-tree init.
	initCmd.NewInitCommand().Execute()

	return testEnv{
		repo: repo,
	}
}

func (env *testEnv) tearDown(t *testing.T) {
	os.Chdir(env.testDir)
	env.repo.Free()
}

func TestDrop_DeletesRootBranch(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	NewDropCommand().Execute()

	rootBranch := env.repo.LookupBranch("git-tree-root")
	if rootBranch != nil {
		t.Errorf("Expected nil root branch but got %v", rootBranch)
	}
}

func TestDrop_DeletesGitTreeSubdir(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	NewDropCommand().Execute()

	dirName := env.repo.Repo.Path() + "/tree"
	if store.DirExists(dirName) {
		t.Errorf("Expected directory %q not to exist, but it does", dirName)
	}
}

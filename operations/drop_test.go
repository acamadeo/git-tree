package operations

import (
	"testing"

	"github.com/acamadeo/git-tree/store"
	"github.com/acamadeo/git-tree/testutil"
)

type testEnvDrop struct {
	repo testutil.TestRepository
}

func setUpDrop(t *testing.T) testEnvDrop {
	repo := testutil.CreateTestRepo()

	// Run git-tree init.
	Init(repo.Repo)

	return testEnvDrop{
		repo: repo,
	}
}

func (env *testEnvDrop) tearDown(t *testing.T) {
	env.repo.Free()
}

func TestDrop_DeletesRootBranch(t *testing.T) {
	env := setUpDrop(t)
	defer env.tearDown(t)

	Drop(env.repo.Repo)

	rootBranch := env.repo.LookupBranch("git-tree-root")
	if rootBranch != nil {
		t.Errorf("Expected nil root branch but got %v", rootBranch)
	}
}

func TestDrop_DeletesGitTreeSubdir(t *testing.T) {
	env := setUpDrop(t)
	defer env.tearDown(t)

	Drop(env.repo.Repo)

	dirName := env.repo.Repo.Path() + "/tree"
	if store.DirExists(dirName) {
		t.Errorf("Expected directory %q not to exist, but it does", dirName)
	}
}

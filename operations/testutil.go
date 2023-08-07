package operations

import (
	"testing"

	"github.com/acamadeo/git-tree/testutil"
)

type testEnv struct {
	repo testutil.TestRepository
}

func setUp(t *testing.T) testEnv {
	return testEnv{
		repo: testutil.CreateTestRepo(),
	}
}

func (env *testEnv) tearDown(t *testing.T) {
	env.repo.Free()
}

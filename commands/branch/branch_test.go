package branch

import (
	"errors"
	"os"
	"strings"
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
func setUpWithGitTreeInit(t *testing.T) testEnv {
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

func TestBranch_BranchAlreadyExists(t *testing.T) {
	env := setUpWithGitTreeInit(t)
	defer env.tearDown(t)

	env.repo.BranchWithCommit("treecko")

	cmd := NewBranchCommand()
	cmd.SetArgs([]string{"treecko"})
	gotError := cmd.Execute()

	wantError := errors.New("Branch \"treecko\" already exists in the git repository.")
	if gotError.Error() != wantError.Error() {
		t.Errorf("Command got error %v, but want error %v", gotError, wantError)
	}
}

// Branches:
//
//	                    ┌───── *mudkip
//	                    ▼
//	master ─── mud -> kip
func TestBranch_NotOnTipCommit(t *testing.T) {
	env := setUpWithGitTreeInit(t)
	defer env.tearDown(t)

	env.repo.CreateBranch("mudkip")
	env.repo.SwitchBranch("mudkip")
	env.repo.WriteAndCommitFile("mud", "mud", "mud")
	env.repo.WriteAndCommitFile("kip", "kip", "kip")
	env.repo.SwitchCommit("mud")

	cmd := NewBranchCommand()
	cmd.SetArgs([]string{"treecko"})
	gotError := cmd.Execute()

	if !containsAll(gotError.Error(), "HEAD commit", "is not pointed to by any tracked branches") {
		t.Errorf("Command got invalid error %v", gotError)
	}
}

// Branches:
//
//	master ─── treecko
func TestBranch_NewBranchIsChildOfCurrentBranch(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	env.repo.BranchWithCommit("treecko")
	initCmd.NewInitCommand().Execute()

	cmd := NewBranchCommand()
	cmd.SetArgs([]string{"grovyle"})
	cmd.Execute()

	if !env.repo.IsBranchAncestor("treecko", "grovyle") {
		t.Errorf("Expected branch %q to be an ancestor of %q, but it is not", "treecko", "grovyle")
	}
}

// Branches:
//
//	master ─── treecko
func TestBranch_HeadIsAtNewBranch(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	env.repo.BranchWithCommit("treecko")
	initCmd.NewInitCommand().Execute()

	cmd := NewBranchCommand()
	cmd.SetArgs([]string{"grovyle"})
	cmd.Execute()

	head, _ := env.repo.Repo.Head()
	newBranch := env.repo.LookupBranch("grovyle")

	if head.Cmp(newBranch.Reference) != 0 {
		t.Errorf("Expected HEAD to be at branch %q, but it is not", "grovyle")
	}
}

// Branches:
//
//	master ─── treecko
func TestBranch_UpdatesBranchMapFile(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	env.repo.BranchWithCommit("treecko")
	initCmd.NewInitCommand().Execute()

	cmd := NewBranchCommand()
	cmd.SetArgs([]string{"grovyle"})
	cmd.Execute()

	gotString := env.repo.ReadFile(".git/tree/branches")
	wantString :=
		`git-tree-root
git-tree-root master
master treecko
treecko grovyle`

	if gotString != wantString {
		t.Errorf("Got branch map file: %v, but want file: %v", gotString, wantString)
	}
}

func containsAll(s string, substr ...string) bool {
	for _, sub := range substr {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}

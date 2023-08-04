package testutil

import (
	"os"
	"time"

	gitutil "github.com/abaresk/git-tree/git"
	"github.com/abaresk/git-tree/store"
	git "github.com/libgit2/git2go/v34"
)

var signature = &git.Signature{
	Name:  "test",
	Email: "test@gmail.com",
	When:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
}

type TestRepository struct {
	Repo *git.Repository
}

func CreateTestRepo() TestRepository {
	tempDir, _ := os.MkdirTemp("", "test-git")
	repo, _ := git.InitRepository(tempDir, false)

	// Create the initial commit.
	testRepo := TestRepository{Repo: repo}
	testRepo.WriteAndCommitFile("README.md", "", "Initial commit")
	return testRepo
}

func (t *TestRepository) Free() {
	path := t.Repo.Workdir()
	t.Repo.Free()
	os.RemoveAll(path)
}

// Create a branch at the current HEAD.
func (t *TestRepository) CreateBranch(name string) *git.Branch {
	head, _ := t.Repo.Head()
	headCommit, _ := t.Repo.LookupCommit(head.Target())
	branch, _ := t.Repo.CreateBranch(name, headCommit, false)
	return branch
}

// Move HEAD to the specified branch. This assumes the specified branch exists.
func (t *TestRepository) SwitchBranch(name string) {
	gitutil.CheckoutBranchByName(t.Repo, name)
}

func (t *TestRepository) LookupBranch(name string) *git.Branch {
	branch, _ := t.Repo.LookupBranch(name, git.BranchLocal)
	return branch
}

// Write to `contents` to file `filename` under the working directory.
//
// This overwrites the contents of the file if the file already exists.
func (t *TestRepository) WriteFile(filename string, contents string) {
	store.OverwriteFile(t.Repo.Workdir()+filename, contents)
}

// Return contents of a file as a string.
func (t *TestRepository) ReadFile(filename string) string {
	return store.ReadFile(t.Repo.Workdir() + filename)
}

// Stage the the specified files. If no argument is provided, all unstaged files
// are staged.
func (t *TestRepository) StageFiles(names ...string) {
	if len(names) == 0 {
		names = []string{"."}
	}

	index, _ := t.Repo.Index()
	index.AddAll(names, git.IndexAddDefault, nil)
	index.Write()
}

// Commit the staged files with the specified `message`.
func (t *TestRepository) WriteCommit(message string) {
	index, _ := t.Repo.Index()
	treeOid, _ := index.WriteTree()
	index.Write()

	tree, _ := t.Repo.LookupTree(treeOid)

	parentObj, _, _ := t.Repo.RevparseExt("HEAD")
	if parentObj != nil {
		// Not the first commit (i.e., there's a parent).
		parent, _ := t.Repo.LookupCommit(parentObj.Id())
		t.Repo.CreateCommit("HEAD", signature, signature, message, tree, parent)
	} else {
		t.Repo.CreateCommit("HEAD", signature, signature, message, tree)
	}
}

// Write to a file and commit the changes to the HEAD branch.
func (t *TestRepository) WriteAndCommitFile(filename string, contents string, message string) {
	t.WriteFile(filename, contents)
	t.StageFiles()
	t.WriteCommit(message)
}

// Creates a new branch off HEAD, adding and committing a file to the new branch.
//
// The branch name, file name, file contents, and commit message are all `name`.
func (t *TestRepository) BranchWithCommit(name string) {
	t.CreateBranch(name)
	t.SwitchBranch(name)
	t.WriteAndCommitFile(name, name, name)
}

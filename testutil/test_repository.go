package testutil

import (
	"os"
	"time"

	gitutil "github.com/acamadeo/git-tree/git"
	"github.com/acamadeo/git-tree/utils"
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

// Create a branch off current HEAD and switch to the new branch.
func (t *TestRepository) CreateAndSwitchBranch(name string) {
	t.CreateBranch(name)
	t.SwitchBranch(name)
}

func (t *TestRepository) LookupBranch(name string) *git.Branch {
	branch, _ := t.Repo.LookupBranch(name, git.BranchLocal)
	return branch
}

// Returns true if branch `a` is an ancestor of branch `b`.
func (t *TestRepository) IsBranchAncestor(a string, b string) bool {
	branchA := t.LookupBranch(a)
	branchB := t.LookupBranch(b)
	return gitutil.IsBranchAncestor(t.Repo, branchA, branchB)
}

// Write to `contents` to file `filename` under the working directory.
//
// This overwrites the contents of the file if the file already exists.
func (t *TestRepository) WriteFile(filename string, contents string) {
	utils.OverwriteFile(t.Repo.Workdir()+filename, contents)
}

// Return contents of a file as a string.
func (t *TestRepository) ReadFile(filename string) string {
	return utils.ReadFile(t.Repo.Workdir() + filename)
}

// Returns whether the file exists.
func (t *TestRepository) FileExists(filename string) bool {
	return utils.FileExists(t.Repo.Workdir() + filename)
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

// Amend the commit at HEAD with a new commit message.
func (t *TestRepository) AmendCommit(message string) {
	headRef, _ := t.Repo.Head()
	headCommit, _ := t.Repo.LookupCommit(headRef.Target())

	index, _ := t.Repo.Index()
	treeOid, _ := index.WriteTree()
	index.Write()
	tree, _ := t.Repo.LookupTree(treeOid)

	headCommit.Amend("HEAD", nil, nil, message, tree)
}

// Moved HEAD to the commit with message `message`. Assumes a single commit with
// the specified message exists.
func (t *TestRepository) SwitchCommit(message string) {
	allCommits := gitutil.AllLocalCommits(t.Repo, nil)
	for _, commit := range allCommits {
		if commit.Message() == message {
			gitutil.CheckoutCommit(t.Repo, commit)
			break
		}
	}
}

// Creates a new branch off HEAD, adding and committing a file to the new branch.
//
// The branch name, file name, file contents, and commit message are all `name`.
func (t *TestRepository) BranchWithCommit(name string) {
	t.CreateBranch(name)
	t.SwitchBranch(name)
	t.WriteAndCommitFile(name, name, name)
}

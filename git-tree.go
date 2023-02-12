package main

import (
	"fmt"
	"os"

	"github.com/abaresk/git-tree/commands"
	git "github.com/libgit2/git2go/v34"
	"github.com/zyedidia/generic/queue"
)

func main() {
	q := queue.New[int]()
	q.Enqueue(1)

	fmt.Println(q.Dequeue())

	pwd, _ := os.Getwd()
	repo, err := git.OpenRepository(pwd)
	if err != nil {
		fmt.Printf("Current directory %q is not a git repository.", pwd)
	}

	context := &commands.Context{
		Repo: repo,
	}
	_ = context
}

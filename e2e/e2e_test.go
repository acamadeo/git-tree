package e2e

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/acamadeo/git-tree/commands"
	"github.com/acamadeo/git-tree/store"
	"github.com/rogpeppe/go-internal/testscript"
)

var reCarriageReturn = regexp.MustCompile(`\r`)

// Paste the contents in file `$EDITOR_INPUT` into Git's editor file.
func runEditor() int {
	editorOut := os.Args[1]
	input := store.ReadFile(os.Getenv("EDITOR_INPUT"))
	store.OverwriteFile(editorOut, input)
	return 0
}

// Paste the contents in file `$SEQ_EDITOR_INPUT` into Git's sequence editor file.
func runSequenceEditor() int {
	editorOut := os.Args[1]
	input := store.ReadFile(os.Getenv("SEQ_EDITOR_INPUT"))
	store.OverwriteFile(editorOut, input)
	return 0
}

func runWriteFile() int {
	filename, contents := os.Args[1], os.Args[2]
	store.OverwriteFile(filename, contents)
	return 0
}

// Compare whether two files have the same contents.
//
// Strips carriage return characters.
func runCompare() int {
	file1, file2 := os.Args[1], os.Args[2]
	contents1, contents2 := store.ReadFile(file1), store.ReadFile(file2)
	contents1, contents2 = reCarriageReturn.ReplaceAllString(contents1, ""),
		reCarriageReturn.ReplaceAllString(contents2, "")
	if contents1 != contents2 {
		fmt.Fprintf(os.Stderr, `file contents differ::
%q:
%s
%q:
%s`, file1, contents1, file2, contents2)
		return 1
	}
	return 0
}

// TODO: Add remaining test cases from the document.
func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"git-tree":   commands.Main,
		"editor":     runEditor,
		"seq_editor": runSequenceEditor,
		"write_file": runWriteFile,
		"compare":    runCompare,
	}))
}

func TestGitTree(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}

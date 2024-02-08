package e2e

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/acamadeo/git-tree/commands"
	"github.com/acamadeo/git-tree/utils"
	"github.com/rogpeppe/go-internal/testscript"
)

const timeout = 10 * time.Second

var reCarriageReturn = regexp.MustCompile(`\r`)

func testscriptParams(dir string) testscript.Params {
	return testscript.Params{
		Dir:      dir,
		Deadline: time.Now().Add(timeout),
	}
}

// Paste the contents in file `$EDITOR_INPUT` into Git's editor file.
func runEditor() int {
	editorOut := os.Args[1]
	input := utils.ReadFile(os.Getenv("EDITOR_INPUT"))
	utils.OverwriteFile(editorOut, input)
	return 0
}

// Paste the contents in file `$SEQ_EDITOR_INPUT` into Git's sequence editor file.
func runSequenceEditor() int {
	editorOut := os.Args[1]
	input := utils.ReadFile(os.Getenv("SEQ_EDITOR_INPUT"))
	utils.OverwriteFile(editorOut, input)
	return 0
}

func runWriteFile() int {
	filename, contents := os.Args[1], os.Args[2]
	utils.OverwriteFile(filename, contents)
	return 0
}

// Compare whether two files have the same contents.
//
// Strips carriage return characters.
func runCompare() int {
	file1, file2 := os.Args[1], os.Args[2]
	contents1, contents2 := utils.ReadFile(file1), utils.ReadFile(file2)
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

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"git-tree":   commands.Main,
		"editor":     runEditor,
		"seq_editor": runSequenceEditor,
		"write_file": runWriteFile,
		"compare":    runCompare,
	}))
}

func TestE2eObsolete(t *testing.T) {
	testscript.Run(t, testscriptParams("obsolete"))
}

func TestE2eEvolveSimple(t *testing.T) {
	testscript.Run(t, testscriptParams("evolve/simple"))
}

func TestE2eEvolveNested(t *testing.T) {
	testscript.Run(t, testscriptParams("evolve/nested"))
}

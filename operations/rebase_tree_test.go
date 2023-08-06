package operations

import (
	"errors"
	"testing"

	gitutil "github.com/abaresk/git-tree/git"
	"github.com/abaresk/git-tree/testutil"
)

// -------------------------------------------------------------------------- \
// RebaseTree                                                                 |
// -------------------------------------------------------------------------- /

// Initial:
//
//	master ─── mew
func TestRebaseTree_SourceAndDestCannotBeTheSame(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("mew")
	dest := env.repo.LookupBranch("mew")
	gotResult := RebaseTree(env.repo.Repo, source, dest)

	wantError := errors.New("Source and destination cannot be the same")

	if !(gotResult.Type == RebaseTreeError && gotResult.Error.Error() == wantError.Error()) {
		t.Errorf("Operation got error %v, but want error %v", gotResult.Error, wantError)
	}
}

// Initial:
//
//	master ─── treecko ─── grovyle
func TestRebaseTree_SourceCannotBeAncestorOfDest(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("treecko")
	env.repo.BranchWithCommit("grovyle")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("grovyle")
	gotResult := RebaseTree(env.repo.Repo, source, dest)

	wantError := errors.New("Source cannot be an ancestor of destination")

	if !(gotResult.Type == RebaseTreeError && gotResult.Error.Error() == wantError.Error()) {
		t.Errorf("Operation got error %v, but want error %v", gotResult.Error, wantError)
	}
}

// Initial:
//
//	master ─── treecko ─── grovyle
func TestRebaseTree_SourceCannotBeDirectChildOfDest(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("treecko")
	env.repo.BranchWithCommit("grovyle")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("grovyle")
	dest := env.repo.LookupBranch("treecko")
	gotResult := RebaseTree(env.repo.Repo, source, dest)

	wantError := errors.New("Source is already a child of destination")

	if !(gotResult.Type == RebaseTreeError && gotResult.Error.Error() == wantError.Error()) {
		t.Errorf("Operation got error %v, but want error %v", gotResult.Error, wantError)
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
//
// Result:
//
//	master ─── mew ─── mudkip ─── treecko
func TestRebaseTree_RebaseOneChild(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("treecko")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("mudkip")
	// TODO: Figure out why root is at mew instead of master!
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mudkip")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("mudkip")
	expectedRepo.BranchWithCommit("treecko")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle
//	                └─ mudkip
//
// Result:
//
//	master ─── mew ─── mudkip ─── treecko ─── grovyle
func TestRebaseTree_RebaseMultipleChildren(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("treecko")
	env.repo.BranchWithCommit("grovyle")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("mudkip")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mudkip")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("mudkip")
	expectedRepo.BranchWithCommit("treecko")
	expectedRepo.BranchWithCommit("grovyle")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle
//	                └─ mudkip
//
// Result:
//
//	master ─── mew ─── treecko ─── grovyle ─── mudkip
func TestRebaseTree_RebaseOntoNestedBranch(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("treecko")
	env.repo.BranchWithCommit("grovyle")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("mudkip")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("mudkip")
	dest := env.repo.LookupBranch("grovyle")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("treecko")
	expectedRepo.BranchWithCommit("grovyle")
	expectedRepo.BranchWithCommit("mudkip")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─── mudkip ─── treecko ─── grovyle
//
// Result:
//
//	master ─── mew ─┬─ treecko ─── grovyle
//	                └─ mudkip
func TestRebaseTree_ForkBranchLine(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("mudkip")
	env.repo.BranchWithCommit("treecko")
	env.repo.BranchWithCommit("grovyle")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mew")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("treecko")
	expectedRepo.BranchWithCommit("grovyle")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("mudkip")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── eevee ─── flareon ─── jolteon ─── vaporeon
//
// Result:
//
//	master ─── eevee ─┬─ vaporeon
//	                  ├─ jolteon
//	                  └─ flareon
func TestRebaseTree_MultipleRebases_Fork(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("eevee")
	env.repo.BranchWithCommit("flareon")
	env.repo.BranchWithCommit("jolteon")
	env.repo.BranchWithCommit("vaporeon")
	Init(env.repo.Repo)

	// Rebase tree operations
	source := env.repo.LookupBranch("jolteon")
	dest := env.repo.LookupBranch("eevee")
	RebaseTree(env.repo.Repo, source, dest)

	source = env.repo.LookupBranch("vaporeon")
	dest = env.repo.LookupBranch("eevee")
	RebaseTree(env.repo.Repo, source, dest)

	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("eevee")
	expectedRepo.BranchWithCommit("vaporeon")
	expectedRepo.SwitchBranch("eevee")
	expectedRepo.BranchWithCommit("jolteon")
	expectedRepo.SwitchBranch("eevee")
	expectedRepo.BranchWithCommit("flareon")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── eevee ─┬─ vaporeon
//	                  ├─ jolteon
//	                  └─ flareon
//
// Result:
//
//	master ─── eevee ─── flareon ─── jolteon ─── vaporeon
func TestRebaseTree_MultipleRebases_Merge(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("eevee")
	env.repo.BranchWithCommit("vaporeon")
	env.repo.SwitchBranch("eevee")
	env.repo.BranchWithCommit("jolteon")
	env.repo.SwitchBranch("eevee")
	env.repo.BranchWithCommit("flareon")
	Init(env.repo.Repo)

	// Rebase tree operations
	source := env.repo.LookupBranch("jolteon")
	dest := env.repo.LookupBranch("flareon")
	RebaseTree(env.repo.Repo, source, dest)

	source = env.repo.LookupBranch("vaporeon")
	dest = env.repo.LookupBranch("jolteon")
	RebaseTree(env.repo.Repo, source, dest)

	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("eevee")
	expectedRepo.BranchWithCommit("flareon")
	expectedRepo.BranchWithCommit("jolteon")
	expectedRepo.BranchWithCommit("vaporeon")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─── treecko ─── grovyle
//
// Result:
//
//	master ─┬─ mew
//	        └─ treecko ───grovyle
func TestRebaseTree_RebaseOntoFirstBranch(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("treecko")
	env.repo.BranchWithCommit("grovyle")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("master")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.SwitchBranch("master")
	expectedRepo.BranchWithCommit("treecko")
	expectedRepo.BranchWithCommit("grovyle")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts
//	                └─ snorunt ─┬─ glalie ───  kirlia ─┬─ gardevoir
//	                            └─ froslass            └─ gallade
func TestRebaseTree_KirliaOntoGlalie(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("ralts")
	env.repo.BranchWithCommit("kirlia")
	env.repo.BranchWithCommit("gardevoir")
	env.repo.SwitchBranch("kirlia")
	env.repo.BranchWithCommit("gallade")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("snorunt")
	env.repo.BranchWithCommit("glalie")
	env.repo.SwitchBranch("snorunt")
	env.repo.BranchWithCommit("froslass")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("kirlia")
	dest := env.repo.LookupBranch("glalie")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─── ralts ─── kirlia ─┬─ gardevoir ─── snorunt ─┬─ glalie
//	                                     └─ gallade                └─ froslass
func TestRebaseTree_SnoruntOntoGardevoir(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("ralts")
	env.repo.BranchWithCommit("kirlia")
	env.repo.BranchWithCommit("gardevoir")
	env.repo.SwitchBranch("kirlia")
	env.repo.BranchWithCommit("gallade")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("snorunt")
	env.repo.BranchWithCommit("glalie")
	env.repo.SwitchBranch("snorunt")
	env.repo.BranchWithCommit("froslass")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("snorunt")
	dest := env.repo.LookupBranch("gardevoir")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts
//	                |
//	                └─ snorunt ─┬─ glalie
//	                            ├─ froslass
//	                            └─ kirlia ─┬─ gardevoir
//	                                       └─ gallade
func TestRebaseTree_KirliaOntoSnorunt(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("ralts")
	env.repo.BranchWithCommit("kirlia")
	env.repo.BranchWithCommit("gardevoir")
	env.repo.SwitchBranch("kirlia")
	env.repo.BranchWithCommit("gallade")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("snorunt")
	env.repo.BranchWithCommit("glalie")
	env.repo.SwitchBranch("snorunt")
	env.repo.BranchWithCommit("froslass")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("kirlia")
	dest := env.repo.LookupBranch("snorunt")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─── ralts ─── kirlia ─┬─ gardevoir
//	                                     ├─ gallade
//	                                     └─ snorunt ─┬─ glalie
//	                                                 └─ froslass
func TestRebaseTree_SnoruntOntoKirlia(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("ralts")
	env.repo.BranchWithCommit("kirlia")
	env.repo.BranchWithCommit("gardevoir")
	env.repo.SwitchBranch("kirlia")
	env.repo.BranchWithCommit("gallade")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("snorunt")
	env.repo.BranchWithCommit("glalie")
	env.repo.SwitchBranch("snorunt")
	env.repo.BranchWithCommit("froslass")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("snorunt")
	dest := env.repo.LookupBranch("kirlia")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      ├─ gallade
//	                |                      └─ glalie
//	                └─ snorunt ─── froslass
func TestRebaseTree_GlalieOntoKirlia(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("ralts")
	env.repo.BranchWithCommit("kirlia")
	env.repo.BranchWithCommit("gardevoir")
	env.repo.SwitchBranch("kirlia")
	env.repo.BranchWithCommit("gallade")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("snorunt")
	env.repo.BranchWithCommit("glalie")
	env.repo.SwitchBranch("snorunt")
	env.repo.BranchWithCommit("froslass")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("glalie")
	dest := env.repo.LookupBranch("kirlia")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("froslass")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─── gallade
//	                └─ snorunt ─┬─ glalie
//	                            ├─ froslass
//	                            └─ gardevoir
func TestRebaseTree_GardevoirOntoSnorunt(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("ralts")
	env.repo.BranchWithCommit("kirlia")
	env.repo.BranchWithCommit("gardevoir")
	env.repo.SwitchBranch("kirlia")
	env.repo.BranchWithCommit("gallade")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("snorunt")
	env.repo.BranchWithCommit("glalie")
	env.repo.SwitchBranch("snorunt")
	env.repo.BranchWithCommit("froslass")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("gardevoir")
	dest := env.repo.LookupBranch("snorunt")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("gardevoir")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir ─── glalie
//	                |                      └─ gallade
//	                └─ snorunt ─── froslass
func TestRebaseTree_GlalieOntoGardevoir(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("ralts")
	env.repo.BranchWithCommit("kirlia")
	env.repo.BranchWithCommit("gardevoir")
	env.repo.SwitchBranch("kirlia")
	env.repo.BranchWithCommit("gallade")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("snorunt")
	env.repo.BranchWithCommit("glalie")
	env.repo.SwitchBranch("snorunt")
	env.repo.BranchWithCommit("froslass")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("glalie")
	dest := env.repo.LookupBranch("gardevoir")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.SwitchBranch("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("froslass")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─┬─ gardevoir
//	                |                      └─ gallade
//	                └─ snorunt ─┬─ glalie
//	                            └─ froslass
//
// Result:
//
//	master ─── mew ─┬─ ralts ───── kirlia ─── gallade
//	                |
//	                └─ snorunt ─┬─ glalie ─── gardevoir
//	                            └─ froslass
func TestRebaseTree_GardevoirOntoGlalie(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("ralts")
	env.repo.BranchWithCommit("kirlia")
	env.repo.BranchWithCommit("gardevoir")
	env.repo.SwitchBranch("kirlia")
	env.repo.BranchWithCommit("gallade")
	env.repo.SwitchBranch("mew")
	env.repo.BranchWithCommit("snorunt")
	env.repo.BranchWithCommit("glalie")
	env.repo.SwitchBranch("snorunt")
	env.repo.BranchWithCommit("froslass")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("gardevoir")
	dest := env.repo.LookupBranch("glalie")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	Drop(env.repo.Repo)

	// Setup expected
	expectedRepo := testutil.CreateTestRepo()
	defer expectedRepo.Free()

	expectedRepo.BranchWithCommit("mew")
	expectedRepo.BranchWithCommit("ralts")
	expectedRepo.BranchWithCommit("kirlia")
	expectedRepo.BranchWithCommit("gallade")
	expectedRepo.SwitchBranch("mew")
	expectedRepo.BranchWithCommit("snorunt")
	expectedRepo.BranchWithCommit("glalie")
	expectedRepo.BranchWithCommit("gardevoir")
	expectedRepo.SwitchBranch("snorunt")
	expectedRepo.BranchWithCommit("froslass")

	if !gitutil.TreesEqual(env.repo.Repo, expectedRepo.Repo) {
		t.Error("Expected rebased repository to match expected, but it does not")
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
func TestRebaseTree_MergeConflict_Result(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial - write conflicting contents to the same file.
	env.repo.BranchWithCommit("mew")
	env.repo.CreateAndSwitchBranch("treecko")
	env.repo.WriteAndCommitFile("starter", "treecko", "treecko")
	env.repo.SwitchBranch("mew")
	env.repo.CreateAndSwitchBranch("mudkip")
	env.repo.WriteAndCommitFile("starter", "mudkip", "mudkip")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mudkip")
	gotResult := RebaseTree(env.repo.Repo, source, dest)

	if gotResult.Type != RebaseTreeMergeConflict {
		t.Error("Operation did not yield merge conflict, but merge conflict expected")
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
func TestRebaseTree_MergeConflict_CannotCallRebaseTreeAgain(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial - write conflicting contents to the same file.
	env.repo.BranchWithCommit("mew")
	env.repo.CreateAndSwitchBranch("treecko")
	env.repo.WriteAndCommitFile("starter", "treecko", "treecko")
	env.repo.SwitchBranch("mew")
	env.repo.CreateAndSwitchBranch("mudkip")
	env.repo.WriteAndCommitFile("starter", "mudkip", "mudkip")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mudkip")
	RebaseTree(env.repo.Repo, source, dest)

	// Try doing Rebase tree again
	gotResult := RebaseTree(env.repo.Repo, source, dest)

	wantError := errors.New("Cannot rebase while another rebase is in progress. Abort or continue the existing rebase")

	if !(gotResult.Type == RebaseTreeError && gotResult.Error.Error() == wantError.Error()) {
		t.Errorf("Operation got error %v, but want error %v", gotResult.Error, wantError)
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko
//	                └─ mudkip
func TestRebaseTree_MergeConflict_CreatesFiles(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial - write conflicting contents to the same file.
	env.repo.BranchWithCommit("mew")
	env.repo.CreateAndSwitchBranch("treecko")
	env.repo.WriteAndCommitFile("starter", "treecko", "treecko")
	env.repo.SwitchBranch("mew")
	env.repo.CreateAndSwitchBranch("mudkip")
	env.repo.WriteAndCommitFile("starter", "mudkip", "mudkip")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mudkip")
	RebaseTree(env.repo.Repo, source, dest)

	gotString := env.repo.ReadFile(".git/tree/rebasing")
	wantString := ""
	if gotString != wantString {
		t.Errorf("Got rebasing file: %v, but want file: %v", gotString, wantString)
	}

	gotString = env.repo.ReadFile(".git/tree/rebasing-source")
	wantString = "treecko"
	if gotString != wantString {
		t.Errorf("Got rebasing-source file: %v, but want file: %v", gotString, wantString)
	}

	gotString = env.repo.ReadFile(".git/tree/rebasing-dest")
	wantString = "mudkip"
	if gotString != wantString {
		t.Errorf("Got rebasing-dest file: %v, but want file: %v", gotString, wantString)
	}

	gotString = env.repo.ReadFile(".git/tree/rebasing-temps")
	wantString = "rebase-treecko treecko"
	if gotString != wantString {
		t.Errorf("Got rebasing-temps file: %v, but want file: %v", gotString, wantString)
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func TestRebaseTree_MergeConflict_RebasingTempsContainsProperBranches(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial - write conflicting contents to the same file.
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("treecko")
	env.repo.CreateAndSwitchBranch("grovyle")
	env.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	env.repo.BranchWithCommit("sceptile")
	env.repo.SwitchBranch("mew")
	env.repo.CreateAndSwitchBranch("mudkip")
	env.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mudkip")
	RebaseTree(env.repo.Repo, source, dest)

	// Should contain only branches that we attempted to rebase (we never reached `sceptile`).
	gotString := env.repo.ReadFile(".git/tree/rebasing-temps")
	wantString := `rebase-treecko treecko
rebase-grovyle grovyle`
	if gotString != wantString {
		t.Errorf("Got rebasing-temps file: %v, but want file: %v", gotString, wantString)
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func TestRebaseTree_MergeConflict_TemporaryBranchesPointToProperCommits(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial - write conflicting contents to the same file.
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("treecko")
	env.repo.CreateAndSwitchBranch("grovyle")
	env.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	env.repo.BranchWithCommit("sceptile")
	env.repo.SwitchBranch("mew")
	env.repo.CreateAndSwitchBranch("mudkip")
	env.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(env.repo.Repo)

	// Get the oid's of the commits the branches currently point to.
	treeckoOid := env.repo.LookupBranch("treecko").Target()
	grovyleOid := env.repo.LookupBranch("grovyle").Target()

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mudkip")
	RebaseTree(env.repo.Repo, source, dest)

	tempTreeckoOid := env.repo.LookupBranch("rebase-treecko").Target()
	tempGrovyleOid := env.repo.LookupBranch("rebase-grovyle").Target()

	// The temporary branches should point to the commits where the rebased
	// branches used to point to.
	if *treeckoOid != *tempTreeckoOid {
		t.Errorf("Expected temporary branch to point to %v, but it points to %v", *treeckoOid, *tempTreeckoOid)
	}
	if *grovyleOid != *tempGrovyleOid {
		t.Errorf("Expected temporary branch to point to %v, but it points to %v", *grovyleOid, *tempGrovyleOid)
	}
}

// -------------------------------------------------------------------------- \
// RebaseTreeAbort                                                            |
// -------------------------------------------------------------------------- /

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func TestRebaseTreeAbort_MovesRebasedBranchesToOriginalLocation(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial - write conflicting contents to the same file.
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("treecko")
	env.repo.CreateAndSwitchBranch("grovyle")
	env.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	env.repo.BranchWithCommit("sceptile")
	env.repo.SwitchBranch("mew")
	env.repo.CreateAndSwitchBranch("mudkip")
	env.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(env.repo.Repo)

	// Get the oid's of the commits the branches currently point to.
	treeckoOid := env.repo.LookupBranch("treecko").Target()
	grovyleOid := env.repo.LookupBranch("grovyle").Target()

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mudkip")
	RebaseTree(env.repo.Repo, source, dest)

	// Abort the rebase
	RebaseTreeAbort(env.repo.Repo)

	newTreeckoOid := env.repo.LookupBranch("treecko").Target()
	newGrovyleOid := env.repo.LookupBranch("grovyle").Target()

	// After aborting, the rebased branches should point back to their original
	// commits.
	if *treeckoOid != *newTreeckoOid {
		t.Errorf("Expected temporary branch to point to %v, but it points to %v", *treeckoOid, *newTreeckoOid)
	}
	if *grovyleOid != *newGrovyleOid {
		t.Errorf("Expected temporary branch to point to %v, but it points to %v", *grovyleOid, *newGrovyleOid)
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func TestRebaseTreeAbort_DeletesFiles(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial - write conflicting contents to the same file.
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("treecko")
	env.repo.CreateAndSwitchBranch("grovyle")
	env.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	env.repo.BranchWithCommit("sceptile")
	env.repo.SwitchBranch("mew")
	env.repo.CreateAndSwitchBranch("mudkip")
	env.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mudkip")
	RebaseTree(env.repo.Repo, source, dest)

	// Abort the rebase
	RebaseTreeAbort(env.repo.Repo)

	filename := ".git/tree/rebasing"
	if env.repo.FileExists(filename) {
		t.Errorf("Expected file %q not to exist, but it does", filename)
	}

	filename = ".git/tree/rebasing-source"
	if env.repo.FileExists(filename) {
		t.Errorf("Expected file %q not to exist, but it does", filename)
	}

	filename = ".git/tree/rebasing-dest"
	if env.repo.FileExists(filename) {
		t.Errorf("Expected file %q not to exist, but it does", filename)
	}

	filename = ".git/tree/rebasing-temps"
	if env.repo.FileExists(filename) {
		t.Errorf("Expected file %q not to exist, but it does", filename)
	}
}

// Initial:
//
//	master ─── mew ─┬─ treecko ─── grovyle ─── sceptile
//	                └─ mudkip
func TestRebaseTreeAbort_DeletesTemporaryBranches(t *testing.T) {
	env := setUp(t)
	defer env.tearDown(t)

	// Setup initial - write conflicting contents to the same file.
	env.repo.BranchWithCommit("mew")
	env.repo.BranchWithCommit("treecko")
	env.repo.CreateAndSwitchBranch("grovyle")
	env.repo.WriteAndCommitFile("favorite", "grovyle", "grovyle")
	env.repo.BranchWithCommit("sceptile")
	env.repo.SwitchBranch("mew")
	env.repo.CreateAndSwitchBranch("mudkip")
	env.repo.WriteAndCommitFile("favorite", "mudkip", "mudkip")
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mudkip")
	RebaseTree(env.repo.Repo, source, dest)

	// Abort the rebase
	RebaseTreeAbort(env.repo.Repo)

	tempTreecko := env.repo.LookupBranch("rebase-treecko")
	tempGrovyle := env.repo.LookupBranch("rebase-grovyle")

	if tempTreecko != nil {
		t.Errorf("Expected temporary branch %q to not exist, but it does", "rebase-treecko")
	}
	if tempGrovyle != nil {
		t.Errorf("Expected temporary branch %q to not exist, but it does", "rebase-grovyle")
	}
}

// TESTS TO ADD:
//  - Continuing a MergeConflict

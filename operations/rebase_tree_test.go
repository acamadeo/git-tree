package operations

import (
	"testing"

	dropCmd "github.com/abaresk/git-tree/commands/drop"
	gitutil "github.com/abaresk/git-tree/git"
	"github.com/abaresk/git-tree/testutil"
)

// -------------------------------------------------------------------------- \
// RebaseTree                                                                 |
// -------------------------------------------------------------------------- /

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
	// TODO: Turn init and drop into operations (so we don't need to import commands here)!
	Init(env.repo.Repo)

	// Rebase tree
	source := env.repo.LookupBranch("treecko")
	dest := env.repo.LookupBranch("mudkip")
	RebaseTree(env.repo.Repo, source, dest)
	// Clean up extra branches from `git-tree init`.
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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
	dropCmd.NewDropCommand().Execute()

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

// TESTS TO ADD:
//  - MergeConflict
//  - Aborting a MergeConflict
//  - Continuing a MergeConflict

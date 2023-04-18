#!/bin/bash

append-branch() {
    git checkout -b $1
    if [ $2 == "conflict" ]; then
        echo $1 > shared.txt
    else
        echo $1 > $1.txt
    fi
    git add .
    git commit -m "Add $1"
}

switch-branch() {
    git checkout $1
}

reset() {
    rm *
    rm -rf .git/
}

# ===========================
# Rebase with merge conflicts
# ===========================
# 
# In the examples below, `*` denotes conflicting files.


# `git-tree rebase -s d -d e`
# ===========================
# Tree:
#   a - b - c ┬ d*
#             └ e* - f

if false; then

# Construct the initial tree
git init
append-branch a
append-branch b
append-branch c
append-branch d conflict
switch-branch c
append-branch e conflict
append-branch f

# First rebase branch `d` onto `e`.
# 
# !! MERGE CONFLICT ENCOUNTERED !!
#  ~ Must manually address merge conflict and continue merge. ~
git checkout d
git checkout -b d2
git rebase --onto e c d

# Manually address merge conflict.
echo d > shared.txt
git add .
git rebase --continue

# Branch `d` has no descendants.

# Delete the temporary branches
git branch -D d2

fi


# `git-tree rebase -s e -d d`
# ===========================
# Tree:
#   a - b - c ┬ d*
#             └ e* - f

if false; then

# Construct the initial tree
git init
append-branch a
append-branch b
append-branch c
append-branch d conflict
switch-branch c
append-branch e conflict
append-branch f

# First rebase branch `e` onto `d`.
# 
# !! MERGE CONFLICT ENCOUNTERED !!
#  ~ Must manually address merge conflict and continue merge. ~
git checkout e
git checkout -b e2
git rebase --onto d c e

# Manually address merge conflict.
echo e > shared.txt
git add .
git rebase --continue

# Then do DFS down each of `e`'s descendants. First, rebase `f`.
git checkout f
git checkout -b f2
git rebase --onto e e2 f

# Delete the temporary branches
git branch -D e2 f2

fi


# `git-tree rebase -s e -d d`
# ===========================
# Tree:
#   a - b - c ┬ d*
#             └ e* - f*

if false; then

# Construct the initial tree
git init
append-branch a
append-branch b
append-branch c
append-branch d conflict
switch-branch c
append-branch e conflict
append-branch f conflict

# First rebase branch `e` onto `d`.
# 
# !! MERGE CONFLICT ENCOUNTERED !!
#  ~ Must manually address merge conflict and continue merge. ~
git checkout e
git checkout -b e2
git rebase --onto d c e

# Manually address merge conflict.
echo e > shared.txt
git add .
git rebase --continue

# Then do DFS down each of `e`'s descendants. First, rebase `f`.
git checkout f
git checkout -b f2
git rebase --onto e e2 f

# No merge conflict this time 🎉
#   The diff of branch `f` from `e` is the same as the diff from its current
#   parent `e2`. No conflict occurs.

# Delete the temporary branches
git branch -D e2 f2

fi


# `git-tree rebase -s d -d a`
# ===========================
# Tree:
#   a - b* - c - d*

if false; then

# Construct the initial tree
git init
append-branch a
append-branch b conflict
append-branch c
append-branch d conflict

# First rebase branch `d` onto `a`.
# 
# !! MERGE CONFLICT ENCOUNTERED !!
#  ~ Must manually address merge conflict and continue merge. ~
git checkout d
git checkout -b d2
git rebase --onto a c d

# Manually address merge conflict.
#  - Conflict was: "deleted by us:  shared.txt"
#  - NOTE: Ideally, this shouldn't trigger a merge conflict. Just use the
#          version that exists at `d`. But this is a future optimization.
git add .
git rebase --continue

# Branch `d` has no descendants.

# Delete the temporary branches
git branch -D d2

fi


# `git-tree rebase -s d -d a`
# ===========================
# Tree:
#   a* - b* - c - d*

if false; then

# Construct the initial tree
git init
append-branch a conflict
append-branch b conflict
append-branch c
append-branch d conflict

# First rebase branch `d` onto `a`.
# 
# !! MERGE CONFLICT ENCOUNTERED !!
#  ~ Must manually address merge conflict and continue merge. ~
git checkout d
git checkout -b d2
git rebase --onto a c d

# Manually address merge conflict.
echo d > shared.txt
git add .
git rebase --continue

# Branch `d` has no descendants.

# Delete the temporary branches
git branch -D d2

fi

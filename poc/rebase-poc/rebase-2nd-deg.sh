#!/bin/bash

append-branch() {
    git checkout -b $1
    echo $1 > $1.txt
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

# Rebasing a 2nd-degree tree onto another 2nd-degree tree
# =======================================================
# Tree:
#   a - b ┬ c - d ┬ e
#         │       └ f
#         └ g ┬ h
#             └ i

construct-tree () {
    git init

    append-branch a
    append-branch b
    append-branch c
    append-branch d
    append-branch e
    switch-branch d
    append-branch f
    switch-branch b
    append-branch g
    append-branch h
    switch-branch g
    append-branch i
}

# `git-tree rebase -s d -d h`
# ===========================
if false; then

# Construct the initial tree
construct-tree

# First rebase branch `d` onto `h`.
git checkout d
git checkout -b d2
git rebase --onto h c d

# Then do DFS down each of `d`'s descendants. First, rebase `e`.
git checkout e
git checkout -b e2
git rebase --onto d d2 e

# Then rebase `f`.
git checkout f
git checkout -b f2
git rebase --onto d d2 f

# Delete the temporary branches
git branch -D d2 e2 f2

fi


# `git-tree rebase -s g -d e`
# ===========================
# Tree:
#   a - b ┬ c - d ┬ e
#         │       └ f
#         └ g ┬ h
#             └ i

if false; then

# Construct the initial tree
construct-tree

# First rebase branch `g` onto `e`.
git checkout g
git checkout -b g2
git rebase --onto e b g

# Then do DFS down each of `g`'s descendants. First, rebase `h`.
git checkout h
git checkout -b h2
git rebase --onto g g2 h

# Then rebase `i`.
git checkout i
git checkout -b i2
git rebase --onto g g2 i

# Delete the temporary branches
git branch -D g2 h2 i2

fi


# `git-tree rebase -s d -d g`
# ===========================
# Tree:
#   a - b ┬ c - d ┬ e
#         │       └ f
#         └ g ┬ h
#             └ i

if false; then

# Construct the initial tree
construct-tree

# First rebase branch `d` onto `g`.
git checkout d
git checkout -b d2
git rebase --onto g c d

# Then do DFS down each of `d`'s descendants. First, rebase `e`.
git checkout e
git checkout -b e2
git rebase --onto d d2 e

# Then rebase `f`.
git checkout f
git checkout -b f2
git rebase --onto d d2 f

# Delete the temporary branches
git branch -D d2 e2 f2

fi


# `git-tree rebase -s g -d d`
# ===========================
# Tree:
#   a - b ┬ c - d ┬ e
#         │       └ f
#         └ g ┬ h
#             └ i

if false; then

# Construct the initial tree
construct-tree

# First rebase branch `g` onto `d`.
git checkout g
git checkout -b g2
git rebase --onto d b g

# Then do DFS down each of `g`'s descendants. First, rebase `h`.
git checkout h
git checkout -b h2
git rebase --onto g g2 h

# Then rebase `i`.
git checkout i
git checkout -b i2
git rebase --onto g g2 i

# Delete the temporary branches
git branch -D g2 h2 i2

fi


# `git-tree rebase -s h -d d`
# ===========================
# Tree:
#   a - b ┬ c - d ┬ e
#         │       └ f
#         └ g ┬ h
#             └ i

if false; then

# Construct the initial tree
construct-tree

# First rebase branch `h` onto `d`.
git checkout h
git checkout -b h2
git rebase --onto d g h

# Branch `h` has no descendants.

# Delete the temporary branches
git branch -D h2

fi


# `git-tree rebase -s e -d g`
# ===========================
# Tree:
#   a - b ┬ c - d ┬ e
#         │       └ f
#         └ g ┬ h
#             └ i

if false; then

# Construct the initial tree
construct-tree

# First rebase branch `e` onto `g`.
git checkout e
git checkout -b e2
git rebase --onto g d e

# Branch `e` has no descendants.

# Delete the temporary branches
git branch -D e2

fi


# `git-tree rebase -s h -d e`
# ===========================
# Tree:
#   a - b ┬ c - d ┬ e
#         │       └ f
#         └ g ┬ h
#             └ i

if false; then

# Construct the initial tree
construct-tree

# First rebase branch `h` onto `e`.
git checkout h
git checkout -b h2
git rebase --onto e g h

# Branch `h` has no descendants.

# Delete the temporary branches
git branch -D h2

fi


# `git-tree rebase -s e -d h`
# ===========================
# Tree:
#   a - b ┬ c - d ┬ e
#         │       └ f
#         └ g ┬ h
#             └ i

if true; then

# Construct the initial tree
construct-tree

# First rebase branch `e` onto `h`.
git checkout e
git checkout -b e2
git rebase --onto h d e

# Branch `e` has no descendants.

# Delete the temporary branches
git branch -D e2

fi

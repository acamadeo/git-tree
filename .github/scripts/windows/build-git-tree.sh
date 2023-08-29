#!/bin/bash

set -ex

export PATH=/d/a/_temp/msys64/mingw64/bin:/d/a/_temp/msys64/usr/bin:$PATH

LIBGIT2_PATH="/d/a/git-tree/git-tree/git2go/vendor/libgit2"
export CGO_CFLAGS="-I${LIBGIT2_PATH}/include"
export CGO_LDFLAGS="-LD:\a\git-tree\git-tree\git2go\vendor\libgit2\build -lws2_32"

go get -d ./...
go build -v --tags static

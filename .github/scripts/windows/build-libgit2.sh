#!/bin/bash

set -ex

export PATH=/d/a/_temp/msys64/mingw64/bin:/d/a/_temp/msys64/usr/bin:$PATH

LIBGIT2_BUILD="/d/a/git-tree/git-tree/git2go/vendor/libgit2/build"
cd "${LIBGIT2_BUILD}"

cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_BUILD_TYPE="RelWithDebInfo" \
      -DCMAKE_INSTALL_PREFIX=../install \
      -DWINHTTP=OFF \
      -DUSE_BUNDLED_ZLIB=ON \
      -DUSE_HTTPS=OFF \
      -DUSE_SSH=OFF \
      -DCURL=OFF \
      -DBUILD_TESTS=OFF \
      -G "MSYS Makefiles" \
      .. &&
cmake --build .

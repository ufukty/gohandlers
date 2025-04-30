#!/bin/bash

set -ve

(
  cd "$(mktemp -d)"
  git clone https://github.com/ufukty/kask
  cd kask
  git fetch --tags --quiet
  git checkout "$(git tag --list 'v*' | sort -Vr | head -n 1)"
  make install
)

~/bin/kask build -in docs -out docs-build -domain / -v

#!/usr/bin/env bash

if [ $# -ne 1 ]; then
  echo 'please use from Makefile'
  exit 1
fi

VERSION=$1

git tag "$VERSION"
git tag "v$VERSION" "$VERSION" # tagging another one which has "v" prefix for Go modules.
git push origin "$VERSION"
git push origin "v$VERSION"

ghr -parallel=1 "$VERSION" pkg

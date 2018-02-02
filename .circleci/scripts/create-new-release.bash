#!/usr/bin/env bash

if [ $# -ne 1 ]; then
  echo 'please use from Makefile'
  exit 1
fi

VERSION=$1

git tag "$VERSION"
git push origin "$VERSION"

ghr -parallel=1 "$VERSION" pkg

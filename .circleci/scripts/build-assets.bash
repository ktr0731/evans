#!/usr/bin/env bash

OSARCH='darwin/386 darwin/amd64 linux/386 linux/amd64 linux/arm linux/arm64'

mkdir pkg
cd pkg

make deps

gox -osarch "$OSARCH" \
  -ldflags="-X github.com/ktr0731/evans/vendor/github.com/ktr0731/go-updater/github.isGitHubReleasedBinary=true" \
  ..


for f in *; do
  mv "$f" evans
  tar cvf "$f.tar.gz" evans
  rm -f evans
done

#!/usr/bin/env bash

OSARCH='darwin/386 darwin/amd64 linux/386 linux/amd64 linux/arm linux/arm64'

mkdir pkg
cd pkg

gox -osarch "$OSARCH" ..

for f in *; do
  mv "$f" evans
  tar cvf "$f.tar.gz" evans
  rm -f evans
done

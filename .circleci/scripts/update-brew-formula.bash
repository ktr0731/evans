#!/bin/bash

set -e -o pipefail

if [ $# -ne 1 ]; then
  echo 'please use from Makefile'
  exit 1
fi

if [ "$GIT_EMAIL" = "" ] || [ "$GIT_NAME" = "" ]; then
  echo 'please set $GIT_EMAIL and $GIT_NAME'
  exit 1
fi

VERSION=$1

git clone https://github.com/ktr0731/homebrew-evans brew

SHA256_AMD64=$(shasum -a 256 pkg/evans_darwin_amd64.tar.gz | awk '{ print $1 }')
SHA256_386=$(shasum -a 256 pkg/evans_darwin_386.tar.gz | awk '{ print $1 }')

cd brew

cp evans.rb.backup evans.rb
sed -i -r "s/VERSION/${VERSION}/" evans.rb
sed -i -r "s/SHA256_AMD64/${SHA256_AMD64}/" evans.rb
sed -i -r "s/SHA256_386/${SHA256_386}/" evans.rb
git add --all
git commit -m "bump ${VERSION}"
git push origin master

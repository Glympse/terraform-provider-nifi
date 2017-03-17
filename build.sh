#!/bin/sh
set -e

# Install dependencies
glide update

mkdir -p ./bin/macosx/
go build -o ./bin/macosx/terraforn-provider-nifi || exit 1

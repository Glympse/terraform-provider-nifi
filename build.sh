#!/bin/sh
set -e

# Install dependencies
glide update

# Build binary for local platform
mkdir -p ./bin/local/
go build -o ./bin/local/terraforn-provider-nifi || exit 1

#!/bin/sh
set -e

# Build plugin locally
./build.sh

# Register the plugin
cat > ~/.terraformrc <<EOL
providers {
   nifi = "$GOPATH/src/github.com/glympse/terraform-provider-nifi/bin/macosx/terraforn-provider-nifi"
}
EOL
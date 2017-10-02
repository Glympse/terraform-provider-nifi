#!/bin/sh
set -e

# Build plugin locally
./build.sh

# Register the plugin
cat > ~/.terraformrc <<EOL
providers {
   nifi = "$GOPATH/src/github.com/af6140/terraform-provider-nifi/bin/local/terraforn-provider-nifi"
}
EOL

#!/bin/sh
set -e

# Usage:
# $ ./run.sh <example_folder>

# Enable the most detailed logging (useful for debugging)
# export TF_LOG=TRACE

pushd $1 > /dev/null
    # Wipe previous state (useful for debugging)
    # rm -f terraform.tfstate

    # Plan changes
    terraform plan -var-file=example.tfvars

    # Apply changes
    terraform apply -var-file=example.tfvars
popd > /dev/null

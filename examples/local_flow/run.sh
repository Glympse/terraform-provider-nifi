#!/bin/sh
set -e

# Wipe previous state (useful for debugging)
# rm -f terraform.tfstate

# Enable the most detailed logging (useful for debugging)
# export TF_LOG=TRACE

# Plan changes
terraform plan -var-file=example.tfvars

# Apply changes
terraform apply -var-file=example.tfvars

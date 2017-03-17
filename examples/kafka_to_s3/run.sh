#!/bin/sh
set -e

rm -f terraform.tfstate

export TF_LOG=TRACE
terraform apply -var-file=example.tfvars

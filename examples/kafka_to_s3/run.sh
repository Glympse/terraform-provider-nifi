#!/bin/sh
set -e

rm terraform.tfstate

export TF_LOG=TRACE
terraform apply -var-file=example.tfvars

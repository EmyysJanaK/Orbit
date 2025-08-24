#!/bin/bash

# Terraform destroy script
set -e

ENVIRONMENT=${1:-dev}

if [[ "$ENVIRONMENT" != "dev" && "$ENVIRONMENT" != "prod" ]]; then
    echo "Error: Environment must be 'dev' or 'prod'"
    exit 1
fi

echo "WARNING: This will destroy all resources in the $ENVIRONMENT environment!"
read -p "Are you sure you want to continue? (yes/no): " confirmation

if [[ "$confirmation" != "yes" ]]; then
    echo "Operation cancelled."
    exit 0
fi

echo "Destroying $ENVIRONMENT environment..."

cd "$(dirname "$0")/../environments/$ENVIRONMENT"

terraform init
terraform destroy -var-file=terraform.tfvars -auto-approve

echo "Destroy completed successfully!"

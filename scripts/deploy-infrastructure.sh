#!/bin/bash

# Infrastructure Deployment Script
set -e

ENVIRONMENT=${1:-dev}
ACTION=${2:-apply}

if [[ "$ENVIRONMENT" != "dev" && "$ENVIRONMENT" != "prod" ]]; then
    echo "❌ Error: Environment must be 'dev' or 'prod'"
    echo "Usage: $0 <environment> [action]"
    echo "Example: $0 dev apply"
    exit 1
fi

if [[ "$ACTION" != "plan" && "$ACTION" != "apply" && "$ACTION" != "destroy" ]]; then
    echo "❌ Error: Action must be 'plan', 'apply', or 'destroy'"
    exit 1
fi

echo "🏗️ Deploying infrastructure for $ENVIRONMENT environment..."

# Check AWS credentials
if ! aws sts get-caller-identity >/dev/null 2>&1; then
    echo "❌ AWS credentials not configured. Please run 'aws configure' first."
    exit 1
fi

# Navigate to terraform directory
cd terraform/scripts

# Run terraform deployment
chmod +x deploy.sh
./deploy.sh $ENVIRONMENT $ACTION

if [[ "$ACTION" == "apply" ]]; then
    echo "✅ Infrastructure deployment completed!"
    echo "📝 Save the ALB DNS name for application deployment."
fi

cd ../../

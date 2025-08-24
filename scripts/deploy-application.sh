#!/bin/bash

# Application Deployment Script
set -e

ENVIRONMENT=${1:-dev}
VERSION=${2:-latest}

if [[ "$ENVIRONMENT" != "dev" && "$ENVIRONMENT" != "prod" ]]; then
    echo "❌ Error: Environment must be 'dev' or 'prod'"
    echo "Usage: $0 <environment> [version]"
    echo "Example: $0 dev latest"
    exit 1
fi

echo "🚀 Deploying application to $ENVIRONMENT environment..."

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Build the application
echo "🏗️ Building application..."
cd java-app

# Clean and build
mvn clean package -DskipTests

# Build Docker image
echo "🐳 Building Docker image..."
IMAGE_NAME="employee-management:$VERSION"
docker build -t $IMAGE_NAME .

# Tag for ECR (if deploying to AWS)
if [[ "$ENVIRONMENT" == "prod" ]]; then
    AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
    AWS_REGION=$(aws configure get region)
    ECR_REPOSITORY="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/employee-management"

    # Login to ECR
    aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $ECR_REPOSITORY

    # Tag and push to ECR
    docker tag $IMAGE_NAME $ECR_REPOSITORY:$VERSION
    docker push $ECR_REPOSITORY:$VERSION

    echo "✅ Application deployed to ECR: $ECR_REPOSITORY:$VERSION"
else
    echo "✅ Application built locally: $IMAGE_NAME"
    echo "💡 For local testing, run: docker run -p 8080:8080 $IMAGE_NAME"
fi

cd ..

echo "🎉 Application deployment completed!"

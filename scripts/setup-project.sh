#!/bin/bash

# Employee Management System Setup Script
set -e

echo "🚀 Setting up Employee Management System..."

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo "📋 Checking prerequisites..."

if ! command_exists java; then
    echo "❌ Java is not installed. Please install Java 11 or higher."
    exit 1
fi

if ! command_exists mvn; then
    echo "❌ Maven is not installed. Please install Maven."
    exit 1
fi

if ! command_exists docker; then
    echo "❌ Docker is not installed. Please install Docker."
    exit 1
fi

if ! command_exists terraform; then
    echo "❌ Terraform is not installed. Please install Terraform."
    exit 1
fi

echo "✅ All prerequisites are installed!"

# Setup local development environment
echo "🔧 Setting up local development environment..."

# Create local directories
mkdir -p logs
mkdir -p data/mysql

# Copy environment specific configs
if [ ! -f java-app/src/main/resources/application-local.properties ]; then
    cat > java-app/src/main/resources/application-local.properties << EOF
# Local Development Configuration
spring.profiles.active=local

# Local MySQL Database
spring.datasource.url=jdbc:mysql://localhost:3306/employeedb_local?useSSL=false&allowPublicKeyRetrieval=true&serverTimezone=UTC
spring.datasource.username=root
spring.datasource.password=password

# JPA Configuration
spring.jpa.hibernate.ddl-auto=update
spring.jpa.show-sql=true

# Logging
logging.level.com.employeemgmt=DEBUG
EOF
fi

# Build the application
echo "🏗️ Building the application..."
cd java-app
mvn clean compile
cd ..

echo "✅ Setup completed successfully!"
echo ""
echo "📚 Next steps:"
echo "1. Start local services: docker-compose up -d"
echo "2. Run the application: cd java-app && mvn spring-boot:run"
echo "3. Access the API at: http://localhost:8080/api/employees"
echo "4. Check health: http://localhost:8080/actuator/health"
echo ""
echo "🔧 For infrastructure deployment:"
echo "1. Configure AWS credentials"
echo "2. Run: ./scripts/deploy-infrastructure.sh dev"
echo "3. Run: ./scripts/deploy-application.sh dev"

#!/bin/bash

# User data script for EC2 instances
yum update -y
yum install -y java-11-openjdk-devel docker

# Start and enable Docker
systemctl start docker
systemctl enable docker
usermod -a -G docker ec2-user

# Install AWS CLI v2
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Create application directory
mkdir -p /opt/employee-app
chown ec2-user:ec2-user /opt/employee-app

# Download and run the application (this would be replaced with actual deployment logic)
# For now, we'll create a placeholder service
cat > /etc/systemd/system/employee-app.service << EOF
[Unit]
Description=Employee Management Application
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/opt/employee-app
ExecStart=/usr/bin/java -jar /opt/employee-app/employee-management.jar
Restart=always
Environment=SPRING_PROFILES_ACTIVE=prod
Environment=DB_ENDPOINT=${db_endpoint}

[Install]
WantedBy=multi-user.target
EOF

systemctl enable employee-app

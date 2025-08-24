aws_region        = "us-west-2"
environment       = "dev"
project_name      = "employee-mgmt"
vpc_cidr          = "10.0.0.0/16"
az_count          = 2
instance_type     = "t3.micro"
min_size          = 1
max_size          = 3
desired_capacity  = 1
db_instance_class = "db.t3.micro"
db_name           = "employeedb"
db_username       = "admin"


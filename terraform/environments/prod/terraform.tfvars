aws_region        = "us-west-2"
environment       = "prod"
project_name      = "employee-mgmt"
vpc_cidr          = "10.1.0.0/16"
az_count          = 3
instance_type     = "t3.small"
min_size          = 2
max_size          = 6
desired_capacity  = 3
db_instance_class = "db.t3.small"
db_name           = "employeedb"
db_username       = "admin"


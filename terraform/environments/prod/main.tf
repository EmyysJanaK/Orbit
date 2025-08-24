terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

module "networking" {
  source = "../../modules/networking"

  environment    = var.environment
  vpc_cidr       = var.vpc_cidr
  az_count       = var.az_count
  project_name   = var.project_name
}

module "security" {
  source = "../../modules/security"

  environment  = var.environment
  vpc_id       = module.networking.vpc_id
  project_name = var.project_name
}

module "database" {
  source = "../../modules/database"

  environment         = var.environment
  vpc_id              = module.networking.vpc_id
  private_subnet_ids  = module.networking.private_subnet_ids
  db_security_group_id = module.security.db_security_group_id
  db_instance_class   = var.db_instance_class
  db_name             = var.db_name
  db_username         = var.db_username
  project_name        = var.project_name
}

module "compute" {
  source = "../../modules/compute"

  environment           = var.environment
  vpc_id                = module.networking.vpc_id
  public_subnet_ids     = module.networking.public_subnet_ids
  private_subnet_ids    = module.networking.private_subnet_ids
  app_security_group_id = module.security.app_security_group_id
  alb_security_group_id = module.security.alb_security_group_id
  instance_type         = var.instance_type
  min_size              = var.min_size
  max_size              = var.max_size
  desired_capacity      = var.desired_capacity
  project_name          = var.project_name
  db_endpoint           = module.database.db_endpoint
}

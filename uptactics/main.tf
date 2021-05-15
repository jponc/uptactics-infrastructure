terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.27"
    }
  }

  backend "s3" {
    bucket         = "uptactics-terraform-up-and-running-state"
    key            = "global/s3/terraform.tfstate"
    region         = "us-east-1"
    profile        = "uptactics"
    dynamodb_table = "uptactics-terraform-up-and-running-locks"
    encrypt        = true
  }
}

provider "aws" {
  profile = var.aws_profile
  region  = var.aws_region
}

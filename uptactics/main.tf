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

# VPC & Subnet configuration

resource "aws_vpc" "vpc" {
  cidr_block = "172.30.0.0/16"

  tags = {
    Name    = "${local.service_name_env}-vpc"
  }
}

resource "aws_subnet" "private-subnet-1" {
  vpc_id     = aws_vpc.vpc.id
  cidr_block = "172.30.1.0/24"

  tags = {
    Name    = "${local.service_name_env}-private-subnet-1"
  }
}

resource "aws_subnet" "private-subnet-2" {
  vpc_id     = aws_vpc.vpc.id
  cidr_block = "172.30.3.0/24"

  tags = {
    Name    = "${local.service_name_env}-private-subnet-2"
  }
}

resource "aws_subnet" "public-subnet" {
  vpc_id     = aws_vpc.vpc.id
  cidr_block = "172.30.2.0/24"

  tags = {
    Name    = "${local.service_name_env}-public-subnet"
  }
}

# Internet Gateway and NAT Gateway Configuration

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.vpc.id

  tags = {
    Name    = "${local.service_name_env}-internet-gateway"
  }
}

resource "aws_eip" "eip" {
  vpc        = true
  depends_on = [aws_internet_gateway.igw]

  tags = {
    Name    = "${local.service_name_env}-eip"
  }
}

resource "aws_nat_gateway" "ngw" {
  allocation_id = aws_eip.eip.id
  subnet_id     = aws_subnet.public-subnet.id

  tags = {
    Name    = "${local.service_name_env}-nat-gateway"
  }
}

# Public route table configuration

resource "aws_route_table" "route-table-public" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Name = "${local.service_name_env}-route-table-public"
  }
}

resource "aws_route_table_association" "route-table-public-subnet-association" {
  subnet_id      = aws_subnet.public-subnet.id
  route_table_id = aws_route_table.route-table-public.id
}

# Private route table configuration

resource "aws_route_table" "route-table-private" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.ngw.id
  }

  tags = {
    Name = "${local.service_name_env}-route-table-private"
  }
}

resource "aws_route_table_association" "route-table-private-subnet-1-association" {
  subnet_id      = aws_subnet.private-subnet-1.id
  route_table_id = aws_route_table.route-table-private.id
}

resource "aws_route_table_association" "route-table-private-subnet-2-association" {
  subnet_id      = aws_subnet.private-subnet-2.id
  route_table_id = aws_route_table.route-table-private.id
}

# Add Private subnet's and default security group to SSM
resource "aws_ssm_parameter" "private_subnets_ssm" {
  name  = "/${var.service_name}/${var.environment}/PRIVATE_SUBNET_IDS"
  type  = "StringList"
  value = join(",", [aws_subnet.private-subnet-1.id, aws_subnet.private-subnet-2.id])
}

resource "aws_ssm_parameter" "default_security_group_ssm" {
  name  = "/${var.service_name}/${var.environment}/DEFAULT_SECURITY_GROUP"
  type  = "String"
  value = aws_vpc.vpc.default_security_group_id
}

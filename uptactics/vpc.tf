resource "aws_vpc" "vpc" {
  cidr_block = "172.30.0.0/16"

  tags = {
    Name = "${local.service_name_env}-vpc"
  }
}

resource "aws_subnet" "private-subnet-1" {
  vpc_id     = aws_vpc.vpc.id
  cidr_block = "172.30.1.0/24"
  map_public_ip_on_launch = "false"

  tags = {
    Name = "${local.service_name_env}-private-subnet-1"
  }
}

resource "aws_subnet" "private-subnet-2" {
  vpc_id     = aws_vpc.vpc.id
  cidr_block = "172.30.3.0/24"
  map_public_ip_on_launch = "false"

  tags = {
    Name = "${local.service_name_env}-private-subnet-2"
  }
}

resource "aws_subnet" "public-subnet" {
  vpc_id     = aws_vpc.vpc.id
  cidr_block = "172.30.2.0/24"
  map_public_ip_on_launch = "true"

  tags = {
    Name = "${local.service_name_env}-public-subnet"
  }
}

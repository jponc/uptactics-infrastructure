resource "aws_security_group" "allow-ssh" {
  vpc_id      = aws_vpc.vpc.id
  name        = "${local.service_name_env}-allow-ssh"
  description = "security group that allows ssh and all egress traffic"
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = {
    Name = "${local.service_name_env}-allow-ssh"
  }
}

resource "aws_instance" "bastion-instance" {
  ami           = var.instance_ami
  instance_type = var.instance_type

  subnet_id              = aws_subnet.public-subnet.id
  vpc_security_group_ids = [aws_security_group.allow-ssh.id]
  key_name               = aws_key_pair.mykeypair.key_name

  tags = {
    Name = "${local.service_name_env}-bastion-instance"
  }
}

resource "aws_instance" "private-instance" {
  ami           = var.instance_ami
  instance_type = var.instance_type

  subnet_id              = aws_subnet.private-subnet-1.id
  vpc_security_group_ids = [
    aws_security_group.allow-ssh.id,
    aws_vpc.vpc.default_security_group_id,
  ]

  key_name = aws_key_pair.mykeypair.key_name

  tags = {
    Name = "${local.service_name_env}-private-instance"
  }
}

resource "aws_key_pair" "mykeypair" {
  key_name   = "${local.service_name_env}-keypair"
  public_key = file(var.key_path)
}

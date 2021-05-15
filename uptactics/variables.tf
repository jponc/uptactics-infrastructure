variable "environment" {
  default     = "staging"
  description = "Environment"
}

variable "service_name" {
  description = "This is the name of the service"
}

variable "aws_profile" {
  description = "AWS Profile"
}

variable "aws_region" {
  description = "AWS Region"
}

variable "instance_ami" {
  default = "ami-09e67e426f25ce0d7"
  description = "Ubuntu Server 20.04 LTS"
}

variable "instance_type" {
  default = "t2.micro"
}

variable "key_path" {
  default = "keys/mykeypair.pub"
}

locals {
  service_name_env = "${var.service_name}-${var.environment}"
}

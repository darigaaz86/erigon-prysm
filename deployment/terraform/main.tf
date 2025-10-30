terraform {
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

# Get default VPC
data "aws_vpc" "default" {
  default = true
}

# Get default security group
data "aws_security_group" "default" {
  vpc_id = data.aws_vpc.default.id
  name   = "default"
}

# Blockchain Node EC2
resource "aws_instance" "blockchain_node" {
  ami           = var.ami_id
  instance_type = var.instance_type
  key_name      = var.key_name

  vpc_security_group_ids = [data.aws_security_group.default.id]

  root_block_device {
    volume_size = 500
    volume_type = "gp3"
    iops        = 10000
    throughput  = 500
  }

  tags = {
    Name = "blockchain-node"
    Type = "node"
  }
}

# TPS Test EC2
resource "aws_instance" "tps_tester" {
  ami           = var.ami_id
  instance_type = "t3.medium"
  key_name      = var.key_name

  vpc_security_group_ids = [data.aws_security_group.default.id]

  tags = {
    Name = "tps-tester"
    Type = "tester"
  }
}

output "blockchain_node_ip" {
  value = aws_instance.blockchain_node.public_ip
}

output "tps_tester_ip" {
  value = aws_instance.tps_tester.public_ip
}

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

# Security Group
resource "aws_security_group" "blockchain_sg" {
  name        = "blockchain-node-sg"
  description = "Security group for blockchain nodes"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8545
    to_port     = 8545
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8546
    to_port     = 8546
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 30303
    to_port     = 30303
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 30303
    to_port     = 30303
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "blockchain-node-sg"
  }
}

# Key Pair
resource "aws_key_pair" "blockchain_key" {
  key_name   = "blockchain-key"
  public_key = file(var.public_key_path)
}

# Blockchain Node EC2
resource "aws_instance" "blockchain_node" {
  ami           = var.ami_id
  instance_type = var.instance_type
  key_name      = aws_key_pair.blockchain_key.key_name

  vpc_security_group_ids = [aws_security_group.blockchain_sg.id]

  root_block_device {
    volume_size = 100
    volume_type = "gp3"
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
  key_name      = aws_key_pair.blockchain_key.key_name

  vpc_security_group_ids = [aws_security_group.blockchain_sg.id]

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

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-southeast-1"
}

variable "ami_id" {
  description = "AMI ID for Ubuntu 22.04"
  type        = string
  default     = "ami-0497a974f8d5dcef8" # Ubuntu 22.04 LTS in ap-southeast-1
}

variable "instance_type" {
  description = "EC2 instance type for blockchain node"
  type        = string
  default     = "t3.2xlarge"
}

variable "public_key_path" {
  description = "Path to SSH public key"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

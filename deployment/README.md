# Blockchain Deployment with Terraform and Ansible

Automated deployment of blockchain nodes and TPS testing infrastructure on AWS.

## Prerequisites

1. AWS CLI configured with credentials
2. Terraform installed
3. Ansible installed
4. SSH key pair generated (`~/.ssh/id_rsa` and `~/.ssh/id_rsa.pub`)

## Quick Start

```bash
# Repository URLs are pre-configured:
# - Main: git@github.com:darigaaz86/erigon-prysm.git
# - Prysm: git@github.com:darigaaz86/prysm.git (develop branch)

# Run deployment
cd deployment
chmod +x deploy.sh
./deploy.sh
```

## Manual Steps

### 1. Create Infrastructure

```bash
cd deployment/terraform
terraform init
terraform apply
```

### 2. Configure and Deploy

```bash
cd ../ansible

# Update inventory with IPs from Terraform output
# Then run playbook
ansible-playbook playbook.yml -e "repo_url=YOUR_REPO_URL"
```

### 3. Run TPS Test

```bash
# SSH to tester instance
ssh ubuntu@<TPS_TESTER_IP>

# Run test
./run-tps-test.sh
```

## Configuration

### Terraform Variables

Edit `terraform/variables.tf`:
- `aws_region`: AWS region (default: ap-southeast-1)
- `instance_type`: EC2 instance type (default: t3.xlarge)
- `ami_id`: Ubuntu AMI ID

### Ansible Variables

Edit `ansible/playbook.yml`:
- `repo_url`: Main repository URL (default: git@github.com:darigaaz86/erigon-prysm.git)
- `prysm_repo_url`: Prysm repository URL (default: git@github.com:darigaaz86/prysm.git, develop branch)

## Cleanup

```bash
cd deployment/terraform
terraform destroy
```

## Architecture

- **Blockchain Node**: Runs Erigon + Prysm in Docker Compose
- **TPS Tester**: Runs Node.js scripts to generate transactions

## Ports

- 8545: HTTP RPC
- 8546: WebSocket RPC
- 30303: P2P
- 22: SSH

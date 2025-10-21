#!/bin/bash
set -e

echo "=== Blockchain Deployment Script ==="

# Configuration
REPO_URL="https://github.com/darigaaz86/erigon-prysm.git"
PRYSM_REPO_URL="https://github.com/darigaaz86/prysm.git"

# Step 1: Terraform
echo "Step 1: Creating EC2 instances with Terraform..."
cd terraform

if [ ! -f "terraform.tfstate" ]; then
    terraform init
fi

terraform apply -auto-approve

# Get outputs
BLOCKCHAIN_NODE_IP=$(terraform output -raw blockchain_node_ip)
TPS_TESTER_IP=$(terraform output -raw tps_tester_ip)

echo "Blockchain Node IP: $BLOCKCHAIN_NODE_IP"
echo "TPS Tester IP: $TPS_TESTER_IP"

cd ..

# Step 2: Wait for instances to be ready
echo "Step 2: Waiting for instances to be ready..."
sleep 30

# Step 3: Generate Ansible inventory
echo "Step 3: Generating Ansible inventory..."
cd ansible

cat > inventory.yml <<EOF
all:
  children:
    blockchain_nodes:
      hosts:
        node:
          ansible_host: $BLOCKCHAIN_NODE_IP
          ansible_user: ubuntu
          ansible_ssh_private_key_file: ~/.ssh/id_rsa
    
    tps_testers:
      hosts:
        tester:
          ansible_host: $TPS_TESTER_IP
          ansible_user: ubuntu
          ansible_ssh_private_key_file: ~/.ssh/id_rsa
          blockchain_rpc_url: http://$BLOCKCHAIN_NODE_IP:8545
EOF

# Step 4: Run Ansible playbook
echo "Step 4: Running Ansible playbook..."
ansible-playbook playbook.yml -e "repo_url=$REPO_URL" -e "prysm_repo_url=$PRYSM_REPO_URL"

cd ..

echo ""
echo "=== Deployment Complete ==="
echo "Blockchain Node: http://$BLOCKCHAIN_NODE_IP:8545"
echo "TPS Tester: $TPS_TESTER_IP"
echo ""
echo "To run TPS test:"
echo "  ssh ubuntu@$TPS_TESTER_IP"
echo "  ./run-tps-test.sh"
echo ""
echo "To destroy infrastructure:"
echo "  cd deployment/terraform && terraform destroy"

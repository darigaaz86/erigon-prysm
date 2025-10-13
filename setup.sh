#!/bin/bash

echo "🚀 Setting up Erigon + Prysm Private Chain..."
echo ""

# Step 1: Create directories
echo "Step 1: Creating directories..."
mkdir -p config genesis
echo "✅ Directories created"
echo ""

# Step 2: Create Prysm config
echo "Step 2: Creating Prysm configuration..."
cat > config/config.yml << 'EOF'
CONFIG_NAME: private-chain
PRESET_BASE: mainnet
GENESIS_FORK_VERSION: 0x20000089
GENESIS_DELAY: 0
ALTAIR_FORK_EPOCH: 0
ALTAIR_FORK_VERSION: 0x20000090
BELLATRIX_FORK_EPOCH: 0
BELLATRIX_FORK_VERSION: 0x20000091
TERMINAL_TOTAL_DIFFICULTY: 0
CAPELLA_FORK_EPOCH: 0
CAPELLA_FORK_VERSION: 0x20000092
DENEB_FORK_EPOCH: 0
DENEB_FORK_VERSION: 0x20000093
ELECTRA_FORK_EPOCH: 18446744073709551615
ELECTRA_FORK_VERSION: 0x20000094
FULU_FORK_EPOCH: 18446744073709551615
FULU_FORK_VERSION: 0x20000095
SECONDS_PER_SLOT: 12
SLOTS_PER_EPOCH: 32
MIN_GENESIS_TIME: 0
MIN_GENESIS_ACTIVE_VALIDATOR_COUNT: 1
DEPOSIT_CHAIN_ID: 32382
DEPOSIT_NETWORK_ID: 32382
DEPOSIT_CONTRACT_ADDRESS: 0x4242424242424242424242424242424242424242
EOF
echo "✅ Prysm config created"
echo ""

# Step 3: Generate genesis state
echo "Step 3: Generating genesis state with prysmctl..."
docker run --rm \
  -v $(pwd)/genesis:/genesis \
  -v $(pwd)/config:/config \
  -v $(pwd)/genesis.json:/genesis.json \
  gcr.io/prysmaticlabs/prysm/cmd/prysmctl:stable \
  testnet generate-genesis \
  --fork=deneb \
  --num-validators=64 \
  --genesis-time-delay=0 \
  --chain-config-file=/config/config.yml \
  --geth-genesis-json-in=/genesis.json \
  --geth-genesis-json-out=/genesis/genesis.json \
  --output-ssz=/genesis/genesis.ssz

if [ $? -eq 0 ]; then
  echo "✅ Genesis state generated"
else
  echo "❌ Genesis generation failed"
  exit 1
fi

echo ""
echo "========================================="
echo "✅ Setup Complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Start services: docker compose -f docker-compose-private.yml up -d"
echo "2. View logs: docker compose -f docker-compose-private.yml logs -f"
echo "3. Test RPC: curl -X POST http://localhost:8545 -H 'Content-Type: application/json' --data '{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[],\"id\":1}'"
echo ""
echo "Pre-funded accounts:"
echo "  0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
echo "  0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
echo "  0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"
echo ""

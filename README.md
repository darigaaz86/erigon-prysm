# Erigon + Prysm Private Chain

Private Ethereum network with **Prysm built from source** and Erigon execution layer.

## Quick Start

### Using Custom-Built Prysm (Recommended)
```bash
# 1. Build Prysm from source
./build-prysm-docker.sh

# 2. Start the network
docker compose -f docker-compose-prysm-local.yml up -d

# 3. Check status
docker ps
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

### Using Official Prysm Images
```bash
docker compose -f docker-compose-private.yml up -d
```

See **[QUICK-START.md](QUICK-START.md)** for detailed instructions.

## Architecture

```
┌─────────────────────────────────────┐
│  Prysm Validator                    │
│  - 64 validators                    │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Prysm Beacon Chain                 │
│  - Built from ./prysm source        │
│  - Ports: 4000, 3500, 8080          │
└──────────────┬──────────────────────┘
               │ Engine API (JWT)
               ▼
┌─────────────────────────────────────┐
│  Erigon Execution Layer             │
│  - Port 8545 (JSON-RPC)             │
│  - Blocks every 12 seconds          │
└─────────────────────────────────────┘
```

## Modify Prysm Consensus

```bash
# 1. Edit Prysm source code
cd prysm
vim beacon-chain/...

# 2. Rebuild
cd ..
./build-prysm-docker.sh

# 3. Restart with clean state
docker compose -f docker-compose-prysm-local.yml down -v
docker compose -f docker-compose-prysm-local.yml up -d
```

## Configuration

- **Network ID**: 32382
- **Chain ID**: 32382
- **Block Time**: 12 seconds
- **Validators**: 64 (interop mode)
- **Pre-funded Accounts**: 10 accounts with 10,000 ETH each

## Pre-funded Accounts

```
Account 1: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
Private Key: 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80

Account 2: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
Private Key: 0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d

Account 3: 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
Private Key: 0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a
```

## Files

### Essential
- **`prysm/`** - Prysm source code
- **`build-prysm-docker.sh`** - Build Prysm from source
- **`docker-compose-prysm-local.yml`** - Custom Prysm setup
- **`docker-compose-private.yml`** - Official Prysm images
- **`QUICK-START.md`** - Detailed guide

### Configuration
- **`genesis/genesis.json`** - Genesis block
- **`config/config.yml`** - Prysm chain config
- **`jwt.hex`** - Engine API authentication

## Ports

- **8545** - Erigon JSON-RPC (HTTP)
- **8546** - Erigon WebSocket
- **8551** - Erigon Engine API
- **4000** - Prysm gRPC
- **3500** - Prysm REST API
- **8080** - Prysm metrics

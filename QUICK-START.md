# Prysm from Source - Quick Start Guide

## ✅ Setup Complete

Your private Ethereum network is running with **Prysm built from source**!

### Architecture

```
┌─────────────────────────────────────┐
│  Prysm Validator (Your Build)      │
│  - 64 validators                    │
│  - Submitting attestations          │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│  Prysm Beacon Chain (Your Build)   │
│  - Ports: 4000, 3500, 8080          │
│  - Built from ./prysm source        │
└──────────────┬──────────────────────┘
               │ Engine API (JWT)
               ▼
┌─────────────────────────────────────┐
│  Erigon Execution Layer             │
│  - Port 8545 (JSON-RPC)             │
│  - Producing blocks every 12s       │
└─────────────────────────────────────┘
```

### Files

- **`./prysm/`** - Prysm source code (modify this!)
- **`build-prysm-docker.sh`** - Build script
- **`docker-compose-prysm-local.yml`** - Docker compose config

## 🚀 Quick Commands

### Check Status
```bash
# View containers
docker ps

# Check block number
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Watch logs
docker logs -f prysm-beacon-local
docker logs -f prysm-validator-local
docker logs -f erigon-private
```

### Modify Prysm

```bash
# 1. Edit source code
cd prysm
vim beacon-chain/...  # or any file

# 2. Rebuild
cd ..
./build-prysm-docker.sh

# 3. Restart
docker compose -f docker-compose-prysm-local.yml restart beacon-chain validator

# 4. Watch your changes
docker logs -f prysm-beacon-local
```

### Control Network

```bash
# Stop
docker compose -f docker-compose-prysm-local.yml down

# Start
docker compose -f docker-compose-prysm-local.yml up -d

# Restart
docker compose -f docker-compose-prysm-local.yml restart

# Clean restart (removes data)
docker compose -f docker-compose-prysm-local.yml down -v
docker compose -f docker-compose-prysm-local.yml up -d
```



## 🎯 Common Tasks

### Add Custom Logging
```bash
# Edit any Prysm file
vim prysm/beacon-chain/blockchain/service.go

# Add logging
import "github.com/sirupsen/logrus"
log.Info("My custom message")

# Rebuild and restart
./build-prysm-docker.sh
docker compose -f docker-compose-prysm-local.yml restart beacon-chain
```

### Change Block Time
```bash
# Edit config
vim prysm/config/params/config.go

# Change SECONDS_PER_SLOT
const SecondsPerSlot = 6  // was 12

# Rebuild and restart
./build-prysm-docker.sh
docker compose -f docker-compose-prysm-local.yml down -v
docker compose -f docker-compose-prysm-local.yml up -d
```

### Deploy Smart Contract
```bash
# Your chain is ready!
cast send --rpc-url http://localhost:8545 \
  --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
  --create <bytecode>
```

## 🔍 Explore Prysm Source

```bash
cd prysm

# Consensus logic
ls beacon-chain/blockchain/
ls beacon-chain/core/

# Validator logic
ls validator/client/

# Networking
ls beacon-chain/p2p/

# Configuration
ls config/params/
```

## 📊 Monitoring

### Beacon Chain API
```bash
# Health check
curl http://localhost:3500/eth/v1/node/health

# Validators
curl http://localhost:3500/eth/v1/beacon/states/head/validators
```

### Execution Layer
```bash
# Block number
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Latest block
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}'
```

## 🐛 Troubleshooting

### Build Fails
```bash
docker system prune -a
./build-prysm-docker.sh
```

### Services Crash
```bash
docker compose -f docker-compose-prysm-local.yml logs
docker compose -f docker-compose-prysm-local.yml restart
```

### Port Conflicts
```bash
# Stop other services
docker compose -f docker-compose-private.yml down

# Check ports
lsof -i :8545
```

## 📚 Learn More

- **Prysm Docs**: https://docs.prylabs.network
- **Prysm GitHub**: https://github.com/prysmaticlabs/prysm
- **Ethereum Consensus Specs**: https://github.com/ethereum/consensus-specs

---

**🎉 You're all set! Start modifying Prysm and building your custom consensus layer!**

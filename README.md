# ✅ Erigon + Prysm Private Chain

## Status: Running

All services are up and running:
- ✅ Erigon (Execution Layer) - Port 8545
- ✅ Prysm Beacon (Consensus Layer) - Port 4000/3500
- ✅ Prysm Validator - 64 interop validators

## Quick Start

Services are already running. To restart:

```bash
# Stop
docker compose -f docker-compose-private.yml down

# Start
docker compose -f docker-compose-private.yml up -d

# View logs
docker compose -f docker-compose-private.yml logs -f
```

## Test Commands

```bash
# Check Erigon RPC
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}'

# Check block number
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Check Prysm beacon
curl http://localhost:3500/eth/v1/node/health

# Check account balance
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","latest"],"id":1}'
```

## Pre-funded Accounts

```
Account 1: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
Private Key: 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80

Account 2: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
Private Key: 0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d

Account 3: 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
Private Key: 0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a
```

## Network Parameters

- Chain ID: 32382
- Network ID: 32382
- Block Time: 12 seconds
- Consensus: Proof of Stake
- Validators: 64 (interop mode)

## Files

- `genesis.json` - Execution layer genesis
- `jwt.hex` - JWT secret for EL/CL communication
- `config/config.yml` - Prysm configuration
- `genesis/genesis.ssz` - Consensus layer genesis
- `docker-compose-private.yml` - Docker compose configuration
- `SETUP-PLAN.md` - Complete setup documentation

## Waiting for Block Production

The network is initializing. Blocks should start being produced within 2-3 minutes.

Monitor with:
```bash
watch -n 3 'curl -s -X POST http://localhost:8545 -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[],\"id\":1}"'
```

## Success! 🎉

Your private Ethereum chain with Erigon + Prysm is running!

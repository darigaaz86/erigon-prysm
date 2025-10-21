# Private Ethereum Blockchain - 5 Validators with HotStuff

High-performance private Ethereum blockchain using Erigon execution layer and Prysm consensus layer with HotStuff consensus algorithm.

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Node.js (for testing scripts)

### Start the Network

```bash
# Start all services (5 validators + beacon chain + execution layer)
docker-compose up -d

# Check status
docker ps

# View logs
docker logs -f prysm-beacon-chain-ultra
docker logs -f erigon-ultra-optimized
docker logs -f prysm-validator-1
```

### Verify Block Production

```bash
# Check beacon chain
curl -s http://localhost:3500/eth/v1/beacon/headers/head | jq

# Check execution layer
curl -s http://localhost:8545 -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
```

## Architecture

### Network Configuration
- **Validators**: 5 (distributed across 5 separate services)
- **Block Time**: 6 seconds (`SECONDS_PER_SLOT: 6`)
- **Epoch Length**: 32 slots (~3.2 minutes)
- **Consensus**: HotStuff (Byzantine Fault Tolerant)
- **Gas Limit**: 30,000,000 per block
- **Target TPS**: 2,500-5,000 for simple transfers

### Components

**Execution Layer (Erigon)**
- RPC: `http://localhost:8545`
- Engine API: `http://localhost:8551`
- Handles transaction execution and state management

**Consensus Layer (Prysm Beacon Chain)**
- REST API: `http://localhost:3500`
- gRPC: `localhost:4000`
- Manages consensus and block production with HotStuff

**Validators (5 instances)**
- Each service runs 1 validator
- Participates in block proposals and attestations
- Distributed for better performance and fault tolerance

## Configuration Files

```
docker-compose.yml                  # Main deployment (5 validators)
config/
  ├── config.yml                    # Main consensus config
  └── consensus-hotstuff.yml        # HotStuff settings
genesis/
  ├── genesis.json                  # Execution layer genesis
  └── genesis-5val.ssz              # Beacon chain genesis (5 validators)
scripts/
  ├── deploy-and-transfer.js        # Deploy contracts and test
  └── native-transfer-test.js       # Native ETH transfer test
```

## Testing

### Deploy Smart Contract and Test
```bash
npm install
node scripts/deploy-and-transfer.js
```

### Native Transfer Throughput Test
```bash
node scripts/native-transfer-test.js
```

## Important Operations

### Clean Restart (Required When)
- Changing validator count
- Modifying genesis configuration
- Seeing genesis mismatch errors

```bash
# Stop and remove all data
docker-compose down -v

# Start fresh
docker-compose up -d
```

### Simple Restart (For)
- Configuration tweaks
- Network issues
- Log inspection

```bash
docker-compose restart
```

### Stop Network
```bash
docker-compose down
```

## Monitoring

### Watch Block Production
```bash
# Beacon chain slots
watch -n 1 'curl -s http://localhost:3500/eth/v1/beacon/headers/head | jq .data.header.message.slot'

# Execution blocks
watch -n 1 'curl -s http://localhost:8545 -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[],\"id\":1}"'
```

### Check Validator Status
```bash
# All validators
for i in {1..5}; do echo "=== Validator $i ===" && docker logs prysm-validator-$i 2>&1 | tail -5; done
```

## Documentation

- **[5-VALIDATOR-SUCCESS.md](5-VALIDATOR-SUCCESS.md)** - Current setup details and troubleshooting
- **[HOTSTUFF-IMPLEMENTATION-PLAN.md](HOTSTUFF-IMPLEMENTATION-PLAN.md)** - HotStuff consensus implementation
- **[HOTSTUFF-ACTIVATION-GUIDE.md](HOTSTUFF-ACTIVATION-GUIDE.md)** - How to activate HotStuff
- **[CONSENSUS-CONFIGURATION-GUIDE.md](CONSENSUS-CONFIGURATION-GUIDE.md)** - Consensus settings
- **[TEST-RESULTS.md](TEST-RESULTS.md)** - Performance test results

## Troubleshooting

### Blocks Not Producing

**Symptom**: No new blocks, logs show "payload status is SYNCING or ACCEPTED"

**Cause**: Genesis mismatch between beacon chain and execution layer

**Solution**: Clean restart
```bash
docker-compose down -v
docker-compose up -d
```

### Validators Not Attesting

**Check**: 
```bash
docker logs prysm-validator-1 2>&1 | grep -i "error\|failed"
```

**Common fixes**:
- Ensure beacon node is running: `docker ps | grep beacon`
- Check validator keys are loaded
- Verify validators are in genesis state

### Performance Issues

**Check resource usage**:
```bash
docker stats
```

**Optimize**:
- Increase Docker memory allocation
- Reduce validator count if needed
- Adjust gas limit in genesis

## Network Details

- **Chain ID**: 32382
- **Network ID**: 32382
- **Minimum Validators**: 1 (currently running 5)
- **Finality Time**: ~2 epochs (~6.4 minutes)
- **Fork Versions**: All forks enabled (Altair, Bellatrix, Capella, Deneb)

## RPC Endpoints

```bash
# Execution Layer (Erigon)
http://localhost:8545

# Consensus Layer (Prysm)
http://localhost:3500  # REST API
localhost:4000         # gRPC
```

## License

See LICENSE file for details.

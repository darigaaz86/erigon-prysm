# Erigon + Prysm Private Chain Setup

## Goal
Create a working docker-compose setup with:
- **Execution Layer (EL)**: Erigon
- **Consensus Layer (CL)**: Prysm
- **Network Type**: Private Development Chain

## Architecture

```
┌─────────────────────────────────────┐
│   Private Chain Architecture       │
├─────────────────────────────────────┤
│                                     │
│  ┌──────────────┐  ┌─────────────┐ │
│  │    Prysm     │  │    Prysm    │ │
│  │  Validator   │──│   Beacon    │ │
│  └──────────────┘  │   Chain     │ │
│                    └──────┬──────┘ │
│                           │        │
│                    ┌──────▼──────┐ │
│                    │   Erigon    │ │
│                    │ (Execution) │ │
│                    └─────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

## Why Erigon + Prysm?

### Prysm Benefits
- **Best genesis tooling**: `prysmctl` makes private chain setup easy
- **Well documented**: Excellent guides for custom networks
- **Stable and tested**: Production-ready
- **Interop mode**: Built-in validator keys for testing

### Private Chain Benefits
- **Instant startup**: No sync required
- **Pre-funded accounts**: Ready for testing
- **Full control**: Custom chain ID, block time, validators
- **Fast blocks**: Configurable block time (e.g., 5-12 seconds)
- **No external dependencies**: Completely isolated network

### Private Chain Configuration

#### Network Parameters
- **Chain ID**: 32382 (custom)
- **Network ID**: 32382
- **Block Time**: 12 seconds (Ethereum-like)
- **Consensus**: Proof of Stake (PoS)
- **Validators**: 64 (configurable)
- **Pre-funded Accounts**: 3-5 accounts with large balances

#### Resource Requirements
- **CPU**: 2 cores
- **RAM**: 4-8GB
- **Disk**: 10-20GB
- **Network**: Local only (no internet required)

## Components for Private Chain

### 1. Erigon (Execution Layer)
```yaml
services:
  erigon:
    image: erigontech/erigon:latest
    command:
      - --datadir=/data
      - --networkid=32382
      - --http
      - --http.addr=0.0.0.0
      - --http.port=8545
      - --http.api=eth,net,web3,debug,trace,txpool,erigon
      - --http.vhosts=*
      - --http.corsdomain=*
      - --ws
      - --externalcl
      - --authrpc.addr=0.0.0.0
      - --authrpc.port=8551
      - --authrpc.vhosts=*
      - --authrpc.jwtsecret=/jwt/jwt.hex
      - --nodiscover
      - --maxpeers=0
    ports:
      - "8545:8545"  # HTTP RPC
      - "8546:8546"  # WebSocket
      - "8551:8551"  # Engine API
    volumes:
      - erigon-data:/data
      - ./jwt.hex:/jwt/jwt.hex:ro
      - ./genesis.json:/genesis.json:ro
```

### 2. Prysm Beacon Chain (Consensus Layer)
```yaml
  beacon:
    image: gcr.io/prysmaticlabs/prysm/beacon-chain:stable
    command:
      - --datadir=/data
      - --min-sync-peers=0
      - --genesis-state=/genesis/genesis.ssz
      - --bootstrap-node=
      - --interop-eth1data-votes
      - --chain-config-file=/config/config.yml
      - --contract-deployment-block=0
      - --chain-id=32382
      - --accept-terms-of-use
      - --jwt-secret=/jwt/jwt.hex
      - --suggested-fee-recipient=0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
      - --minimum-peers-per-subnet=0
      - --execution-endpoint=http://erigon:8551
      - --rpc-host=0.0.0.0
      - --grpc-gateway-host=0.0.0.0
    ports:
      - "4000:4000"  # RPC
      - "3500:3500"  # gRPC Gateway
    volumes:
      - beacon-data:/data
      - ./jwt.hex:/jwt/jwt.hex:ro
      - ./genesis:/genesis:ro
      - ./config:/config:ro
    depends_on:
      - erigon
```

### 3. Prysm Validator (Block Production)
```yaml
  validator:
    image: gcr.io/prysmaticlabs/prysm/validator:stable
    command:
      - --datadir=/data
      - --accept-terms-of-use
      - --interop-num-validators=64
      - --interop-start-index=0
      - --chain-config-file=/config/config.yml
      - --beacon-rpc-provider=beacon:4000
      - --suggested-fee-recipient=0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
    volumes:
      - validator-data:/data
      - ./config:/config:ro
    depends_on:
      - beacon
```

## Setup Steps for Private Chain

### Step 1: Generate JWT Secret
```bash
# Create jwt directory
mkdir -p jwt

# Generate JWT secret (32 bytes hex)
openssl rand -hex 32 > jwt/jwt.hex
```

### Step 2: Create Genesis Configuration
Create `genesis.json` with custom chain parameters:
```json
{
  "config": {
    "chainId": 32382,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "berlinBlock": 0,
    "londonBlock": 0,
    "shanghaiTime": 0,
    "cancunTime": 0,
    "terminalTotalDifficulty": 0,
    "terminalTotalDifficultyPassed": true
  },
  "alloc": {
    "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266": {
      "balance": "0x200000000000000000000000000000000000000000000000000000000000000"
    }
  }
}
```

### Step 3: Create Prysm Configuration
```bash
# Create directories
mkdir -p config genesis

# Create beacon chain config
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
```

### Step 4: Generate Genesis State
```bash
# Use prysmctl to generate genesis
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
```

### Step 5: Initialize Erigon with Genesis
```bash
# Initialize Erigon database with custom genesis
docker run --rm -v $(pwd)/erigon-data:/data -v $(pwd)/genesis.json:/genesis.json \
  erigontech/erigon:latest \
  init --datadir=/data /genesis.json
```

### Step 6: Start Services
```bash
docker compose up -d
```

## Key Configuration Points

### 1. JWT Authentication
- Both EL and CL must share the same JWT secret
- Located at `./jwt/jwt.hex`
- Used for secure communication via Engine API

### 2. Network Isolation
- `--nodiscover`: Prevents discovery of external peers
- `--maxpeers=0`: No external connections
- Completely isolated private network

### 3. Network Ports
- **8545**: HTTP RPC (for applications)
- **8546**: WebSocket (for subscriptions)
- **8551**: Engine API (EL ↔ CL communication)
- **4000**: Prysm Beacon RPC
- **3500**: Prysm gRPC Gateway

## Resource Requirements (Private Chain)

- **CPU**: 2 cores (minimal)
- **RAM**: 4-8GB
- **Disk**: 10-20GB (grows slowly)
- **Network**: None (local only)

## Implementation Plan

### Phase 1: Setup Files ✓
- [x] Generate JWT secret
- [x] Create genesis.json
- [x] Generate Lighthouse testnet config
- [x] Create validator keys
- [x] Create docker-compose.yml

### Phase 2: Initialize
- [ ] Initialize Erigon with genesis
- [ ] Verify JWT secret is shared
- [ ] Check file permissions

### Phase 3: Start Services
- [ ] Start Erigon (EL)
- [ ] Start Lighthouse beacon (CL)
- [ ] Start Lighthouse validator
- [ ] Verify all services running

### Phase 4: Verify
- [ ] Check Erigon RPC responding
- [ ] Check Lighthouse beacon API
- [ ] Verify EL/CL connection
- [ ] Wait for block production
- [ ] Test transactions

### Phase 5: Test
- [ ] Deploy smart contract
- [ ] Send transactions
- [ ] Query balances
- [ ] Verify block production

## Pre-funded Accounts

The genesis will include these accounts with large balances:

```
Account 1: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
Private Key: 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80

Account 2: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
Private Key: 0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d

Account 3: 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC
Private Key: 0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a
```

Each account will have: `~10^77 wei` (essentially unlimited for testing)

## Useful Commands

```bash
# Generate JWT secret
openssl rand -hex 32 > jwt/jwt.hex

# Start services
docker compose up -d

# View logs
docker compose logs -f erigon
docker compose logs -f beacon
docker compose logs -f validator

# Check Erigon RPC
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Check Prysm beacon
curl http://localhost:3500/eth/v1/node/health

# Check block production
watch -n 3 'curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[],\"id\":1}"'

# Check account balance
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_getBalance","params":["0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266","latest"],"id":1}'

# Stop services
docker compose down

# Clean restart (removes all data)
docker compose down -v
rm -rf erigon-data lighthouse-data validator-data
```

## Expected Behavior

### Startup Sequence
1. **Erigon starts** → Initializes with genesis → Waits for CL
2. **Lighthouse beacon starts** → Connects to Erigon via Engine API
3. **Lighthouse validator starts** → Connects to beacon → Starts proposing blocks
4. **Blocks produced** → Every 12 seconds (configurable)

### Timeline
- **0-30 seconds**: Services starting up
- **30-60 seconds**: EL/CL connection established
- **60-120 seconds**: First blocks produced
- **2+ minutes**: Steady block production

### Success Indicators
- ✅ Erigon logs show "Engine API" connection
- ✅ Lighthouse logs show "Execution engine online"
- ✅ Block number increases every 12 seconds
- ✅ RPC responds to queries
- ✅ Transactions can be sent

## Troubleshooting

### Issue: Services won't start
- Check JWT secret exists and is readable
- Verify genesis.json is valid
- Check Docker has enough resources

### Issue: EL/CL not connecting
- Verify JWT secret is identical in both services
- Check Engine API port (8551) is accessible
- Look for "JWT" errors in logs

### Issue: No blocks being produced
- Check validator keys are loaded
- Verify validator is connected to beacon
- Check beacon is connected to execution layer
- Wait 2-3 minutes for initialization

### Issue: Transactions failing
- Verify account has balance
- Check gas price is reasonable
- Ensure chain ID matches (32382)

## References

- Erigon: https://github.com/erigontech/erigon
- Prysm: https://docs.prylabs.network/
- Prysm Custom Networks: https://docs.prylabs.network/docs/advanced/proof-of-stake-devnet
- Engine API: https://github.com/ethereum/execution-apis/tree/main/src/engine

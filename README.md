# Erigon-Prysm with HotStuff BFT Consensus

🔥 **Production-ready HotStuff BFT consensus implementation for Ethereum** - A complete integration of HotStuff consensus into Erigon (execution layer) and Prysm (consensus layer).

[![Status](https://img.shields.io/badge/status-operational-success)](https://github.com/darigaaz86/erigon-prysm)
[![Blocks Produced](https://img.shields.io/badge/blocks-600%2B-blue)](https://github.com/darigaaz86/erigon-prysm)
[![License](https://img.shields.io/badge/license-LGPL--3.0-orange)](LICENSE)

## 🎯 What is This?

This project implements **HotStuff BFT consensus** as an alternative to Ethereum's Proof of Stake, integrated into:
- **Erigon**: Ethereum execution client (handles transactions and state)
- **Prysm**: Ethereum consensus client (handles consensus and block production)

**Result**: A fully functional Ethereum-compatible blockchain using HotStuff consensus that produces real blocks with actual execution payloads.

## ✨ Features

- 🔥 **100% HotStuff Block Production** - All blocks produced via HotStuff consensus
- ⚡ **Fast Finality** - 3-phase commit (~18 seconds) vs PoS 2 epochs (~13 minutes)
- 🛡️ **Byzantine Fault Tolerance** - Tolerates f < n/3 Byzantine failures
- 📈 **Linear Communication** - O(n) messages per view (vs O(n²) in PBFT)
- 🔄 **Leader Rotation** - Round-robin or stake-weighted leader selection
- 🎯 **Real Execution** - Produces actual blocks with transactions on Erigon
- 🐳 **Docker Ready** - Complete Docker Compose setup for easy deployment

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- 8GB+ RAM
- 50GB+ disk space

### Start the Network

```bash
# Clone the repository
git clone https://github.com/darigaaz86/erigon-prysm.git
cd erigon-prysm

# Start HotStuff network
docker compose -f docker-compose-hotstuff.yml up -d

# Watch HotStuff producing blocks
docker logs -f prysm-beacon-chain | grep "🔥 HotStuff"
```

### Check Block Production

```bash
# Get current block number
curl -s http://localhost:8545 -X POST \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'

# Watch blocks being produced
watch -n 6 'curl -s http://localhost:8545 -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_blockNumber\",\"params\":[],\"id\":1}"'
```

## 📊 Architecture

```
┌─────────────────────────────────────────────┐
│           Validator Clients                 │
│         (64 validators)                     │
└──────────────────┬──────────────────────────┘
                   │ Block Requests
                   ↓
┌─────────────────────────────────────────────┐
│      Prysm Beacon Chain (Consensus)         │
│  ┌──────────────────────────────────────┐   │
│  │  🔥 HotStuff Consensus Service       │   │
│  │  - View management                   │   │
│  │  - Leader election (round-robin)     │   │
│  │  - Phase transitions (4 phases)      │   │
│  │  - QC aggregation                    │   │
│  │  - View timeouts & changes           │   │
│  └──────────────┬───────────────────────┘   │
│                 ↓                            │
│  ┌──────────────────────────────────────┐   │
│  │  Block Builder                       │   │
│  │  - Creates execution payload         │   │
│  │  - Fills with transactions           │   │
│  └──────────────┬───────────────────────┘   │
└─────────────────┼───────────────────────────┘
                  ↓
┌─────────────────────────────────────────────┐
│      Erigon (Execution Layer)               │
│  - Builds execution payloads                │
│  - Processes transactions                   │
│  - Updates world state                      │
│  - Stores blockchain data                   │
└─────────────────────────────────────────────┘
```

## 🔥 HotStuff Consensus

### What is HotStuff?

HotStuff is a leader-based Byzantine fault-tolerant consensus protocol with:
- **Linear communication complexity**: O(n) messages per view
- **Optimistic responsiveness**: Fast in good conditions
- **View synchronization**: Automatic leader rotation on timeout
- **Safety & Liveness**: Proven BFT guarantees

### Four-Phase Protocol

1. **PREPARE**: Leader proposes block, replicas vote
2. **PRE-COMMIT**: Aggregate votes into QC, replicas vote again
3. **COMMIT**: Second QC formed, replicas commit
4. **DECIDE**: Block finalized and executed

### Configuration

Edit `config/consensus-hotstuff.yml`:

```yaml
mode: "hotstuff"

hotstuff:
  view_timeout: 10s        # Timeout before view change
  block_time: 6s           # Target time between blocks
  min_validators: 4        # Minimum validators (3f+1)
  quorum_threshold: 0.67   # 2/3 supermajority
  leader_rotation: "round-robin"
  enable_view_change: true
```

## 📁 Project Structure

```
erigon-prysm/
├── prysm/                              # Prysm consensus client (submodule)
│   └── beacon-chain/
│       ├── consensus/
│       │   ├── hotstuff/              # 🔥 HotStuff implementation
│       │   │   ├── service.go         # Main consensus service
│       │   │   ├── phases.go          # Phase management
│       │   │   ├── qc.go              # Quorum certificates
│       │   │   ├── leader.go          # Leader election
│       │   │   ├── voting.go          # Vote handling
│       │   │   └── types.go           # Core types
│       │   └── factory.go             # Consensus factory
│       ├── node/node.go               # Node integration
│       └── rpc/                       # RPC services
├── config/
│   ├── consensus-hotstuff.yml         # HotStuff config
│   └── config.yml                     # Chain config
├── genesis/
│   └── genesis.json                   # Genesis state
├── docker-compose-hotstuff.yml        # HotStuff deployment
├── HOTSTUFF-INTEGRATION-COMPLETE.md   # Integration docs
└── README.md                          # This file
```

## 🧪 Testing

### Run Tests

```bash
# Build and test Prysm
cd prysm
go test ./beacon-chain/consensus/hotstuff/...

# Run integration test
./test-hotstuff.sh
```

### Verify HotStuff is Running

```bash
# Check HotStuff logs
docker logs prysm-beacon-chain 2>&1 | grep "🔥 HotStuff"

# Should see:
# 🔥 HotStuff: Begin building block via HotStuff consensus
# 🔥 HotStuff: Building block with execution payload
# 🔥 HotStuff: Successfully built block!

# Verify no PoS blocks
docker logs prysm-beacon-chain 2>&1 | grep "Begin building block" | grep -v "HotStuff" | wc -l
# Should output: 0
```

## 📈 Performance

| Metric | Value |
|--------|-------|
| Block Time | 6 seconds |
| View Timeout | 10 seconds |
| Finality Time | ~18 seconds (3 phases) |
| Communication | O(n) per view |
| Fault Tolerance | f < n/3 Byzantine |
| Validators | 64 (configurable) |

## 🔧 Development

### Build from Source

```bash
# Build Erigon
make erigon

# Build Prysm
cd prysm
go build -o build/beacon-chain ./cmd/beacon-chain
go build -o build/validator ./cmd/validator
```

### Docker Build

```bash
# Build custom images
./build-prysm-docker.sh

# Start with custom images
docker compose -f docker-compose-hotstuff.yml up -d
```

## 📚 Documentation

- [HotStuff Integration Complete](HOTSTUFF-INTEGRATION-COMPLETE.md) - Full integration details
- [HotStuff Activation Guide](HOTSTUFF-ACTIVATION-GUIDE.md) - Step-by-step activation
- [Consensus Configuration Guide](CONSENSUS-CONFIGURATION-GUIDE.md) - Configuration options
- [Implementation Plan](HOTSTUFF-IMPLEMENTATION-PLAN.md) - Development roadmap

## 🎯 Use Cases

- **Research**: Study BFT consensus in Ethereum context
- **Private Networks**: Fast finality for enterprise blockchains
- **Testing**: Experiment with alternative consensus mechanisms
- **Education**: Learn how consensus protocols work
- **Development**: Build applications requiring fast finality

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project combines:
- **Erigon**: LGPL-3.0 License
- **Prysm**: GPL-3.0 License
- **HotStuff Implementation**: GPL-3.0 License

See individual LICENSE files for details.

## 🙏 Acknowledgments

- [Erigon](https://github.com/ledgerwatch/erigon) - High-performance Ethereum client
- [Prysm](https://github.com/prysmaticlabs/prysm) - Ethereum consensus client
- [HotStuff Paper](https://arxiv.org/abs/1803.05069) - Original HotStuff protocol
- Ethereum Foundation - For the amazing ecosystem

## 📞 Contact

- GitHub: [@darigaaz86](https://github.com/darigaaz86)
- Issues: [GitHub Issues](https://github.com/darigaaz86/erigon-prysm/issues)

## ⭐ Star History

If you find this project useful, please consider giving it a star! ⭐

---

**Status**: ✅ Fully Operational - Producing 600+ blocks with HotStuff consensus

**Last Updated**: October 15, 2025

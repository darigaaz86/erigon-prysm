# HotStuff Consensus Activation Guide

## Current Status

✅ **Implemented**: HotStuff consensus code (Steps 1-4)
- Consensus abstraction layer
- HotStuff types and QC
- Leader election
- Three-phase protocol

❌ **Not Integrated**: HotStuff is not active yet (Step 5 incomplete)

**Currently Running**: Standard Ethereum PoS (Gasper) consensus

## How to Verify Current Consensus

### Method 1: Check Prysm Logs
```bash
docker logs prysm-beacon-chain 2>&1 | grep -i "consensus"
```

**Current Output**: No HotStuff-specific messages (using default PoS)

### Method 2: Check for HotStuff Messages
```bash
docker logs prysm-beacon-chain 2>&1 | grep -i "hotstuff\|prepare\|pre-commit\|view"
```

**Expected if HotStuff Active**: 
- "HotStuff consensus initialized"
- "View: X, Phase: PREPARE"
- "QC collected for block X"
- "Leader for view X: validator Y"

**Current Output**: None (HotStuff not active)

### Method 3: Check Attestation Pattern
```bash
docker logs prysm-beacon-chain 2>&1 | grep "attestation" | tail -10
```

**PoS Pattern**: Standard attestations every slot
**HotStuff Pattern**: Would show PREPARE/PRE-COMMIT/COMMIT votes

## What Needs to Be Done (Step 5: Integration)

### 1. Add Command-Line Flags

**File**: `prysm/cmd/beacon-chain/flags/base.go`

Add:
```go
var (
    // ConsensusMode flag
    ConsensusModeFlag = &cli.StringFlag{
        Name:  "consensus-mode",
        Usage: "Consensus mode to use: 'pos' or 'hotstuff'",
        Value: "pos",
    }
    
    // ConsensusConfig flag
    ConsensusConfigFlag = &cli.StringFlag{
        Name:  "consensus-config",
        Usage: "Path to consensus configuration YAML file",
        Value: "",
    }
)
```

### 2. Update Main Beacon Chain Entry Point

**File**: `prysm/cmd/beacon-chain/main.go`

Add flags to `appFlags`:
```go
var appFlags = []cli.Flag{
    // ... existing flags ...
    flags.ConsensusModeFlag,
    flags.ConsensusConfigFlag,
}
```

### 3. Initialize Consensus Factory

**File**: `prysm/beacon-chain/node/node.go` (or similar)

Add consensus initialization:
```go
import (
    "github.com/OffchainLabs/prysm/v6/beacon-chain/consensus"
)

func New(ctx *cli.Context, cancel context.CancelFunc, opts ...Option) (*BeaconNode, error) {
    // ... existing code ...
    
    // Load consensus configuration
    consensusMode := ctx.String(flags.ConsensusModeFlag.Name)
    consensusConfigPath := ctx.String(flags.ConsensusConfigFlag.Name)
    
    var consensusConfig *consensus.Config
    if consensusConfigPath != "" {
        // Load from YAML file
        consensusConfig, err = consensus.LoadConfigFromFile(consensusConfigPath)
        if err != nil {
            return nil, err
        }
    } else {
        // Use default config based on mode
        if consensusMode == "hotstuff" {
            consensusConfig = consensus.DefaultHotStuffConfig()
        } else {
            consensusConfig = consensus.DefaultPoSConfig()
        }
    }
    
    // Create consensus factory
    factory, err := consensus.NewFactory(consensusConfig)
    if err != nil {
        return nil, err
    }
    
    // Create consensus instance
    consensusService, err := factory.Create(ctx.Context)
    if err != nil {
        return nil, err
    }
    
    // Use consensusService instead of blockchain.Service
    // ... rest of initialization ...
}
```

### 4. Implement HotStuff Factory Method

**File**: `prysm/beacon-chain/consensus/factory.go`

Update `createHotStuff`:
```go
func (f *Factory) createHotStuff(ctx context.Context, opts ...Option) (Consensus, error) {
    // Import the hotstuff service
    import "github.com/OffchainLabs/prysm/v6/beacon-chain/consensus/hotstuff"
    
    // Create HotStuff service with config
    service, err := hotstuff.NewService(ctx, f.config.HotStuff)
    if err != nil {
        return nil, err
    }
    
    return service, nil
}
```

### 5. Wire Up HotStuff Service

**File**: `prysm/beacon-chain/consensus/hotstuff/service.go`

The service already exists but needs to be connected to:
- Block production pipeline
- Attestation processing
- Fork choice
- State management

### 6. Add Configuration Loading

**File**: `prysm/beacon-chain/consensus/config.go`

Add:
```go
func LoadConfigFromFile(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    if err := config.Validate(); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

## Testing HotStuff After Integration

### 1. Start with HotStuff Mode

```bash
docker compose -f docker-compose-hotstuff.yml down -v
docker compose -f docker-compose-hotstuff.yml up -d
```

### 2. Verify HotStuff is Active

```bash
# Check for HotStuff initialization
docker logs prysm-beacon-chain 2>&1 | grep "HotStuff"

# Expected output:
# "HotStuff consensus initialized"
# "Using round-robin leader election"
# "View timeout: 10s"
```

### 3. Monitor HotStuff Operation

```bash
# Watch for HotStuff phases
docker logs -f prysm-beacon-chain | grep -E "PREPARE|PRE-COMMIT|COMMIT|DECIDE"

# Expected output:
# "Phase: PREPARE, View: 1, Block: 0x..."
# "Phase: PRE-COMMIT, View: 1, QC collected"
# "Phase: COMMIT, View: 1, Block committed"
# "Phase: DECIDE, View: 1, Block finalized"
```

### 4. Check Leader Election

```bash
# Watch leader changes
docker logs -f prysm-beacon-chain | grep "Leader"

# Expected output:
# "Leader for view 1: validator 0"
# "Leader for view 2: validator 1"
# "Leader for view 3: validator 2"
```

### 5. Verify QC Collection

```bash
# Watch QC formation
docker logs -f prysm-beacon-chain | grep "QC"

# Expected output:
# "QC Builder: Added vote from validator 5"
# "QC Builder: Quorum reached (43/64 votes)"
# "QC created for block 0x... at view 10"
```

## Current vs. HotStuff Behavior

### Current (PoS/Gasper)
- **Consensus**: LMD-GHOST + Casper FFG
- **Finality**: 2 epochs (~12.8 minutes)
- **Messages**: Attestations
- **Leader**: Proposer per slot
- **Logs**: "Synced new block", "attestation", "justified", "finalized"

### With HotStuff Active
- **Consensus**: HotStuff BFT
- **Finality**: 3 phases (~18 seconds with 6s slots)
- **Messages**: PREPARE, PRE-COMMIT, COMMIT votes
- **Leader**: Round-robin or stake-weighted
- **Logs**: "View", "Phase", "QC", "Leader"

## Quick Test Commands

```bash
# Current consensus (should show PoS behavior)
docker logs prysm-beacon-chain 2>&1 | grep -c "attestation"  # High count
docker logs prysm-beacon-chain 2>&1 | grep -c "PREPARE"      # Zero

# After HotStuff integration (should show HotStuff behavior)
docker logs prysm-beacon-chain 2>&1 | grep -c "PREPARE"      # High count
docker logs prysm-beacon-chain 2>&1 | grep -c "View:"        # High count
```

## Summary

**Current State**: 
- ✅ HotStuff code implemented (Steps 1-4)
- ❌ HotStuff not integrated (Step 5 incomplete)
- 🔄 Running standard PoS consensus

**To Activate HotStuff**:
1. Complete Step 5 integration tasks above
2. Rebuild Prysm Docker images
3. Start with `--consensus-mode=hotstuff` flag
4. Verify with log messages showing HotStuff phases

**Estimated Effort**: 2-3 days for full integration and testing

---

**Status**: Documentation complete, ready for Step 5 implementation

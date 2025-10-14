# HotStuff Consensus Implementation Plan

## Goal

Implement HotStuff consensus as an alternative to Ethereum's PoS consensus in Prysm, with the ability to switch between them via configuration.

## Phase 1: Understanding Current PoS Consensus

### 1.1 Core Consensus Components

#### Beacon Chain Core (`beacon-chain/blockchain/`)
- **`service.go`** - Main blockchain service (Service struct manages full PoS beacon chain)
  - Key fields: `head`, `ForkChoiceStore`, `AttPool`, `StateGen`
  - Handles block processing, attestations, and chain state
- **`process_block.go`** - Block processing logic
- **`process_attestation.go`** - Attestation handling
- **`chain_info.go`** - Chain state management

#### Fork Choice (`beacon-chain/forkchoice/`)
- **`interfaces.go`** - Fork choice interface definition
  - `ForkChoicer` - Main interface with HeadRetriever, BlockProcessor, AttestationProcessor
  - `Head()` - Computes current chain head
  - `ProcessAttestation()` - Processes validator attestations
- **`doubly-linked-tree/`** - Fork choice implementation (LMD-GHOST)
- Current: Uses **LMD-GHOST** (Latest Message Driven Greedy Heaviest Observed SubTree)

#### Block Production (`beacon-chain/core/blocks/`)
- **`block.go`** - Block creation and validation
- **`proposer.go`** - Block proposer selection
- **`attestation.go`** - Attestation processing

#### Validator Duties (`validator/client/`)
- **`validator.go`** - Main validator client
- **`propose.go`** - Block proposal logic
- **`attest.go`** - Attestation submission
- **`aggregate.go`** - Attestation aggregation

### 1.2 Current PoS Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    Ethereum PoS Consensus                    │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. Slot Timer (12s)                                        │
│     ↓                                                        │
│  2. Proposer Selection (deterministic based on slot)        │
│     ↓                                                        │
│  3. Block Proposal                                          │
│     - Get execution payload from Erigon                     │
│     - Create beacon block                                   │
│     - Sign and broadcast                                    │
│     ↓                                                        │
│  4. Attestations (validators vote)                          │
│     - Vote on head block                                    │
│     - Vote on justified checkpoint                          │
│     - Vote on finalized checkpoint                          │
│     ↓                                                        │
│  5. Fork Choice (LMD-GHOST)                                 │
│     - Weight blocks by attestations                         │
│     - Select heaviest subtree                               │
│     ↓                                                        │
│  6. Finality (Casper FFG)                                   │
│     - 2/3 majority → justified                              │
│     - Justified + next epoch 2/3 → finalized                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 1.3 Existing PoS Code Structure

#### Core Components and Their Roles

**1. Blockchain Service (`beacon-chain/blockchain/service.go`)**
- **Role**: Main orchestrator of consensus
- **Key Fields**:
  - `cfg.ForkChoiceStore` - Fork choice algorithm
  - `head` - Current chain head (root, block, state)
  - `cfg.AttPool` - Attestation pool
  - `cfg.StateGen` - State generator
- **Key Methods**:
  - `ReceiveBlock()` - Receives and validates new blocks
  - `ProcessBlock()` - Processes blocks through state transition
  - `ProcessAttestation()` - Processes validator attestations
  - `UpdateHead()` - Updates chain head based on fork choice

**2. Fork Choice (`beacon-chain/forkchoice/doubly-linked-tree/`)**
- **Role**: Implements LMD-GHOST fork choice algorithm
- **Key Files**:
  - `forkchoice.go` - Main fork choice logic
    - `Head()` - Computes chain head (most important!)
    - `ProcessAttestation()` - Updates vote accounting
  - `store.go` - Fork choice store with block tree
  - `node.go` - Block tree node structure
  - `ffg.go` - Casper FFG justification/finalization
- **Algorithm**: LMD-GHOST (Latest Message Driven Greedy Heaviest Observed SubTree)
  - Weights blocks by attestations
  - Selects heaviest subtree as head

**3. Block Processing (`beacon-chain/blockchain/process_block.go`)**
- **Role**: Validates and processes blocks
- **Key Functions**:
  - `onBlock()` - Main block processing entry point
  - `onBlockBatch()` - Batch block processing
  - Validates block structure, signatures, state transition
  - Updates fork choice with new block

**4. Attestation Processing (`beacon-chain/blockchain/process_attestation.go`)**
- **Role**: Processes validator votes (attestations)
- **Key Functions**:
  - `OnAttestation()` - Processes single attestation
  - Validates attestation
  - Updates fork choice vote accounting
  - Aggregates attestations

**5. Head Selection (`beacon-chain/blockchain/head.go`)**
- **Role**: Manages chain head
- **Key Functions**:
  - `saveHead()` - Saves new head to cache and DB
  - `UpdateAndSaveHeadWithBalances()` - Updates head with validator balances
- **Head Structure**:
  ```go
  type head struct {
      root       [32]byte                    // current head root
      block      interfaces.ReadOnlySignedBeaconBlock
      state      state.BeaconState
      slot       primitives.Slot
      optimistic bool
  }
  ```

**6. State Transition (`beacon-chain/core/transition/`)**
- **Role**: Applies state transitions
- **Key Functions**:
  - `ExecuteStateTransition()` - Main state transition function
  - Processes slots, blocks, epochs
  - Updates validator balances, justification, finalization

**7. Epoch Processing (`beacon-chain/core/epoch/`)**
- **Role**: Processes epoch boundaries
- **Key Functions**:
  - `ProcessEpoch()` - Processes epoch transition
  - Updates justification and finalization
  - Processes rewards and penalties
  - Updates validator registry

#### How PoS Consensus Works in Prysm

```
┌─────────────────────────────────────────────────────────────┐
│                  Current PoS Flow in Prysm                   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. Block Reception (blockchain/receive_block.go)           │
│     - ReceiveBlock() called                                 │
│     - Validates block structure                             │
│     - Checks signatures                                     │
│     ↓                                                        │
│  2. Block Processing (blockchain/process_block.go)          │
│     - onBlock() processes block                             │
│     - Executes state transition                             │
│     - Updates execution payload                             │
│     ↓                                                        │
│  3. Fork Choice Update (forkchoice/doubly-linked-tree/)     │
│     - InsertNode() adds block to tree                       │
│     - Updates block weights                                 │
│     ↓                                                        │
│  4. Attestation Processing (blockchain/process_attestation.go) │
│     - OnAttestation() processes votes                       │
│     - ProcessAttestation() updates fork choice              │
│     - Aggregates attestations                               │
│     ↓                                                        │
│  5. Head Computation (forkchoice/doubly-linked-tree/)       │
│     - Head() computes new head                              │
│     - updateBalances() updates validator balances           │
│     - applyWeightChanges() applies attestation weights      │
│     - updateBestDescendant() finds heaviest subtree         │
│     ↓                                                        │
│  6. Head Update (blockchain/head.go)                        │
│     - saveHead() saves new head                             │
│     - Updates head cache                                    │
│     - Notifies subscribers                                  │
│     ↓                                                        │
│  7. Epoch Processing (core/epoch/)                          │
│     - ProcessEpoch() at epoch boundaries                    │
│     - Updates justification (2/3 majority)                  │
│     - Updates finalization (justified + next epoch)         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

#### Key Files to Study

```
prysm/beacon-chain/
├── blockchain/
│   ├── service.go              # Main service - START HERE
│   ├── receive_block.go        # Block reception
│   ├── process_block.go        # Block processing
│   ├── process_attestation.go  # Attestation processing
│   ├── head.go                 # Head management
│   └── chain_info.go           # Chain state queries
│
├── forkchoice/
│   ├── interfaces.go           # Fork choice interface
│   └── doubly-linked-tree/
│       ├── forkchoice.go       # Main fork choice - IMPORTANT
│       ├── store.go            # Fork choice store
│       ├── node.go             # Block tree nodes
│       └── ffg.go              # Casper FFG finality
│
└── core/
    ├── blocks/
    │   ├── block.go            # Block validation
    │   └── proposer.go         # Proposer selection
    ├── transition/
    │   └── transition.go       # State transition
    └── epoch/
        └── epoch_processing.go # Epoch processing
```

## Phase 2: HotStuff Consensus Overview

### 2.1 HotStuff Algorithm

HotStuff is a **leader-based Byzantine Fault Tolerant (BFT)** consensus protocol.

#### Key Characteristics:
- **Linear communication complexity**: O(n) messages per view
- **Optimistic responsiveness**: Progress in network delay time
- **Simplicity**: Three-phase commit protocol
- **Safety**: Byzantine fault tolerance (tolerates f < n/3 failures)
- **Liveness**: Guaranteed progress with synchrony

#### Three-Phase Protocol:

```
┌─────────────────────────────────────────────────────────────┐
│                    HotStuff Consensus                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  View v (Leader: replica i)                                 │
│                                                              │
│  Phase 1: PREPARE                                           │
│    Leader → All: PREPARE(v, block, qc_high)                 │
│    All → Leader: VOTE-PREPARE(v, block_hash)                │
│    Leader collects 2f+1 votes → prepare_qc                  │
│                                                              │
│  Phase 2: PRE-COMMIT                                        │
│    Leader → All: PRE-COMMIT(v, prepare_qc)                  │
│    All → Leader: VOTE-PRE-COMMIT(v, block_hash)             │
│    Leader collects 2f+1 votes → precommit_qc                │
│                                                              │
│  Phase 3: COMMIT                                            │
│    Leader → All: COMMIT(v, precommit_qc)                    │
│    All → Leader: VOTE-COMMIT(v, block_hash)                 │
│    Leader collects 2f+1 votes → commit_qc                   │
│                                                              │
│  Phase 4: DECIDE                                            │
│    Leader → All: DECIDE(v, commit_qc)                       │
│    All: Execute block, update state                         │
│                                                              │
│  View Change (if timeout or leader failure)                 │
│    Replicas → New Leader: VIEW-CHANGE(v+1, qc_high)         │
│    New Leader: Start new view with highest QC               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Key Concepts

#### Quorum Certificate (QC)
- Collection of 2f+1 signatures on a block
- Proves that a supermajority agreed on the block
- Types: prepare_qc, precommit_qc, commit_qc

#### View
- Consensus round with a designated leader
- Increments on timeout or successful commit
- Leader rotation: deterministic or round-robin

#### Safety Rule
- A block is committed only after three consecutive QCs
- Ensures Byzantine fault tolerance

#### Liveness Rule
- View change mechanism ensures progress
- Timeout triggers leader change
- New leader uses highest QC to continue

## Phase 3: Implementation Strategy

### 3.1 Configuration System

Create a consensus mode selector:

```yaml
# config/consensus.yml
consensus:
  mode: "hotstuff"  # or "pos" for Ethereum PoS
  
  hotstuff:
    view_timeout: 10s
    block_time: 6s
    min_validators: 4
    quorum_threshold: 0.67  # 2f+1 out of 3f+1
    leader_rotation: "round-robin"  # or "stake-weighted"
    
  pos:
    # Existing Ethereum PoS config
    slot_duration: 12s
    epochs_per_sync_committee: 256
```

### 3.2 Architecture Design

#### Current PoS Consensus Code Location

The existing PoS consensus is **NOT in a single package** but spread across multiple packages:

```
prysm/beacon-chain/
├── blockchain/                       # EXISTING: Main consensus orchestration
│   ├── service.go                    # Main blockchain service (orchestrates everything)
│   ├── head.go                       # Head selection and saving
│   ├── process_block.go              # Block processing logic
│   ├── process_attestation.go        # Attestation processing
│   └── receive_block.go              # Block reception and validation
│
├── forkchoice/                       # EXISTING: Fork choice algorithm
│   ├── interfaces.go                 # Fork choice interface (ForkChoicer)
│   └── doubly-linked-tree/           # LMD-GHOST implementation
│       ├── forkchoice.go             # Main fork choice logic
│       │   - Head() - Computes chain head
│       │   - ProcessAttestation() - Processes votes
│       ├── store.go                  # Fork choice store
│       ├── node.go                   # Block tree nodes
│       └── ffg.go                    # Casper FFG finality
│
└── core/                             # EXISTING: Core consensus rules
    ├── blocks/                       # Block operations
    │   ├── block.go                  # Block validation
    │   └── proposer.go               # Proposer selection
    ├── transition/                   # State transition
    │   └── transition.go             # State transition function
    └── epoch/                        # Epoch processing
        └── epoch_processing.go       # Justification & finalization
```

#### Proposed New Architecture

We'll create a **consensus abstraction layer** that wraps the existing code:

```
prysm/beacon-chain/
├── consensus/                        # NEW: Consensus abstraction
│   ├── interface.go                  # Consensus interface
│   ├── factory.go                    # Consensus factory
│   ├── config.go                     # Consensus configuration
│   │
│   ├── pos/                          # NEW: PoS wrapper (wraps existing code)
│   │   ├── service.go                # Wraps blockchain.Service
│   │   ├── forkchoice.go             # Wraps forkchoice.ForkChoicer
│   │   └── adapter.go                # Adapts existing PoS to interface
│   │
│   └── hotstuff/                     # NEW: HotStuff implementation
│       ├── service.go                # Main HotStuff service
│       ├── types.go                  # HotStuff types (QC, View, etc)
│       ├── leader.go                 # Leader election
│       ├── phases.go                 # Three-phase protocol
│       ├── voting.go                 # Vote collection
│       ├── qc.go                     # Quorum certificate
│       ├── view_change.go            # View change protocol
│       └── safety.go                 # Safety rules
│
├── blockchain/                       # MODIFIED: Use consensus interface
│   └── service.go                    # Modified to delegate to consensus
│
└── forkchoice/                       # UNCHANGED: Keep existing
    └── doubly-linked-tree/           # Still used by PoS mode
```

**Key Insight**: We're NOT moving existing PoS code, we're **wrapping it** with an interface so we can plug in HotStuff alongside it.

### 3.3 Core Interfaces

```go
// beacon-chain/consensus/interface.go

package consensus

import (
    "context"
    ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
)

// Consensus defines the interface for consensus mechanisms
type Consensus interface {
    // Block Processing
    ProcessBlock(ctx context.Context, block *ethpb.SignedBeaconBlock) error
    ProposeBlock(ctx context.Context, slot uint64) (*ethpb.SignedBeaconBlock, error)
    
    // Voting/Attestation
    ProcessVote(ctx context.Context, vote Vote) error
    CreateVote(ctx context.Context, block *ethpb.BeaconBlock) (Vote, error)
    
    // Chain Selection
    Head(ctx context.Context) ([32]byte, error)
    UpdateHead(ctx context.Context) error
    
    // Finality
    FinalizedCheckpoint() *ethpb.Checkpoint
    JustifiedCheckpoint() *ethpb.Checkpoint
    
    // Lifecycle
    Start() error
    Stop() error
}

// Vote represents a validator's vote (attestation in PoS, vote in HotStuff)
type Vote interface {
    BlockRoot() [32]byte
    Signature() []byte
    ValidatorIndex() uint64
}
```

## Phase 4: Implementation Steps

### Step 1: Create Consensus Abstraction (Week 1)

**Files to create:**
- `beacon-chain/consensus/interface.go`
- `beacon-chain/consensus/factory.go`
- `beacon-chain/consensus/config.go`

**Tasks:**
1. Define `Consensus` interface
2. Create factory pattern for consensus selection
3. Add configuration loading

### Step 2: Wrap Existing PoS (Week 1-2)

**Files to modify:**
- `beacon-chain/blockchain/service.go`
- `beacon-chain/forkchoice/protoarray/store.go`

**Tasks:**
1. Create `beacon-chain/consensus/pos/` package
2. Wrap existing PoS logic into interface implementation
3. Refactor blockchain service to use consensus interface
4. Test that existing PoS still works

### Step 3: Implement HotStuff Types (Week 2)

**Files to create:**
- `beacon-chain/consensus/hotstuff/types.go`
- `beacon-chain/consensus/hotstuff/qc.go`

**Tasks:**
1. Define HotStuff data structures:
   - `View` - consensus round
   - `QuorumCertificate` - 2f+1 signatures
   - `HotStuffBlock` - block with QC
   - `Vote` - validator vote
2. Implement QC creation and verification
3. Add serialization/deserialization

### Step 4: Implement Leader Election (Week 3)

**Files to create:**
- `beacon-chain/consensus/hotstuff/leader.go`

**Tasks:**
1. Implement round-robin leader selection
2. Implement stake-weighted leader selection
3. Add leader verification
4. Handle leader rotation on view change

### Step 5: Implement Three-Phase Protocol (Week 3-4)

**Files to create:**
- `beacon-chain/consensus/hotstuff/phases.go`
- `beacon-chain/consensus/hotstuff/voting.go`

**Tasks:**
1. Implement PREPARE phase
2. Implement PRE-COMMIT phase
3. Implement COMMIT phase
4. Implement DECIDE phase
5. Add vote collection and aggregation
6. Add timeout handling

### Step 6: Implement View Change (Week 4)

**Files to create:**
- `beacon-chain/consensus/hotstuff/view_change.go`

**Tasks:**
1. Implement view change protocol
2. Add timeout detection
3. Handle VIEW-CHANGE messages
4. Implement new view initialization

### Step 7: Implement Safety Rules (Week 5)

**Files to create:**
- `beacon-chain/consensus/hotstuff/safety.go`

**Tasks:**
1. Implement voting rules
2. Implement commit rules
3. Add Byzantine fault detection
4. Ensure safety guarantees

### Step 8: Integration (Week 5-6)

**Files to modify:**
- `beacon-chain/blockchain/service.go`
- `validator/client/validator.go`
- `validator/client/propose.go`

**Tasks:**
1. Integrate HotStuff with blockchain service
2. Update validator client for HotStuff
3. Add consensus mode switching
4. Test both modes

### Step 9: Testing (Week 6-7)

**Files to create:**
- `beacon-chain/consensus/hotstuff/*_test.go`
- `testing/endtoend/hotstuff_test.go`

**Tasks:**
1. Unit tests for all HotStuff components
2. Integration tests
3. Byzantine fault tests
4. Performance benchmarks
5. End-to-end tests

### Step 10: Documentation (Week 7)

**Tasks:**
1. API documentation
2. Configuration guide
3. Migration guide
4. Performance comparison

## Phase 5: Detailed Implementation

### 5.1 HotStuff Types

```go
// beacon-chain/consensus/hotstuff/types.go

package hotstuff

import (
    "github.com/prysmaticlabs/prysm/v5/crypto/bls"
)

// View represents a consensus round
type View struct {
    Number uint64
    Leader uint64
}

// QuorumCertificate represents 2f+1 signatures on a block
type QuorumCertificate struct {
    View       uint64
    BlockHash  [32]byte
    Signatures []bls.Signature
    Signers    []uint64
}

// HotStuffBlock extends beacon block with HotStuff metadata
type HotStuffBlock struct {
    BeaconBlock *ethpb.BeaconBlock
    View        uint64
    QC          *QuorumCertificate  // QC for parent block
    Justify     *QuorumCertificate  // Highest QC known
}

// Vote represents a validator's vote in a phase
type Vote struct {
    View          uint64
    BlockHash     [32]byte
    Phase         Phase
    ValidatorIdx  uint64
    Signature     bls.Signature
}

// Phase represents the current phase in the protocol
type Phase uint8

const (
    PhasePrepare Phase = iota
    PhasePreCommit
    PhaseCommit
    PhaseDecide
)

// ViewChangeMsg represents a view change message
type ViewChangeMsg struct {
    NewView       uint64
    HighestQC     *QuorumCertificate
    ValidatorIdx  uint64
    Signature     bls.Signature
}
```

### 5.2 HotStuff Service

```go
// beacon-chain/consensus/hotstuff/service.go

package hotstuff

import (
    "context"
    "sync"
    "time"
)

type Service struct {
    // Configuration
    cfg *Config
    
    // State
    currentView   uint64
    currentPhase  Phase
    currentLeader uint64
    
    // Block tree
    blocks        map[[32]byte]*HotStuffBlock
    highestQC     *QuorumCertificate
    lockedQC      *QuorumCertificate
    
    // Voting
    votes         map[uint64]map[[32]byte][]*Vote  // view -> blockHash -> votes
    
    // Validators
    validators    []uint64
    validatorIdx  uint64
    
    // Channels
    blockChan     chan *HotStuffBlock
    voteChan      chan *Vote
    viewChangeCh  chan *ViewChangeMsg
    
    // Timers
    viewTimer     *time.Timer
    
    // Synchronization
    mu            sync.RWMutex
    
    // Execution
    execEngine    ExecutionEngine
}

func NewService(cfg *Config) *Service {
    return &Service{
        cfg:          cfg,
        currentView:  0,
        currentPhase: PhasePrepare,
        blocks:       make(map[[32]byte]*HotStuffBlock),
        votes:        make(map[uint64]map[[32]byte][]*Vote),
        blockChan:    make(chan *HotStuffBlock, 100),
        voteChan:     make(chan *Vote, 1000),
        viewChangeCh: make(chan *ViewChangeMsg, 100),
    }
}

func (s *Service) Start() error {
    go s.run()
    return nil
}

func (s *Service) run() {
    s.startViewTimer()
    
    for {
        select {
        case block := <-s.blockChan:
            s.handleBlock(block)
        case vote := <-s.voteChan:
            s.handleVote(vote)
        case msg := <-s.viewChangeCh:
            s.handleViewChange(msg)
        case <-s.viewTimer.C:
            s.handleTimeout()
        }
    }
}
```

### 5.3 Three-Phase Protocol

```go
// beacon-chain/consensus/hotstuff/phases.go

package hotstuff

func (s *Service) handleBlock(block *HotStuffBlock) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Verify block
    if err := s.verifyBlock(block); err != nil {
        return err
    }
    
    // Store block
    blockHash := hashBlock(block)
    s.blocks[blockHash] = block
    
    // Process based on current phase
    switch s.currentPhase {
    case PhasePrepare:
        return s.handlePrepare(block)
    case PhasePreCommit:
        return s.handlePreCommit(block)
    case PhaseCommit:
        return s.handleCommit(block)
    case PhaseDecide:
        return s.handleDecide(block)
    }
    
    return nil
}

func (s *Service) handlePrepare(block *HotStuffBlock) error {
    // Create PREPARE vote
    vote := s.createVote(block, PhasePrepare)
    
    // Send vote to leader
    s.sendVote(vote)
    
    // If we're the leader, collect votes
    if s.isLeader() {
        s.collectVotes(block, PhasePrepare)
    }
    
    return nil
}

func (s *Service) collectVotes(block *HotStuffBlock, phase Phase) {
    blockHash := hashBlock(block)
    votes := s.votes[s.currentView][blockHash]
    
    // Check if we have 2f+1 votes
    if len(votes) >= s.quorumSize() {
        // Create QC
        qc := s.createQC(votes)
        
        // Move to next phase
        s.advancePhase(qc)
    }
}

func (s *Service) advancePhase(qc *QuorumCertificate) {
    switch s.currentPhase {
    case PhasePrepare:
        s.currentPhase = PhasePreCommit
        s.broadcastPreCommit(qc)
    case PhasePreCommit:
        s.currentPhase = PhaseCommit
        s.lockedQC = qc  // Lock on this QC
        s.broadcastCommit(qc)
    case PhaseCommit:
        s.currentPhase = PhaseDecide
        s.broadcastDecide(qc)
        s.executeBlock(qc.BlockHash)
        s.advanceView()
    }
}
```

## Phase 6: Testing Strategy

### 6.1 Unit Tests
- QC creation and verification
- Leader election
- Vote collection
- Phase transitions
- View changes

### 6.2 Integration Tests
- Full three-phase protocol
- View change scenarios
- Byzantine fault scenarios
- Network partition recovery

### 6.3 Performance Tests
- Throughput (blocks/second)
- Latency (time to finality)
- Scalability (validators)
- Resource usage

### 6.4 Comparison Tests
- HotStuff vs PoS performance
- Finality time comparison
- Resource usage comparison

## Phase 7: Timeline

| Week | Phase | Deliverables |
|------|-------|--------------|
| 1 | Abstraction | Consensus interface, PoS wrapper |
| 2 | Types | HotStuff data structures |
| 3 | Leader | Leader election implementation |
| 4 | Protocol | Three-phase protocol |
| 5 | View Change | View change mechanism |
| 6 | Safety | Safety rules, integration |
| 7 | Testing | Comprehensive tests |
| 8 | Documentation | Complete documentation |

## Phase 8: Success Criteria

- ✅ Both PoS and HotStuff modes work
- ✅ Configuration switching works seamlessly
- ✅ HotStuff achieves Byzantine fault tolerance
- ✅ Performance is acceptable (>10 blocks/sec)
- ✅ All tests pass
- ✅ Documentation is complete

## References

- **HotStuff Paper**: https://arxiv.org/abs/1803.05069
- **Prysm Documentation**: https://docs.prylabs.network
- **Ethereum Consensus Specs**: https://github.com/ethereum/consensus-specs
- **BFT Consensus**: https://pmg.csail.mit.edu/papers/osdi99.pdf

## Next Steps

1. Study current Prysm PoS implementation
2. Create consensus interface
3. Begin HotStuff implementation
4. Iterate and test

---

**Status**: Planning Phase
**Last Updated**: 2025-10-14

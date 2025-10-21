# Design Document: HotStuff 2-Phase Optimization

## Overview

This design document describes the implementation approach for optimizing the HotStuff consensus protocol from 4 phases to 2 phases. The optimization combines PREPARE+PRE-COMMIT into PROPOSE and COMMIT+DECIDE into COMMIT, reducing consensus latency by approximately 50% and enabling 3-second block times.

## Architecture

### Current 4-Phase Architecture

```
Block Proposal → PREPARE → PRE-COMMIT → COMMIT → DECIDE → Execute
     (Leader)      (Vote)     (Vote)      (Vote)   (Execute)
                    ↓ QC       ↓ QC        ↓ QC
                  ~500ms     ~500ms      ~500ms    ~200ms
                  
Total: ~1700ms minimum
```

### New 2-Phase Architecture

```
Block Proposal → PROPOSE → COMMIT → Execute
     (Leader)     (Vote)    (Vote+Execute)
                   ↓ QC      ↓ QC
                 ~700ms    ~700ms
                 
Total: ~1400ms minimum (50% faster)
```

### Key Differences

1. **PROPOSE Phase** combines:
   - Block verification (from PREPARE)
   - Safety rule checking (from PRE-COMMIT)
   - Immediate voting

2. **COMMIT Phase** combines:
   - Final commitment (from COMMIT)
   - Block execution (from DECIDE)
   - View advancement

## Components and Interfaces

### 1. Phase Enumeration (`types.go`)

**Current:**
```go
type Phase string

const (
    PhasePrepare    Phase = "PREPARE"
    PhasePreCommit  Phase = "PRE-COMMIT"
    PhaseCommit     Phase = "COMMIT"
    PhaseDecide     Phase = "DECIDE"
)
```

**New:**
```go
type Phase string

const (
    PhasePropose Phase = "PROPOSE"
    PhaseCommit  Phase = "COMMIT"
)
```

### 2. Block Node Structure (`types.go`)

**Current:**
```go
type BlockNode struct {
    Block        *HotStuffBlock
    PrepareQC    *QuorumCertificate
    PreCommitQC  *QuorumCertificate
    CommitQC     *QuorumCertificate
    Status       BlockStatus
}
```

**New:**
```go
type BlockNode struct {
    Block      *HotStuffBlock
    ProposeQC  *QuorumCertificate
    CommitQC   *QuorumCertificate
    Status     BlockStatus
}
```

### 3. Block Status Enumeration (`types.go`)

**Current:**
```go
const (
    StatusProposed      BlockStatus = "PROPOSED"
    StatusPrepared      BlockStatus = "PREPARED"
    StatusPreCommitted  BlockStatus = "PRE-COMMITTED"
    StatusCommitted     BlockStatus = "COMMITTED"
    StatusDecided       BlockStatus = "DECIDED"
)
```

**New:**
```go
const (
    StatusProposed   BlockStatus = "PROPOSED"
    StatusCommitted  BlockStatus = "COMMITTED"
    StatusExecuted   BlockStatus = "EXECUTED"
)
```

### 4. Phase Handler Interface (`phases.go`)

**New Structure:**
```go
// handleBlock routes to appropriate phase handler
func (s *Service) handleBlock(block *HotStuffBlock) error

// handlePropose combines PREPARE + PRE-COMMIT logic
func (s *Service) handlePropose(block *HotStuffBlock, blockHash [32]byte) error

// handleCommit combines COMMIT + DECIDE logic
func (s *Service) handleCommit(block *HotStuffBlock, blockHash [32]byte) error
```

## Data Models

### Vote Structure (Unchanged)

```go
type Vote struct {
    View           uint64
    Phase          Phase  // Now only PhasePropose or PhaseCommit
    BlockHash      [32]byte
    ValidatorIndex primitives.ValidatorIndex
    Signature      bls.Signature
}
```

### Quorum Certificate Structure (Unchanged)

```go
type QuorumCertificate struct {
    View          uint64
    Phase         Phase  // Now only PhasePropose or PhaseCommit
    BlockHash     [32]byte
    Signatures    []bls.Signature
    SignerIndices []uint64
}
```

## Detailed Component Design

### 1. handlePropose Phase

**Purpose:** Verify block and collect votes for proposal

**Logic Flow:**
```
1. Verify block extends from highest QC (PREPARE safety rule)
2. Verify block extends from locked QC or higher (PRE-COMMIT safety rule)
3. Verify block validity (signatures, state transition)
4. Create PROPOSE vote
5. If leader: collect vote
6. If validator: send vote to leader
7. On quorum: build ProposeQC and advance to COMMIT
```

**Safety Rules:**
- Block must extend from highest known QC
- Block must extend from locked QC or have higher QC
- Block must be valid (signatures, state)

**Code Structure:**
```go
func (s *Service) handlePropose(block *HotStuffBlock, blockHash [32]byte) error {
    // Combined safety checks from PREPARE + PRE-COMMIT
    if !s.shouldVotePropose(block) {
        return nil
    }
    
    // Create and send vote
    vote, err := s.createVote(block, blockHash, PhasePropose)
    if err != nil {
        return err
    }
    
    if s.isLeader() {
        return s.collectVote(vote)
    }
    return s.sendVoteToLeader(vote)
}
```

### 2. handleCommit Phase

**Purpose:** Finalize block and execute

**Logic Flow:**
```
1. Verify ProposeQC exists
2. Update locked QC (COMMIT safety rule)
3. Create COMMIT vote
4. If leader: collect vote
5. If validator: send vote to leader
6. On quorum: build CommitQC, execute block, advance view
```

**Safety Rules:**
- Must have valid ProposeQC
- Update locked QC to prevent rollback
- Execute block atomically with commit

**Code Structure:**
```go
func (s *Service) handleCommit(block *HotStuffBlock, blockHash [32]byte) error {
    // Verify ProposeQC exists
    node := s.blocks[blockHash]
    if node.ProposeQC == nil {
        return errors.New("missing ProposeQC")
    }
    
    // Update locked QC
    s.lockedQC = node.ProposeQC
    
    // Create and send vote
    vote, err := s.createVote(block, blockHash, PhaseCommit)
    if err != nil {
        return err
    }
    
    if s.isLeader() {
        return s.collectVote(vote)
    }
    return s.sendVoteToLeader(vote)
}
```

### 3. onQuorumReached Updates

**Purpose:** Handle quorum for each phase

**Logic Flow:**
```
PROPOSE quorum reached:
  1. Build ProposeQC
  2. Update block status to StatusCommitted
  3. Advance to COMMIT phase
  4. Broadcast phase message

COMMIT quorum reached:
  1. Build CommitQC
  2. Execute block immediately
  3. Update block status to StatusExecuted
  4. Advance to next view
  5. If new leader: propose next block
```

**Code Structure:**
```go
func (s *Service) onQuorumReached(view uint64, phase Phase, blockHash [32]byte, builder *QCBuilder) error {
    qc, err := builder.Build()
    if err != nil {
        return err
    }
    
    node := s.blocks[blockHash]
    
    switch phase {
    case PhasePropose:
        node.ProposeQC = qc
        node.Status = StatusCommitted
        return s.advancePhase(PhaseCommit, qc)
        
    case PhaseCommit:
        node.CommitQC = qc
        // Execute immediately
        if err := s.executeBlock(node.Block, blockHash); err != nil {
            return err
        }
        node.Status = StatusExecuted
        return s.advanceView()
        
    default:
        return fmt.Errorf("unexpected phase: %s", phase)
    }
}
```

### 4. Safety Rule Consolidation

**shouldVotePropose:** Combines PREPARE + PRE-COMMIT safety checks

```go
func (s *Service) shouldVotePropose(block *HotStuffBlock) bool {
    // PREPARE rule: extends from highest QC
    if block.JustifyQC == nil {
        return false
    }
    if CompareQC(block.JustifyQC, s.highestQC) < 0 {
        return false
    }
    
    // PRE-COMMIT rule: extends from locked QC or higher
    if s.lockedQC != nil {
        if CompareQC(block.JustifyQC, s.lockedQC) < 0 {
            return false
        }
    }
    
    return true
}
```

**shouldVoteCommit:** Simplified commit check

```go
func (s *Service) shouldVoteCommit(block *HotStuffBlock, node *BlockNode) bool {
    // Must have ProposeQC
    if node.ProposeQC == nil {
        return false
    }
    
    // Verify ProposeQC is valid
    if err := VerifyQC(node.ProposeQC, s.publicKeys, s.quorumSize); err != nil {
        return false
    }
    
    return true
}
```

## Error Handling

### Phase Transition Errors

1. **Missing QC**: If ProposeQC is missing when entering COMMIT phase
   - Log error with block hash and view
   - Do not vote
   - Wait for QC or timeout

2. **Invalid QC**: If QC verification fails
   - Log error with QC details
   - Reject block
   - Trigger view change if enabled

3. **Execution Failure**: If block execution fails in COMMIT phase
   - Log error with execution details
   - Mark block as failed
   - Trigger view change
   - Do not advance view

### Vote Collection Errors

1. **Duplicate Vote**: Validator votes twice for same view/phase
   - Ignore duplicate
   - Log warning
   - Continue processing

2. **Invalid Signature**: Vote signature verification fails
   - Reject vote
   - Log error with validator index
   - Continue processing

3. **Quorum Timeout**: Quorum not reached within view timeout
   - Trigger view change
   - Rotate leader
   - Start new view

## Testing Strategy

### Unit Tests

1. **Phase Handler Tests**
   - Test handlePropose with valid/invalid blocks
   - Test handleCommit with valid/invalid ProposeQC
   - Test safety rule enforcement
   - Test vote creation and collection

2. **QC Builder Tests**
   - Test quorum detection for 2 phases
   - Test QC building with correct signatures
   - Test duplicate vote handling

3. **State Transition Tests**
   - Test phase advancement (Propose → Commit)
   - Test view advancement (Commit → next Propose)
   - Test status updates

### Integration Tests

1. **Single Validator Tests**
   - Test 2-phase consensus with 1 validator
   - Verify block execution
   - Verify view advancement

2. **Multi-Validator Tests**
   - Test 2-phase consensus with 4 validators
   - Test vote collection and QC building
   - Test leader rotation

3. **Byzantine Fault Tests**
   - Test with 1 Byzantine validator (f=1, n=4)
   - Verify safety with conflicting votes
   - Verify liveness with non-voting validators

### Performance Tests

1. **Latency Tests**
   - Measure time for each phase
   - Verify <1.5s per phase
   - Verify total consensus <3s

2. **Throughput Tests**
   - Test with 3-second blocks
   - Verify 400+ TPS with 500 tx/block
   - Verify stability under sustained load

3. **Stress Tests**
   - Test with 1-second blocks (should fail gracefully)
   - Test with 2-second blocks (stretch goal)
   - Test with high transaction volume

## Migration Strategy

### Phase 1: Code Changes (No Deployment)
1. Update type definitions
2. Implement new phase handlers
3. Update QC builder logic
4. Add comprehensive tests

### Phase 2: Testing (Testnet)
1. Deploy to isolated testnet
2. Run performance benchmarks
3. Verify safety and liveness
4. Collect metrics

### Phase 3: Configuration Update
1. Update block time to 3 seconds
2. Monitor consensus performance
3. Verify TPS improvements
4. Collect production metrics

### Phase 4: Documentation
1. Update README files
2. Update configuration guides
3. Document performance characteristics
4. Create troubleshooting guide

## Performance Expectations

### Before Optimization (4-Phase, 6s blocks)
- Consensus time: ~1.7-2.0s
- Block time: 6s
- TPS: ~230 (with current tx/block)
- Phase overhead: 4 phases × 500ms = 2000ms

### After Optimization (2-Phase, 3s blocks)
- Consensus time: ~1.0-1.4s
- Block time: 3s
- TPS: ~400-500 (with 500 tx/block)
- Phase overhead: 2 phases × 700ms = 1400ms

### Improvement Metrics
- **Consensus latency**: 30-40% reduction
- **Block time**: 50% reduction (6s → 3s)
- **TPS**: 75-120% increase (230 → 400-500)
- **Phase count**: 50% reduction (4 → 2)

## Risks and Mitigation

### Risk 1: Safety Violation
**Risk**: Combining phases might introduce safety bugs
**Mitigation**: 
- Comprehensive safety rule testing
- Formal verification of safety properties
- Extensive Byzantine fault testing

### Risk 2: Performance Regression
**Risk**: 2-phase might not be faster in practice
**Mitigation**:
- Benchmark before and after
- Monitor phase completion times
- Have rollback plan

### Risk 3: Network Instability
**Risk**: Faster blocks might cause network congestion
**Mitigation**:
- Gradual rollout (6s → 4s → 3s)
- Monitor network metrics
- Adjust block time if needed

### Risk 4: Execution Layer Bottleneck
**Risk**: Erigon might not keep up with 3s blocks
**Mitigation**:
- Test execution layer performance
- Optimize block building pipeline
- Consider parallel execution

## Future Optimizations

1. **Pipelined Consensus**: Start building block N+1 while finalizing block N
2. **Parallel Validation**: Validate multiple blocks concurrently
3. **Optimistic Execution**: Execute speculatively before commit
4. **Batch Verification**: Verify multiple signatures in parallel
5. **1-Phase Consensus**: Research single-phase consensus for even faster blocks

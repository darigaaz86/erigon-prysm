# Requirements Document

## Introduction

This specification defines the optimization of the HotStuff consensus protocol from a 4-phase to a 2-phase implementation. The current implementation uses PREPARE, PRE-COMMIT, COMMIT, and DECIDE phases, which creates significant latency (800-2000ms per block). By combining phases, we can reduce consensus overhead by 50% and enable faster block times (3 seconds instead of 6 seconds), ultimately achieving 500+ TPS.

## Glossary

- **HotStuff Service**: The consensus service implementing the HotStuff BFT protocol in `prysm/beacon-chain/consensus/hotstuff/`
- **Phase**: A stage in the consensus protocol where validators vote and build quorum certificates
- **QC (Quorum Certificate)**: A certificate proving that a quorum of validators voted for a block
- **Fast-HotStuff**: An optimized 2-phase variant of HotStuff that combines phases for faster finality
- **Block Node**: A data structure storing a block and its associated QCs and status
- **View**: A numbered consensus round in which a leader proposes a block

## Requirements

### Requirement 1: Reduce Consensus Phases from 4 to 2

**User Story:** As a blockchain operator, I want the consensus protocol to use only 2 phases instead of 4, so that blocks can be finalized faster and achieve higher throughput.

#### Acceptance Criteria

1. WHEN the HotStuff Service processes a block, THE HotStuff Service SHALL use exactly two phases: PROPOSE and COMMIT
2. WHEN a validator receives a block proposal, THE HotStuff Service SHALL immediately vote in the PROPOSE phase without requiring a PREPARE phase
3. WHEN a PROPOSE phase reaches quorum, THE HotStuff Service SHALL advance directly to the COMMIT phase
4. WHEN a COMMIT phase reaches quorum, THE HotStuff Service SHALL immediately execute the block without requiring a DECIDE phase
5. THE HotStuff Service SHALL maintain the same safety guarantees as the 4-phase implementation

### Requirement 2: Update Phase Enumeration and Types

**User Story:** As a developer, I want the phase types to reflect the 2-phase model, so that the code is clear and maintainable.

#### Acceptance Criteria

1. THE HotStuff Service SHALL define exactly two phase constants: PhasePropose and PhaseCommit
2. THE HotStuff Service SHALL remove the PhasePrepare, PhasePreCommit, and PhaseDecide constants
3. WHEN storing block state, THE Block Node SHALL track ProposeQC and CommitQC instead of PrepareQC, PreCommitQC, and CommitQC
4. THE HotStuff Service SHALL update all phase-related logging to use the new phase names
5. THE HotStuff Service SHALL update status enumerations to reflect the 2-phase model (StatusProposed, StatusCommitted, StatusExecuted)

### Requirement 3: Combine PREPARE and PRE-COMMIT Logic

**User Story:** As a validator, I want to vote immediately when I receive a valid block proposal, so that consensus can proceed faster.

#### Acceptance Criteria

1. WHEN a validator receives a block in the PROPOSE phase, THE HotStuff Service SHALL verify the block and vote immediately if valid
2. THE HotStuff Service SHALL combine the safety checks from both PREPARE and PRE-COMMIT phases into the PROPOSE phase
3. WHEN creating a PROPOSE vote, THE HotStuff Service SHALL include all necessary information for safety verification
4. THE HotStuff Service SHALL verify that the block extends from the highest known QC
5. THE HotStuff Service SHALL verify that the block extends from the locked QC or has a higher QC

### Requirement 4: Combine COMMIT and DECIDE Logic

**User Story:** As a validator, I want blocks to be executed immediately after commit quorum is reached, so that finality is achieved faster.

#### Acceptance Criteria

1. WHEN the COMMIT phase reaches quorum, THE HotStuff Service SHALL immediately execute the block
2. THE HotStuff Service SHALL update the locked QC when processing the COMMIT phase
3. WHEN executing a block, THE HotStuff Service SHALL update the block status to StatusExecuted
4. THE HotStuff Service SHALL advance to the next view immediately after block execution
5. THE HotStuff Service SHALL maintain the same execution guarantees as the 4-phase implementation

### Requirement 5: Update QC Builder and Vote Collection

**User Story:** As a leader validator, I want to collect votes and build QCs for only 2 phases, so that the consensus process is streamlined.

#### Acceptance Criteria

1. THE HotStuff Service SHALL create QC builders for only PhasePropose and PhaseCommit
2. WHEN collecting votes, THE HotStuff Service SHALL route votes to the appropriate phase builder (Propose or Commit)
3. WHEN a PROPOSE quorum is reached, THE HotStuff Service SHALL build a ProposeQC and advance to COMMIT phase
4. WHEN a COMMIT quorum is reached, THE HotStuff Service SHALL build a CommitQC and execute the block
5. THE HotStuff Service SHALL clean up QC builders for completed views to prevent memory leaks

### Requirement 6: Update Block Proposal Logic

**User Story:** As a leader validator, I want to propose blocks with the new 2-phase structure, so that validators can process them correctly.

#### Acceptance Criteria

1. WHEN proposing a block, THE HotStuff Service SHALL set the initial phase to PhasePropose
2. THE HotStuff Service SHALL include the highest known QC as the JustifyQC in the proposal
3. WHEN broadcasting a block, THE HotStuff Service SHALL ensure it contains all necessary information for 2-phase consensus
4. THE HotStuff Service SHALL log block proposals with the new phase terminology
5. THE HotStuff Service SHALL maintain backward compatibility with the execution layer integration

### Requirement 7: Update Phase Advancement Logic

**User Story:** As a validator, I want the system to advance through phases correctly in the 2-phase model, so that consensus proceeds smoothly.

#### Acceptance Criteria

1. WHEN advancing from PROPOSE to COMMIT, THE HotStuff Service SHALL update the current phase to PhaseCommit
2. WHEN advancing from COMMIT to next view, THE HotStuff Service SHALL reset the phase to PhasePropose
3. THE HotStuff Service SHALL reset the view timer when advancing phases
4. THE HotStuff Service SHALL update the highest QC when advancing phases
5. THE HotStuff Service SHALL broadcast phase messages to all validators when advancing phases

### Requirement 8: Maintain Safety and Liveness Properties

**User Story:** As a blockchain operator, I want the 2-phase consensus to maintain the same safety and liveness guarantees as the original 4-phase implementation, so that the blockchain remains secure and reliable.

#### Acceptance Criteria

1. THE HotStuff Service SHALL ensure that no two conflicting blocks can be committed
2. THE HotStuff Service SHALL ensure that the locked QC mechanism prevents safety violations
3. THE HotStuff Service SHALL ensure that validators can make progress even with up to f Byzantine validators (where 3f+1 is the total)
4. THE HotStuff Service SHALL ensure that view changes can occur if the leader fails
5. THE HotStuff Service SHALL maintain the same quorum threshold (2f+1) for both phases

### Requirement 9: Update Configuration and Documentation

**User Story:** As a system administrator, I want the configuration to reflect the 2-phase optimization, so that I can tune the system appropriately.

#### Acceptance Criteria

1. THE HotStuff Service SHALL update configuration comments to reflect 2-phase consensus
2. THE HotStuff Service SHALL update logging messages to use 2-phase terminology
3. THE HotStuff Service SHALL provide clear documentation of the phase reduction
4. THE HotStuff Service SHALL update any README files to explain the 2-phase model
5. THE HotStuff Service SHALL maintain configuration compatibility with existing deployments

### Requirement 10: Enable Faster Block Times

**User Story:** As a blockchain operator, I want to reduce block time to 3 seconds after the 2-phase optimization, so that I can achieve higher transaction throughput.

#### Acceptance Criteria

1. WHEN the 2-phase optimization is complete, THE System SHALL support 3-second block times without consensus failures
2. THE System SHALL achieve at least 400 TPS with 3-second blocks and 500 transactions per block
3. THE System SHALL maintain stability under sustained load with 3-second blocks
4. THE System SHALL complete both consensus phases within 2.5 seconds, leaving 0.5 seconds buffer
5. THE System SHALL log performance metrics showing phase completion times

## Non-Functional Requirements

### Performance
- Each phase should complete in less than 1.5 seconds under normal conditions
- The 2-phase consensus should reduce total consensus time by at least 40% compared to 4-phase
- Memory usage should not increase due to the optimization

### Reliability
- The system should maintain 99.9% uptime during the transition
- No data loss should occur during the phase reduction
- The system should gracefully handle edge cases and Byzantine behavior

### Maintainability
- Code should be well-documented with clear comments
- Phase logic should be modular and easy to test
- Changes should not create technical debt

## Out of Scope

- Changes to the execution layer (Erigon)
- Changes to the P2P networking layer
- Changes to the validator client
- Implementation of sharding or parallel chains
- Changes to the cryptographic primitives (BLS signatures)

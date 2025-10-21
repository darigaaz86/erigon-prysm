# Implementation Plan

- [x] 1. Update type definitions and constants
  - [x] 1.1 Update Phase enumeration to use PhasePropose and PhaseCommit only
    - Modify `prysm/beacon-chain/consensus/hotstuff/types.go`
    - Remove PhasePrepare, PhasePreCommit, PhaseDecide constants
    - Add PhasePropose and PhaseCommit constants
    - _Requirements: 1.1, 2.1, 2.2_
  
  - [x] 1.2 Update BlockNode structure to use 2-phase QCs
    - Modify `BlockNode` struct in `types.go`
    - Replace PrepareQC, PreCommitQC with ProposeQC
    - Keep CommitQC
    - _Requirements: 2.3_
  
  - [x] 1.3 Update BlockStatus enumeration
    - Modify `BlockStatus` constants in `types.go`
    - Remove StatusPrepared, StatusPreCommitted, StatusDecided
    - Keep StatusProposed, StatusCommitted
    - Add StatusExecuted
    - _Requirements: 2.5_

- [x] 2. Implement handlePropose phase logic
  - [x] 2.1 Create shouldVotePropose safety check function
    - Implement combined PREPARE + PRE-COMMIT safety rules
    - Check block extends from highest QC
    - Check block extends from locked QC or higher
    - Add comprehensive logging
    - _Requirements: 3.2, 3.4, 3.5_
  
  - [x] 2.2 Implement handlePropose function
    - Create new `handlePropose` function in `phases.go`
    - Call `shouldVotePropose` for safety verification
    - Create PROPOSE vote
    - Route vote to leader or collect if leader
    - _Requirements: 3.1, 3.3_
  
  - [x] 2.3 Update handleBlock routing logic
    - Modify `handleBlock` function in `phases.go`
    - Remove cases for PhasePrepare, PhasePreCommit, PhaseDecide
    - Add case for PhasePropose
    - Keep case for PhaseCommit
    - _Requirements: 1.1_

- [x] 3. Implement handleCommit phase logic
  - [x] 3.1 Create shouldVoteCommit safety check function
    - Verify ProposeQC exists
    - Verify ProposeQC is valid
    - Add comprehensive logging
    - _Requirements: 4.1, 4.2_
  
  - [x] 3.2 Implement handleCommit function
    - Create new `handleCommit` function in `phases.go`
    - Verify ProposeQC exists
    - Update locked QC
    - Create COMMIT vote
    - Route vote to leader or collect if leader
    - _Requirements: 4.2, 4.3_
  
  - [x] 3.3 Integrate block execution into COMMIT phase
    - Call `executeBlock` when COMMIT quorum reached
    - Update block status to StatusExecuted
    - Advance to next view after execution
    - _Requirements: 4.1, 4.4, 4.5_

- [x] 4. Update QC builder and vote collection logic
  - [x] 4.1 Update onQuorumReached for 2 phases
    - Modify `onQuorumReached` function in `phases.go`
    - Handle PhasePropose: build ProposeQC, advance to COMMIT
    - Handle PhaseCommit: build CommitQC, execute block, advance view
    - Remove cases for PhasePrepare, PhasePreCommit
    - _Requirements: 5.3, 5.4_
  
  - [x] 4.2 Update QC builder initialization
    - Modify `collectVote` function in `phases.go`
    - Create builders only for PhasePropose and PhaseCommit
    - Update builder cleanup logic
    - _Requirements: 5.1, 5.2, 5.5_

- [x] 5. Update phase advancement logic
  - [x] 5.1 Update advancePhase function
    - Modify `advancePhase` function in `phases.go`
    - Handle only PhasePropose → PhaseCommit transition
    - Update highest QC
    - Reset view timer
    - Broadcast phase message
    - _Requirements: 7.1, 7.3, 7.4, 7.5_
  
  - [x] 5.2 Update advanceView function
    - Modify `advanceView` function in `phases.go`
    - Reset phase to PhasePropose
    - Increment view counter
    - Reset view timer
    - Propose block if new leader
    - _Requirements: 7.2_

- [x] 6. Update block proposal logic
  - [x] 6.1 Update proposeBlock function
    - Modify `proposeBlock` function in `phases.go`
    - Set initial phase to PhasePropose
    - Include highest QC as JustifyQC
    - Update logging to use new phase names
    - _Requirements: 6.1, 6.2, 6.4_
  
  - [x] 6.2 Verify block proposal structure
    - Ensure block contains all necessary fields for 2-phase consensus
    - Verify JustifyQC is properly set
    - Test with execution layer integration
    - _Requirements: 6.3, 6.5_

- [x] 7. Update logging and error messages
  - [x] 7.1 Update all phase-related log messages
    - Search for all log statements mentioning phases
    - Update to use PhasePropose and PhaseCommit terminology
    - Ensure consistent logging format
    - _Requirements: 2.4, 9.2_
  
  - [x] 7.2 Update error messages
    - Update error messages to reflect 2-phase model
    - Add clear error messages for missing ProposeQC
    - Add clear error messages for phase transition failures
    - _Requirements: 9.2_

- [x] 8. Update configuration and documentation
  - [x] 8.1 Update configuration comments
    - Modify `service.go` Config struct comments
    - Update to reflect 2-phase consensus
    - Document expected phase timings
    - _Requirements: 9.1_
  
  - [x] 8.2 Update README documentation
    - Update `prysm/beacon-chain/consensus/hotstuff/README.md`
    - Document 2-phase consensus model
    - Explain phase reduction benefits
    - Add performance expectations
    - _Requirements: 9.3, 9.4_
  
  - [x] 8.3 Create migration guide
    - Document changes from 4-phase to 2-phase
    - Explain configuration updates needed
    - Provide troubleshooting tips
    - _Requirements: 9.4_

- [x] 9. Write unit tests for phase handlers
  - [x] 9.1 Test handlePropose function
    - Test with valid block extending from highest QC
    - Test with invalid block (doesn't extend from highest QC)
    - Test with block not extending from locked QC
    - Test vote creation and collection
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_
  
  - [x] 9.2 Test handleCommit function
    - Test with valid ProposeQC
    - Test with missing ProposeQC
    - Test with invalid ProposeQC
    - Test locked QC update
    - Test block execution trigger
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_
  
  - [x] 9.3 Test shouldVotePropose safety rules
    - Test highest QC check
    - Test locked QC check
    - Test with various QC combinations
    - Verify safety guarantees
    - _Requirements: 8.1, 8.2_
  
  - [x] 9.4 Test onQuorumReached for both phases
    - Test PROPOSE quorum handling
    - Test COMMIT quorum handling
    - Test QC building
    - Test phase advancement
    - Test view advancement
    - _Requirements: 5.3, 5.4_

- [x] 10. Write integration tests
  - [x] 10.1 Test single validator 2-phase consensus
    - Set up single validator environment
    - Propose and finalize blocks
    - Verify both phases complete
    - Verify block execution
    - _Requirements: 1.1, 8.3_
  
  - [x] 10.2 Test multi-validator 2-phase consensus
    - Set up 4-validator environment
    - Test vote collection across validators
    - Test QC building with multiple signatures
    - Test leader rotation
    - _Requirements: 5.1, 5.2, 8.3_
  
  - [x] 10.3 Test Byzantine fault tolerance
    - Set up 4-validator environment with 1 Byzantine
    - Test with non-voting validator
    - Test with conflicting votes
    - Verify safety maintained
    - Verify liveness maintained
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [x] 11. Write performance tests
  - [x] 11.1 Measure phase completion times
    - Add timing instrumentation to each phase
    - Run consensus with various loads
    - Verify each phase completes in <1.5s
    - Log performance metrics
    - _Requirements: 10.4_
  
  - [x] 11.2 Test with 3-second blocks
    - Update configuration to 3-second blocks
    - Run sustained load test
    - Verify consensus completes within block time
    - Measure TPS achieved
    - _Requirements: 10.1, 10.2, 10.3_
  
  - [x] 11.3 Stress test with various block times
    - Test with 2-second blocks (stretch goal)
    - Test with 1-second blocks (should fail gracefully)
    - Test with 4-second blocks (should work easily)
    - Document minimum viable block time
    - _Requirements: 10.1_

- [x] 12. Update block time configuration
  - [x] 12.1 Update consensus configuration files
    - Modify `config/consensus-hotstuff-dev.yml`
    - Change block_time from 6s to 3s
    - Update comments to reflect 2-phase optimization
    - _Requirements: 10.1_
  
  - [x] 12.2 Update beacon chain configuration
    - Modify `config/config-optimized.yml`
    - Change SECONDS_PER_SLOT from 6 to 3
    - Verify other timing parameters are compatible
    - _Requirements: 10.1_
  
  - [x] 12.3 Test configuration changes
    - Deploy with new configuration
    - Verify blocks are produced every 3 seconds
    - Monitor for consensus failures
    - Collect performance metrics
    - _Requirements: 10.1, 10.2, 10.3, 10.5_

- [x] 13. End-to-end validation
  - [x] 13.1 Deploy to clean testnet
    - Stop existing services
    - Clean data directories
    - Deploy with 2-phase code and 3s blocks
    - Verify startup successful
    - _Requirements: 10.1_
  
  - [x] 13.2 Run sustained load test
    - Submit transactions at 400+ TPS
    - Run for 10+ minutes
    - Verify all transactions processed
    - Verify no consensus failures
    - _Requirements: 10.2, 10.3_
  
  - [x] 13.3 Verify transaction queryability
    - Query random transactions by hash
    - Verify all transactions are queryable
    - Verify receipts are available
    - Check RPC performance
    - _Requirements: 10.2_
  
  - [x] 13.4 Collect and analyze metrics
    - Measure actual TPS achieved
    - Measure phase completion times
    - Measure block finalization time
    - Compare with 4-phase baseline
    - Document performance improvements
    - _Requirements: 10.2, 10.4, 10.5_

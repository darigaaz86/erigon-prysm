#!/bin/bash

# Script to capture Erigon and Beacon logs during a test
# Usage: ./capture-logs-during-test.sh <test_name>

TEST_NAME=${1:-"test"}
LOG_DIR="logs/${TEST_NAME}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "=== Log Capture Script ==="
echo "Test Name: $TEST_NAME"
echo "Log Directory: $LOG_DIR"
echo "Timestamp: $TIMESTAMP"
echo ""

# Create log directory
mkdir -p "$LOG_DIR"

# Find Erigon process
ERIGON_PID=$(pgrep -f "erigon.*--datadir" | head -1)
if [ -n "$ERIGON_PID" ]; then
  echo "Found Erigon process: PID $ERIGON_PID"
  ERIGON_LOG="${LOG_DIR}/erigon_${TIMESTAMP}.log"
  echo "Erigon log will be saved to: $ERIGON_LOG"
else
  echo "⚠️  Erigon process not found"
fi

# Find Beacon process
BEACON_PID=$(pgrep -f "beacon-chain" | head -1)
if [ -n "$BEACON_PID" ]; then
  echo "Found Beacon process: PID $BEACON_PID"
  BEACON_LOG="${LOG_DIR}/beacon_${TIMESTAMP}.log"
  echo "Beacon log will be saved to: $BEACON_LOG"
else
  echo "⚠️  Beacon process not found"
fi

echo ""
echo "Note: This script identifies the processes but cannot capture their existing logs."
echo "To capture logs for future tests:"
echo "1. Restart Erigon with: erigon ... 2>&1 | tee erigon.log"
echo "2. Restart Beacon with: beacon-chain ... 2>&1 | tee beacon.log"
echo ""
echo "Or check if logs are being written to:"
echo "- Erigon data directory"
echo "- Beacon data directory"
echo "- System logs (journalctl on Linux, Console.app on macOS)"

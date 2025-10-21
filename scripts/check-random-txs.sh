#!/bin/bash

# Check random transactions from the hash file
HASH_FILE="transaction_hashes_5tps_5min.txt"
NUM_TO_CHECK=${1:-10}

echo "=== CHECKING $NUM_TO_CHECK RANDOM TRANSACTIONS ==="
echo ""

# Extract only transaction hashes (lines starting with 0x)
grep "^0x" "$HASH_FILE" | shuf -n "$NUM_TO_CHECK" | while read -r hash; do
  echo "Checking: $hash"
  
  result=$(curl -s http://localhost:8545 -X POST -H "Content-Type: application/json" \
    --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getTransactionByHash\",\"params\":[\"$hash\"],\"id\":1}")
  
  blockNum=$(echo "$result" | jq -r '.result.blockNumber')
  nonce=$(echo "$result" | jq -r '.result.nonce')
  
  if [ "$blockNum" != "null" ]; then
    blockDec=$(printf "%d" $blockNum 2>/dev/null || echo "?")
    nonceDec=$(printf "%d" $nonce 2>/dev/null || echo "?")
    echo "  ✅ Found in block $blockDec (nonce $nonceDec)"
  else
    echo "  ❌ Not found"
  fi
  echo ""
done

echo "=== CHECK COMPLETE ==="

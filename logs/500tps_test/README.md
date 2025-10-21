# 500 TPS Test Logs

## Test Summary
- **Test Name**: 500 TPS Stress Test
- **Date**: October 16, 2025
- **Duration**: 64 seconds
- **Total Transactions**: 15,000
- **Average TPS**: 234.38
- **Errors**: 0
- **Start Nonce**: 13,614
- **End Nonce**: 28,614

## Log Files

### Erigon Logs
- **File**: `erigon_500tps.log`
- **Size**: 142 KB
- **Lines**: 1,063
- **Time Range**: Last 10 minutes from log capture

### Beacon Chain Logs
- **File**: `beacon_500tps.log`
- **Size**: 64 KB
- **Lines**: 386
- **Time Range**: Last 10 minutes from log capture

## Transaction Hashes
All 15,000 transaction hashes are stored in:
`scripts/transaction_hashes_500tps_30sec.txt`

## Docker Containers
- **Erigon**: `erigon-ultra-optimized`
- **Beacon**: `prysm-beacon-chain-ultra`

## How to View Logs
```bash
# View Erigon logs
cat logs/500tps_test/erigon_500tps.log

# View Beacon logs
cat logs/500tps_test/beacon_500tps.log

# Search for specific patterns
grep "block" logs/500tps_test/erigon_500tps.log
grep "HotStuff" logs/500tps_test/beacon_500tps.log
```

## Notes
- Logs captured using: `docker logs <container> --since 10m`
- Both containers were running during the entire test
- Zero errors reported during the test execution

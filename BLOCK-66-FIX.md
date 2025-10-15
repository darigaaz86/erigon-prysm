# Block 66 Stuck Issue - Root Cause and Fix

## Issue Summary

The blockchain consistently stopped producing blocks at block 66, with Erigon's block builder timing out and never recovering.

## Symptoms

- Blocks 1-65: ✅ Built successfully (20-70ms build time)
- Block 66: ❌ Started building but never completed
- Error: `"Stopping block builder due to max build time exceeded"`
- After timeout: Block builder never restarted, chain permanently stalled

## Root Cause

**Incorrect genesis configuration for EIP-1559 and EIP-4844 (Cancun fork)**

The `genesis/genesis.json` file had null values for post-London/Cancun fields:

```json
{
  "baseFeePerGas": null,      // ❌ Should be a hex value
  "excessBlobGas": null,      // ❌ Should be "0x0"
  "blobGasUsed": null         // ❌ Should be "0x0"
}
```

### Why It Failed at Block 66

With null values, Erigon's base fee calculation worked initially but failed when:
1. The accumulated state reached a specific threshold (~block 66)
2. Base fee calculations tried to process null values
3. The block builder hung during the transaction filtration phase
4. Never logged "Filtration" message, indicating it hung before that point

## The Fix

Updated `genesis/genesis.json` with proper values:

```json
{
  "baseFeePerGas": "0x3b9aca00",  // ✅ 1 Gwei (1,000,000,000 wei)
  "excessBlobGas": "0x0",         // ✅ 0 (no excess blob gas at genesis)
  "blobGasUsed": "0x0"            // ✅ 0 (no blobs used at genesis)
}
```

## Verification

After the fix:
- ✅ Blocks 66, 67, 68, 69, 70+ all build successfully
- ✅ Build times: 20-50ms (normal)
- ✅ No timeouts or hangs
- ✅ Chain continues indefinitely

## Additional Improvements

1. **Faster slot time**: Changed `SECONDS_PER_SLOT` from 12 to 6 seconds for faster testing
2. **Increased timeout**: Added `--engine-endpoint-timeout-seconds=30` to Prysm beacon config

## Testing Methodology

To identify the root cause, we:

1. **Confirmed block-number specificity**: Changed slot time from 12s to 6s - issue still occurred at block 66
2. **Ruled out JWT issues**: Verified authentication was working correctly
3. **Added debug logging**: Instrumented Erigon's block building code to track execution flow
4. **Identified hang point**: Found that `ProvideTxns` was being called but filtration never started
5. **Fixed genesis**: Updated null values to proper hex values
6. **Verified fix**: Blocks now build successfully past block 100+

## Files Changed

- `genesis/genesis.json` - Fixed EIP-1559/4844 values
- `config/config.yml` - Changed slot time to 6 seconds
- `docker-compose-private.yml` - Added engine timeout parameter

## Lessons Learned

1. **Genesis validation is critical**: Null values in genesis can cause subtle bugs that only appear after many blocks
2. **EIP-1559 requires baseFeePerGas**: Post-London forks must have a valid base fee, even at genesis
3. **EIP-4844 requires blob gas fields**: Cancun fork requires proper blob gas initialization
4. **Block-specific bugs**: Some bugs only manifest at specific block heights due to accumulated state

## Related EIPs

- **EIP-1559** (London): Introduced `baseFeePerGas` for fee market
- **EIP-4844** (Cancun): Introduced `excessBlobGas` and `blobGasUsed` for blob transactions

## Commit

```
commit cb4a0b2698
Fix: Resolve block 66 stuck issue by setting proper genesis EIP-1559/4844 values
```

---

**Status**: ✅ RESOLVED  
**Date**: 2025-10-15  
**Impact**: Critical - Chain could not progress past block 65  
**Solution**: Genesis configuration fix

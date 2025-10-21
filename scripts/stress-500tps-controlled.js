const { ethers } = require('ethers');
const fs = require('fs');

// Configuration
const RPC_URL = 'http://localhost:8545';
const PRIVATE_KEY = '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80';
const RECIPIENT = '0x70997970C51812dc3A010C7d01b50e0d17dc79C8';
const TPS = 500; // Transactions per second (rate-limited)
const DURATION_SECONDS = 30;
const TOTAL_TRANSACTIONS = TPS * DURATION_SECONDS; // 500 TPS * 30 sec = 15000 txs
const OUTPUT_FILE = 'transaction_hashes_500tps_controlled.txt';

// Rate limiting: send in batches every 100ms
const BATCH_INTERVAL_MS = 100; // Send every 100ms
const BATCH_SIZE = Math.floor(TPS * (BATCH_INTERVAL_MS / 1000)); // 500 * 0.1 = 50 txs per batch

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function main() {
  const provider = new ethers.JsonRpcProvider(RPC_URL);
  const wallet = new ethers.Wallet(PRIVATE_KEY, provider);
  
  console.log('=== CONTROLLED STRESS TEST: 500 TPS FOR 30 SECONDS ===');
  console.log(`From: ${wallet.address}`);
  console.log(`To: ${RECIPIENT}`);
  console.log(`Target TPS: ${TPS} (rate-limited)`);
  console.log(`Duration: ${DURATION_SECONDS} seconds`);
  console.log(`Total Transactions: ${TOTAL_TRANSACTIONS}`);
  console.log(`Batch Size: ${BATCH_SIZE} txs every ${BATCH_INTERVAL_MS}ms`);
  console.log(`Output File: ${OUTPUT_FILE}`);
  console.log('');
  
  // Get starting nonce
  const startNonce = await provider.getTransactionCount(wallet.address);
  console.log(`Starting nonce: ${startNonce}`);
  console.log('');
  
  const txHashes = [];
  const startTime = Date.now();
  let sentCount = 0;
  let errorCount = 0;
  let batchCount = 0;
  
  // Clear output file
  fs.writeFileSync(OUTPUT_FILE, `# 500 TPS Controlled Stress Test - ${new Date().toISOString()}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# From: ${wallet.address}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# To: ${RECIPIENT}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Target TPS: ${TPS} (rate-limited)\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Batch Size: ${BATCH_SIZE} every ${BATCH_INTERVAL_MS}ms\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Total Transactions: ${TOTAL_TRANSACTIONS}\n\n`);
  
  console.log('Starting rate-limited transaction submission...');
  console.log(`Sending ${BATCH_SIZE} transactions every ${BATCH_INTERVAL_MS}ms\n`);
  
  // Send transactions in controlled batches
  for (let batch = 0; batch < TOTAL_TRANSACTIONS; batch += BATCH_SIZE) {
    const batchStartTime = Date.now();
    const promises = [];
    const batchEnd = Math.min(batch + BATCH_SIZE, TOTAL_TRANSACTIONS);
    
    // Send batch
    for (let i = batch; i < batchEnd; i++) {
      const promise = (async (index) => {
        try {
          const tx = await wallet.sendTransaction({
            to: RECIPIENT,
            value: ethers.parseEther('0.001'),
            nonce: startNonce + index,
            gasLimit: 21000
          });
          
          txHashes.push(tx.hash);
          sentCount++;
          
          // Write to file immediately
          fs.appendFileSync(OUTPUT_FILE, `${tx.hash}\n`);
          
          return tx.hash;
        } catch (error) {
          errorCount++;
          if (errorCount <= 10) {
            console.error(`Error sending transaction ${index + 1}: ${error.message}`);
          }
          fs.appendFileSync(OUTPUT_FILE, `# ERROR at tx ${index + 1}: ${error.message}\n`);
          return null;
        }
      })(i);
      
      promises.push(promise);
    }
    
    await Promise.all(promises);
    batchCount++;
    
    // Progress update every 500 transactions (every second)
    if (sentCount % 500 === 0 || sentCount >= TOTAL_TRANSACTIONS) {
      const elapsed = (Date.now() - startTime) / 1000;
      const avgTps = sentCount / elapsed;
      const targetTps = TPS;
      const deviation = ((avgTps - targetTps) / targetTps * 100).toFixed(1);
      console.log(`Progress: ${sentCount}/${TOTAL_TRANSACTIONS} txs | Elapsed: ${elapsed.toFixed(2)}s | Avg TPS: ${avgTps.toFixed(2)} | Target: ${targetTps} | Deviation: ${deviation}% | Errors: ${errorCount}`);
    }
    
    // Rate limiting: wait until next batch time
    const batchDuration = Date.now() - batchStartTime;
    const waitTime = BATCH_INTERVAL_MS - batchDuration;
    
    if (waitTime > 0 && batch + BATCH_SIZE < TOTAL_TRANSACTIONS) {
      await sleep(waitTime);
    }
  }
  
  const totalTime = (Date.now() - startTime) / 1000;
  const actualTps = sentCount / totalTime;
  const deviation = ((actualTps - TPS) / TPS * 100).toFixed(2);
  
  console.log('\n=== TEST COMPLETE ===');
  console.log(`Total Transactions Sent: ${sentCount}/${TOTAL_TRANSACTIONS}`);
  console.log(`Total Errors: ${errorCount}`);
  console.log(`Total Time: ${totalTime.toFixed(2)} seconds`);
  console.log(`Target TPS: ${TPS}`);
  console.log(`Actual TPS: ${actualTps.toFixed(2)}`);
  console.log(`Deviation: ${deviation}%`);
  console.log(`Batches Sent: ${batchCount}`);
  console.log(`Transaction hashes saved to: ${OUTPUT_FILE}`);
  console.log(`Final nonce: ${startNonce + sentCount}`);
  
  // Write summary to file
  fs.appendFileSync(OUTPUT_FILE, `\n# === SUMMARY ===\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Total Sent: ${sentCount}/${TOTAL_TRANSACTIONS}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Errors: ${errorCount}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Duration: ${totalTime.toFixed(2)}s\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Target TPS: ${TPS}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Actual TPS: ${actualTps.toFixed(2)}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Deviation: ${deviation}%\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Start Nonce: ${startNonce}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# End Nonce: ${startNonce + sentCount}\n`);
}

main()
  .then(() => process.exit(0))
  .catch(error => {
    console.error('Fatal error:', error);
    process.exit(1);
  });

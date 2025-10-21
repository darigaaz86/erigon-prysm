const { ethers } = require('ethers');
const fs = require('fs');

// Configuration
const RPC_URL = 'http://localhost:8545';
const PRIVATE_KEY = '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80';
const RECIPIENT = '0x70997970C51812dc3A010C7d01b50e0d17dc79C8';
const TPS = 5; // Transactions per second
const DURATION_MINUTES = 5;
const TOTAL_TRANSACTIONS = TPS * 60 * DURATION_MINUTES; // 5 TPS * 60 sec * 5 min = 1500 txs
const OUTPUT_FILE = 'transaction_hashes_5tps_5min.txt';

async function main() {
  const provider = new ethers.JsonRpcProvider(RPC_URL);
  const wallet = new ethers.Wallet(PRIVATE_KEY, provider);
  
  console.log('=== SUSTAINED 5 TPS TEST FOR 5 MINUTES ===');
  console.log(`From: ${wallet.address}`);
  console.log(`To: ${RECIPIENT}`);
  console.log(`TPS: ${TPS}`);
  console.log(`Duration: ${DURATION_MINUTES} minutes`);
  console.log(`Total Transactions: ${TOTAL_TRANSACTIONS}`);
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
  
  // Clear output file
  fs.writeFileSync(OUTPUT_FILE, `# 5 TPS Test - ${new Date().toISOString()}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# From: ${wallet.address}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# To: ${RECIPIENT}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Total Transactions: ${TOTAL_TRANSACTIONS}\n\n`);
  
  console.log('Starting transaction submission...');
  console.log('Press Ctrl+C to stop early\n');
  
  // Send transactions in batches
  const intervalMs = 1000 / TPS; // 200ms between transactions for 5 TPS
  
  for (let i = 0; i < TOTAL_TRANSACTIONS; i++) {
    const txStartTime = Date.now();
    
    try {
      const tx = await wallet.sendTransaction({
        to: RECIPIENT,
        value: ethers.parseEther('0.001'),
        nonce: startNonce + i,
        gasLimit: 21000
      });
      
      txHashes.push(tx.hash);
      sentCount++;
      
      // Write to file immediately
      fs.appendFileSync(OUTPUT_FILE, `${tx.hash}\n`);
      
      // Progress update every 50 transactions
      if ((i + 1) % 50 === 0) {
        const elapsed = (Date.now() - startTime) / 1000;
        const avgTps = sentCount / elapsed;
        console.log(`Progress: ${i + 1}/${TOTAL_TRANSACTIONS} txs | Elapsed: ${elapsed.toFixed(1)}s | Avg TPS: ${avgTps.toFixed(2)} | Errors: ${errorCount}`);
      }
      
      // Wait to maintain TPS rate
      const txElapsed = Date.now() - txStartTime;
      const waitTime = Math.max(0, intervalMs - txElapsed);
      if (waitTime > 0) {
        await new Promise(resolve => setTimeout(resolve, waitTime));
      }
      
    } catch (error) {
      errorCount++;
      console.error(`Error sending transaction ${i + 1}: ${error.message}`);
      fs.appendFileSync(OUTPUT_FILE, `# ERROR at tx ${i + 1}: ${error.message}\n`);
    }
  }
  
  const totalTime = (Date.now() - startTime) / 1000;
  const actualTps = sentCount / totalTime;
  
  console.log('\n=== TEST COMPLETE ===');
  console.log(`Total Transactions Sent: ${sentCount}/${TOTAL_TRANSACTIONS}`);
  console.log(`Total Errors: ${errorCount}`);
  console.log(`Total Time: ${totalTime.toFixed(2)} seconds (${(totalTime / 60).toFixed(2)} minutes)`);
  console.log(`Average TPS: ${actualTps.toFixed(2)}`);
  console.log(`Transaction hashes saved to: ${OUTPUT_FILE}`);
  console.log(`Final nonce: ${startNonce + sentCount}`);
  
  // Write summary to file
  fs.appendFileSync(OUTPUT_FILE, `\n# === SUMMARY ===\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Total Sent: ${sentCount}/${TOTAL_TRANSACTIONS}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Errors: ${errorCount}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Duration: ${totalTime.toFixed(2)}s\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Average TPS: ${actualTps.toFixed(2)}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Start Nonce: ${startNonce}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# End Nonce: ${startNonce + sentCount}\n`);
}

main()
  .then(() => process.exit(0))
  .catch(error => {
    console.error('Fatal error:', error);
    process.exit(1);
  });

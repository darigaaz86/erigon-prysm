const { ethers } = require('ethers');
const fs = require('fs');

// Configuration
const RPC_URL = 'http://localhost:8545';
const PRIVATE_KEY = '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80';
const RECIPIENT = '0x70997970C51812dc3A010C7d01b50e0d17dc79C8';
const TPS = 100; // Transactions per second
const DURATION_SECONDS = 30;
const TOTAL_TRANSACTIONS = TPS * DURATION_SECONDS; // 100 TPS * 30 sec = 3000 txs
const OUTPUT_FILE = 'transaction_hashes_100tps_30sec.txt';

async function main() {
  const provider = new ethers.JsonRpcProvider(RPC_URL);
  const wallet = new ethers.Wallet(PRIVATE_KEY, provider);
  
  console.log('=== STRESS TEST: 100 TPS FOR 30 SECONDS ===');
  console.log(`From: ${wallet.address}`);
  console.log(`To: ${RECIPIENT}`);
  console.log(`TPS: ${TPS}`);
  console.log(`Duration: ${DURATION_SECONDS} seconds`);
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
  fs.writeFileSync(OUTPUT_FILE, `# 100 TPS Stress Test - ${new Date().toISOString()}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# From: ${wallet.address}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# To: ${RECIPIENT}\n`);
  fs.appendFileSync(OUTPUT_FILE, `# Total Transactions: ${TOTAL_TRANSACTIONS}\n\n`);
  
  console.log('Starting transaction submission...');
  console.log('Sending transactions as fast as possible!\n');
  
  // Send all transactions as fast as possible
  const promises = [];
  
  for (let i = 0; i < TOTAL_TRANSACTIONS; i++) {
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
        
        // Progress update every 100 transactions
        if ((index + 1) % 100 === 0) {
          const elapsed = (Date.now() - startTime) / 1000;
          const avgTps = sentCount / elapsed;
          console.log(`Progress: ${index + 1}/${TOTAL_TRANSACTIONS} txs | Elapsed: ${elapsed.toFixed(2)}s | Avg TPS: ${avgTps.toFixed(2)} | Errors: ${errorCount}`);
        }
        
        return tx.hash;
      } catch (error) {
        errorCount++;
        console.error(`Error sending transaction ${index + 1}: ${error.message}`);
        fs.appendFileSync(OUTPUT_FILE, `# ERROR at tx ${index + 1}: ${error.message}\n`);
        return null;
      }
    })(i);
    
    promises.push(promise);
    
    // Small delay every 10 transactions to avoid overwhelming the RPC
    if ((i + 1) % 10 === 0) {
      await Promise.all(promises.splice(0, 10));
    }
  }
  
  // Wait for remaining transactions
  await Promise.all(promises);
  
  const totalTime = (Date.now() - startTime) / 1000;
  const actualTps = sentCount / totalTime;
  
  console.log('\n=== TEST COMPLETE ===');
  console.log(`Total Transactions Sent: ${sentCount}/${TOTAL_TRANSACTIONS}`);
  console.log(`Total Errors: ${errorCount}`);
  console.log(`Total Time: ${totalTime.toFixed(2)} seconds`);
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

const { ethers } = require('ethers');

async function slowTpsTest() {
    const provider = new ethers.JsonRpcProvider('http://localhost:8545');
    const wallet = new ethers.Wallet('0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80', provider);
    
    const recipient = '0x70997970C51812dc3A010C7d01b50e0d17dc79C8';
    const amount = ethers.parseEther('0.001'); // Small amount
    
    console.log('🚀 Slow TPS Test - 10 transactions per second');
    console.log('👛 Wallet:', wallet.address);
    
    const startNonce = await provider.getTransactionCount(wallet.address);
    console.log('📊 Starting nonce:', startNonce);
    
    const totalTx = 100; // Send 100 transactions total
    const txPerSecond = 10;
    const delayMs = 1000 / txPerSecond; // 100ms between transactions
    
    console.log(`\n📤 Sending ${totalTx} transactions at ${txPerSecond} TPS (${delayMs}ms delay)\n`);
    
    const txHashes = [];
    const startTime = Date.now();
    
    // Send transactions
    for (let i = 0; i < totalTx; i++) {
        const nonce = startNonce + i;
        
        try {
            const tx = await wallet.sendTransaction({
                to: recipient,
                value: amount,
                nonce: nonce,
                gasLimit: 21000,
                maxPriorityFeePerGas: ethers.parseUnits('1', 'gwei'),
                maxFeePerGas: ethers.parseUnits('50', 'gwei'),
            });
            
            txHashes.push(tx.hash);
            
            if ((i + 1) % 10 === 0) {
                console.log(`   Sent ${i + 1}/${totalTx} transactions...`);
            }
            
            // Wait before sending next transaction
            if (i < totalTx - 1) {
                await new Promise(resolve => setTimeout(resolve, delayMs));
            }
        } catch (error) {
            console.error(`❌ Failed to send transaction ${i}:`, error.message);
        }
    }
    
    const sendDuration = (Date.now() - startTime) / 1000;
    console.log(`\n✅ Sent ${txHashes.length} transactions in ${sendDuration.toFixed(2)}s`);
    console.log(`   Actual rate: ${(txHashes.length / sendDuration).toFixed(2)} TPS\n`);
    
    // Wait for confirmations
    console.log('⏳ Waiting for confirmations...\n');
    
    let confirmed = 0;
    let failed = 0;
    const confirmStartTime = Date.now();
    
    for (let i = 0; i < txHashes.length; i++) {
        try {
            const receipt = await provider.getTransactionReceipt(txHashes[i]);
            if (receipt && receipt.status === 1) {
                confirmed++;
            } else if (receipt && receipt.status === 0) {
                failed++;
                console.log(`❌ Transaction ${i} failed:`, txHashes[i]);
            }
            
            if ((i + 1) % 10 === 0) {
                console.log(`   Checked ${i + 1}/${txHashes.length} transactions...`);
            }
        } catch (error) {
            console.log(`⚠️  Transaction ${i} not found yet:`, txHashes[i]);
        }
    }
    
    const totalDuration = (Date.now() - startTime) / 1000;
    
    console.log('\n' + '='.repeat(70));
    console.log('📊 SLOW TPS TEST RESULTS');
    console.log('='.repeat(70));
    console.log(`⏱️  Total Duration: ${totalDuration.toFixed(2)}s`);
    console.log(`📤 Sent: ${txHashes.length}`);
    console.log(`✅ Confirmed: ${confirmed}`);
    console.log(`❌ Failed: ${failed}`);
    console.log(`⏳ Pending: ${txHashes.length - confirmed - failed}`);
    console.log(`🎯 Success Rate: ${((confirmed / txHashes.length) * 100).toFixed(2)}%`);
    console.log(`📈 Confirmed TPS: ${(confirmed / totalDuration).toFixed(2)} TPS`);
    console.log('='.repeat(70));
    
    // Verify a few random transactions
    console.log('\n🔍 Verifying random transactions...\n');
    const samplesToVerify = 5;
    for (let i = 0; i < samplesToVerify && i < txHashes.length; i++) {
        const randomIndex = Math.floor(Math.random() * txHashes.length);
        const hash = txHashes[randomIndex];
        
        try {
            const receipt = await provider.getTransactionReceipt(hash);
            if (receipt) {
                console.log(`✅ TX ${randomIndex}: Block ${receipt.blockNumber}, Status: ${receipt.status === 1 ? 'Success' : 'Failed'}`);
                console.log(`   Hash: ${hash}`);
            } else {
                console.log(`⚠️  TX ${randomIndex}: Not found`);
                console.log(`   Hash: ${hash}`);
            }
        } catch (error) {
            console.log(`❌ TX ${randomIndex}: Error - ${error.message}`);
        }
    }
    
    return {
        sent: txHashes.length,
        confirmed,
        failed,
        pending: txHashes.length - confirmed - failed,
        txHashes: txHashes.slice(0, 10) // Return first 10 hashes for verification
    };
}

slowTpsTest()
    .then(result => {
        console.log('\n✅ Test completed!');
        console.log('Sample transaction hashes:', result.txHashes);
        process.exit(0);
    })
    .catch(error => {
        console.error('\n❌ Test failed:', error);
        process.exit(1);
    });

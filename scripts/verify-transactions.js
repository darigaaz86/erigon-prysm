const { ethers } = require('ethers');

async function verifyTransactions() {
    const provider = new ethers.JsonRpcProvider('http://localhost:8545');
    const wallet = new ethers.Wallet('0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80', provider);
    
    console.log('🔍 Verifying 100 transactions from slow TPS test\n');
    
    const startNonce = 3; // From the test
    const totalTx = 100;
    
    const results = {
        confirmed: [],
        pending: [],
        failed: []
    };
    
    console.log('Checking transactions...\n');
    
    for (let i = 0; i < totalTx; i++) {
        const nonce = startNonce + i;
        
        try {
            // Get transaction by sender and nonce
            // We need to reconstruct or query each transaction
            // Since we don't have the hashes, let's check the account's transaction count
            // and query recent blocks
            
            // For now, let's just check if transactions with these nonces exist
            // by checking the current nonce
            const currentNonce = await provider.getTransactionCount(wallet.address);
            
            if (nonce < currentNonce) {
                results.confirmed.push(nonce);
            } else {
                results.pending.push(nonce);
            }
        } catch (error) {
            results.failed.push(nonce);
        }
        
        if ((i + 1) % 20 === 0) {
            console.log(`   Checked ${i + 1}/${totalTx}...`);
        }
    }
    
    console.log('\n' + '='.repeat(70));
    console.log('📊 VERIFICATION RESULTS');
    console.log('='.repeat(70));
    console.log(`✅ Confirmed (nonce used): ${results.confirmed.length}`);
    console.log(`⏳ Pending (nonce not used): ${results.pending.length}`);
    console.log(`❌ Failed: ${results.failed.length}`);
    console.log('='.repeat(70));
    
    // Now let's actually query some blocks to find the transactions
    console.log('\n🔍 Searching for transactions in recent blocks...\n');
    
    const currentBlock = await provider.getBlockNumber();
    console.log(`Current block: ${currentBlock}`);
    
    const txFound = [];
    const blocksToCheck = 20; // Check last 20 blocks
    
    for (let blockNum = Math.max(0, currentBlock - blocksToCheck); blockNum <= currentBlock; blockNum++) {
        try {
            const block = await provider.getBlock(blockNum, true);
            if (block && block.transactions) {
                const ourTxs = block.transactions.filter(tx => 
                    tx.from && tx.from.toLowerCase() === wallet.address.toLowerCase() &&
                    tx.nonce >= startNonce && tx.nonce < startNonce + totalTx
                );
                
                if (ourTxs.length > 0) {
                    console.log(`Block ${blockNum}: Found ${ourTxs.length} transactions`);
                    txFound.push(...ourTxs.map(tx => ({
                        hash: tx.hash,
                        nonce: tx.nonce,
                        block: blockNum
                    })));
                }
            }
        } catch (error) {
            // Block might not be available yet
        }
    }
    
    console.log(`\n📊 Found ${txFound.length} transactions in recent blocks`);
    
    // Verify a sample of found transactions
    if (txFound.length > 0) {
        console.log('\n✅ Sample verified transactions:\n');
        const samplesToShow = Math.min(10, txFound.length);
        
        for (let i = 0; i < samplesToShow; i++) {
            const tx = txFound[i];
            console.log(`Nonce ${tx.nonce}: ${tx.hash}`);
            console.log(`   Block: ${tx.block}`);
            
            // Verify with curl command
            console.log(`   Curl: curl -s http://localhost:8545 -X POST -H "Content-Type: application/json" \\`);
            console.log(`     --data '{"jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["${tx.hash}"],"id":1}' | jq`);
            console.log('');
        }
    }
    
    return {
        totalChecked: totalTx,
        confirmedByNonce: results.confirmed.length,
        foundInBlocks: txFound.length,
        sampleHashes: txFound.slice(0, 10).map(tx => tx.hash)
    };
}

verifyTransactions()
    .then(result => {
        console.log('\n✅ Verification completed!');
        console.log(`Total: ${result.totalChecked}`);
        console.log(`Confirmed by nonce: ${result.confirmedByNonce}`);
        console.log(`Found in blocks: ${result.foundInBlocks}`);
        process.exit(0);
    })
    .catch(error => {
        console.error('\n❌ Verification failed:', error);
        process.exit(1);
    });

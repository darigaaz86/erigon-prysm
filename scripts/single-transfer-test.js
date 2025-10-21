const { ethers } = require('ethers');

async function sendSingleTransfer() {
    const provider = new ethers.JsonRpcProvider('http://localhost:8545');
    const wallet = new ethers.Wallet('0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80', provider);
    
    console.log('👛 Wallet address:', wallet.address);
    
    // Get current nonce
    const nonce = await provider.getTransactionCount(wallet.address);
    console.log('📊 Current nonce:', nonce);
    
    // Get balance
    const balance = await provider.getBalance(wallet.address);
    console.log('💰 Balance:', ethers.formatEther(balance), 'ETH');
    
    // Send a single transaction
    const recipient = '0x70997970C51812dc3A010C7d01b50e0d17dc79C8';
    const amount = ethers.parseEther('0.1');
    
    console.log('\n🚀 Sending transaction...');
    console.log('   To:', recipient);
    console.log('   Amount: 0.1 ETH');
    console.log('   Nonce:', nonce);
    
    const tx = await wallet.sendTransaction({
        to: recipient,
        value: amount,
        nonce: nonce,
        gasLimit: 21000,
        maxPriorityFeePerGas: ethers.parseUnits('1', 'gwei'),
        maxFeePerGas: ethers.parseUnits('50', 'gwei'),
    });
    
    console.log('📤 Transaction sent!');
    console.log('   Hash:', tx.hash);
    console.log('   Nonce:', tx.nonce);
    
    console.log('\n⏳ Waiting for confirmation...');
    const receipt = await tx.wait(1, 60000); // Wait up to 60 seconds
    
    console.log('\n✅ Transaction confirmed!');
    console.log('   Block:', receipt.blockNumber);
    console.log('   Block Hash:', receipt.blockHash);
    console.log('   Gas Used:', receipt.gasUsed.toString());
    console.log('   Status:', receipt.status === 1 ? 'Success' : 'Failed');
    
    // Verify by querying the transaction
    console.log('\n🔍 Verifying transaction...');
    const txData = await provider.getTransaction(tx.hash);
    console.log('   Found in block:', txData.blockNumber);
    
    const receiptData = await provider.getTransactionReceipt(tx.hash);
    console.log('   Receipt status:', receiptData.status === 1 ? 'Success ✅' : 'Failed ❌');
    
    return {
        hash: tx.hash,
        blockNumber: receipt.blockNumber,
        nonce: tx.nonce
    };
}

sendSingleTransfer()
    .then(result => {
        console.log('\n✅ Test completed successfully!');
        console.log('Result:', result);
        process.exit(0);
    })
    .catch(error => {
        console.error('\n❌ Error:', error.message);
        process.exit(1);
    });

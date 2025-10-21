const { ethers } = require('ethers');

async function main() {
  const provider = new ethers.JsonRpcProvider('http://localhost:8545');
  const wallet = new ethers.Wallet('0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80', provider);
  
  console.log('=== SIMPLE TRANSACTION TEST ===');
  console.log(`From: ${wallet.address}`);
  console.log('');
  
  // Get current nonce
  const nonce = await provider.getTransactionCount(wallet.address);
  console.log(`Current nonce: ${nonce}`);
  console.log('');
  
  // Send a simple ETH transfer
  console.log('Sending 0.001 ETH transaction...');
  const tx = await wallet.sendTransaction({
    to: '0x70997970C51812dc3A010C7d01b50e0d17dc79C8',
    value: ethers.parseEther('0.001'),
    gasLimit: 21000
  });
  
  console.log(`Transaction hash: ${tx.hash}`);
  console.log('Waiting for confirmation...');
  
  const receipt = await tx.wait();
  
  console.log('');
  console.log('✅ Transaction confirmed!');
  console.log(`Block number: ${receipt.blockNumber}`);
  console.log(`Gas used: ${receipt.gasUsed.toString()}`);
  console.log(`Status: ${receipt.status === 1 ? 'Success' : 'Failed'}`);
  
  return tx.hash;
}

main()
  .then(hash => {
    console.log('');
    console.log(`Transaction hash for curl: ${hash}`);
    process.exit(0);
  })
  .catch(error => {
    console.error('Error:', error.message);
    process.exit(1);
  });

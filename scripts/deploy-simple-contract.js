const { ethers } = require('ethers');

// Simple Storage Contract
const contractABI = [
  "function set(uint256 value) public",
  "function get() public view returns (uint256)"
];

const contractBytecode = "0x608060405234801561001057600080fd5b5060f78061001f6000396000f3fe6080604052348015600f57600080fd5b5060043610603c5760003560e01c80631865c57d1460415780636d4ce63c14605d578063b3de648b146077575b600080fd5b6047608f565b6040518082815260200191505060405180910390f35b60636098565b604051808281526020019150506040518091039060a1565b608d60048036036020811015608b57600080fd5b503560aa565b005b60005481565b60005490565b60005556fea2646970667358221220c7f729e1c0f3c3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e3e364736f6c63430007060033";

async function main() {
  const provider = new ethers.JsonRpcProvider('http://localhost:8545');
  const wallet = new ethers.Wallet('0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80', provider);
  
  console.log('=== DEPLOYING SIMPLE STORAGE CONTRACT ===');
  console.log(`Deployer: ${wallet.address}`);
  console.log('');
  
  // Get current nonce
  const nonce = await provider.getTransactionCount(wallet.address);
  console.log(`Current nonce: ${nonce}`);
  
  // Deploy contract
  console.log('Deploying contract...');
  const factory = new ethers.ContractFactory(contractABI, contractBytecode, wallet);
  const contract = await factory.deploy();
  
  console.log(`Transaction hash: ${contract.deploymentTransaction().hash}`);
  console.log('Waiting for deployment...');
  
  await contract.waitForDeployment();
  const contractAddress = await contract.getAddress();
  
  console.log('');
  console.log('✅ Contract deployed successfully!');
  console.log(`Contract address: ${contractAddress}`);
  console.log(`Deployment tx hash: ${contract.deploymentTransaction().hash}`);
  
  // Now send a transaction to the contract
  console.log('');
  console.log('=== SENDING TRANSACTION TO CONTRACT ===');
  console.log('Calling set(42)...');
  
  const tx = await contract.set(42);
  console.log(`Transaction hash: ${tx.hash}`);
  console.log('Waiting for confirmation...');
  
  const receipt = await tx.wait();
  console.log('');
  console.log('✅ Transaction confirmed!');
  console.log(`Block number: ${receipt.blockNumber}`);
  console.log(`Gas used: ${receipt.gasUsed.toString()}`);
  
  // Verify the value
  const value = await contract.get();
  console.log(`Stored value: ${value}`);
  
  console.log('');
  console.log('=== SUMMARY ===');
  console.log(`Contract Address: ${contractAddress}`);
  console.log(`Deployment TX: ${contract.deploymentTransaction().hash}`);
  console.log(`Set TX: ${tx.hash}`);
  
  return tx.hash;
}

main()
  .then(hash => {
    console.log('');
    console.log(`Final transaction hash: ${hash}`);
    process.exit(0);
  })
  .catch(error => {
    console.error('Error:', error);
    process.exit(1);
  });

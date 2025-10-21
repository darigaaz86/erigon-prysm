const { ethers } = require('ethers');
const fs = require('fs');
const solc = require('solc');

// Configuration
const RPC_URL = 'http://localhost:8545';
const PRIVATE_KEY = '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80'; // Default Hardhat account #0
const RECIPIENT_ADDRESS = '0x70997970C51812dc3A010C7d01b50e0d17dc79C8'; // Default Hardhat account #1

async function compileContract() {
    console.log('📝 Compiling contract...');
    const source = fs.readFileSync('contracts/SimpleToken.sol', 'utf8');

    const input = {
        language: 'Solidity',
        sources: {
            'SimpleToken.sol': {
                content: source
            }
        },
        settings: {
            outputSelection: {
                '*': {
                    '*': ['abi', 'evm.bytecode']
                }
            }
        }
    };

    const output = JSON.parse(solc.compile(JSON.stringify(input)));

    if (output.errors) {
        output.errors.forEach(err => {
            console.error(err.formattedMessage);
        });
        if (output.errors.some(err => err.severity === 'error')) {
            throw new Error('Compilation failed');
        }
    }

    const contract = output.contracts['SimpleToken.sol']['SimpleToken'];
    return {
        abi: contract.abi,
        bytecode: contract.evm.bytecode.object
    };
}

async function main() {
    try {
        // Compile contract
        const { abi, bytecode } = await compileContract();
        console.log('✅ Contract compiled successfully\n');

        // Connect to network
        console.log('🔌 Connecting to network...');
        const provider = new ethers.JsonRpcProvider(RPC_URL);
        const wallet = new ethers.Wallet(PRIVATE_KEY, provider);

        console.log('👛 Deployer address:', wallet.address);
        const balance = await provider.getBalance(wallet.address);
        console.log('💰 Balance:', ethers.formatEther(balance), 'ETH\n');

        // Deploy contract
        console.log('🚀 Deploying contract...');
        const factory = new ethers.ContractFactory(abi, bytecode, wallet);
        const initialSupply = 1000000; // 1 million tokens
        const contract = await factory.deploy(initialSupply);

        await contract.waitForDeployment();
        const contractAddress = await contract.getAddress();

        console.log('✅ Contract deployed at:', contractAddress);
        console.log('📊 Initial supply:', initialSupply, 'tokens\n');

        // Check deployer balance
        const deployerBalance = await contract.balanceOf(wallet.address);
        console.log('💼 Deployer token balance:', ethers.formatEther(deployerBalance), 'STK\n');

        // Transfer tokens
        console.log('💸 Transferring 1000 tokens to', RECIPIENT_ADDRESS);
        const transferAmount = ethers.parseEther('1000');
        const tx = await contract.transfer(RECIPIENT_ADDRESS, transferAmount);

        console.log('⏳ Transaction hash:', tx.hash);
        await tx.wait();
        console.log('✅ Transfer confirmed!\n');

        // Check balances after transfer
        const deployerBalanceAfter = await contract.balanceOf(wallet.address);
        const recipientBalance = await contract.balanceOf(RECIPIENT_ADDRESS);

        console.log('📊 Final balances:');
        console.log('   Deployer:', ethers.formatEther(deployerBalanceAfter), 'STK');
        console.log('   Recipient:', ethers.formatEther(recipientBalance), 'STK');

        // Save contract info
        const contractInfo = {
            address: contractAddress,
            abi: abi,
            deploymentBlock: tx.blockNumber,
            deployer: wallet.address
        };

        fs.writeFileSync('contract-info.json', JSON.stringify(contractInfo, null, 2));
        console.log('\n💾 Contract info saved to contract-info.json');

    } catch (error) {
        console.error('❌ Error:', error.message);
        process.exit(1);
    }
}

main();

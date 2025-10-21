const { ethers } = require('ethers');
const fs = require('fs');

// Test with native ETH transfers (21K gas vs 100K gas for ERC20)
const RPC_URL = 'http://localhost:8545';
const PRIVATE_KEY = '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80';
const TEST_DURATION_SECONDS = 60;
const BATCH_SIZE = 500;
const CONCURRENT_BATCHES = 2;
const BATCH_DELAY_MS = 100;

class NativeTransferTest {
    constructor() {
        this.provider = new ethers.JsonRpcProvider(RPC_URL, undefined, {
            staticNetwork: true,
            batchMaxCount: 1000,
            batchStallTime: 10,
            polling: false
        });
        this.wallet = new ethers.Wallet(PRIVATE_KEY, this.provider);
        this.results = {
            totalSent: 0,
            totalConfirmed: 0,
            totalFailed: 0,
            startTime: null,
            endTime: null,
            sendingEndTime: null,
            blockStats: new Map(),
            errors: []
        };
        this.pendingTxHashes = [];
    }

    async sendNativeTransfer(nonce, recipient, amount) {
        try {
            const tx = await this.wallet.sendTransaction({
                to: recipient,
                value: amount,
                nonce: nonce,
                gasLimit: 21000,
                maxPriorityFeePerGas: ethers.parseUnits('1', 'gwei'),
                maxFeePerGas: ethers.parseUnits('50', 'gwei'),
                chainId: 32382,
                type: 2
            });

            this.results.totalSent++;
            this.pendingTxHashes.push(tx.hash);
            return { hash: tx.hash, nonce: nonce };
        } catch (error) {
            this.results.totalFailed++;
            if (this.results.errors.length < 100) {
                this.results.errors.push({
                    nonce: nonce,
                    error: error.message.substring(0, 100)
                });
            }
            return null;
        }
    }

    async sendBatch(startNonce, batchSize, recipient, amount) {
        const promises = [];
        for (let i = 0; i < batchSize; i++) {
            promises.push(this.sendNativeTransfer(startNonce + i, recipient, amount));
        }
        return await Promise.allSettled(promises);
    }

    async sendConcurrentBatches(startNonce, batchSize, concurrentCount, recipient, amount) {
        const batchPromises = [];
        for (let i = 0; i < concurrentCount; i++) {
            const batchStartNonce = startNonce + (i * batchSize);
            batchPromises.push(this.sendBatch(batchStartNonce, batchSize, recipient, amount));
        }
        return await Promise.all(batchPromises);
    }

    async waitForConfirmations() {
        console.log(`⏳ Waiting for ${this.pendingTxHashes.length} transactions...\n`);

        const checkBatchSize = 50;
        let lastUpdate = 0;

        for (let i = 0; i < this.pendingTxHashes.length; i += checkBatchSize) {
            const batch = this.pendingTxHashes.slice(i, i + checkBatchSize);
            const promises = batch.map(hash =>
                this.provider.waitForTransaction(hash, 1, 180000)
                    .then(receipt => {
                        if (receipt) {
                            this.results.totalConfirmed++;
                            const blockCount = this.results.blockStats.get(receipt.blockNumber) || 0;
                            this.results.blockStats.set(receipt.blockNumber, blockCount + 1);
                        }
                    })
                    .catch(() => {
                        this.results.totalFailed++;
                    })
            );

            await Promise.allSettled(promises);

            if (this.results.totalConfirmed - lastUpdate >= 500) {
                console.log(`   ✅ ${this.results.totalConfirmed}/${this.pendingTxHashes.length} confirmed...`);
                lastUpdate = this.results.totalConfirmed;
            }
        }
    }

    async runTest() {
        console.log('🚀 NATIVE TRANSFER THROUGHPUT TEST');
        console.log('='.repeat(70));
        console.log('Using native ETH transfers (21K gas vs 100K for ERC20)');
        console.log(`Duration: ${TEST_DURATION_SECONDS} seconds`);
        console.log(`Batch size: ${BATCH_SIZE} transactions`);
        console.log(`Concurrent batches: ${CONCURRENT_BATCHES}`);
        console.log(`Effective rate: ${BATCH_SIZE * CONCURRENT_BATCHES} tx/round`);
        console.log('='.repeat(70), '\n');

        let currentNonce = await this.provider.getTransactionCount(this.wallet.address, 'pending');
        console.log('📊 Starting nonce:', currentNonce, '\n');

        const recipient = '0x70997970C51812dc3A010C7d01b50e0d17dc79C8';
        const amount = ethers.parseEther('0.0001');

        this.results.startTime = Date.now();
        const endTime = this.results.startTime + (TEST_DURATION_SECONDS * 1000);

        let roundCount = 0;
        const txPerRound = BATCH_SIZE * CONCURRENT_BATCHES;

        console.log('🔥 SENDING PHASE\n');

        while (Date.now() < endTime) {
            roundCount++;
            const roundStart = Date.now();

            try {
                await this.sendConcurrentBatches(currentNonce, BATCH_SIZE, CONCURRENT_BATCHES, recipient, amount);

                const roundDuration = Date.now() - roundStart;
                const roundTps = (txPerRound / roundDuration) * 1000;
                const successRate = ((this.results.totalSent / (this.results.totalSent + this.results.totalFailed)) * 100).toFixed(1);

                console.log(`📤 Round ${roundCount}: ${roundDuration}ms | ${roundTps.toFixed(0)} TPS | Total: ${this.results.totalSent} | Success: ${successRate}%`);

                currentNonce += txPerRound;

                if (BATCH_DELAY_MS > 0) {
                    await new Promise(resolve => setTimeout(resolve, BATCH_DELAY_MS));
                }
            } catch (error) {
                console.error(`❌ Round ${roundCount} error:`, error.message);
            }
        }

        this.results.sendingEndTime = Date.now();
        const sendingDuration = (this.results.sendingEndTime - this.results.startTime) / 1000;
        const sendingTps = this.results.totalSent / sendingDuration;

        console.log('\n' + '='.repeat(70));
        console.log(`⏸️  SENDING COMPLETE`);
        console.log(`   Duration: ${sendingDuration.toFixed(2)}s`);
        console.log(`   Sent: ${this.results.totalSent.toLocaleString()}`);
        console.log(`   Rate: ${sendingTps.toFixed(2)} TPS`);
        console.log('='.repeat(70), '\n');

        await this.waitForConfirmations();

        this.results.endTime = Date.now();

        this.printResults();
        this.saveResults();
    }

    printResults() {
        const totalDuration = (this.results.endTime - this.results.startTime) / 1000;
        const sendingDuration = (this.results.sendingEndTime - this.results.startTime) / 1000;
        const confirmedTps = this.results.totalConfirmed / totalDuration;

        console.log('\n');
        console.log('='.repeat(70));
        console.log('📊 NATIVE TRANSFER TEST RESULTS');
        console.log('='.repeat(70));
        console.log(`⏱️  Total Duration: ${totalDuration.toFixed(2)}s`);
        console.log(`📤 Sent: ${this.results.totalSent.toLocaleString()}`);
        console.log(`✅ Confirmed: ${this.results.totalConfirmed.toLocaleString()}`);
        console.log(`❌ Failed: ${this.results.totalFailed.toLocaleString()}`);
        console.log(`🎯 Confirmed TPS: ${confirmedTps.toFixed(2)} TPS`);
        console.log('');

        const sortedBlocks = Array.from(this.results.blockStats.entries()).sort((a, b) => a[0] - b[0]);

        if (sortedBlocks.length > 0) {
            const txPerBlock = sortedBlocks.map(([_, count]) => count);
            const avgTxPerBlock = txPerBlock.reduce((a, b) => a + b, 0) / txPerBlock.length;
            const maxTxPerBlock = Math.max(...txPerBlock);
            const blockTime = sendingDuration / sortedBlocks.length;

            console.log('📦 Block Statistics:');
            console.log(`   Blocks: ${sortedBlocks.length}`);
            console.log(`   Avg TX/Block: ${avgTxPerBlock.toFixed(2)}`);
            console.log(`   Max TX/Block: ${maxTxPerBlock}`);
            console.log(`   Avg Block Time: ${blockTime.toFixed(2)}s`);
            console.log(`   Theoretical Max: ${(maxTxPerBlock / blockTime).toFixed(2)} TPS`);
            console.log('');
            console.log('💡 Native transfers use 21K gas vs 100K for ERC20');
            console.log(`   Gas savings: ${((1 - 21000 / 100000) * 100).toFixed(0)}%`);
            console.log(`   Potential 5x more transactions per block!`);
        }

        console.log('='.repeat(70));
    }

    saveResults() {
        const summary = {
            testType: 'native-transfer',
            metrics: {
                totalSent: this.results.totalSent,
                totalConfirmed: this.results.totalConfirmed,
                totalFailed: this.results.totalFailed,
                confirmedTps: this.results.totalConfirmed / ((this.results.endTime - this.results.startTime) / 1000)
            },
            blockStats: Object.fromEntries(this.results.blockStats),
            timestamp: new Date().toISOString()
        };

        fs.writeFileSync('native-transfer-results.json', JSON.stringify(summary, null, 2));
        console.log('\n💾 Results saved to native-transfer-results.json');
    }
}

async function main() {
    const test = new NativeTransferTest();

    try {
        await test.runTest();
    } catch (error) {
        console.error('❌ Test failed:', error);
        process.exit(1);
    }
}

main();

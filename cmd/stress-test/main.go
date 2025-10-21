package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	rpcURL     = flag.String("rpc", getEnv("RPC_URL", "http://localhost:8545"), "RPC endpoint URL")
	privateKey = flag.String("key", "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", "Private key (without 0x)")
	recipient  = flag.String("to", "0x70997970C51812dc3A010C7d01b50e0d17dc79C8", "Recipient address")
	targetTPS  = flag.Int("tps", 500, "Target transactions per second")
	duration   = flag.Int("duration", 30, "Test duration in seconds")
	workers    = flag.Int("workers", 10, "Number of concurrent workers")
	chainID    = flag.Int64("chainid", 32382, "Chain ID")
	gasPrice   = flag.Int64("gasprice", 1000000000, "Gas price in wei (1 gwei default)")
	outputFile = flag.String("output", "tx_hashes.txt", "Output file for transaction hashes")
)

type Stats struct {
	sent   uint64
	errors uint64
	start  time.Time
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	flag.Parse()

	fmt.Println("=== HIGH-PERFORMANCE STRESS TEST ===")
	fmt.Printf("RPC URL: %s\n", *rpcURL)
	fmt.Printf("Target TPS: %d\n", *targetTPS)
	fmt.Printf("Duration: %d seconds\n", *duration)
	fmt.Printf("Workers: %d\n", *workers)
	fmt.Printf("Total Transactions: %d\n", *targetTPS**duration)
	fmt.Printf("Chain ID: %d\n", *chainID)
	fmt.Println()

	// Connect to Ethereum node
	client, err := ethclient.Dial(*rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	// Load private key
	key, err := crypto.HexToECDSA(*privateKey)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	publicKey := key.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Failed to cast public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	toAddress := common.HexToAddress(*recipient)

	fmt.Printf("From: %s\n", fromAddress.Hex())
	fmt.Printf("To: %s\n", toAddress.Hex())
	fmt.Println()

	// Get starting nonce
	ctx := context.Background()
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}
	fmt.Printf("Starting nonce: %d\n\n", nonce)

	// Open output file
	file, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "# Stress Test - %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "# From: %s\n", fromAddress.Hex())
	fmt.Fprintf(file, "# To: %s\n", toAddress.Hex())
	fmt.Fprintf(file, "# Target TPS: %d\n", *targetTPS)
	fmt.Fprintf(file, "# Duration: %d seconds\n\n", *duration)

	// Statistics
	stats := &Stats{start: time.Now()}
	var fileMutex sync.Mutex

	// Calculate transactions per worker
	totalTxs := *targetTPS * *duration
	txsPerWorker := totalTxs / *workers

	// Start workers
	var wg sync.WaitGroup
	startTime := time.Now()

	fmt.Println("Starting transaction submission...")

	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func(workerID int, startNonce uint64) {
			defer wg.Done()

			currentNonce := startNonce
			sent := 0
			targetPerWorker := txsPerWorker

			// Calculate delay between transactions for this worker
			delayNs := time.Duration(float64(time.Second) / float64(*targetTPS / *workers))

			for sent < targetPerWorker {
				txStart := time.Now()

				// Create transaction
				tx := types.NewTransaction(
					currentNonce,
					toAddress,
					big.NewInt(1000000000000000), // 0.001 ETH
					21000,
					big.NewInt(*gasPrice),
					nil,
				)

				// Sign transaction
				signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(*chainID)), key)
				if err != nil {
					atomic.AddUint64(&stats.errors, 1)
					log.Printf("Worker %d: Failed to sign tx: %v", workerID, err)
					currentNonce++
					sent++
					continue
				}

				// Send transaction
				err = client.SendTransaction(ctx, signedTx)
				if err != nil {
					atomic.AddUint64(&stats.errors, 1)
					if atomic.LoadUint64(&stats.errors) <= 10 {
						log.Printf("Worker %d: Failed to send tx: %v", workerID, err)
					}
				} else {
					atomic.AddUint64(&stats.sent, 1)

					// Write hash to file
					fileMutex.Lock()
					fmt.Fprintf(file, "%s\n", signedTx.Hash().Hex())
					fileMutex.Unlock()
				}

				currentNonce++
				sent++

				// Rate limiting
				elapsed := time.Since(txStart)
				if elapsed < delayNs {
					time.Sleep(delayNs - elapsed)
				}
			}
		}(w, nonce+uint64(w*txsPerWorker))
	}

	// Progress reporter
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sent := atomic.LoadUint64(&stats.sent)
				errors := atomic.LoadUint64(&stats.errors)
				elapsed := time.Since(startTime).Seconds()
				currentTPS := float64(sent) / elapsed
				deviation := ((currentTPS - float64(*targetTPS)) / float64(*targetTPS)) * 100

				fmt.Printf("Progress: %d/%d txs | Elapsed: %.2fs | TPS: %.2f | Target: %d | Deviation: %.1f%% | Errors: %d\n",
					sent, totalTxs, elapsed, currentTPS, *targetTPS, deviation, errors)
			case <-done:
				return
			}
		}
	}()

	// Wait for all workers
	wg.Wait()
	close(done)

	// Final statistics
	totalTime := time.Since(startTime).Seconds()
	sent := atomic.LoadUint64(&stats.sent)
	errors := atomic.LoadUint64(&stats.errors)
	actualTPS := float64(sent) / totalTime
	deviation := ((actualTPS - float64(*targetTPS)) / float64(*targetTPS)) * 100

	fmt.Println("\n=== TEST COMPLETE ===")
	fmt.Printf("Total Transactions Sent: %d/%d\n", sent, totalTxs)
	fmt.Printf("Total Errors: %d\n", errors)
	fmt.Printf("Total Time: %.2f seconds\n", totalTime)
	fmt.Printf("Target TPS: %d\n", *targetTPS)
	fmt.Printf("Actual TPS: %.2f\n", actualTPS)
	fmt.Printf("Deviation: %.2f%%\n", deviation)
	fmt.Printf("Transaction hashes saved to: %s\n", *outputFile)

	// Write summary to file
	fmt.Fprintf(file, "\n# === SUMMARY ===\n")
	fmt.Fprintf(file, "# Total Sent: %d/%d\n", sent, totalTxs)
	fmt.Fprintf(file, "# Errors: %d\n", errors)
	fmt.Fprintf(file, "# Duration: %.2fs\n", totalTime)
	fmt.Fprintf(file, "# Target TPS: %d\n", *targetTPS)
	fmt.Fprintf(file, "# Actual TPS: %.2f\n", actualTPS)
	fmt.Fprintf(file, "# Deviation: %.2f%%\n", deviation)
}

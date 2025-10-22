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
	rpcURL        = flag.String("rpc", getEnv("RPC_URL", "http://localhost:8545"), "RPC endpoint URL")
	privateKey    = flag.String("key", "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80", "Private key (without 0x)")
	targetTPS     = flag.Int("tps", 500, "Target transactions per second")
	duration      = flag.Int("duration", 30, "Test duration in seconds")
	numSenders    = flag.Int("senders", 20, "Number of sender addresses to use")
	numRecipients = flag.Int("recipients", 100, "Number of recipient addresses")
	chainID       = flag.Int64("chainid", 32382, "Chain ID")
	gasPrice      = flag.Int64("gasprice", 1000000000, "Gas price in wei (1 gwei default)")
	outputFile    = flag.String("output", "tx_hashes.txt", "Output file for transaction hashes")
	skipSetup     = flag.Bool("skip-setup", false, "Skip the setup phase (addresses already funded)")
)

type Stats struct {
	sent   uint64
	errors uint64
	start  time.Time
}

type Account struct {
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
	Nonce      uint64
	mu         sync.Mutex
}

func (a *Account) GetAndIncrementNonce() uint64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	nonce := a.Nonce
	a.Nonce++
	return nonce
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func generateAccounts(count int) []*Account {
	accounts := make([]*Account, count)
	for i := 0; i < count; i++ {
		key, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("Failed to generate key: %v", err)
		}
		publicKey := key.Public()
		publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
		address := crypto.PubkeyToAddress(*publicKeyECDSA)
		accounts[i] = &Account{
			PrivateKey: key,
			Address:    address,
			Nonce:      0,
		}
	}
	return accounts
}

func main() {
	flag.Parse()

	fmt.Println("=== MULTI-ADDRESS STRESS TEST ===")
	fmt.Printf("RPC URL: %s\n", *rpcURL)
	fmt.Printf("Target TPS: %d\n", *targetTPS)
	fmt.Printf("Duration: %d seconds\n", *duration)
	fmt.Printf("Sender Addresses: %d\n", *numSenders)
	fmt.Printf("Recipient Addresses: %d\n", *numRecipients)
	fmt.Printf("Total Transactions: %d\n", *targetTPS**duration)
	fmt.Printf("Chain ID: %d\n", *chainID)
	fmt.Println()

	// Connect to Ethereum node
	client, err := ethclient.Dial(*rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Load genesis private key
	genesisKey, err := crypto.HexToECDSA(*privateKey)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	publicKey := genesisKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Failed to cast public key to ECDSA")
	}

	genesisAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("Genesis Address: %s\n\n", genesisAddress.Hex())

	// Generate sender and recipient accounts
	fmt.Println("Generating accounts...")
	senderAccounts := generateAccounts(*numSenders)
	recipientAccounts := generateAccounts(*numRecipients)
	fmt.Printf("Generated %d sender accounts and %d recipient accounts\n\n", *numSenders, *numRecipients)

	// Setup phase: Fund sender accounts
	if !*skipSetup {
		fmt.Println("=== SETUP PHASE: Funding Sender Accounts ===")
		err = fundAccounts(ctx, client, genesisKey, genesisAddress, senderAccounts)
		if err != nil {
			log.Fatalf("Failed to fund accounts: %v", err)
		}
		fmt.Println("All sender accounts funded successfully!\n")

		// Wait for transactions to be mined by checking balances
		fmt.Println("Waiting for funding transactions to be mined (checking balances)...")
		err = waitForFunding(ctx, client, senderAccounts)
		if err != nil {
			log.Fatalf("Failed to verify funding: %v", err)
		}
		fmt.Println("All accounts verified and funded!\n")

		// Update nonces for sender accounts
		for _, acc := range senderAccounts {
			nonce, err := client.PendingNonceAt(ctx, acc.Address)
			if err != nil {
				log.Printf("Warning: Failed to get nonce for %s: %v", acc.Address.Hex(), err)
			} else {
				acc.Nonce = nonce
			}
		}
	} else {
		fmt.Println("Skipping setup phase (--skip-setup flag set)\n")
		// Get current nonces
		for _, acc := range senderAccounts {
			nonce, err := client.PendingNonceAt(ctx, acc.Address)
			if err != nil {
				log.Printf("Warning: Failed to get nonce for %s: %v", acc.Address.Hex(), err)
			} else {
				acc.Nonce = nonce
			}
		}
	}

	// Test phase: Send transactions from multiple senders to multiple recipients
	fmt.Println("=== TEST PHASE: Multi-Address Transaction Test ===")
	err = runStressTest(ctx, client, senderAccounts, recipientAccounts)
	if err != nil {
		log.Fatalf("Stress test failed: %v", err)
	}
}

func fundAccounts(ctx context.Context, client *ethclient.Client, genesisKey *ecdsa.PrivateKey, genesisAddress common.Address, accounts []*Account) error {
	nonce, err := client.PendingNonceAt(ctx, genesisAddress)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %v", err)
	}

	// Fund each account with 10000 ETH
	fundAmount := new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e18))

	fmt.Printf("Funding %d accounts with 10000 ETH each...\n", len(accounts))

	for i, acc := range accounts {
		tx := types.NewTransaction(
			nonce,
			acc.Address,
			fundAmount,
			21000,
			big.NewInt(*gasPrice),
			nil,
		)

		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(*chainID)), genesisKey)
		if err != nil {
			return fmt.Errorf("failed to sign funding tx: %v", err)
		}

		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			return fmt.Errorf("failed to send funding tx: %v", err)
		}

		nonce++
		if (i+1)%10 == 0 || i == len(accounts)-1 {
			fmt.Printf("  Funded %d/%d accounts\n", i+1, len(accounts))
		}
	}

	return nil
}

func runStressTest(ctx context.Context, client *ethclient.Client, senders []*Account, recipients []*Account) error {
	// Open output file
	file, err := os.Create(*outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "# Multi-Address Stress Test - %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(file, "# Senders: %d\n", len(senders))
	fmt.Fprintf(file, "# Recipients: %d\n", len(recipients))
	fmt.Fprintf(file, "# Target TPS: %d\n", *targetTPS)
	fmt.Fprintf(file, "# Duration: %d seconds\n\n", *duration)

	// Statistics
	stats := &Stats{start: time.Now()}
	var fileMutex sync.Mutex

	totalTxs := *targetTPS * *duration
	txsPerSender := totalTxs / len(senders)

	// Start workers (one per sender)
	var wg sync.WaitGroup
	startTime := time.Now()

	fmt.Printf("Starting transaction submission with %d senders...\n", len(senders))

	for senderIdx, sender := range senders {
		wg.Add(1)
		go func(senderID int, senderAcc *Account) {
			defer wg.Done()

			sent := 0
			targetForSender := txsPerSender

			// Calculate delay between transactions for this sender
			delayNs := time.Duration(float64(time.Second) / float64(*targetTPS/len(senders)))

			for sent < targetForSender {
				txStart := time.Now()

				// Pick a random recipient
				recipientIdx := (senderID + sent) % len(recipients)
				recipient := recipients[recipientIdx]

				// Get nonce
				nonce := senderAcc.GetAndIncrementNonce()

				// Create transaction
				tx := types.NewTransaction(
					nonce,
					recipient.Address,
					big.NewInt(1000000000000), // 0.0001 ETH
					21000,
					big.NewInt(*gasPrice),
					nil,
				)

				// Sign transaction
				signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(*chainID)), senderAcc.PrivateKey)
				if err != nil {
					atomic.AddUint64(&stats.errors, 1)
					sent++
					continue
				}

				// Send transaction
				err = client.SendTransaction(ctx, signedTx)
				if err != nil {
					atomic.AddUint64(&stats.errors, 1)
					if atomic.LoadUint64(&stats.errors) <= 10 {
						log.Printf("Sender %d: Failed to send tx: %v", senderID, err)
					}
				} else {
					atomic.AddUint64(&stats.sent, 1)

					// Write hash to file
					fileMutex.Lock()
					fmt.Fprintf(file, "%s\n", signedTx.Hash().Hex())
					fileMutex.Unlock()
				}

				sent++

				// Rate limiting
				elapsed := time.Since(txStart)
				if elapsed < delayNs {
					time.Sleep(delayNs - elapsed)
				}
			}
		}(senderIdx, sender)
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

	return nil
}

func waitForFunding(ctx context.Context, client *ethclient.Client, accounts []*Account) error {
	minBalance := new(big.Int).Mul(big.NewInt(1), big.NewInt(1e18)) // At least 1 ETH
	maxWait := 120 * time.Second
	checkInterval := 3 * time.Second
	startTime := time.Now()

	for {
		allFunded := true
		fundedCount := 0

		for _, acc := range accounts {
			balance, err := client.BalanceAt(ctx, acc.Address, nil)
			if err != nil {
				log.Printf("Warning: Failed to check balance for %s: %v", acc.Address.Hex(), err)
				allFunded = false
				continue
			}

			if balance.Cmp(minBalance) >= 0 {
				fundedCount++
			} else {
				allFunded = false
			}
		}

		fmt.Printf("  Verified %d/%d accounts funded\n", fundedCount, len(accounts))

		if allFunded {
			return nil
		}

		if time.Since(startTime) > maxWait {
			return fmt.Errorf("timeout waiting for accounts to be funded (only %d/%d funded after %v)", fundedCount, len(accounts), maxWait)
		}

		time.Sleep(checkInterval)
	}
}

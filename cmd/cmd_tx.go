package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

func cmdTx() *cli.Command {
	return &cli.Command{
		Name:  "tx",
		Usage: "transaction related commands",
		Subcommands: []*cli.Command{
			cmdFindCosmosHash(),
		},
	}
}

func cmdFindCosmosHash() *cli.Command {
	return &cli.Command{
		Name:      "find-cosmos-hash",
		Action:    findCosmosHash,
		Usage:     "find Cosmos transaction hash from recent blocks",
		ArgsUsage: "",
		Description: `
Find the Cosmos transaction hash by searching recent blocks.
This is useful when you have an EVM transaction hash from moca-cmd but need the Cosmos hash for mocad query.

Examples:
# Find recent transactions and their Cosmos hashes
$ moca-cmd tx find-cosmos-hash`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "blocks",
				Value: 20,
				Usage: "number of recent blocks to search",
			},
		},
	}
}

func findCosmosHash(ctx *cli.Context) error {
	client, err := NewClient(ctx, ClientOptions{IsQueryCmd: true})
	if err != nil {
		return toCmdErr(err)
	}

	c, cancel := context.WithCancel(globalContext)
	defer cancel()

	blocksToSearch := ctx.Int("blocks")
	
	// Get latest block height
	latestBlock, err := client.GetLatestBlock(c)
	if err != nil {
		return toCmdErr(err)
	}
	
	latestHeight := latestBlock.Header.Height
	fmt.Printf("Searching last %d blocks (from height %d)...\n\n", blocksToSearch, latestHeight)
	
	foundTxs := 0
	for i := int64(0); i < int64(blocksToSearch); i++ {
		height := latestHeight - i
		if height < 1 {
			break
		}
		
		block, err := client.GetBlockByHeight(c, height)
		if err != nil {
			continue
		}
		
		if len(block.Data.Txs) > 0 {
			for txIdx, txBytes := range block.Data.Txs {
				// Calculate Cosmos hash (SHA256 of tx bytes)
				cosmosHash := calculateCosmosHash(txBytes)
				
				// Try to decode and get some info about the tx
				txBase64 := base64.StdEncoding.EncodeToString(txBytes)
				
				fmt.Printf("Block %d, TX #%d:\n", height, txIdx)
				fmt.Printf("  Cosmos Hash: %s\n", cosmosHash)
				fmt.Printf("  TX Size: %d bytes\n", len(txBytes))
				fmt.Printf("  Query command: mocad query tx %s\n", cosmosHash)
				fmt.Printf("  TX Base64 (first 100 chars): %s...\n\n", truncateString(txBase64, 100))
				
				foundTxs++
			}
		}
	}
	
	if foundTxs == 0 {
		fmt.Printf("No transactions found in the last %d blocks.\n", blocksToSearch)
	} else {
		fmt.Printf("Found %d transaction(s) in the last %d blocks.\n", foundTxs, blocksToSearch)
	}
	
	return nil
}

func calculateCosmosHash(txBytes []byte) string {
	hash := sha256.Sum256(txBytes)
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

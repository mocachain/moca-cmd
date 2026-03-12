package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// GetCosmosHashFromTxBytes calculates the Cosmos transaction hash from transaction bytes
// Cosmos hash is SHA256(txBytes)
func GetCosmosHashFromTxBytes(txBytes []byte) string {
	hash := sha256.Sum256(txBytes)
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// GetCosmosHashFromBase64 calculates the Cosmos transaction hash from base64 encoded tx bytes
func GetCosmosHashFromBase64(txBase64 string) (string, error) {
	txBytes, err := base64.StdEncoding.DecodeString(txBase64)
	if err != nil {
		return "", err
	}
	return GetCosmosHashFromTxBytes(txBytes), nil
}

// FormatTxHashOutput formats transaction hash output to show both EVM and Cosmos formats
func FormatTxHashOutput(evmHash string, cosmosHash string) string {
	var output strings.Builder
	output.WriteString(fmt.Sprintf("transaction hash: %s\n", evmHash))
	if cosmosHash != "" && cosmosHash != evmHash {
		output.WriteString(fmt.Sprintf("cosmos tx hash (for mocad query): %s\n", cosmosHash))
		output.WriteString(fmt.Sprintf("\nTo query this transaction:\n"))
		output.WriteString(fmt.Sprintf("  mocad query tx %s\n", cosmosHash))
	}
	return output.String()
}

// NormalizeEvmHash ensures EVM hash has 0x prefix
func NormalizeEvmHash(hash string) string {
	if !strings.HasPrefix(hash, "0x") {
		return "0x" + hash
	}
	return hash
}

// NormalizeCosmosHash ensures Cosmos hash is uppercase without 0x prefix
func NormalizeCosmosHash(hash string) string {
	hash = strings.TrimPrefix(hash, "0x")
	return strings.ToUpper(hash)
}

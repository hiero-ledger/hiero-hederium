package hedera

import "time"

// Temorary file for constants

const (
	GetBlockByHashOrNumber = "getBlockByHashOrNumber"
	GetContractResult      = "getContractResult"
	GetContractById        = "getContractById"
	GetAccountById         = "getAccountById"
	GetTokenById           = "getTokenById"

	DefaultExpiration = 1 * time.Hour

	// Maximum gas that can be used per second
	maxGasPerSec = 15000000
	// Transaction size limit in bytes (128KB)
	transactionSizeLimit = 128 * 1024
	// Default file append chunk size
	fileAppendChunkSize = 5120
	// Maximum number of chunks for file append
	maxChunks = 20

	maxRetries = 2

	retryDelay = 1 * time.Second

	Limit = 100

	MaxPages = 100
)

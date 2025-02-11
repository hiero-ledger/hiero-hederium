package domain

import (
	"fmt"
)

// EthGetBlockByHashParams represents parameters for eth_getBlockByHash
type EthGetBlockByHashParams struct {
	BlockHash   string `json:"blockHash" binding:"required,len=66,hexadecimal,startswith=0x"`
	ShowDetails bool   `json:"showDetails" binding:"required"`
}

// EthGetBlockByNumberParams represents parameters for eth_getBlockByNumber
type EthGetBlockByNumberParams struct {
	BlockNumber string `json:"blockNumber" binding:"required,block_number_or_tag"`
	ShowDetails bool   `json:"showDetails" binding:"required"`
}

// EthGetBalanceParams represents parameters for eth_getBalance
type EthGetBalanceParams struct {
	Address     string `json:"address" binding:"required,eth_address"`
	BlockNumber string `json:"blockNumber" binding:"omitempty,block_number_or_tag"`
}

// EthGetTransactionCountParams represents parameters for eth_getTransactionCount
type EthGetTransactionCountParams struct {
	Address     string `json:"address" binding:"required,eth_address"`
	BlockNumber string `json:"blockNumber" binding:"omitempty,block_number_or_tag"`
}

// EthEstimateGasParams represents parameters for eth_estimateGas
type EthEstimateGasParams struct {
	CallObject     map[string]interface{} `json:"callObject" binding:"required"`
	BlockParameter string                 `json:"blockParameter" binding:"omitempty,block_number_or_tag"`
}

// EthCallParams represents parameters for eth_call
type EthCallParams struct {
	CallObject map[string]interface{} `json:"callObject" binding:"required"`
	Block      string                 `json:"block" binding:"required,block_number_or_tag"`
}

// EthGetTransactionByHashParams represents parameters for eth_getTransactionByHash
type EthGetTransactionByHashParams struct {
	TransactionHash string `json:"transactionHash" binding:"required,len=66,hexadecimal,startswith=0x"`
}

// EthGetTransactionReceiptParams represents parameters for eth_getTransactionReceipt
type EthGetTransactionReceiptParams struct {
	TransactionHash string `json:"transactionHash" binding:"required,len=66,hexadecimal,startswith=0x"`
}

// EthFeeHistoryParams represents parameters for eth_feeHistory
type EthFeeHistoryParams struct {
	BlockCount        string   `json:"blockCount" binding:"required,hexadecimal,startswith=0x"`
	NewestBlock       string   `json:"newestBlock" binding:"required,block_number_or_tag"`
	RewardPercentiles []string `json:"rewardPercentiles" binding:"omitempty,dive,hexadecimal,startswith=0x"`
}

// EthGetStorageAtParams represents parameters for eth_getStorageAt
type EthGetStorageAtParams struct {
	Address         string `json:"address" binding:"required,eth_address"`
	StoragePosition string `json:"storagePosition" binding:"required,hexadecimal,startswith=0x"`
	BlockNumber     string `json:"blockNumber" binding:"omitempty,block_number_or_tag"`
}

// EthGetLogsParams represents parameters for eth_getLogs
type EthGetLogsParams struct {
	Address   []string `json:"address" binding:"omitempty,dive,eth_address"`      
	Topics    []string `json:"topics" binding:"omitempty,dive,hexadecimal,len=66"`
	BlockHash string   `json:"blockHash" binding:"omitempty,hexadecimal,len=66"` 
	FromBlock string   `json:"fromBlock" binding:"omitempty,block_number_or_tag"`
	ToBlock   string   `json:"toBlock" binding:"omitempty,block_number_or_tag"`
}

// EthGetBlockTransactionCountByHashParams represents parameters for eth_getBlockTransactionCountByHash
type EthGetBlockTransactionCountByHashParams struct {
	BlockHash string `json:"blockHash" binding:"required,len=66,hexadecimal,startswith=0x"`
}

// EthGetBlockTransactionCountByNumberParams represents parameters for eth_getBlockTransactionCountByNumber
type EthGetBlockTransactionCountByNumberParams struct {
	BlockNumber string `json:"blockNumber" binding:"required,block_number_or_tag"`
}

// EthGetTransactionByBlockHashAndIndexParams represents parameters for eth_getTransactionByBlockHashAndIndex
type EthGetTransactionByBlockHashAndIndexParams struct {
	BlockHash        string `json:"blockHash" binding:"required,len=66,hexadecimal,startswith=0x"`
	TransactionIndex string `json:"transactionIndex" binding:"required,hexadecimal,startswith=0x"`
}

// EthGetTransactionByBlockNumberAndIndexParams represents parameters for eth_getTransactionByBlockNumberAndIndex
type EthGetTransactionByBlockNumberAndIndexParams struct {
	BlockNumber      string `json:"blockNumber" binding:"required,block_number_or_tag"`
	TransactionIndex string `json:"transactionIndex" binding:"required,hexadecimal,startswith=0x"`
}

// EthSendRawTransactionParams represents parameters for eth_sendRawTransaction
type EthSendRawTransactionParams struct {
	SignedTransaction string `json:"signedTransaction" binding:"required,hexadecimal,startswith=0x"`
}

// EthGetCodeParams represents parameters for eth_getCode
type EthGetCodeParams struct {
	Address     string `json:"address" binding:"required,eth_address"`
	BlockNumber string `json:"blockNumber" binding:"required,block_number_or_tag"`
}

// EthGetUncleCountByBlockHashParams represents parameters for eth_getUncleCountByBlockHash
type EthGetUncleCountByBlockHashParams struct {
	BlockHash string `json:"blockHash" binding:"required,len=66,hexadecimal,startswith=0x"`
}

// EthGetUncleCountByBlockNumberParams represents parameters for eth_getUncleCountByBlockNumber
type EthGetUncleCountByBlockNumberParams struct {
	BlockNumber string `json:"blockNumber" binding:"required,block_number_or_tag"`
}

// EthGetUncleByBlockHashAndIndexParams represents parameters for eth_getUncleByBlockHashAndIndex
type EthGetUncleByBlockHashAndIndexParams struct {
	BlockHash string `json:"blockHash" binding:"required,len=66,hexadecimal,startswith=0x"`
	Index     string `json:"index" binding:"required,hexadecimal,startswith=0x"`
}

// EthGetUncleByBlockNumberAndIndexParams represents parameters for eth_getUncleByBlockNumberAndIndex
type EthGetUncleByBlockNumberAndIndexParams struct {
	BlockNumber string `json:"blockNumber" binding:"required,block_number_or_tag"`
	Index       string `json:"index" binding:"required,hexadecimal,startswith=0x"`
}
package service

import "time"

// Constants for the Ethereum JSON-RPC API methods + other constants
// Temporary place for these constants until we have a better place for them
const (
	GetBlockByHash                      = "eth_getBlockByHash"
	GetBlockByNumber                    = "eth_getBlockByNumber"
	GetBlockTransactionCount            = "eth_getBlockTransactionCount"
	GetUncleByBlockHash                 = "eth_getUncleByBlockHash"
	GetUncleByBlockNumber               = "eth_getUncleByBlockNumber"
	GetUncleCountByBlockHash            = "eth_getUncleCountByBlockHash"
	GetBlockTransactionCountByHash      = "eth_getBlockTransactionCountByHash"
	GetBlockTransactionCountByNumber    = "eth_getBlockTransactionCountByNumber"
	GetTransactionByHash                = "eth_getTransactionByHash"
	GetTransactionCount                 = "eth_getTransactionCount"
	SendTransaction                     = "eth_sendTransaction"
	SendRawTransaction                  = "eth_sendRawTransaction"
	GetPendingTransactions              = "eth_getPendingTransactions"
	GetAccounts                         = "eth_accounts"
	GetTransactionByBlockHashAndIndex   = "eth_getTransactionByBlockHashAndIndex"
	GetTransactionByBlockNumberAndIndex = "eth_getTransactionByBlockNumberAndIndex"
	GetBalance                          = "eth_getBalance"
	GetCode                             = "eth_getCode"
	GetStorageAt                        = "eth_getStorageAt"
	GetTransactionReceipt               = "eth_getTransactionReceipt"
	GetGasPrice                         = "eth_gasPrice"
	EstimateGas                         = "eth_estimateGas"
	GetLogs                             = "eth_getLogs"
	GetChainId                          = "eth_chainId"
	GetProtocolVersion                  = "eth_protocolVersion"
	GetSyncing                          = "eth_syncing"
	Call                                = "eth_call"
	ProtocolVersion                     = "eth_protocolVersion"
	NetVersion                          = "net_version"
	NetListening                        = "net_listening"
	NetPeerCount                        = "net_peerCount"

	DefaultExpiration = 1 * time.Hour
)

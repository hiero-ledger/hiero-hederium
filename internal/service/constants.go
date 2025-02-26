package service

import "time"

// Constants for the Ethereum JSON-RPC API methods + other constants
// Temporary place for these constants until we have a better place for them
const (
	GetBlockNumber                      = "eth_blockNumber"
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
	ShortExpiration   = 1 * time.Second

	// Fungible token creation selectors
	CreateFungibleTokenV1         string = "0x83062e38"
	CreateFungibleTokenV2         string = "0x6577761c"
	CreateFungibleTokenV3         string = "0x6c42689c"
	CreateFungibleTokenWithFeesV1 string = "0x6446e17e"
	CreateFungibleTokenWithFeesV2 string = "0x8f4eb604"
	CreateFungibleTokenWithFeesV3 string = "0x5c14dd49"

	// Non-fungible token creation selectors
	CreateNonFungibleTokenV1         string = "0x5e724461"
	CreateNonFungibleTokenV2         string = "0x5a3c15af"
	CreateNonFungibleTokenV3         string = "0x5d24ea56"
	CreateNonFungibleTokenWithFeesV1 string = "0x5e9a79c9"
	CreateNonFungibleTokenWithFeesV2 string = "0x5f99d676"
	CreateNonFungibleTokenWithFeesV3 string = "0x5e0c7ee3"

	MaxTimestampParamRange = 604800 // 7 days in seconds

	maxBlockCountForResult  = 10
	defaultUsedGasRatio     = 0.5
	zeroHex32Bytes          = "0x0000000000000000000000000000000000000000000000000000000000000000"
	blockRangeLimit         = 1000
	redirectBytecodePrefix  = "6080604052348015600f57600080fd5b506000610167905077618dc65e"
	redirectBytecodePostfix = "600052366000602037600080366018016008845af43d806000803e8160008114605857816000f35b816000fdfea2646970667358221220d8378feed472ba49a0005514ef7087017f707b45fb9bf56bb81bb93ff19a238b64736f6c634300080b0033"
	iHTSAddress             = "0x0000000000000000000000000000000000000167"
)

var HTSCreateFuncSelectors = map[string]struct{}{
	CreateFungibleTokenV1:            {},
	CreateFungibleTokenV2:            {},
	CreateFungibleTokenV3:            {},
	CreateFungibleTokenWithFeesV1:    {},
	CreateFungibleTokenWithFeesV2:    {},
	CreateFungibleTokenWithFeesV3:    {},
	CreateNonFungibleTokenV1:         {},
	CreateNonFungibleTokenV2:         {},
	CreateNonFungibleTokenV3:         {},
	CreateNonFungibleTokenWithFeesV1: {},
	CreateNonFungibleTokenWithFeesV2: {},
	CreateNonFungibleTokenWithFeesV3: {},
}

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
	maxBlockRange           = 5 // This is for the transactionCount function
	defaultUsedGasRatio     = 0.5
	zeroHex32Bytes          = "0x0000000000000000000000000000000000000000000000000000000000000000"
	zeroHexAddress          = "0x0000000000000000000000000000000000000000"
	zeroHex                 = "0x0"
	oneHex                  = "0x1"
	blockRangeLimit         = 1000
	redirectBytecodePrefix  = "6080604052348015600f57600080fd5b506000610167905077618dc65e"
	redirectBytecodePostfix = "600052366000602037600080366018016008845af43d806000803e8160008114605857816000f35b816000fdfea2646970667358221220d8378feed472ba49a0005514ef7087017f707b45fb9bf56bb81bb93ff19a238b64736f6c634300080b0033"
	iHTSAddress             = "0x0000000000000000000000000000000000000167"

	DefaultPollingInterval = 500

	EventNewHeads = "newHeads"
	EventLogs     = "logs"

	TinybarToWeibarCoef = 10000000000

	emptyHex        = "0x"
	emptyBloom      = "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	defaultRootHash = "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"

	BalancesUpdateInterval = 900 // 15 minutes in seconds
	LatestBlockTolerance   = 1
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

// Gas-related constants
const (
	// Base gas cost for a transaction
	TxBaseCost = 21000

	// Minimum gas for hollow account creation
	MinTxHollowAccountCreationGas = 587000

	// Average gas for contract calls
	TxContractCallAverageGas = 500000

	// Default gas for unknown transactions
	TxDefaultGas = 400000

	// Extra gas for contract creation
	TxCreateExtra = 32000

	// Gas cost for zero bytes in transaction data
	TxDataZeroCost = 4

	// Gas cost for non-zero bytes in transaction data
	TxDataNonZeroCost = 16

	// Function selector character length (including 0x prefix)
	FunctionSelectorCharLength int = 10

	IstanbulTxDataNonZeroCost = 16
	MaxGasPerSec              = 15000000
	GasPriceTinyBarBuffer     = 1

	GasLimit = 30000000
)

const (
	BloomByteSize = 256
	BloomMask     = 0x7ff
)

const (
	DETERMINISTIC_DEPLOYER_TRANSACTION = "0xf8a58085174876e800830186a08080b853604580600e600039806000f350fe7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe03601600081602082378035828234f58015156039578182fd5b8082525050506014600cf31ba02222222222222222222222222222222222222222222222222222222222222222a02222222222222222222222222222222222222222222222222222222222222222"
	DETERMINISTIC_DEPLOYMENT_SIGNER    = "0x3fab184622dc19b6109349b94811493bf2a45362"
	DETERMINISTIC_PROXY_CONTRACT       = "0x4e59b44847b379578588920ca78fbf26c0b4956c"
)

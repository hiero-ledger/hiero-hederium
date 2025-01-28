package service

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/LimeChain/Hederium/internal/domain"
	infrahedera "github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/LimeChain/Hederium/internal/infrastructure/limiter"
	"go.uber.org/zap"
)

const (
	maxBlockCountForResult = 10
	defaultUsedGasRatio    = 0.5
	zeroHex32Bytes         = "0x0000000000000000000000000000000000000000000000000000000000000000"
	blockRangeLimit        = 1000
)

type EthService struct {
	hClient       infrahedera.HederaNodeClient
	mClient       infrahedera.MirrorNodeClient
	logger        *zap.Logger
	tieredLimiter *limiter.TieredLimiter
	chainId       string
}

func NewEthService(
	hClient infrahedera.HederaNodeClient,
	mClient infrahedera.MirrorNodeClient,
	log *zap.Logger,
	l *limiter.TieredLimiter,
	chainId string,
) *EthService {
	return &EthService{
		hClient:       hClient,
		mClient:       mClient,
		logger:        log,
		tieredLimiter: l,
		chainId:       chainId,
	}
}

// GetBlockNumber retrieves the latest block number from the Hedera network and returns it
// in hexadecimal format, compatible with Ethereum JSON-RPC specifications.
// It returns two values:
//   - interface{}: A hex string representing the block number (e.g., "0x1234") on success,
//     or nil on failure
//   - map[string]interface{}: Error details if the operation fails, nil on success.
//     Error format follows Ethereum JSON-RPC error specifications.
func (s *EthService) GetBlockNumber() (interface{}, map[string]interface{}) {
	s.logger.Info("Getting block number")
	block, err := s.mClient.GetLatestBlock()
	if err != nil {
		s.logger.Error("Failed to fetch latest block", zap.Error(err))
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to fetch block data",
		}
	}

	s.logger.Debug("Received block data", zap.Any("block", block))

	if blockNumber, ok := block["number"].(float64); ok {
		s.logger.Debug("Found block number", zap.Float64("blockNumber", blockNumber))

		blockNum := uint64(blockNumber)
		hexBlockNum := "0x" + strconv.FormatUint(blockNum, 16)
		s.logger.Debug("Successfully converted to hex", zap.String("hexBlockNum", hexBlockNum))
		s.logger.Info("Successfully returned block number", zap.String("blockNumber", hexBlockNum))
		return hexBlockNum, nil
	}

	s.logger.Error("Block number not found or invalid type", zap.Any("block", block))
	return nil, map[string]interface{}{
		"code":    -32000,
		"message": "Invalid block data",
	}
}

// GetGasPrice returns the current gas price in wei with a 10% buffer added.
// The gas price is fetched from the network in tinybars, converted to weibars,
// and returned as a hex string with "0x" prefix.
func (s *EthService) GetGasPrice() (interface{}, map[string]interface{}) {
	s.logger.Info("Getting gas price")

	timestampTo := "" // We pass empty, because we want gas from latest block
	order := ""

	weibars, err := GetFeeWeibars(s, timestampTo, order)
	if err != nil {
		errMsg := "Failed to fetch gas price"
		s.logger.Error(errMsg)
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": errMsg,
		}
	}
	// Add 10% buffer to the gas price
	buffer := new(big.Int).Div(weibars, big.NewInt(10))
	weibars.Add(weibars, buffer)
	gasPrice := "0x" + strconv.FormatUint(weibars.Uint64(), 16)

	s.logger.Info("Successfully returned gas price", zap.String("gasPrice", gasPrice))
	return gasPrice, nil
}

// GetChainId returns the network's chain ID as configured in the service.
// The chain ID is returned as a hex string with "0x" prefix.
func (s *EthService) GetChainId() (interface{}, map[string]interface{}) {
	s.logger.Info("Getting chain ID")
	s.logger.Info("Returning chain ID", zap.String("chainId", s.chainId))
	return s.chainId, nil
}

// retrieves a block by its hash and optionally includes detailed transaction information.
// Parameters:
//   - hash: The hash of the block to retrieve
//   - showDetails: If true, returns full transaction objects; if false, only transaction hashes
//
// Returns nil for both return values if the block is not found.
func (s *EthService) GetBlockByHash(hash string, showDetails bool) (interface{}, map[string]interface{}) {
	s.logger.Info("Getting block by hash", zap.String("hash", hash), zap.Bool("showDetails", showDetails))
	block := s.mClient.GetBlockByHashOrNumber(hash)
	if block == nil {
		return nil, nil
	}
	return ProcessBlock(s, block, showDetails)
}

// GetBlockByHash retrieves a block by its hash from the Hedera network and returns it
// in an Ethereum-compatible format.
//
// Parameters:
//   - hash: The hash of the block to retrieve
//   - showDetails: If true, includes full transaction details in the response.
//     If false, only includes transaction hashes.
//
// Returns:
//   - interface{}: The block data in Ethereum format (*domain.Block), or nil if not found
//   - map[string]interface{}: Error information if any occurred, nil otherwise
func (s *EthService) GetBlockByNumber(numberOrTag string, showDetails bool) (interface{}, map[string]interface{}) {
	s.logger.Info("Getting block by number", zap.String("numberOrTag", numberOrTag), zap.Bool("showDetails", showDetails))
	var blockNumber string
	switch numberOrTag {
	case "latest", "pending":
		latestBlock, errMap := s.GetBlockNumber()
		if errMap != nil {
			s.logger.Debug("Failed to get latest block number")
			return nil, nil
		}

		latestBlockStr, ok := latestBlock.(string)
		if !ok {
			s.logger.Debug("Invalid block number format")
			return nil, nil
		}

		// Convert hex string to int, remove "0x" prefix
		latestBlockNum, err := strconv.ParseInt(latestBlockStr[2:], 16, 64)
		if err != nil {
			s.logger.Debug("Failed to parse latest block number", zap.Error(err))
			return nil, nil
		}
		blockNumber = strconv.FormatInt(latestBlockNum, 10)
	case "earliest":
		blockNumber = "0"
	default:
		// If it's a hex number, convert it to decimal
		if strings.HasPrefix(numberOrTag, "0x") {
			num, err := strconv.ParseInt(numberOrTag[2:], 16, 64)
			if err != nil {
				s.logger.Debug("Failed to parse block number", zap.Error(err))
				return nil, nil
			}
			blockNumber = strconv.FormatInt(num, 10)
		} else {
			blockNumber = numberOrTag
		}
	}

	block := s.mClient.GetBlockByHashOrNumber(blockNumber)
	if block == nil {
		return nil, nil
	}

	return ProcessBlock(s, block, showDetails)
}

func (s *EthService) GetBalance(address string, blockNumberTagOrHash string) string {
	s.logger.Info("Getting balance", zap.String("address", address), zap.String("blockNumberTagOrHash", blockNumberTagOrHash))
	var block *domain.BlockResponse

	switch blockNumberTagOrHash {
	case "latest", "pending":
		latestBlock, errMap := s.GetBlockNumber()
		if errMap != nil {
			s.logger.Debug("Failed to get latest block number")
			return "0x0"
		}

		latestBlockStr, ok := latestBlock.(string)
		if !ok {
			s.logger.Debug("Invalid block number format")
			return "0x0"
		}

		// Convert hex string to int, remove "0x" prefix
		latestBlockNum, err := strconv.ParseInt(latestBlockStr[2:], 16, 64)
		if err != nil {
			s.logger.Debug("Failed to parse latest block number", zap.Error(err))
			return "0x0"
		}
		block = s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(latestBlockNum, 10))
		if block == nil {
			s.logger.Debug("Latest block not found")
			return "0x0"
		}
	case "earliest":
		block = s.mClient.GetBlockByHashOrNumber("0")
		if block == nil {
			s.logger.Debug("Earliest block not found")
			return "0x0"
		}
	default:
		// Check if it's a 32 byte hash (0x + 64 hex chars)
		if len(blockNumberTagOrHash) == 66 && strings.HasPrefix(blockNumberTagOrHash, "0x") {
			block = s.mClient.GetBlockByHashOrNumber(blockNumberTagOrHash)
			if block == nil {
				s.logger.Debug("Block not found for hash", zap.String("hash", blockNumberTagOrHash))
				return "0x0"
			}
		} else if strings.HasPrefix(blockNumberTagOrHash, "0x") {
			// If it's a hex number, convert it to decimal
			num, err := strconv.ParseInt(blockNumberTagOrHash[2:], 16, 64)
			if err != nil {
				s.logger.Debug("Failed to parse block number", zap.Error(err))
				return "0x0"
			}
			block = s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(num, 10))
			if block == nil {
				s.logger.Debug("Block not found for number", zap.String("number", blockNumberTagOrHash))
				return "0x0"
			}
		} else {
			block = s.mClient.GetBlockByHashOrNumber(blockNumberTagOrHash)
			if block == nil {
				s.logger.Debug("Block not found for number", zap.String("number", blockNumberTagOrHash))
				return "0x0"
			}
		}
	}
	balance := s.mClient.GetBalance(address, block.Timestamp.To)

	return balance
}

func (s *EthService) GetTransactionCount(address string, blockNumberOrTag string) string {
	s.logger.Info("Getting transaction count", zap.String("address", address), zap.String("blockNumberOrTag", blockNumberOrTag))
	// TODO: This whole flow, could be optimized.

	var block *domain.BlockResponse
	requestingLatest := false
	switch blockNumberOrTag {
	case "latest", "pending":
		requestingLatest = true
		latestBlock, errMap := s.GetBlockNumber()
		if errMap != nil {
			s.logger.Debug("Failed to get latest block number")
			return "0x0"
		}

		latestBlockStr, ok := latestBlock.(string)
		if !ok {
			s.logger.Debug("Invalid block number format")
			return "0x0"
		}

		// Convert hex string to int, remove "0x" prefix
		latestBlockNum, err := strconv.ParseInt(latestBlockStr[2:], 16, 64)
		if err != nil {
			s.logger.Debug("Failed to parse latest block number", zap.Error(err))
			return "0x0"
		}
		block = s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(latestBlockNum, 10))

	case "earliest":
		block = s.mClient.GetBlockByHashOrNumber("0")
	default:

		latestBlock, errMap := s.GetBlockNumber()
		if errMap != nil {
			s.logger.Debug("Failed to get latest block number")
			return "0x0"
		}

		latestBlockStr, ok := latestBlock.(string)
		if !ok {
			s.logger.Debug("Invalid block number format")
			return "0x0"
		}

		// Convert hex string to int, remove "0x" prefix
		latestBlockNum, err := strconv.ParseInt(latestBlockStr[2:], 16, 64)
		if err != nil {
			s.logger.Debug("Failed to parse latest block number", zap.Error(err))
			return "0x0"
		}

		// If it's a hex number, convert it to decimal
		num, err := strconv.ParseInt(blockNumberOrTag[2:], 16, 64)
		if err != nil {
			s.logger.Debug("Failed to parse block number", zap.Error(err))
			return "0x0"
		}
		if num+10 > latestBlockNum {
			requestingLatest = true
		}
		block = s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(num, 10))

	}

	if block == nil {
		return "0x0"
	}
	account := s.mClient.GetAccount(address, block.Timestamp.To)
	if account == nil {
		return "0x0"
	}
	accountResponse := account.(domain.AccountResponse)

	if requestingLatest {
		return "0x" + strconv.FormatUint(uint64(accountResponse.EthereumNonce), 16)
	}

	contractResult := s.mClient.GetContractResult(accountResponse.Transactions[0].TransactionId)
	if contractResult == nil {
		return "0x0"
	}
	contractResultResponse := contractResult.(domain.ContractResultResponse)
	nonce := "0x" + strconv.FormatUint(uint64(contractResultResponse.Nonce+1), 16) // We add 1 here, because of the nature nonce is incremented.

	s.logger.Info("Returning nonce", zap.String("nonce", nonce), zap.String("address", address))
	return nonce
}

func (s *EthService) EstimateGas(transaction interface{}, blockParam interface{}) (string, map[string]interface{}) {
	s.logger.Info("Estimating gas", zap.Any("transaction", transaction))
	errorObject := map[string]interface{}{
		"code":    -32000,
		"message": "Error encountered while estimating gas",
	}

	txObj, err := ParseTransactionCallObject(s, transaction)
	if err != nil {
		return "0x0", errorObject
	}

	formatResult, err := FormatTransactionCallObject(s, txObj, blockParam, true)
	if err != nil {
		return "0x0", errorObject
	}

	callResult := s.mClient.PostCall(formatResult)
	if callResult == nil {
		return "0x0", errorObject
	}

	// Remove leading zeros from the result string
	result := NormalizeHexString(callResult.(string))

	s.logger.Info("Returning gas", zap.Any("gas", result))
	return result, nil
}

func (s *EthService) Call(transaction interface{}, blockParam interface{}) (interface{}, map[string]interface{}) {
	s.logger.Info("Performing eth_call", zap.Any("transaction", transaction))
	errorObject := map[string]interface{}{
		"code":    -32000,
		"message": "Error encountered while performing eth_call",
	}

	txObj, err := ParseTransactionCallObject(s, transaction)
	if err != nil {
		return "0x0", errorObject
	}

	result, err := FormatTransactionCallObject(s, txObj, blockParam, false)
	if err != nil {
		return "0x0", errorObject
	}

	callResult := s.mClient.PostCall(result)
	if callResult == nil {
		return "0x0", errorObject
	}

	normalizedResult := NormalizeHexString(callResult.(string))

	s.logger.Info("Returning formatted transaction call result", zap.Any("result", normalizedResult))
	return normalizedResult, nil
}

func (s *EthService) GetTransactionByHash(hash string) interface{} {
	s.logger.Info("Getting transaction by hash", zap.String("hash", hash))
	contractResult := s.mClient.GetContractResult(hash)
	if contractResult == nil {
		// TODO: Here we should handle synthetic transactions
		return nil
	}
	contractResultResponse := contractResult.(domain.ContractResultResponse)

	// TODO: Resolve evm addresses
	transaction := ProcessTransactionResponse(contractResultResponse)

	return transaction
}

func (s *EthService) GetTransactionReceipt(hash string) interface{} {
	s.logger.Info("Getting transaction receipt", zap.String("hash", hash))

	contractResult := s.mClient.GetContractResult(hash)
	if contractResult == nil {
		// TODO: Here we should handle synthetic transactions
		return nil
	}
	contractResultResponse := contractResult.(domain.ContractResultResponse)

	// Convert logs
	logs := make([]domain.Log, len(contractResultResponse.Logs))
	for i, log := range contractResultResponse.Logs {
		logs[i] = domain.Log{
			Address:          log.Address,
			BlockHash:        contractResultResponse.BlockHash[:66],
			BlockNumber:      "0x" + strconv.FormatInt(contractResultResponse.BlockNumber, 16),
			Data:             log.Data,
			LogIndex:         "0x" + strconv.FormatInt(int64(i), 16),
			Removed:          false,
			Topics:           log.Topics,
			TransactionHash:  hash,
			TransactionIndex: "0x" + strconv.FormatInt(int64(contractResultResponse.TransactionIndex), 16),
		}
	}

	// TODO: Check if the address is a system contract here
	// Default values
	const emptyHex = "0x"
	const emptyBloom = "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	const defaultRootHash = "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
	// TODO: Check revert reason, if matches error_message, return it, else it's ASCII so make it hex and return then
	// TODO: Implement resolveEvmAddress for from/to addresses

	// Create receipt
	// TODO: add utility function to convert to hex
	receipt := domain.TransactionReceipt{
		BlockHash:         contractResultResponse.BlockHash[:66],
		BlockNumber:       "0x" + strconv.FormatInt(contractResultResponse.BlockNumber, 16),
		From:              contractResultResponse.From, // TODO: resolve EVM address
		To:                contractResultResponse.To,   // TODO: resolve EVM address
		CumulativeGasUsed: "0x" + strconv.FormatInt(contractResultResponse.BlockGasUsed, 16),
		GasUsed:           "0x" + strconv.FormatInt(contractResultResponse.GasUsed, 16),
		ContractAddress:   contractResultResponse.Address, // TODO: Set if contract creation
		Logs:              logs,
		LogsBloom: func() string {
			if contractResultResponse.Bloom == emptyHex {
				return emptyBloom
			}
			return contractResultResponse.Bloom
		}(),
		TransactionHash:   hash,
		TransactionIndex:  "0x" + strconv.FormatInt(int64(contractResultResponse.TransactionIndex), 16),
		EffectiveGasPrice: "0x" + contractResultResponse.GasPrice, // TODO: Calculate effective gas price if needed
		Root:              defaultRootHash,
		Status:            contractResultResponse.Status,
		Type: func() *string {
			if contractResultResponse.Type == nil {
				return nil
			}
			hexType := "0x" + strconv.FormatInt(int64(*contractResultResponse.Type), 16)
			return &hexType
		}(),
	}

	s.logger.Info("Returning transaction receipt", zap.Any("receipt", receipt))
	return receipt
}

func (s *EthService) FeeHistory(blockCount string, newestBlock string, rewardPercentiles []string) (interface{}, map[string]interface{}) {
	s.logger.Info("Getting fee history", zap.String("blockCount", blockCount), zap.String("newestBlock", newestBlock), zap.Any("rewardPercentiles", rewardPercentiles))

	//Get the block number of the newest block
	latestBlockNumber, err := s.GetBlockNumber()
	if err != nil {
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to get latest block number1",
		}
	}

	latestBlockHex, ok := latestBlockNumber.(string)
	if !ok {
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to parse latest block number2",
		}
	}
	latestBlockInt, errMap := HexToDec(latestBlockHex)
	if errMap != nil {
		return nil, errMap
	}
	newestBlockNumber, errMap := s.getBlockNumberByHashOrTag(newestBlock)
	if errMap != nil {
		return nil, errMap
	}

	newestBlockInt, ok := newestBlockNumber.(int64)
	if !ok {
		return nil, errMap
	}

	//Convert the block number to decimal
	blockCountInt, errMap := HexToDec(blockCount)
	if errMap != nil {
		return nil, errMap
	}

	//Check if the blockCount is greater then the one we need
	if blockCountInt > int64(maxBlockCountForResult) {
		blockCountInt = int64(maxBlockCountForResult)
	}

	if newestBlockInt > latestBlockInt {
		newestBlockInt = latestBlockInt
	}

	oldestBlockInt := newestBlockInt - blockCountInt + 1

	fixed_Fee := false // TODO: see how to get this fixed_Fee
	if fixed_Fee {
		if oldestBlockInt < 0 {
			blockCountInt = 1
			oldestBlockInt = 1
		}
		fee, errMap := s.GetGasPrice()
		if errMap != nil {
			return nil, errMap
		}
		feeHex, ok := fee.(string)
		if !ok {
			return nil, map[string]interface{}{
				"code":    -32000,
				"message": "Failed to parse fee",
			}
		}

		feeHistory := s.getRepeatedFeeHistory(blockCountInt, oldestBlockInt, rewardPercentiles, feeHex)
		return feeHistory, nil
	}

	feeHistory, errMap := s.getFeeHistory(blockCountInt, newestBlockInt, latestBlockInt, rewardPercentiles)
	if errMap != nil {
		return nil, errMap
	}

	return feeHistory, nil
}

func (s *EthService) GetStorageAt(address, slot, blockNumberOrHash string) (interface{}, map[string]interface{}) {
	s.logger.Info("Getting storage at", zap.String("address", address), zap.String("slot", slot), zap.String("blockNumberOrHash", blockNumberOrHash))
	blockInt, errMap := s.getBlockNumberByHashOrTag(blockNumberOrHash)
	if errMap != nil {
		return nil, errMap
	}

	blockResponse := s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(blockInt.(int64), 10))

	if blockResponse == nil {
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to get block data",
		}
	}

	timestampTo := blockResponse.Timestamp.To

	result, err := s.mClient.GetContractStateByAddressAndSlot(address, slot, timestampTo)
	if err != nil {
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to get storage data",
		}
	}

	if result == nil || len(result.State) == 0 {
		s.logger.Info("Returning default storage value")
		return zeroHex32Bytes, nil // Default value
	}
	s.logger.Info("Returning storage", zap.Any("storage", result))

	return result.State[0].Value, nil
}

func (s *EthService) GetLogs(logParams domain.LogParams) (interface{}, map[string]interface{}) {
	s.logger.Info("Getting logs", zap.Any("logParams", logParams))
	params := make(map[string]interface{})

	if logParams.BlockHash != "" {
		if !s.validateBlockHashAndAddTimestampToParams(params, logParams.BlockHash) {
			return []domain.Log{}, nil
		}
	} else {
		if !s.validateBlockRangeAndAddTimestampToParams(params, logParams.FromBlock, logParams.ToBlock, logParams.Address) {
			return []domain.Log{}, nil
		}
	}

	if logParams.Topics != nil {
		for i, topic := range logParams.Topics {
			if topic != "" {
				params[fmt.Sprintf("topic%d", i)] = topic
			}
		}
	}

	s.logger.Debug("Received log parameters", zap.Any("params", params))

	logs, err := s.getLogsWithParams(logParams.Address, params)
	if err != nil {
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to get logs",
		}
	}

	return logs, nil
}

// GetAccounts returns an empty array of accounts, similar to Infura's implementation
func (s *EthService) GetAccounts() (interface{}, map[string]interface{}) {
	s.logger.Info("Getting accounts")
	s.logger.Debug("Returning empty accounts array as per specification")
	return []string{}, nil
}

// Syncing returns false, because the Hedera network does not support syncing
func (s *EthService) Syncing() (interface{}, map[string]interface{}) {
	s.logger.Info("Syncing")
	s.logger.Debug("Returning false as per specification")
	return false, nil
}

// Mining returns false, because the Hedera network does not support mining
func (s *EthService) Mining() (interface{}, map[string]interface{}) {
	s.logger.Info("Mining")
	s.logger.Debug("Returning false as per specification")
	return false, nil
}

// MaxPriorityFeePerGas returns 0x0, because the Hedera network does not support it
func (s *EthService) MaxPriorityFeePerGas() (interface{}, map[string]interface{}) {
	s.logger.Info("MaxPriorityFeePerGas")
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// Hashrate returns 0x0, because the Hedera network does not support it
func (s *EthService) Hashrate() (interface{}, map[string]interface{}) {
	s.logger.Info("Hashrate")
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// GetUncleCountByBlockNumber returns 0x0, because the Hedera network does not support it
func (s *EthService) GetUncleCountByBlockNumber() (interface{}, map[string]interface{}) {
	s.logger.Info("GetUncleCountByBlockNumber")
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// GetUncleByBlockNumberAndIndex returns nil, because the Hedera network does not support it
func (s *EthService) GetUncleByBlockNumberAndIndex() (interface{}, map[string]interface{}) {
	s.logger.Info("GetUncleByBlockNumberAndIndex")
	s.logger.Debug("Returning nil as per specification")
	return nil, nil
}

// GetUncleCountByBlockHash returns 0x0, because the Hedera network does not support it
func (s *EthService) GetUncleCountByBlockHash() (interface{}, map[string]interface{}) {
	s.logger.Info("GetUncleCountByBlockHash")
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// GetUncleByBlockHashAndIndex returns nil, because the Hedera network does not support it
func (s *EthService) GetUncleByBlockHashAndIndex() (interface{}, map[string]interface{}) {
	s.logger.Info("GetUncleByBlockHashAndIndex")
	s.logger.Debug("Returning nil as per specification")
	return nil, nil
}

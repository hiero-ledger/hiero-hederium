package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	infrahedera "github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/LimeChain/Hederium/internal/infrastructure/limiter"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
)

// TODO: Refactor the EthService struct.
// Decide which methods should be private, public,
// and if any should be helper functions.
type EthServicer interface {
	Call(transaction interface{}, blockParam interface{}) (interface{}, *domain.RPCError)
	EstimateGas(transaction interface{}, blockParam interface{}) (string, *domain.RPCError)
	FeeHistory(blockCount string, newestBlock string, rewardPercentiles []string) (interface{}, *domain.RPCError)
	GetAccounts() (interface{}, *domain.RPCError)
	GetBalance(address string, blockNumberTagOrHash string) string
	GetBlockByHash(hash string, showDetails bool) (interface{}, *domain.RPCError)
	GetBlockByNumber(numberOrTag string, showDetails bool) (interface{}, *domain.RPCError)
	GetBlockNumber() (interface{}, *domain.RPCError)
	GetBlockTransactionCountByHash(blockHash string) (interface{}, *domain.RPCError)
	GetBlockTransactionCountByNumber(blockNumberOrTag string) (interface{}, *domain.RPCError)
	GetChainId() (interface{}, *domain.RPCError)
	GetCode(address string, blockNumberOrTag string) (interface{}, *domain.RPCError)
	GetGasPrice() (interface{}, *domain.RPCError)
	GetLogs(logParams domain.LogParams) (interface{}, *domain.RPCError)
	GetStorageAt(address string, slot string, blockNumberOrHash string) (interface{}, *domain.RPCError)
	GetTransactionByBlockHashAndIndex(blockHash string, txIndex string) (interface{}, *domain.RPCError)
	GetTransactionByBlockNumberAndIndex(blockNumberOrTag string, txIndex string) (interface{}, *domain.RPCError)
	GetTransactionByHash(hash string) (interface{}, *domain.RPCError)
	GetTransactionCount(address string, blockNumberOrTag string) string
	GetTransactionReceipt(hash string) (interface{}, *domain.RPCError)
	GetUncleByBlockHashAndIndex(blockHash string, index string) (interface{}, *domain.RPCError)
	GetUncleByBlockNumberAndIndex(blockNumber string, index string) (interface{}, *domain.RPCError)
	GetUncleCountByBlockHash(blockHash string) (interface{}, *domain.RPCError)
	GetUncleCountByBlockNumber(blockNumber string) (interface{}, *domain.RPCError)
	Hashrate() (interface{}, *domain.RPCError)
	MaxPriorityFeePerGas() (interface{}, *domain.RPCError)
	Mining() (interface{}, *domain.RPCError)
	ProcessTransactionResponse(contractResult domain.ContractResultResponse) interface{}
	SendRawTransaction(data string) (interface{}, *domain.RPCError)
	Syncing() (interface{}, *domain.RPCError)
}

type EthService struct {
	hClient       infrahedera.HederaNodeClient
	mClient       infrahedera.MirrorNodeClient
	commonService CommonService
	logger        *zap.Logger
	tieredLimiter *limiter.TieredLimiter
	chainId       string
	precheck      Precheck
	cacheService  cache.CacheService
	ctx           context.Context
}

func NewEthService(
	hClient infrahedera.HederaNodeClient,
	mClient infrahedera.MirrorNodeClient,
	commonService CommonService,
	log *zap.Logger,
	l *limiter.TieredLimiter,
	chainId string,
	cacheService cache.CacheService,
) *EthService {
	return &EthService{
		hClient:       hClient,
		mClient:       mClient,
		commonService: commonService,
		logger:        log,
		tieredLimiter: l,
		chainId:       chainId,
		precheck:      NewPrecheck(mClient, log, chainId),
		cacheService:  cacheService,
		ctx:           context.Background(),
	}
}

// GetBlockNumber retrieves the latest block number from the Hedera network and returns it
// in hexadecimal format, compatible with Ethereum JSON-RPC specifications.
// It returns two values:
//   - interface{}: A hex string representing the block number (e.g., "0x1234") on success,
//     or nil on failure
//   - map[string]interface{}: Error details if the operation fails, nil on success.
//     Error format follows Ethereum JSON-RPC error specifications.
func (s *EthService) GetBlockNumber() (interface{}, *domain.RPCError) {
	var cachedBlockNumber string
	err := s.cacheService.Get(s.ctx, GetBlockNumber, &cachedBlockNumber)
	if err == nil && cachedBlockNumber != "" {
		s.logger.Info("Block number fetched from cache", zap.String("blockNumber", cachedBlockNumber))
		return cachedBlockNumber, nil
	}

	blockNumber, err := s.commonService.GetBlockNumber()

	if err := s.cacheService.Set(s.ctx, GetBlockNumber, blockNumber, ShortExpiration); err != nil {
		s.logger.Debug("Failed to cache block number", zap.Error(err))
	}

	return blockNumber, nil
}

// GetGasPrice returns the current gas price in wei with a 10% buffer added.
// The gas price is fetched from the network in tinybars, converted to weibars,
// and returned as a hex string with "0x" prefix.
func (s *EthService) GetGasPrice() (interface{}, *domain.RPCError) {
	s.logger.Info("Getting gas price")

	cacheKey := GetGasPrice

	var cachedPrice string
	err := s.cacheService.Get(s.ctx, cacheKey, &cachedPrice)
	if err == nil && cachedPrice != "" {
		s.logger.Info("Gas price fetched from cache", zap.Any("gasPrice", cachedPrice))
		return cachedPrice, nil
	}

	timestampTo := "" // We pass empty, because we want gas from latest block
	order := ""

	weibars, err := GetFeeWeibars(s, timestampTo, order)
	if err != nil {
		s.logger.Error("Failed to fetch gas price", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to fetch gas price")
	}

	gasPrice := fmt.Sprintf("0x%x", weibars)

	if err := s.cacheService.Set(s.ctx, cacheKey, gasPrice, DefaultExpiration); err != nil {
		s.logger.Debug("Failed to cache gas price", zap.Error(err))
	}

	s.logger.Info("Successfully returned gas price", zap.String("gasPrice", gasPrice))
	return gasPrice, nil
}

// GetChainId returns the network's chain ID as configured in the service.
// The chain ID is returned as a hex string with "0x" prefix.
func (s *EthService) GetChainId() (interface{}, *domain.RPCError) {
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
func (s *EthService) GetBlockByHash(hash string, showDetails bool) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting block by hash", zap.String("hash", hash), zap.Bool("showDetails", showDetails))

	cacheKey := fmt.Sprintf("%s_%s_%t", GetBlockByHash, hash, showDetails)

	var cachedBlock domain.Block
	if err := s.cacheService.Get(s.ctx, cacheKey, &cachedBlock); err == nil && cachedBlock.Hash != nil {
		s.logger.Info("Block fetched from cache", zap.Any("block", cachedBlock))
		return cachedBlock, nil
	}

	block := s.mClient.GetBlockByHashOrNumber(hash)
	if block == nil {
		return nil, nil
	}

	processedBlock, err := ProcessBlock(s, block, showDetails)
	if err != nil {
		s.logger.Error("Failed to process block", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to process block")
	}

	if err := s.cacheService.Set(s.ctx, cacheKey, &processedBlock, DefaultExpiration); err != nil {
		s.logger.Debug("Failed to cache block", zap.Error(err))
	}

	return processedBlock, nil
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
func (s *EthService) GetBlockByNumber(numberOrTag string, showDetails bool) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting block by number", zap.String("numberOrTag", numberOrTag), zap.Bool("showDetails", showDetails))

	blockNumberInt, errRpc := s.commonService.GetBlockNumberByNumberOrTag(numberOrTag)
	if errRpc != nil {
		return nil, errRpc
	}

	cachedKey := fmt.Sprintf("%s_%d_%t", GetBlockByNumber, blockNumberInt, showDetails)

	var cachedBlock domain.Block
	if err := s.cacheService.Get(s.ctx, cachedKey, &cachedBlock); err == nil && cachedBlock.Hash != nil {
		s.logger.Info("Block fetched from cache", zap.Any("block", cachedBlock))
		return &cachedBlock, nil
	}

	block := s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(blockNumberInt, 10))
	if block == nil {
		return nil, nil
	}

	processedBlock, err := ProcessBlock(s, block, showDetails)
	if err != nil {
		s.logger.Error("Failed to process block", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to process block")
	}

	if err := s.cacheService.Set(s.ctx, cachedKey, &processedBlock, DefaultExpiration); err != nil {
		s.logger.Debug("Failed to cache block", zap.Error(err))
	}

	return processedBlock, nil
}

// TODO: Add error handling
func (s *EthService) GetBalance(address string, blockNumberTagOrHash string) string {
	s.logger.Info("Getting balance", zap.String("address", address), zap.String("blockNumberTagOrHash", blockNumberTagOrHash))

	var block *domain.BlockResponse

	switch blockNumberTagOrHash {
	case "latest", "pending":
		balance := s.mClient.GetBalance(address, "0")
		return balance
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

// TODO: Add error handling
func (s *EthService) GetTransactionCount(address string, blockNumberOrTag string) string {
	s.logger.Info("Getting transaction count", zap.String("address", address), zap.String("blockNumberOrTag", blockNumberOrTag))

	blockNumberInt, errRpc := s.commonService.GetBlockNumberByNumberOrTag(blockNumberOrTag)
	if errRpc != nil {
		return "0x0"
	}

	requestingLatest := s.isLatestBlockRequest(blockNumberOrTag, blockNumberInt)

	block := s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(blockNumberInt, 10))

	if block == nil {
		return "0x0"
	}

	account := s.mClient.GetAccount(address, block.Timestamp.To)
	if account == nil {
		return "0x0"
	}
	accountResponse := account.(domain.AccountResponse)

	if requestingLatest {
		return fmt.Sprintf("0x%x", accountResponse.EthereumNonce)
	}

	if len(accountResponse.Transactions) == 0 {
		return "0x0"
	}

	contractResult := s.mClient.GetContractResult(accountResponse.Transactions[0].TransactionId)
	if contractResult == nil {
		return "0x0"
	}
	contractResultResponse := contractResult.(domain.ContractResultResponse)

	nonce := fmt.Sprintf("0x%x", contractResultResponse.Nonce+1) // We add 1 here, because of the nature nonce is incremented.

	s.logger.Info("Returning nonce", zap.String("nonce", nonce), zap.String("address", address))
	return nonce
}

func (s *EthService) EstimateGas(transaction interface{}, blockParam interface{}) (string, *domain.RPCError) {
	s.logger.Info("Estimating gas", zap.Any("transaction", transaction))

	txObj, err := ParseTransactionCallObject(s, transaction)
	if err != nil {
		s.logger.Error("Failed to parse transaction call object", zap.Error(err))
		return "0x0", domain.NewRPCError(domain.ServerError, "Failed to parse transaction call object")
	}

	formatResult, err := FormatTransactionCallObject(s, txObj, blockParam, true)
	if err != nil {
		s.logger.Error("Failed to format transaction call object", zap.Error(err))
		return "0x0", domain.NewRPCError(domain.ServerError, "Failed to format transaction call object")
	}

	callResult := s.mClient.PostCall(formatResult)
	if callResult == nil {
		s.logger.Error("Failed to post call", zap.Error(err))
		return "0x0", domain.NewRPCError(domain.ServerError, "Failed to post call")
	}

	// Remove leading zeros from the result string
	result := NormalizeHexString(callResult.(string))

	s.logger.Info("Returning gas", zap.Any("gas", result))
	return result, nil
}

func (s *EthService) Call(transaction interface{}, blockParam interface{}) (interface{}, *domain.RPCError) {
	s.logger.Info("Performing eth_call", zap.Any("transaction", transaction))

	txObj, err := ParseTransactionCallObject(s, transaction)
	if err != nil {
		s.logger.Error("Failed to parse transaction call object", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to parse transaction call object")
	}

	result, err := FormatTransactionCallObject(s, txObj, blockParam, false)
	if err != nil {
		s.logger.Error("Failed to format transaction call object", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to format transaction call object")
	}

	callResult := s.mClient.PostCall(result)
	if callResult == nil {
		s.logger.Error("Failed to post call", zap.Error(err))
		return "0x0", domain.NewRPCError(domain.ServerError, "Failed to post call")
	}

	s.logger.Info("Returning transaction call result", zap.Any("result", callResult))
	return callResult, nil
}

func (s *EthService) GetTransactionByHash(hash string) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting transaction by hash", zap.String("hash", hash))

	cacheKey := fmt.Sprintf("%s_%s", GetTransactionByHash, hash)

	var cachedTx interface{}
	if err := s.cacheService.Get(s.ctx, cacheKey, &cachedTx); err == nil && cachedTx != nil {
		s.logger.Info("Transaction fetched from cache", zap.Any("transaction", cachedTx))
		return cachedTx, nil
	}
	contractResult := s.mClient.GetContractResult(hash)

	if contractResult == nil {
		// TODO: Here we should handle synthetic transactions
		return nil, nil
	}
	contractResultResponse := contractResult.(domain.ContractResultResponse)

	transaction := s.ProcessTransactionResponse(contractResultResponse)

	if err := s.cacheService.Set(s.ctx, cacheKey, &transaction, DefaultExpiration); err != nil {
		s.logger.Debug("Failed to cache transaction", zap.Error(err))
	}

	return transaction, nil
}

func (s *EthService) GetTransactionReceipt(hash string) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting transaction receipt", zap.String("hash", hash))

	cacheKey := fmt.Sprintf("%s_%s", GetTransactionReceipt, hash)

	var cachedReceipt interface{}
	if err := s.cacheService.Get(s.ctx, cacheKey, &cachedReceipt); err == nil && cachedReceipt != nil {
		s.logger.Info("Transaction receipt fetched from cache", zap.Any("receipt", cachedReceipt))
		return cachedReceipt, nil
	}

	contractResult := s.mClient.GetContractResult(hash)
	if contractResult == nil {
		// TODO: Here we should handle synthetic transactions
		return nil, nil
	}
	contractResultResponse := contractResult.(domain.ContractResultResponse)

	// Convert logs
	logs := make([]domain.Log, len(contractResultResponse.Logs))
	for i, log := range contractResultResponse.Logs {
		logs[i] = domain.Log{
			Address:          log.Address,
			BlockHash:        contractResultResponse.BlockHash[:66],
			BlockNumber:      hexify(contractResultResponse.BlockNumber),
			Data:             log.Data,
			LogIndex:         hexify(int64(i)),
			Removed:          false,
			Topics:           log.Topics,
			TransactionHash:  hash,
			TransactionIndex: hexify(int64(contractResultResponse.TransactionIndex)),
		}
	}

	// Default values
	const emptyHex = "0x"
	const emptyBloom = "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	const defaultRootHash = "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"

	evmAddressFrom, err := s.resolveEvmAddress(contractResultResponse.From)
	if err != nil {
		s.logger.Error("Failed to resolve EVM address for from", zap.Any("error", err))
	}

	evmAddressTo, err := s.resolveEvmAddress(contractResultResponse.To)
	if err != nil {
		s.logger.Error("Failed to resolve EVM address for to", zap.Any("error", err))
	}

	effectiveGasPrice, err := s.getCurrentGasPriceForBlock(contractResultResponse.BlockHash[:66])
	if err != nil {
		s.logger.Error("Failed to get gas price for block", zap.Any("error", err))
	}

	logsBloom := contractResultResponse.Bloom
	if logsBloom == emptyHex {
		logsBloom = emptyBloom
	}

	var contractType *string
	if contractResultResponse.Type != nil {
		hexType := hexify(int64(*contractResultResponse.Type))
		contractType = &hexType
	}

	contractAddress := s.getContractAddressFromReceipt(contractResultResponse)

	// Create receipt
	receipt := domain.TransactionReceipt{
		BlockHash:         contractResultResponse.BlockHash[:66],
		BlockNumber:       hexify(contractResultResponse.BlockNumber),
		From:              *evmAddressFrom,
		To:                *evmAddressTo,
		CumulativeGasUsed: hexify(contractResultResponse.BlockGasUsed),
		GasUsed:           hexify(contractResultResponse.GasUsed),
		ContractAddress:   contractAddress,
		Logs:              logs,
		LogsBloom:         logsBloom,
		TransactionHash:   hash,
		TransactionIndex:  hexify(int64(contractResultResponse.TransactionIndex)),
		EffectiveGasPrice: effectiveGasPrice,
		Root:              defaultRootHash,
		Status:            contractResultResponse.Status,
		Type:              contractType,
	}

	if contractResultResponse.ErrorMessage != nil {
		if isHexString(*contractResultResponse.ErrorMessage) {
			receipt.RevertReason = *contractResultResponse.ErrorMessage
		} else {
			receipt.RevertReason = fmt.Sprintf("0x%s", hex.EncodeToString([]byte(*contractResultResponse.ErrorMessage)))
		}
	}

	if err := s.cacheService.Set(s.ctx, cacheKey, &receipt, DefaultExpiration); err != nil {
		s.logger.Debug("Failed to cache transaction receipt", zap.Error(err))
	}

	s.logger.Info("Returning transaction receipt", zap.Any("receipt", receipt))
	return receipt, nil
}

func (s *EthService) FeeHistory(blockCount string, newestBlock string, rewardPercentiles []string) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting fee history", zap.String("blockCount", blockCount), zap.String("newestBlock", newestBlock), zap.Any("rewardPercentiles", rewardPercentiles))

	//Get the block number of the newest block
	latestBlockNumber, errRpc := s.GetBlockNumber()
	if errRpc != nil {
		return nil, errRpc
	}

	latestBlockHex, ok := latestBlockNumber.(string)
	if !ok {
		return nil, domain.NewRPCError(domain.ServerError, "Failed to parse latest block number")
	}
	latestBlockInt, err := HexToDec(latestBlockHex)
	if err != nil {
		return nil, domain.NewRPCError(domain.ServerError, fmt.Sprintf("Failed to parse latest block number: %s", err.Error()))
	}
	newestBlockInt, errRpc := s.commonService.GetBlockNumberByNumberOrTag(newestBlock)
	if errRpc != nil {
		return nil, errRpc
	}

	//Convert the block number to decimal
	blockCountInt, err := HexToDec(blockCount)
	if err != nil {
		s.logger.Error("Failed to parse block count", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to parse block count:")
	}

	//Check if the blockCount is greater then the one we need
	if blockCountInt > int64(maxBlockCountForResult) {
		blockCountInt = int64(maxBlockCountForResult)
	}

	if newestBlockInt > latestBlockInt {
		newestBlockInt = latestBlockInt
	}

	oldestBlockInt := newestBlockInt - blockCountInt + 1

	fixed_Fee := true // The nodejs implementation uses this flag to determine if the fee is fixed or not
	if fixed_Fee {
		if oldestBlockInt <= 0 {
			blockCountInt = 1
			oldestBlockInt = 1
		}
		fee, errRpc := s.GetGasPrice()
		if errRpc != nil {
			return nil, errRpc
		}
		feeHex, ok := fee.(string)
		if !ok {
			return nil, domain.NewRPCError(domain.ServerError, "Failed to parse fee")
		}

		feeHistory := s.getRepeatedFeeHistory(blockCountInt, oldestBlockInt, rewardPercentiles, feeHex)
		return feeHistory, nil
	}

	feeHistory, err := s.getFeeHistory(blockCountInt, newestBlockInt, latestBlockInt, rewardPercentiles)
	if err != nil {
		s.logger.Error("Failed to get fee history", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to get fee history:")
	}

	return feeHistory, nil
}

func (s *EthService) GetStorageAt(address, slot, blockNumberOrHash string) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting storage at", zap.String("address", address), zap.String("slot", slot), zap.String("blockNumberOrHash", blockNumberOrHash))
	blockInt, errRpc := s.commonService.GetBlockNumberByNumberOrTag(blockNumberOrHash)
	if errRpc != nil {
		return nil, errRpc
	}

	blockResponse := s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(blockInt, 10))

	if blockResponse == nil {
		return nil, domain.NewRPCError(domain.ServerError, "Failed to get block data")
	}

	timestampTo := blockResponse.Timestamp.To

	result, err := s.mClient.GetContractStateByAddressAndSlot(address, slot, timestampTo)
	if err != nil {
		return nil, domain.NewRPCError(domain.ServerError, fmt.Sprintf("Failed to get storage data: %s", err.Error()))
	}

	if result == nil || len(result.State) == 0 {
		s.logger.Info("Returning default storage value")
		return zeroHex32Bytes, nil // Default value
	}
	s.logger.Info("Returning storage", zap.Any("storage", result))

	return result.State[0].Value, nil
}

func (s *EthService) GetLogs(logParams domain.LogParams) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting logs", zap.Any("logParams", logParams))

	return s.commonService.GetLogs(logParams)
}

func (s *EthService) GetBlockTransactionCountByHash(blockHash string) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting block transaction count by hash", zap.String("blockHash", blockHash))

	cacheKey := fmt.Sprintf("%s_%s", GetBlockTransactionCountByHash, blockHash)

	var transactionCount string

	if err := s.cacheService.Get(s.ctx, cacheKey, &transactionCount); err == nil && transactionCount != "" {
		s.logger.Info("Transaction count fetched from cache", zap.String("count", transactionCount))
		return transactionCount, nil
	}

	block := s.mClient.GetBlockByHashOrNumber(blockHash)

	if block == nil {
		return nil, nil
	}

	transactionCount = fmt.Sprintf("0x%x", block.Count)

	if err := s.cacheService.Set(s.ctx, cacheKey, transactionCount, DefaultExpiration); err != nil {
		s.logger.Debug("Failed to cache transaction count", zap.Error(err))
	}

	return transactionCount, nil
}

func (s *EthService) GetBlockTransactionCountByNumber(blockNumberOrTag string) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting block transaction count by number", zap.String("blockNumber", blockNumberOrTag))
	blockNumberInt, errRpc := s.commonService.GetBlockNumberByNumberOrTag(blockNumberOrTag)
	if errRpc != nil {
		return nil, errRpc
	}

	cachedKey := fmt.Sprintf("%s_%d", GetBlockTransactionCountByNumber, blockNumberInt)

	var transactionCount string

	if err := s.cacheService.Get(s.ctx, cachedKey, &transactionCount); err == nil && transactionCount != "" {
		s.logger.Info("Transaction count fetched from cache", zap.String("count", transactionCount))
		return transactionCount, nil
	}

	block := s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(blockNumberInt, 10))

	if block == nil {
		return nil, nil
	}

	transactionCount = fmt.Sprintf("0x%x", block.Count)

	if err := s.cacheService.Set(s.ctx, cachedKey, transactionCount, DefaultExpiration); err != nil {
		s.logger.Debug("Failed to cache transaction count", zap.Error(err))
	}

	return transactionCount, nil
}

func (s *EthService) GetTransactionByBlockHashAndIndex(blockHash string, txIndex string) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting transaction by block and index", zap.String("blockHash", blockHash), zap.String("txIndex", txIndex))

	cacheKey := fmt.Sprintf("%s_%s_%s", GetTransactionByBlockHashAndIndex, blockHash, txIndex)

	var cachedTx interface{}
	if err := s.cacheService.Get(s.ctx, cacheKey, &cachedTx); err == nil {
		s.logger.Info("Transaction fetched from cache", zap.Any("transaction", cachedTx))
		return cachedTx, nil
	}

	txIndexInt, err := HexToDec(txIndex)
	if err != nil {
		s.logger.Error("Failed to parse transaction index", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to parse hex value")
	}

	queryParamas := map[string]interface{}{
		"block.hash":        blockHash,
		"transaction.index": txIndexInt,
	}

	tx, err := s.getTransactionByBlockAndIndex(queryParamas)
	if err != nil {
		s.logger.Error("Failed to get transaction by block and index", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to get transaction by block and index")
	}

	if err := s.cacheService.Set(s.ctx, cacheKey, tx, DefaultExpiration); err != nil {
		s.logger.Debug("Failed to cache transaction", zap.Error(err))
	}

	return tx, nil
}

func (s *EthService) GetTransactionByBlockNumberAndIndex(blockNumberOrTag string, txIndex string) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting transaction by block number and index", zap.String("blockNumberOrTag", blockNumberOrTag), zap.String("txIndex", txIndex))

	blockNumberInt, errRpc := s.commonService.GetBlockNumberByNumberOrTag(blockNumberOrTag)
	if errRpc != nil {
		return nil, errRpc
	}

	cacheKey := fmt.Sprintf("%s_%d_%s", GetTransactionByBlockNumberAndIndex, blockNumberInt, txIndex)

	var cachedTx interface{}
	if err := s.cacheService.Get(s.ctx, cacheKey, &cachedTx); err == nil {
		s.logger.Info("Transaction fetched from cache", zap.Any("transaction", cachedTx))
		return cachedTx, nil
	}

	txIndexInt, err := HexToDec(txIndex)
	if err != nil {
		s.logger.Error("Failed to parse transaction index", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to parse hex value")
	}

	queryParamas := map[string]interface{}{
		"block.number":      blockNumberInt,
		"transaction.index": txIndexInt,
	}

	tx, err := s.getTransactionByBlockAndIndex(queryParamas)
	if err != nil {
		s.logger.Error("Failed to get transaction by block and index", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to get transaction by block and index")
	}

	if err := s.cacheService.Set(s.ctx, cacheKey, tx, DefaultExpiration); err != nil {
		s.logger.Debug("Failed to cache transaction", zap.Error(err))
	}

	return tx, nil
}

func (s *EthService) SendRawTransaction(data string) (interface{}, *domain.RPCError) {
	s.logger.Info("Sending raw transaction", zap.String("data", data))

	parsedTx, err := ParseTransaction(data)
	if err != nil {
		s.logger.Error("Failed to parse transaction", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to parse transaction")
	}

	if err = s.precheck.CheckSize(data); err != nil {
		return nil, domain.NewRPCError(domain.ServerError, err.Error())
	}

	gasPriceHex, rpcErr := s.GetGasPrice()
	if rpcErr != nil {
		return nil, rpcErr
	}

	gasPrice, err := HexToDec(gasPriceHex.(string))
	if err != nil {
		s.logger.Error("Failed to parse gas price", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to parse gas price")
	}

	if err = s.precheck.SendRawTransactionCheck(parsedTx, gasPrice); err != nil {
		s.logger.Error("Transaction rejected by precheck", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Transaction rejected by precheck")
	}

	rawTxHex := strings.TrimPrefix(data, "0x")

	rawTx, err := hex.DecodeString(rawTxHex)
	if err != nil {
		s.logger.Error("Failed to decode raw transaction", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to decode raw transaction")
	}

	txHash, err := s.SendRawTransactionProcessor(rawTx, parsedTx, gasPrice)
	if err != nil {
		s.logger.Error("Failed to process transaction", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to process transaction")
	}

	return txHash, nil
}

func (s *EthService) GetCode(address string, blockNumberOrTag string) (interface{}, *domain.RPCError) {
	s.logger.Info("Getting code", zap.String("address", address), zap.String("blockNumberOrTag", blockNumberOrTag))

	// Check for iHTS precompile address first
	if address == iHTSAddress {
		s.logger.Debug("Returning iHTS contract code")
		return "0xfe", nil
	}

	cachedKey := fmt.Sprintf("%s_%s_%s", GetCode, address, blockNumberOrTag)

	var cachedCode string
	if err := s.cacheService.Get(s.ctx, cachedKey, &cachedCode); err == nil && cachedCode != "" {
		s.logger.Info("Code fetched from cache", zap.String("code", cachedCode))
		return cachedCode, nil
	}

	// Resolve the address type (contract or token)
	result, err := s.resolveAddressType(address)
	if err != nil {
		s.logger.Debug("Failed to resolve address type from Mirror node", zap.Any("error", err))
	}

	switch result := result.(type) {
	case *domain.ContractResponse:
		contract := result
		if contract.RuntimeBytecode != nil && *contract.RuntimeBytecode != zeroHex32Bytes {
			bytecode, err := hexutil.Decode(*contract.RuntimeBytecode)
			if err != nil {
				s.logger.Error("Failed to decode bytecode", zap.Error(err))
				return nil, domain.NewRPCError(domain.ServerError, "Failed to decode bytecode")
			}

			if !hasProhibitedOpcodes(bytecode) {
				if err = s.cacheService.Set(s.ctx, cachedKey, *contract.RuntimeBytecode, DefaultExpiration); err != nil {
					s.logger.Debug("Failed to cache contract bytecode", zap.Error(err))
				}

				return *contract.RuntimeBytecode, nil
			}
		}
	case *domain.TokenResponse:
		s.logger.Debug("Token redirect case, returning redirectBytecode")
		redirectBytecode := redirectBytecodePrefix + address[2:] + redirectBytecodePostfix
		return "0x" + redirectBytecode, nil
	}

	result, err = s.hClient.GetContractByteCode(0, 0, address)
	if err != nil {
		// TODO: Handle error better
		s.logger.Error("Failed to get contract bytecode", zap.Error(err))
		return "0x", nil
	}

	response := fmt.Sprintf("0x%x", result)

	if err := s.cacheService.Set(s.ctx, cachedKey, response, DefaultExpiration); err != nil {
		s.logger.Debug("Failed to cache contract bytecode", zap.Error(err))
	}

	return response, nil
}

// GetAccounts returns an empty array of accounts, similar to Infura's implementation
func (s *EthService) GetAccounts() (interface{}, *domain.RPCError) {
	s.logger.Info("Getting accounts")
	s.logger.Debug("Returning empty accounts array as per specification")
	return []string{}, nil
}

// Syncing returns false, because the Hedera network does not support syncing
func (s *EthService) Syncing() (interface{}, *domain.RPCError) {
	s.logger.Info("Syncing")
	s.logger.Debug("Returning false as per specification")
	return false, nil
}

// Mining returns false, because the Hedera network does not support mining
func (s *EthService) Mining() (interface{}, *domain.RPCError) {
	s.logger.Info("Mining")
	s.logger.Debug("Returning false as per specification")
	return false, nil
}

// MaxPriorityFeePerGas returns 0x0, because the Hedera network does not support it
func (s *EthService) MaxPriorityFeePerGas() (interface{}, *domain.RPCError) {
	s.logger.Info("MaxPriorityFeePerGas")
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// Hashrate returns 0x0, because the Hedera network does not support it
func (s *EthService) Hashrate() (interface{}, *domain.RPCError) {
	s.logger.Info("Hashrate")
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// GetUncleCountByBlockNumber returns 0x0, because the Hedera network does not support it
func (s *EthService) GetUncleCountByBlockNumber(blockNumber string) (interface{}, *domain.RPCError) {
	s.logger.Info("GetUncleCountByBlockNumber", zap.String("blockNumber", blockNumber))
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// GetUncleByBlockNumberAndIndex returns nil, because the Hedera network does not support it
func (s *EthService) GetUncleByBlockNumberAndIndex(blockNumber string, index string) (interface{}, *domain.RPCError) {
	s.logger.Info("GetUncleByBlockNumberAndIndex", zap.String("blockNumber", blockNumber), zap.String("index", index))
	s.logger.Debug("Returning nil as per specification")
	return nil, nil
}

// GetUncleCountByBlockHash returns 0x0, because the Hedera network does not support it
func (s *EthService) GetUncleCountByBlockHash(blockHash string) (interface{}, *domain.RPCError) {
	s.logger.Info("GetUncleCountByBlockHash", zap.String("blockHash", blockHash))
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// GetUncleByBlockHashAndIndex returns nil, because the Hedera network does not support it
func (s *EthService) GetUncleByBlockHashAndIndex(blockHash string, index string) (interface{}, *domain.RPCError) {
	s.logger.Info("GetUncleByBlockHashAndIndex", zap.String("blockHash", blockHash), zap.String("index", index))
	s.logger.Debug("Returning nil as per specification")
	return nil, nil
}

func (s *EthService) SubmitWork() (interface{}, *domain.RPCError) {
	s.logger.Debug("Returning nil as per specification")
	return false, nil
}

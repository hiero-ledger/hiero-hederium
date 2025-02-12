package service

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/asm"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/rlp"
	"go.uber.org/zap"
)

// GetFeeWeibars retrieves the current network fees in tinybars from the mirror client
// and converts them to weibars (1 tinybar = 10^10 weibars).
//
// Parameters:
//   - s: Pointer to EthService instance containing the mirror client
//   - params: Optional parameters for timestamp and order
//
// Returns:
//   - *big.Int: The fee amount in weibars, or nil if there was an error
//   - map[string]interface{}: Error details if any occurred, nil otherwise
//     The error map contains:
//   - "code": -32000 for failed requests
//   - "message": Description of the error
func GetFeeWeibars(s *EthService, params ...string) (*big.Int, map[string]interface{}) {
	// Default values
	timestampTo := ""
	order := ""

	if len(params) > 0 {
		timestampTo = params[0]
	}
	if len(params) > 1 {
		order = params[1]
	}

	gasTinybars, err := s.mClient.GetNetworkFees(timestampTo, order)
	if err != nil {
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to fetch gas price",
		}
	}

	// Convert tinybars to weibars
	weibars := big.NewInt(gasTinybars).
		Mul(big.NewInt(gasTinybars), big.NewInt(10000000000)) // 10^10 conversion factor

	return weibars, nil
}

func ProcessBlock(s *EthService, block *domain.BlockResponse, showDetails bool) (*domain.Block, map[string]interface{}) {
	// Create a new Block instance with default values
	ethBlock := domain.NewBlock()

	hexNumber := "0x" + strconv.FormatUint(uint64(block.Number), 16)
	hexGasUsed := "0x" + strconv.FormatUint(uint64(block.GasUsed), 16)
	hexSize := "0x" + strconv.FormatUint(uint64(block.Size), 16)
	timestampStr := strings.Split(block.Timestamp.From, ".")[0]
	timestampInt, _ := strconv.ParseUint(timestampStr, 10, 64)
	hexTimestamp := "0x" + strconv.FormatUint(timestampInt, 16)
	trimmedHash := block.Hash
	if len(trimmedHash) > 66 {
		trimmedHash = trimmedHash[:66]
	}
	trimmedParentHash := block.PreviousHash
	if len(trimmedParentHash) > 66 {
		trimmedParentHash = trimmedParentHash[:66]
	}

	ethBlock.Number = &hexNumber
	ethBlock.GasUsed = hexGasUsed
	ethBlock.GasLimit = "0x" + strconv.FormatUint(15000000, 16) // Hedera's default gas limit
	ethBlock.Hash = &trimmedHash
	ethBlock.LogsBloom = block.LogsBloom
	ethBlock.TransactionsRoot = &trimmedHash
	ethBlock.ParentHash = trimmedParentHash
	ethBlock.Timestamp = hexTimestamp
	ethBlock.Size = hexSize

	contractResults := s.mClient.GetContractResults(block.Timestamp)
	for _, contractResult := range contractResults {
		if contractResult.Result == "WRONG_NONCE" || contractResult.Result == "INVALID_ACCOUNT_ID" {
			continue
		}

		// TODO: Resolve evm addresses

		if showDetails {
			tx := ProcessTransaction(contractResult)
			ethBlock.Transactions = append(ethBlock.Transactions, tx)
		} else {
			ethBlock.Transactions = append(ethBlock.Transactions, contractResult.Hash)
		}
	}

	s.logger.Debug("Returning block data", zap.Any("block", ethBlock))
	s.logger.Info("Successfully returned block data block: "+*ethBlock.Hash,
		zap.Int("txCount", len(ethBlock.Transactions)))
	return ethBlock, nil
}

// ProcessBlock converts a Hedera block response into an Ethereum- block format.
// It takes a poincompatibleter to an EthService, a BlockResponse, and a boolean flag indicating whether
// to include full transaction details.
//
// The function performs the following:
// - Creates a new Ethereum block with default values
// - Converts block numbers, gas values, and timestamps to hex format
// - Trims hash values to standard Ethereum length (66 chars)
// - Retrieves and filters contract results, excluding failed transactions
// - Optionally processes full transaction details based on showDetails flag
//
// Returns:
// - *domain.Block: The converted Ethereum-compatible block
// - map[string]interface{}: Error information if any, nil on success
func ProcessTransaction(contractResult domain.ContractResults) interface{} {
	hexBlockNumber := hexify(contractResult.BlockNumber)
	hexGasUsed := hexify(contractResult.GasUsed)
	hexTransactionIndex := hexify(int64(contractResult.TransactionIndex))
	hexValue := hexify(int64(contractResult.Amount))
	hexV := hexify(int64(contractResult.V))

	// Safe string slicing with length checks
	hexR := "0x0"
	if contractResult.R != "" {
		hexR = truncateString(contractResult.R, 66)
	}

	hexS := "0x0"
	if contractResult.S != "" {
		hexS = truncateString(contractResult.S, 66)
	}

	hexNonce := hexify(contractResult.Nonce)

	hexTo := "0x0"
	if contractResult.To != "" {
		hexTo = truncateString(contractResult.To, 42)
	}

	trimmedBlockHash := "0x0"
	if contractResult.BlockHash != "" {
		trimmedBlockHash = truncateString(contractResult.BlockHash, 66)
	}

	trimmedFrom := "0x0"
	if contractResult.From != "" {
		trimmedFrom = truncateString(contractResult.From, 42)
	}

	trimmedHash := "0x0"
	if contractResult.Hash != "" {
		trimmedHash = truncateString(contractResult.Hash, 66)
	}

	gasPrice := "0x0"
	if contractResult.GasPrice != "" && contractResult.GasPrice != "0x" {
		gasTinybars, err := HexToDec(contractResult.GasPrice)
		if err == nil {
			gasPriceInt := big.NewInt(gasTinybars).
				Mul(big.NewInt(gasTinybars), big.NewInt(10000000000)) // 10^10 conversion factor
			gasPrice = hexify(gasPriceInt.Int64())
		}
	}

	commonFields := domain.Transaction{
		BlockHash:        &trimmedBlockHash,
		BlockNumber:      &hexBlockNumber,
		From:             trimmedFrom,
		Gas:              hexGasUsed,
		GasPrice:         gasPrice,
		Hash:             trimmedHash,
		Input:            contractResult.FunctionParameters,
		Nonce:            hexNonce,
		To:               &hexTo,
		TransactionIndex: &hexTransactionIndex,
		Value:            hexValue,
		V:                hexV,
		R:                hexR,
		S:                hexS,
		Type:             hexify(int64(contractResult.Type)),
	}

	// Handle chain ID
	if contractResult.ChainID != "0x" {
		commonFields.ChainId = &contractResult.ChainID
	}

	switch contractResult.Type {
	case 0:
		return commonFields // Legacy transaction (EIP-155)
	case 1:
		return domain.Transaction2930{
			Transaction: commonFields,
			AccessList:  []domain.AccessListEntry{}, // Empty access list for now
		}
	case 2:
		return domain.Transaction1559{
			Transaction:          commonFields,
			AccessList:           []domain.AccessListEntry{}, // Empty access list for now
			MaxPriorityFeePerGas: contractResult.MaxPriorityFeePerGas,
			MaxFeePerGas:         contractResult.MaxFeePerGas,
		}
	default:
		return commonFields // Default to legacy transaction
	}
}

func (s *EthService) ProcessTransactionResponse(contractResult domain.ContractResultResponse) interface{} {
	hexBlockNumber := hexify(contractResult.BlockNumber)
	hexGasUsed := hexify(contractResult.GasUsed)
	hexTransactionIndex := hexify(int64(contractResult.TransactionIndex))
	hexValue := hexify(int64(contractResult.Amount))
	hexV := hexify(int64(contractResult.V))

	// Safe string slicing with length checks
	hexR := contractResult.R
	if len(contractResult.R) > 66 {
		hexR = contractResult.R[:66]
	}

	hexS := contractResult.S
	if len(contractResult.S) > 66 {
		hexS = contractResult.S[:66]
	}

	hexNonce := hexify(contractResult.Nonce)

	trimmedBlockHash := contractResult.BlockHash
	if len(contractResult.BlockHash) > 66 {
		trimmedBlockHash = contractResult.BlockHash[:66]
	}

	hexTo := contractResult.To
	if len(contractResult.To) > 42 {
		hexTo = contractResult.To[:42]
	}

	var toAddress string
	evmAddressTo, errMap := s.resolveEvmAddress(hexTo)
	if errMap != nil {
		toAddress = hexTo
	} else {
		toAddress = *evmAddressTo
	}

	trimmedFrom := contractResult.From
	if len(contractResult.From) > 42 {
		trimmedFrom = contractResult.From[:42]
	}

	var fromAddress string
	evmAddressFrom, errMap := s.resolveEvmAddress(trimmedFrom)
	if errMap != nil {
		fromAddress = trimmedFrom
	} else {
		fromAddress = *evmAddressFrom
	}

	trimmedHash := contractResult.Hash
	if len(contractResult.Hash) > 66 {
		trimmedHash = contractResult.Hash[:66]
	}

	// Ensure Type is not nil before dereferencing
	var txType string
	var transactionType int64
	if contractResult.Type != nil {
		txType = hexify(int64(*contractResult.Type))
		transactionType = int64(*contractResult.Type)
	} else {
		txType = "0x0" // Default to legacy transaction type
		transactionType = 0
	}

	commonFields := domain.Transaction{
		BlockHash:        &trimmedBlockHash,
		BlockNumber:      &hexBlockNumber,
		From:             fromAddress,
		Gas:              hexGasUsed,
		GasPrice:         contractResult.GasPrice,
		Hash:             trimmedHash,
		Input:            contractResult.FunctionParameters,
		Nonce:            hexNonce,
		To:               &toAddress,
		TransactionIndex: &hexTransactionIndex,
		Value:            hexValue,
		V:                hexV,
		R:                hexR,
		S:                hexS,
		Type:             txType,
	}

	// Handle chain ID
	if contractResult.ChainID != "0x" {
		commonFields.ChainId = &contractResult.ChainID
	}

	switch transactionType {
	case 0:
		return commonFields // Legacy transaction (EIP-155)
	case 1:
		return domain.Transaction2930{
			Transaction: commonFields,
			AccessList:  []domain.AccessListEntry{}, // Empty access list for now
		}
	case 2:
		return domain.Transaction1559{
			Transaction:          commonFields,
			AccessList:           []domain.AccessListEntry{}, // Empty access list for now
			MaxPriorityFeePerGas: contractResult.MaxPriorityFeePerGas,
			MaxFeePerGas:         contractResult.MaxFeePerGas,
		}
	default:
		return commonFields // Default to legacy transaction
	}
}

func ParseTransactionCallObject(s *EthService, transaction interface{}) (*domain.TransactionCallObject, error) {
	var transactionCallObject domain.TransactionCallObject
	jsonBytes, err := json.Marshal(transaction)
	if err != nil {
		s.logger.Error("Error marshaling transaction", zap.Error(err))
		return nil, err
	}

	if err := json.Unmarshal(jsonBytes, &transactionCallObject); err != nil {
		s.logger.Error("Error unmarshaling transaction", zap.Error(err))
		return nil, err
	}

	return &transactionCallObject, nil
}

func FormatTransactionCallObject(s *EthService, transactionCallObject *domain.TransactionCallObject, blockParam interface{}, estimate bool) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Handle value conversion if present
	if transactionCallObject.Value != "" && transactionCallObject.Value != "0" && transactionCallObject.Value != "0x" {
		value, err := WeibarHexToTinyBarInt(transactionCallObject.Value)
		if err != nil {
			return nil, err
		}
		result["value"] = strconv.FormatInt(value, 10)
	}

	// Handle gas price
	if transactionCallObject.GasPrice != "" && transactionCallObject.GasPrice != "0x" {
		gasPrice, err := strconv.ParseInt(strings.TrimPrefix(transactionCallObject.GasPrice, "0x"), 16, 64)
		if err != nil {
			return nil, err
		}
		result["gasPrice"] = strconv.FormatInt(gasPrice, 10)
	}
	// else {
	// 	// Fetch gas price if not provided
	// 	gasPrice, errMap := s.GetGasPrice()
	// 	if errMap != nil {
	// 		return nil, fmt.Errorf("failed to get gas price: %v", errMap["message"])
	// 	}

	// 	// Convert hex string to decimal
	// 	gasPriceStr, ok := gasPrice.(string)
	// 	if !ok {
	// 		return nil, fmt.Errorf("invalid gas price format")
	// 	}

	// 	gasPriceInt, err := strconv.ParseInt(strings.TrimPrefix(gasPriceStr, "0x"), 16, 64)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to parse gas price: %v", err)
	// 	}

	// 	result["gasPrice"] = strconv.FormatInt(gasPriceInt, 10)
	// }
	// TODO: Decide whether gasPrice is needed, commented for now.

	// Handle gas only if present and not empty
	if transactionCallObject.Gas != "" && transactionCallObject.Gas != "0x" {
		gas, err := strconv.ParseInt(strings.TrimPrefix(transactionCallObject.Gas, "0x"), 16, 64)
		if err != nil {
			return nil, err
		}
		result["gas"] = strconv.FormatInt(gas, 10)
	}

	// Handle from address when not provided but value is present
	if transactionCallObject.From != "" || (transactionCallObject.Value != "" && transactionCallObject.Value != "0" && transactionCallObject.Value != "0x") {
		if transactionCallObject.From != "" {
			result["from"] = transactionCallObject.From
		} else {
			result["from"] = "0x17b2b8c63fa35402088640e426c6709a254c7ffb" // TODO: For now we just hardcode random account
		}
	}

	// Handle input/data field consistency - prioritize input over data if both are present
	if transactionCallObject.Input != "" && transactionCallObject.Data != "" {
		// If both are present and different, return error
		if transactionCallObject.Input != transactionCallObject.Data {
			return nil, fmt.Errorf("both input and data fields are present with different values")
		}
		result["data"] = transactionCallObject.Input
	} else if transactionCallObject.Input != "" {
		result["data"] = transactionCallObject.Input
	} else if transactionCallObject.Data != "" {
		result["data"] = transactionCallObject.Data
	}

	if blockParam != nil {
		result["block"] = blockParam
	}

	// Copy any remaining non-empty fields
	if transactionCallObject.To != "" {
		result["to"] = transactionCallObject.To
	}
	if transactionCallObject.Nonce != "" && transactionCallObject.Nonce != "0x" {
		result["nonce"] = transactionCallObject.Nonce
	}
	result["estimate"] = estimate

	return result, nil
}

// Helper function to convert weibar hex to tinybar int
const TINYBAR_TO_WEIBAR_COEF = 10000000000 // 10^10

func WeibarHexToTinyBarInt(value string) (int64, error) {
	// Handle "0x" case
	if value == "0x" {
		return 0, nil
	}

	// Convert the hex string to big.Int
	weiBigInt := new(big.Int)
	if strings.HasPrefix(value, "0x") {
		_, success := weiBigInt.SetString(value[2:], 16)
		if !success {
			return 0, fmt.Errorf("failed to parse hex value: %s", value)
		}
	} else {
		_, success := weiBigInt.SetString(value, 10)
		if !success {
			return 0, fmt.Errorf("failed to parse value: %s", value)
		}
	}

	// Create coefficient as big.Int
	coefBigInt := big.NewInt(TINYBAR_TO_WEIBAR_COEF)

	// Calculate tinybar value
	tinybarValue := new(big.Int).Div(weiBigInt, coefBigInt)

	// Only round up if the value is significant enough
	remainder := new(big.Int).Mod(weiBigInt, coefBigInt)
	if tinybarValue.Cmp(big.NewInt(0)) == 0 && remainder.Cmp(big.NewInt(TINYBAR_TO_WEIBAR_COEF/2)) > 0 {
		return 1, nil // Round up to the smallest unit of tinybar only if remainder is significant
	}

	// Convert to int64 and check if it fits
	if !tinybarValue.IsInt64() {
		return 0, fmt.Errorf("tinybar value exceeds int64 range: %s", tinybarValue.String())
	}

	return tinybarValue.Int64(), nil
}

// Utility functions

func NormalizeHexString(hexStr string) string {
	if len(hexStr) > 2 && hexStr[:2] == "0x" {
		trimmed := strings.TrimLeft(hexStr[2:], "0")
		if trimmed == "" {
			return "0x0"
		}
		return "0x" + trimmed
	}
	if hexStr == "0x" {
		return "0x0"
	}
	return hexStr
}

func hexify(n int64) string {
	return "0x" + strconv.FormatInt(n, 16)
}

func HexToDec(hexStr string) (int64, map[string]interface{}) {
	dec, err := strconv.ParseInt(strings.TrimPrefix(hexStr, "0x"), 16, 64)
	if err != nil {
		return 0, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to parse hex value",
		}
	}
	return dec, nil
}

func (s *EthService) getBlockNumberByHashOrTag(blockNumberOrTag string) (interface{}, map[string]interface{}) {
	s.logger.Debug("Getting block number by hash or tag", zap.String("blockNumberOrTag", blockNumberOrTag))
	switch blockNumberOrTag {
	case "latest", "pending":
		latestBlock, errMap := s.GetBlockNumber()
		if errMap != nil {
			s.logger.Debug("Failed to get latest block number")
			return "0x0", errMap
		}

		latestBlockStr, ok := latestBlock.(string)
		if !ok {
			s.logger.Debug("Invalid block number format")
			return "0x0", errMap
		}

		// Convert hex string to int, remove "0x" prefix
		latestBlockNum, err := HexToDec(latestBlockStr)
		if err != nil {
			s.logger.Debug("Failed to parse latest block number")
			return "0x0", errMap
		}
		return latestBlockNum, nil

	case "earliest":
		return int64(0), nil
	default:
		// Convert hex string to int, remove "0x" prefix
		latestBlockNum, err := HexToDec(blockNumberOrTag)
		if err != nil {
			s.logger.Debug("Failed to parse latest block number")
			return "0x0", err
		}

		return latestBlockNum, nil
	}
}

func (s *EthService) getFeeHistory(blockCount, newestBlockInt, latestBlockInt int64, rewardPercentiles []string) (*domain.FeeHistory, map[string]interface{}) {
	oldestBlockNumber := newestBlockInt - blockCount + 1
	if oldestBlockNumber < 0 {
		oldestBlockNumber = 0
	}

	feeHistory := &domain.FeeHistory{
		BaseFeePerGas: []string{},
		GasUsedRatio:  []float64{},
		OldestBlock:   fmt.Sprintf("0x%x", oldestBlockNumber),
	}

	// Get fees from oldest to newest blocks
	for blockNumber := oldestBlockNumber; blockNumber <= newestBlockInt; blockNumber++ {
		fee, errMap := s.getFeeByBlockNumber(blockNumber)
		if errMap != nil {
			return nil, errMap
		}

		feeHistory.BaseFeePerGas = append(feeHistory.BaseFeePerGas, fee)
		feeHistory.GasUsedRatio = append(feeHistory.GasUsedRatio, defaultUsedGasRatio)
	}

	// Get the fee for the next block if the newest block is not the latest
	var nextBaseFeePerGas string
	var errMap map[string]interface{}
	if latestBlockInt > newestBlockInt {
		nextBaseFeePerGas, errMap = s.getFeeByBlockNumber(newestBlockInt + 1)
		if errMap != nil {
			return nil, errMap
		}
	} else {
		nextBaseFeePerGas = feeHistory.BaseFeePerGas[len(feeHistory.BaseFeePerGas)-1]
	}

	if nextBaseFeePerGas != "" {
		feeHistory.BaseFeePerGas = append(feeHistory.BaseFeePerGas, nextBaseFeePerGas)
	}

	// Check if there are any reward percentiles
	if len(rewardPercentiles) > 0 {
		rewards := make([][]string, blockCount)
		for i := range rewards {
			rewards[i] = make([]string, len(rewardPercentiles))
			for j := range rewards[i] {
				rewards[i][j] = "0x0" // Default reward
			}
		}
		feeHistory.Reward = rewards
	}

	return feeHistory, nil
}

func (s *EthService) getFeeByBlockNumber(blockNumber int64) (string, map[string]interface{}) {
	block := s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(blockNumber, 10))
	if block == nil {
		return "", map[string]interface{}{
			"code":    -32000,
			"message": "Failed to get block data",
		}
	}

	fee, err := GetFeeWeibars(s, block.Timestamp.To, "desc") // Hardcode desc to be sure that we get latest
	if err != nil {
		return "", map[string]interface{}{
			"code":    -32000,
			"message": "Failed to get fee data",
		}
	}

	// Implement dec to hex func
	return "0x" + strconv.FormatUint(fee.Uint64(), 16), nil
}

func (s *EthService) getRepeatedFeeHistory(blockCount, oldestBlockInt int64, rewardPercentiles []string, fee string) *domain.FeeHistory {
	feeHistory := &domain.FeeHistory{
		BaseFeePerGas: make([]string, blockCount+1),
		GasUsedRatio:  make([]float64, blockCount),
		OldestBlock:   fmt.Sprintf("0x%x", oldestBlockInt),
	}

	for i := int64(0); i < blockCount; i++ {
		feeHistory.BaseFeePerGas[i] = fee
		feeHistory.GasUsedRatio[i] = defaultUsedGasRatio
	}

	feeHistory.BaseFeePerGas[blockCount] = fee

	//Check if there are any reward percentiles
	if len(rewardPercentiles) > 0 {
		rewards := make([][]string, blockCount)
		for i := range rewards {
			rewards[i] = make([]string, len(rewardPercentiles))
			for j := range rewards[i] {
				rewards[i][j] = "0x0" // Default reward
			}
		}
		feeHistory.Reward = rewards
	}

	return feeHistory
}

func (s *EthService) validateBlockHashAndAddTimestampToParams(params map[string]interface{}, blockHash string) bool {
	block := s.mClient.GetBlockByHashOrNumber(blockHash)
	if block == nil {
		s.logger.Debug("Failed to get block data")
		return false
	}
	s.logger.Debug("Received block data", zap.Any("block", block))

	params["timestamp"] = fmt.Sprintf("gte:%s&timestamp=lte:%s", block.Timestamp.From, block.Timestamp.To)

	s.logger.Debug("Returning timestamp", zap.Any("timestamp", params["timestamp"]))

	return true
}

func (s *EthService) validateBlockRangeAndAddTimestampToParams(params map[string]interface{}, fromBlock, toBlock string, address []string) bool {
	latestBlock, errMap := s.GetBlockNumber()
	if errMap != nil {
		s.logger.Debug("Failed to get latest block number")
		return false
	}

	latestBlockStr, ok := latestBlock.(string)
	if !ok {
		return false
	}

	if fromBlock == "latest" || fromBlock == "pending" {
		fromBlock = latestBlockStr
	}

	if toBlock == "latest" || toBlock == "pending" {
		toBlock = latestBlockStr
	}

	latestBlockNum, errMap := HexToDec(latestBlockStr)
	if errMap != nil {
		s.logger.Debug("Failed to parse latest block number", zap.Any("error", errMap))
		return false
	}

	toBlockNum, errMap := HexToDec(toBlock)
	if errMap != nil {
		return false
	}

	if toBlockNum < latestBlockNum && fromBlock == "" {
		s.logger.Debug("Invalid block range", zap.String("toBlock", toBlock), zap.String("latestBlock", latestBlockStr))
		return false
	}

	fromBlockNum, errMap := HexToDec(fromBlock)
	if errMap != nil {
		return false
	}

	fromBlockResponse := s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(fromBlockNum, 10))
	if fromBlockResponse == nil {
		s.logger.Debug("Failed to get from block data")
		return false
	}

	var timestamp string

	timestamp = fmt.Sprintf("gte:%s", fromBlockResponse.Timestamp.From)

	if fromBlock == toBlock {
		timestamp += fmt.Sprintf("&timestamp=lte:%s", fromBlockResponse.Timestamp.To)

	} else {
		fromBlockNum := fromBlockResponse.Number
		toBlockResponse := s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(toBlockNum, 10))

		var toBlockNum int

		if toBlockResponse != nil {
			timestamp = fmt.Sprintf("%s&timestamp=lte:%s", timestamp, toBlockResponse.Timestamp.To)
			toBlockNum = toBlockResponse.Number
		}

		if fromBlockNum > toBlockNum {
			return false
		}

		isSingleAddress := len(address) == 1
		if !isSingleAddress && toBlockNum-fromBlockNum > maxBlockCountForResult {
			return false
		}
	}

	s.logger.Debug("Returning timestamp", zap.String("timestamp", timestamp))
	params["timestamp"] = timestamp

	return true
}

func (s *EthService) getLogsWithParams(address []string, params map[string]interface{}) ([]domain.Log, map[string]interface{}) {
	addresses := address

	var logs []domain.Log

	if len(address) == 0 {
		logResults, err := s.mClient.GetContractResultsLogsWithRetry(params)
		if err != nil {
			s.logger.Error("Failed to get logs", zap.Error(err))
			return nil, map[string]interface{}{
				"code":    -32000,
				"message": "Failed to get logs",
			}
		}

		for _, logResult := range logResults {
			logs = append(logs, domain.Log{
				Address:          logResult.Address,
				BlockHash:        logResult.BlockHash,
				BlockNumber:      "0x" + strconv.FormatInt(logResult.BlockNumber, 16),
				Data:             logResult.Result,
				TransactionHash:  logResult.Hash,
				TransactionIndex: strconv.Itoa(logResult.TransactionIndex),
			})
		}

	}

	for _, addr := range addresses {
		logResults, err := s.mClient.GetContractResultsLogsByAddress(addr, params)
		if err != nil {
			s.logger.Error("Failed to get logs", zap.Error(err))
			return nil, map[string]interface{}{
				"code":    -32000,
				"message": "Failed to get logs",
			}
		}
		for _, logResult := range logResults {
			logs = append(logs, domain.Log{
				Address:          logResult.Address,
				BlockHash:        logResult.BlockHash,
				BlockNumber:      "0x" + strconv.FormatInt(logResult.BlockNumber, 16),
				Data:             logResult.Result,
				TransactionHash:  logResult.Hash,
				TransactionIndex: strconv.Itoa(logResult.TransactionIndex),
			})
		}
	}

	if logs == nil {
		return []domain.Log{}, nil
	}

	return logs, nil
}

// Optimise the function to avoid multiple calls to the mirror node
func (s *EthService) resolveEvmAddress(address string) (*string, map[string]interface{}) {
	result, errMap := s.resolveAddressType(address)
	if errMap != nil {
		return nil, errMap
	}

	switch data := result.(type) {
	case *domain.AccountResponse:
		return &data.EvmAddress, nil
	case *domain.ContractResponse:
		return &data.EvmAddress, nil
	case *domain.TokenResponse:
		return &address, nil
	}

	return nil, map[string]interface{}{
		"code":    -32000,
		"message": "Unable to resolve EVM address",
	}
}

func (s *EthService) resolveAddressType(address string) (interface{}, map[string]interface{}) {
	if contractData, _ := s.mClient.GetContractById(address); contractData != nil {
		return contractData, nil
	}

	if accountData, _ := s.mClient.GetAccountById(address); accountData != nil {
		return accountData, nil
	}

	// TODO: Make it in constant
	if strings.HasPrefix(address, "0x000000000000") {
		addressNum, errMap := HexToDec(address)
		if errMap != nil {
			return nil, errMap
		}

		tokenId := "0.0." + strconv.FormatInt(addressNum, 10)
		if tokenData, _ := s.mClient.GetTokenById(tokenId); tokenData != nil {
			return tokenData, nil
		}
	}

	return nil, map[string]interface{}{
		"code":    -32000,
		"message": "Unable to identify address type",
	}
}

func (s *EthService) getTransactionByBlockAndIndex(queryParamas map[string]interface{}) (interface{}, map[string]interface{}) {
	transaction, err := s.mClient.GetContractResultWithRetry(queryParamas)
	if err != nil {
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to get transaction",
		}
	}

	if transaction == nil {
		return nil, nil
	}

	evmAddressTo, errMap := s.resolveEvmAddress(transaction.To)
	if errMap != nil {
		return nil, errMap
	}

	evmAddressFrom, errMap := s.resolveEvmAddress(transaction.From)
	if errMap != nil {
		return nil, errMap
	}

	s.logger.Info("Processing contract result", zap.Any("evmAddressTO", evmAddressTo))
	s.logger.Info("Processing contract result", zap.Any("evmAddressFROM", evmAddressFrom))

	transaction.To = *evmAddressTo
	transaction.From = *evmAddressFrom

	return ProcessTransaction(*transaction), nil
}

func ParseTransaction(rawTxHex string) (*types.Transaction, error) {
	if rawTxHex == "" {
		return nil, errors.New("transaction data is empty")
	}

	rawTxHex = strings.TrimPrefix(rawTxHex, "0x")

	rawTx, err := hex.DecodeString(rawTxHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %w", err)
	}

	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(rawTx, tx); err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	return tx, nil
}

// Add 10% buffer to the gas price
func AddBuffer(weibars *big.Int) *big.Int {
	buffer := new(big.Int).Div(weibars, big.NewInt(10))
	return weibars.Add(weibars, buffer)
}

// ProcessRawTransaction handles the processing of a raw Ethereum transaction for Hedera
func (s *EthService) SendRawTransactionProcessor(transactionData []byte, tx *types.Transaction, gasPrice int64) (*string, error) {
	// Get the sender address for event tracking
	fromAddress, err := GetFromAddress(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender address: %w", err)
	}

	// Get the recipient address for event tracking
	var toAddress string
	if tx.To() != nil {
		toAddress = tx.To().String()
	}

	// Send the raw transaction using the client's implementation
	response, err := s.hClient.SendRawTransaction(transactionData, gasPrice, fromAddress)
	if err != nil {
		s.logger.Error("Failed to send raw transaction",
			zap.Error(err),
			zap.String("from", fromAddress.String()),
			zap.String("to", toAddress),
			zap.Int64("gasPrice", gasPrice))
		return nil, fmt.Errorf("failed to send raw transaction: %w", err)
	}

	subbmitedTransactionId := response.TransactionID

	transactionIDRegex := regexp.MustCompile(`\d{1}\.\d{1}\.\d{1,10}\@\d{1,10}\.\d{1,9}`)
	if !transactionIDRegex.MatchString(subbmitedTransactionId) {
		s.logger.Error("Invalid transaction ID format", zap.String("transactionID", subbmitedTransactionId))
		return nil, fmt.Errorf("invalid transaction ID format: %s", subbmitedTransactionId)
	}

	if subbmitedTransactionId != "" {
		transactionId := ConvertTransactionID(subbmitedTransactionId)
		contractResult := s.mClient.RepeatGetContractResult(transactionId, 10)
		if contractResult == nil {
			s.logger.Error("Failed to get contract result",
				zap.String("transactionID", transactionId))
			return nil, fmt.Errorf("no matching transaction record retrieved: %s", transactionId)
		}

		hash := contractResult.Hash

		if hash == "" {
			s.logger.Error("Transaction returned a null transaction hash:",
				zap.String("transactionID", subbmitedTransactionId))
			return nil, fmt.Errorf("no matching transaction record retrieved: %s", subbmitedTransactionId)
		}

		s.logger.Info("Transaction sent successfully",
			zap.String("transactionID", hash),
			zap.String("from", fromAddress.String()),
			zap.String("to", toAddress),
			zap.Int64("gasPrice", gasPrice))

		return &hash, nil
	}

	return nil, fmt.Errorf("failed to send transaction: %w", err)
}

func (s *EthService) getCurrentGasPriceForBlock(blockHash string) (string, map[string]interface{}) {
	block := s.mClient.GetBlockByHashOrNumber(blockHash)
	gasPriceForTimestamp, errMap := GetFeeWeibars(s, block.Timestamp.From)
	if errMap != nil {
		return "", errMap
	}

	return fmt.Sprintf("0x%x", gasPriceForTimestamp), nil
}
func GetFromAddress(tx *types.Transaction) (*common.Address, error) {
	signer := types.NewEIP155Signer(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return nil, err
	}
	return &from, nil
}

func ConvertTransactionID(transactionID string) string {
	parts := strings.Split(transactionID, "@")

	parts[1] = strings.ReplaceAll(parts[1], ".", "-")

	return parts[0] + "-" + parts[1]
}

// TODO: Move it to a separate file
var prohibitedOpcodes = map[vm.OpCode]bool{
	vm.CALLCODE:     true,
	vm.DELEGATECALL: true,
	vm.SELFDESTRUCT: true,
}

func hasProhibitedOpcodes(bytecode []byte) bool {
	ops, err := asm.Disassemble(bytecode)
	if err != nil {
		log.Printf("Error disassembling bytecode: %v", err)
		return false
	}

	for _, op := range ops {
		if prohibitedOpcodes[vm.OpCode(vm.StringToOp(op))] {
			return true
		}
	}
	return false
}

func truncateString(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength]
	}
	return s
}

func (s *EthService) isLatestBlockRequest(blockNumberOrTag string, blockNumber int64) bool {
	if blockNumberOrTag == "latest" || blockNumberOrTag == "pending" {
		return true
	}
	if blockNumberOrTag == "earliest" {
		return false
	}

	latestBlock, err := s.getBlockNumberByHashOrTag("latest")
	if err != nil {
		return false
	}

	latestBlockInt, ok := latestBlock.(int64)
	if !ok {
		return false
	}

	return blockNumber+10 > latestBlockInt
}

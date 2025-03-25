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
	"sync"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/asm"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/crypto"
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
func GetFeeWeibars(s *EthService, params ...string) (*big.Int, error) {
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
		return nil, fmt.Errorf("failed to fetch gas price: %s", err.Error())
	}

	// Convert tinybars to weibars
	weibars := big.NewInt(gasTinybars).
		Mul(big.NewInt(gasTinybars), big.NewInt(10000000000)) // 10^10 conversion factor

	return weibars, nil
}

func ProcessBlock(s *EthService, block *domain.BlockResponse, showDetails bool) (*domain.Block, error) {
	// Create a new Block instance with default values
	ethBlock := domain.NewBlock()

	hexNumber := hexify(int64(block.Number))
	hexGasUsed := hexify(int64(block.GasUsed))
	hexSize := hexify(int64(block.Size))
	timestampStr := strings.Split(block.Timestamp.From, ".")[0]
	timestampInt, _ := strconv.ParseInt(timestampStr, 10, 64)
	hexTimestamp := hexify(timestampInt)

	trimmedHash := block.Hash
	if len(trimmedHash) > 66 {
		trimmedHash = trimmedHash[:66]
	}

	trimmedParentHash := block.PreviousHash
	if len(trimmedParentHash) > 66 {
		trimmedParentHash = trimmedParentHash[:66]
	}

	ethBlock.WithdrawalsRoot = "0x0000000000000000000000000000000000000000000000000000000000000000"
	ethBlock.MixHash = "0x0000000000000000000000000000000000000000000000000000000000000000"

	gasPrice, err := s.GetGasPrice()
	if err != nil {
		s.logger.Error("Failed to get gas price", zap.Error(err))
	}
	gasPriceStr, ok := gasPrice.(string)
	if !ok {
		s.logger.Error("Gas price is not a string")
		gasPriceStr = "0x0"
	}
	ethBlock.Withdrawals = []string{}
	ethBlock.BaseFeePerGas = gasPriceStr
	ethBlock.Number = &hexNumber
	ethBlock.GasUsed = hexGasUsed
	ethBlock.GasLimit = hexify(GasLimit) // Hedera's default gas limit
	ethBlock.Hash = &trimmedHash
	ethBlock.LogsBloom = block.LogsBloom
	ethBlock.TransactionsRoot = &trimmedHash
	ethBlock.ParentHash = trimmedParentHash
	ethBlock.Timestamp = hexTimestamp
	ethBlock.Size = hexSize
	ethBlock.TotalDifficulty = "0x0"

	contractResults := s.mClient.GetContractResults(block.Timestamp)
	for _, contractResult := range contractResults {
		if contractResult.Result == "WRONG_NONCE" || contractResult.Result == "INVALID_ACCOUNT_ID" {
			continue
		}

		to, err := s.resolveEvmAddress(contractResult.To)
		if err != nil {
			s.logger.Error("Failed to resolve to address", zap.Error(err))
		}

		from, err := s.resolveEvmAddress(contractResult.From)
		if err != nil {
			s.logger.Error("Failed to resolve from address", zap.Error(err))
		}

		contractResult.To = *to
		contractResult.From = *from

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
	hexValue := fmt.Sprintf("0x%x", uint64(contractResult.Amount))
	hexV := hexify(int64(contractResult.V))

	// Safe string slicing with length checks
	hexR := "0x0"
	if contractResult.R != "" {
		hexR = removeLeadingZeroes(truncateString(contractResult.R, 66))
	}

	hexS := "0x0"
	if contractResult.S != "" {
		hexS = removeLeadingZeroes(truncateString(contractResult.S, 66))
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
		MaxPriorityFeePerGas := parseFee(contractResult.MaxPriorityFeePerGas)
		MaxFeePerGas := parseFee(contractResult.MaxFeePerGas)
		return domain.Transaction1559{
			Transaction:          commonFields,
			AccessList:           []domain.AccessListEntry{}, // Empty access list for now
			MaxPriorityFeePerGas: MaxPriorityFeePerGas,
			MaxFeePerGas:         MaxFeePerGas,
		}
	default:
		return commonFields // Default to legacy transaction
	}
}

func (s *EthService) ProcessTransactionResponse(contractResult domain.ContractResultResponse) interface{} {
	hexBlockNumber := hexify(contractResult.BlockNumber)
	hexGasUsed := hexify(contractResult.GasUsed)
	hexTransactionIndex := hexify(int64(contractResult.TransactionIndex))
	value, err := s.tinybarsToWeibars(int64(contractResult.Amount), true)
	if err != nil {
		// TODO: If allowNegative in tinybarsToWeibars can be false - this should return error and be handled in properly
		// return domain.NewRPCError(domain.InternalError, "Invalid value - cannot pass negative number")
		s.logger.Error("Invalid value - cannot pass negative number", zap.Error(err))
		return nil
	}
	hexValue := hexify(value)
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
	evmAddressTo, err := s.resolveEvmAddress(hexTo)
	if err != nil {
		toAddress = hexTo
	} else {
		toAddress = *evmAddressTo
	}

	trimmedFrom := contractResult.From
	if len(contractResult.From) > 42 {
		trimmedFrom = contractResult.From[:42]
	}

	var fromAddress string
	evmAddressFrom, err := s.resolveEvmAddress(trimmedFrom)
	if err != nil {
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

func (s *EthService) tinybarsToWeibars(tinybars int64, allowNegative bool) (int64, error) {
	if tinybars == 0 {
		return 0, nil
	}

	if allowNegative && tinybars < 0 {
		return tinybars, nil
	}

	if tinybars < 0 {
		return 0, fmt.Errorf("tinybars cannot be negative")
	}

	coefBigInt := big.NewInt(TINYBAR_TO_WEIBAR_COEF)
	weiBigInt := new(big.Int).Mul(big.NewInt(tinybars), coefBigInt)

	return weiBigInt.Int64(), nil
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
			result["from"] = s.hClient.GetOperatorPublicKey()
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

func HexToDec(hexStr string) (int64, error) {
	dec, err := strconv.ParseInt(strings.TrimPrefix(hexStr, "0x"), 16, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse hex value: %s", err)
	}
	return dec, nil
}

func (s *EthService) getFeeHistory(blockCount, newestBlockInt, latestBlockInt int64, rewardPercentiles []string) (*domain.FeeHistory, error) {
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
		fee, err := s.getFeeByBlockNumber(blockNumber)
		if err != nil {
			return nil, err
		}

		feeHistory.BaseFeePerGas = append(feeHistory.BaseFeePerGas, fee)
		feeHistory.GasUsedRatio = append(feeHistory.GasUsedRatio, defaultUsedGasRatio)
	}

	// Get the fee for the next block if the newest block is not the latest
	var nextBaseFeePerGas string
	var err error
	if latestBlockInt > newestBlockInt {
		nextBaseFeePerGas, err = s.getFeeByBlockNumber(newestBlockInt + 1)
		if err != nil {
			return nil, err
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

func (s *EthService) getFeeByBlockNumber(blockNumber int64) (string, error) {
	block := s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(blockNumber, 10))
	if block == nil {
		return "", fmt.Errorf("failed to get block data")
	}

	fee, err := GetFeeWeibars(s, block.Timestamp.To, "desc") // Hardcode desc to be sure that we get latest
	if err != nil {
		return "", err
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

func (s *EthService) resolveEvmAddress(address string) (*string, error) {
	if address == "" {
		return &address, fmt.Errorf("address is empty")
	}

	cacheKey := fmt.Sprintf("evm_address_%s", address)
	var cachedResult string
	if err := s.cacheService.Get(s.ctx, cacheKey, &cachedResult); err == nil && cachedResult != "" {
		s.logger.Info("EVM Address fetched from cache", zap.String("address", cachedResult))
		return &cachedResult, nil
	}

	evmAddress := address

	result, err := s.resolveAddressType(address)
	if err == nil {
		switch data := result.(type) {
		case *domain.AccountResponse:
			evmAddress = data.EvmAddress
		case *domain.ContractResponse:
			evmAddress = data.EvmAddress
		}
	}

	if err := s.cacheService.Set(s.ctx, cacheKey, evmAddress, DefaultExpiration); err != nil {
		s.logger.Debug("Failed to cache evm address", zap.Error(err))
	}

	return &evmAddress, nil
}

func (s *EthService) resolveAddressType(address string) (interface{}, error) {
	res := make(chan interface{}, 1)

	var wg sync.WaitGroup

	tryResolve := func(f func() (interface{}, error)) {
		defer wg.Done()
		if data, err := f(); err == nil && data != nil {
			select {
			case res <- data:
			default:
			}
		}
	}

	wg.Add(2)
	go tryResolve(func() (interface{}, error) { return s.mClient.GetContractById(address) })
	go tryResolve(func() (interface{}, error) { return s.mClient.GetAccountById(address) })

	if tokenId, err := checkTokenId(address); err == nil && tokenId != nil {
		wg.Add(1)
		go tryResolve(func() (interface{}, error) { return s.mClient.GetTokenById(*tokenId) })
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	if res, ok := <-res; ok {
		s.logger.Info("Resolved address type", zap.Any("result", res))
		return res, nil
	}

	return nil, fmt.Errorf("unable to identify address type")
}

func checkTokenId(address string) (*string, error) {
	if !strings.HasPrefix(address, "0x000000000000") {
		return nil, fmt.Errorf("not a token address")
	}

	addressNum, err := HexToDec(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hex value: %s", err.Error())
	}

	str := fmt.Sprintf("0.0.%d", addressNum)
	return &str, nil
}

func (s *EthService) getTransactionByBlockAndIndex(queryParamas map[string]interface{}) (interface{}, error) {
	transaction, err := s.mClient.GetContractResultWithRetry(queryParamas)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %s", err.Error())
	}

	if transaction == nil {
		return nil, nil
	}

	evmAddressTo, err := s.resolveEvmAddress(transaction.To)
	if err != nil {
		s.logger.Error("Failed to resolve to address", zap.Error(err))
	}

	evmAddressFrom, err := s.resolveEvmAddress(transaction.From)
	if err != nil {
		s.logger.Error("Failed to resolve from address", zap.Error(err))
	}

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

func (s *EthService) getCurrentGasPriceForBlock(blockHash string) (string, error) {
	block := s.mClient.GetBlockByHashOrNumber(blockHash)
	gasPriceForTimestamp, err := GetFeeWeibars(s, block.Timestamp.From)
	if err != nil {
		return "", err
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

	latestBlockInt, err := s.commonService.GetBlockNumberByNumberOrTag("latest")
	if err != nil {
		return false
	}

	return blockNumber+10 > latestBlockInt
}

func (s *EthService) getContractAddressFromReceipt(receiptResponse domain.ContractResultResponse) string {
	if len(receiptResponse.FunctionParameters) < 10 {
		return receiptResponse.Address
	}

	parameters := receiptResponse.FunctionParameters[:10]
	if _, isHTSCreation := HTSCreateFuncSelectors[parameters]; !isHTSCreation {
		return receiptResponse.Address
	}

	if len(receiptResponse.CallResult) < 40 {
		return receiptResponse.Address
	}

	tokenAddress := receiptResponse.CallResult[len(receiptResponse.CallResult)-40:]

	if !strings.HasPrefix(tokenAddress, "0x") {
		tokenAddress = fmt.Sprintf("0x%s", tokenAddress)
	}

	return tokenAddress
}

func isHexString(str string) bool {
	str = strings.TrimPrefix(str, "0x")

	_, err := hex.DecodeString(str)
	return err == nil
}


func buildLogsBloom(address string, topics []string) string {
	if address == "" || len(topics) == 0 {
		return zeroHex32Bytes
	}

	address = strings.TrimPrefix(address, "0x")

	items := []string{address}
	for _, topic := range topics {
		items = append(items, strings.TrimPrefix(topic, "0x"))
	}

	bitvector := make([]byte, BloomByteSize)

	for _, item := range items {
		itemBytes, _ := hex.DecodeString(item)
		hash := crypto.Keccak256(itemBytes)

		for i := 0; i < 3; i++ {
			// Get first 2 bytes at position i*2
			first2bytes := uint16(hash[i*2])<<8 | uint16(hash[i*2+1])

			// Calculate bit position
			loc := BloomMask & first2bytes
			byteLoc := loc >> 3
			bitLoc := uint8(1 << (loc % 8))

			// Set the bit in the bitvector
			bitvector[BloomByteSize-int(byteLoc)-1] |= bitLoc
		}
	}

	return fmt.Sprintf("0x%s", hex.EncodeToString(bitvector))
}

func removeLeadingZeroes(str string) string {
	return fmt.Sprintf("0x%s", strings.TrimLeft(strings.TrimPrefix(str, "0x"), "0"))
}

func parseFee(fee string) string {
	if fee == "" || fee == "0x" {
		return "0x0"
	}
	return removeLeadingZeroes(fee)
}

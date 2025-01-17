package service

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/LimeChain/Hederium/internal/domain"
	"go.uber.org/zap"
)

// GetFeeWeibars retrieves the current network fees in tinybars from the mirror client
// and converts them to weibars (1 tinybar = 10^8 weibars).
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
		Mul(big.NewInt(gasTinybars), big.NewInt(100000000)) // 10^8 conversion factor

	return weibars, nil
}

// GetFeeWeibars retrieves the current network fees in tinybars from the mirror client
// and converts them to weibars (1 tinybar = 10^8 weibars).
//
// Parameters:
//   - s: Pointer to EthService instance containing the mirror client
//
// Returns:
//   - *big.Int: The fee amount in weibars, or nil if there was an error
//   - map[string]interface{}: Error details if any occurred, nil otherwise
//     The error map contains:
//   - "code": -32000 for failed requests
//   - "message": Description of the error
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

// ProcessBlock converts a Hedera block response into an Ethereum-compatible block format.
// It takes a pointer to an EthService, a BlockResponse, and a boolean flag indicating whether
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
	hexR := contractResult.R
	if len(contractResult.R) > 66 {
		hexR = contractResult.R[:66]
	}

	hexS := contractResult.S
	if len(contractResult.S) > 66 {
		hexS = contractResult.S[:66]
	}

	hexNonce := hexify(contractResult.Nonce)

	hexTo := contractResult.To
	if len(contractResult.To) > 42 {
		hexTo = contractResult.To[:42]
	}

	trimmedBlockHash := contractResult.BlockHash
	if len(contractResult.BlockHash) > 66 {
		trimmedBlockHash = contractResult.BlockHash[:66]
	}

	trimmedFrom := contractResult.From
	if len(contractResult.From) > 42 {
		trimmedFrom = contractResult.From[:42]
	}

	trimmedHash := contractResult.Hash
	if len(contractResult.Hash) > 66 {
		trimmedHash = contractResult.Hash[:66]
	}

	commonFields := domain.Transaction{
		BlockHash:        &trimmedBlockHash,
		BlockNumber:      &hexBlockNumber,
		From:             trimmedFrom,
		Gas:              hexGasUsed,
		GasPrice:         contractResult.GasPrice,
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

func ProcessTransactionResponse(contractResult domain.ContractResultResponse) interface{} {
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

	hexTo := contractResult.To
	if len(contractResult.To) > 42 {
		hexTo = contractResult.To[:42]
	}

	trimmedBlockHash := contractResult.BlockHash
	if len(contractResult.BlockHash) > 66 {
		trimmedBlockHash = contractResult.BlockHash[:66]
	}

	trimmedFrom := contractResult.From
	if len(contractResult.From) > 42 {
		trimmedFrom = contractResult.From[:42]
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
		From:             trimmedFrom,
		Gas:              hexGasUsed,
		GasPrice:         contractResult.GasPrice,
		Hash:             trimmedHash,
		Input:            contractResult.FunctionParameters,
		Nonce:            hexNonce,
		To:               &hexTo,
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

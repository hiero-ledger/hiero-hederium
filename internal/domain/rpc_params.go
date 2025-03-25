package domain

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RPCParams interface defines methods that all RPC parameter structs should implement
type RPCParams interface {
	// FromPositionalParams converts positional parameters (array) to struct fields
	FromPositionalParams(params []interface{}) *RPCError
}

// EthGetBlockByHashParams represents parameters for eth_getBlockByHash
type EthGetBlockByHashParams struct {
	BlockHash   string `json:"blockHash" binding:"required,len=66,hexadecimal,startswith=0x"`
	ShowDetails bool   `json:"showDetails"`
}

// EthGetBlockByNumberParams represents parameters for eth_getBlockByNumber
type EthGetBlockByNumberParams struct {
	BlockNumber string `json:"blockNumber" binding:"required,block_number_or_tag"`
	ShowDetails bool   `json:"showDetails"`
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
	CallObject     CallObject `json:"callObject" binding:"required"`
	BlockParameter string     `json:"blockParameter" binding:"omitempty,block_number_or_tag"`
}

type CallObject struct {
	From     string `json:"from" binding:"omitempty,eth_address"`
	To       string `json:"to" binding:"omitempty,eth_address"`
	Gas      string `json:"gas" binding:"omitempty,hexadecimal"`
	GasPrice string `json:"gasPrice" binding:"omitempty,hexadecimal"`
	Value    string `json:"value" binding:"omitempty,hexadecimal"`
	Data     string `json:"data" binding:"omitempty,data"`
}

// EthCallParams represents parameters for eth_call
type EthCallParams struct {
	CallObject CallObject `json:"callObject" binding:"required"`
	Block      string     `json:"block" binding:"required,block_number_or_tag"`
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
	RewardPercentiles []string `json:"rewardPercentiles" binding:"omitempty"`
}

// EthGetStorageAtParams represents parameters for eth_getStorageAt
type EthGetStorageAtParams struct {
	Address         string `json:"address" binding:"required,eth_address"`
	StoragePosition string `json:"storagePosition" binding:"required,hexadecimal,startswith=0x"`
	BlockNumber     string `json:"blockNumber" binding:"omitempty,block_number_or_tag"`
}

// NoParameters represents a struct with no parameters for endpoints that do not have input parameters
type NoParameters struct{}

// FilterObject represents the filter object for eth_getLogs
type FilterObject struct {
	Address   Address  `json:"address" binding:"omitempty,eth_address_or_array"`
	Topics    []string `json:"topics" binding:"omitempty,dive,hexadecimal,len=66"`
	BlockHash string   `json:"blockHash" binding:"omitempty,hexadecimal,len=66"`
	FromBlock string   `json:"fromBlock" binding:"omitempty,block_number_or_tag"`
	ToBlock   string   `json:"toBlock" binding:"omitempty,block_number_or_tag"`
}

// EthGetLogsParams represents parameters for eth_getLogs
type EthGetLogsParams struct {
	Address   Address  `json:"address" binding:"omitempty,dive,eth_address"`
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

// FromPositionalParams implements parameter conversion for NoParameters
func (p *NoParameters) FromPositionalParams(params []interface{}) *RPCError {
	// No parameters expected
	return nil
}

// FromPositionalParams implements parameter conversion for EthGetBlockByHashParams
func (p *EthGetBlockByHashParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 2 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 2 parameters, got %d", len(params)))
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("blockHash must be a string")
	}
	p.BlockHash = blockHash

	showDetails, ok := params[1].(bool)
	if !ok {
		return NewInvalidParamsError("showDetails must be a boolean")
	}
	p.ShowDetails = showDetails

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetBlockByNumberParams
func (p *EthGetBlockByNumberParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 2 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 2 parameters, got %d", len(params)))
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	showDetails, ok := params[1].(bool)
	if !ok {
		return NewInvalidParamsError("showDetails must be a boolean")
	}
	p.ShowDetails = showDetails

	return nil
}

func (p *EthGetLogsParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 1 {
		return NewInvalidParamsError("eth_getLogs expects exactly one parameter object")
	}

	filterObj, ok := params[0].(map[string]interface{})
	if !ok {
		return NewInvalidParamsError("eth_getLogs expects a filter object parameter")
	}

	validFields := map[string]bool{
		"address":   true,
		"topics":    true,
		"blockHash": true,
		"fromBlock": true,
		"toBlock":   true,
	}

	for field := range filterObj {
		if !validFields[field] {
			return NewInvalidParamsError(fmt.Sprintf("'%s' is not a valid parameter for eth_getLogs", field))
		}
	}

	var filter FilterObject
	filterBytes, err := json.Marshal(filterObj)
	if err != nil {
		return NewInvalidParamsError(fmt.Sprintf("failed to marshal filter object: %v", err))
	}

	if err := json.Unmarshal(filterBytes, &filter); err != nil {
		return NewInvalidParamsError(fmt.Sprintf("failed to unmarshal filter object: %v", err))
	}

	validate := binding.Validator.Engine().(*validator.Validate)
	if err := validate.Struct(&filter); err != nil {
		return NewInvalidParamsError(fmt.Sprintf("invalid filter parameters: %v", err))
	}

	p.Address = filter.Address
	p.Topics = filter.Topics
	p.BlockHash = filter.BlockHash
	p.FromBlock = filter.FromBlock
	p.ToBlock = filter.ToBlock

	if p.BlockHash != "" {
		if p.FromBlock != "" || p.ToBlock != "" {
			return NewInvalidParamsError("can't use both blockHash and toBlock/fromBlock")
		}
	} else {
		if p.ToBlock == "" {
			p.ToBlock = "latest"
		}
	}

	return nil
}

// ToLogParams converts EthGetLogsParams to LogParams
func (p *EthGetLogsParams) ToLogParams() LogParams {
	return LogParams{
		Address:   p.Address,
		Topics:    p.Topics,
		BlockHash: p.BlockHash,
		FromBlock: p.FromBlock,
		ToBlock:   p.ToBlock,
	}
}

// FromPositionalParams implements parameter conversion for EthGetBalanceParams
func (p *EthGetBalanceParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) < 1 || len(params) > 2 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 1 or 2 parameters, got %d", len(params)))
	}

	address, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("address must be a string")
	}
	p.Address = address

	if len(params) > 1 {
		blockNumber, ok := params[1].(string)
		if !ok {
			return NewInvalidParamsError("blockNumber must be a string")
		}
		p.BlockNumber = blockNumber
	} else {
		p.BlockNumber = "latest"
	}

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetTransactionCountParams
func (p *EthGetTransactionCountParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) < 1 || len(params) > 2 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 1 or 2 parameters, got %d", len(params)))
	}

	address, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("address must be a string")
	}
	p.Address = address

	if len(params) > 1 {
		blockNumber, ok := params[1].(string)
		if !ok {
			return NewInvalidParamsError("blockNumber must be a string")
		}
		p.BlockNumber = blockNumber
	} else {
		p.BlockNumber = "latest"
	}

	return nil
}

// FromPositionalParams implements parameter conversion for EthEstimateGasParams
func (p *EthEstimateGasParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) == 0 || len(params) > 2 {
		return NewInvalidParamsError(fmt.Sprintf("Missing value for required parameter 0, expected 1 or 2 parameters, got %d", len(params)))
	}

	callObj, ok := params[0].(map[string]interface{})
	if !ok {
		return NewInvalidParamsError("callObject must be an object")
	}

	var callObject CallObject

	if from, exists := callObj["from"]; exists {
		strFrom, ok := from.(string)
		if !ok {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed string representing the address (20 bytes), value: %v", from))
		}
		callObject.From = strFrom
	}

	if to, exists := callObj["to"]; exists {
		if strTo, ok := to.(string); ok {
			callObject.To = strTo
		} else if to == nil {
			callObject.To = ""
		} else {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed string representing the address (20 bytes), value: %v", to))
		}
	}

	if gas, exists := callObj["gas"]; exists {
		strGas, ok := gas.(string)
		if !ok {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed hexadecimal value, value: %v", gas))
		}
		callObject.Gas = strGas
	}

	if gasPrice, exists := callObj["gasPrice"]; exists {
		strGasPrice, ok := gasPrice.(string)
		if !ok {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed hexadecimal value, value: %v", gasPrice))
		}
		callObject.GasPrice = strGasPrice
	}

	if value, exists := callObj["value"]; exists {
		strValue, ok := value.(string)
		if !ok {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed hexadecimal value, value: %v", value))
		}
		callObject.Value = strValue
	}

	if data, exists := callObj["data"]; exists {
		strData, ok := data.(string)
		if !ok {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed hexadecimal value with even length, value: %v", data))
		}
		callObject.Data = strData
	}

	validate := binding.Validator.Engine().(*validator.Validate)
	if err := validate.Struct(&callObject); err != nil {
		message, tag := translateValidationErrors(err)
		return NewInvalidParamsError(fmt.Sprintf("Invalid parameter '%s' for TransactionObject: %s", tag, message))
	}

	p.CallObject = callObject

	if len(params) > 1 {
		blockParam, ok := params[1].(string)
		if !ok {
			return NewInvalidParamsError("blockParameter must be a string")
		}
		p.BlockParameter = blockParam
	}

	return nil
}

// FromPositionalParams implements parameter conversion for EthCallParams
func (p *EthCallParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 2 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 2 parameters, got %d", len(params)))
	}

	callObj, ok := params[0].(map[string]interface{})
	if !ok {
		return NewInvalidParamsError("callObject must be an object")
	}

	var callObject CallObject

	if from, exists := callObj["from"]; exists {
		strFrom, ok := from.(string)
		if !ok {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed string representing the address (20 bytes), value: %v", from))
		}
		callObject.From = strFrom
	}

	if to, exists := callObj["to"]; exists {
		strTo, ok := to.(string)
		if !ok {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed string representing the address (20 bytes), value: %v", to))
		}
		callObject.To = strTo
	}

	if gas, exists := callObj["gas"]; exists {
		strGas, ok := gas.(string)
		if !ok {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed hexadecimal value, value: %v", gas))
		}
		callObject.Gas = strGas
	}

	if gasPrice, exists := callObj["gasPrice"]; exists {
		strGasPrice, ok := gasPrice.(string)
		if !ok {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed hexadecimal value, value: %v", gasPrice))
		}
		callObject.GasPrice = strGasPrice
	}

	if value, exists := callObj["value"]; exists {
		strValue, ok := value.(string)
		if !ok {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed hexadecimal value, value: %v", value))
		}
		callObject.Value = strValue
	}

	if data, exists := callObj["data"]; exists {
		strData, ok := data.(string)
		if !ok {
			return NewInvalidParamsError(fmt.Sprintf("Expected 0x prefixed hexadecimal value with even length, value: %v", data))
		}
		callObject.Data = strData
	}

	validate := binding.Validator.Engine().(*validator.Validate)
	if err := validate.Struct(&callObject); err != nil {
		message, tag := translateValidationErrors(err)
		return NewInvalidParamsError(fmt.Sprintf("Invalid parameter '%s' for TransactionObject: %s", tag, message))
	}

	p.CallObject = callObject

	block, ok := params[1].(string)
	if !ok {
		return NewInvalidParamsError("block must be a string")
	}
	p.Block = block

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetTransactionByHashParams
func (p *EthGetTransactionByHashParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 1 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 1 parameter, got %d", len(params)))
	}

	txHash, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("transactionHash must be a string")
	}
	p.TransactionHash = txHash

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetTransactionReceiptParams
func (p *EthGetTransactionReceiptParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 1 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 1 parameter, got %d", len(params)))
	}

	txHash, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("transactionHash must be a string")
	}
	p.TransactionHash = txHash

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetBlockTransactionCountByHashParams
func (p *EthGetBlockTransactionCountByHashParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 1 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 1 parameter, got %d", len(params)))
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("blockHash must be a string")
	}
	p.BlockHash = blockHash

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetBlockTransactionCountByNumberParams
func (p *EthGetBlockTransactionCountByNumberParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 1 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 1 parameter, got %d", len(params)))
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetTransactionByBlockHashAndIndexParams
func (p *EthGetTransactionByBlockHashAndIndexParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 2 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 2 parameters, got %d", len(params)))
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("blockHash must be a string")
	}
	p.BlockHash = blockHash

	transactionIndex, ok := params[1].(string)
	if !ok {
		return NewInvalidParamsError("transactionIndex must be a string")
	}
	p.TransactionIndex = transactionIndex

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetTransactionByBlockNumberAndIndexParams
func (p *EthGetTransactionByBlockNumberAndIndexParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 2 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 2 parameters, got %d", len(params)))
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	transactionIndex, ok := params[1].(string)
	if !ok {
		return NewInvalidParamsError("transactionIndex must be a string")
	}
	p.TransactionIndex = transactionIndex

	return nil
}

// FromPositionalParams implements parameter conversion for EthSendRawTransactionParams
func (p *EthSendRawTransactionParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 1 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 1 parameter, got %d", len(params)))
	}

	signedTx, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("signedTransaction must be a string")
	}
	p.SignedTransaction = signedTx

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetCodeParams
func (p *EthGetCodeParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 2 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 2 parameters, got %d", len(params)))
	}

	address, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("address must be a string")
	}
	p.Address = address

	blockNumber, ok := params[1].(string)
	if !ok {
		return NewInvalidParamsError("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetStorageAtParams
func (p *EthGetStorageAtParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) < 2 || len(params) > 3 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 2 or 3 parameters, got %d", len(params)))
	}

	address, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("address must be a string")
	}
	p.Address = address

	storagePosition, ok := params[1].(string)
	if !ok {
		return NewInvalidParamsError("storagePosition must be a string")
	}
	p.StoragePosition = storagePosition

	if len(params) > 2 {
		blockNumber, ok := params[2].(string)
		if !ok {
			return NewInvalidParamsError("blockNumber must be a string")
		}
		p.BlockNumber = blockNumber
	} else {
		p.BlockNumber = "latest"
	}

	return nil
}

// FromPositionalParams implements parameter conversion for EthFeeHistoryParams
func (p *EthFeeHistoryParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) < 2 || len(params) > 3 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 2 or 3 parameters, got %d", len(params)))
	}

	blockCount, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("blockCount must be a string")
	}
	p.BlockCount = blockCount

	newestBlock, ok := params[1].(string)
	if !ok {
		return NewInvalidParamsError("newestBlock must be a string")
	}
	p.NewestBlock = newestBlock
	if len(params) > 2 {
		rawPercentiles, ok := params[2].([]interface{})
		if !ok {
			return NewInvalidParamsError("rewardPercentiles must be an array")
		}

		rewardPercentiles := make([]string, 0, len(rawPercentiles))
		for _, rawPercentile := range rawPercentiles {
			percentile, ok := rawPercentile.(string)
			if !ok {
				return NewInvalidParamsError("each reward percentile must be a string")
			}
			rewardPercentiles = append(rewardPercentiles, percentile)
		}
		p.RewardPercentiles = rewardPercentiles
	}

	return nil
}

type EthNewFilterParams struct {
	FromBlock string   `json:"fromBlock" validate:"omitempty,hexadecimal"`
	ToBlock   string   `json:"toBlock" validate:"omitempty,hexadecimal"`
	Address   Address  `json:"address" validate:"omitempty,dive,eth_address"`
	Topics    []string `json:"topics" validate:"omitempty,dive,hexadecimal"`
}

func (p *EthNewFilterParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) > 0 {
		filterObj, ok := params[0].(map[string]interface{})
		if !ok {
			p.FromBlock = "latest"
			p.ToBlock = "latest"
			return nil
		}
		if fromBlock, ok := filterObj["fromBlock"].(string); ok {
			p.FromBlock = fromBlock
		}
		if toBlock, ok := filterObj["toBlock"].(string); ok {
			p.ToBlock = toBlock
		}
		if address, ok := filterObj["address"].([]string); ok {
			p.Address = address
		}
		if topics, ok := filterObj["topics"].([]string); ok {
			p.Topics = topics
		}
	}
	if p.FromBlock == "" {
		p.FromBlock = "latest"
	}
	if p.ToBlock == "" {
		p.ToBlock = "latest"
	}
	return nil
}

type EthUninstallFilterParams struct {
	FilterID string `json:"filterId" validate:"required,hexadecimal"`
}

func (p *EthUninstallFilterParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) < 1 {
		return NewInvalidParamsError("missing filter ID parameter")
	}
	if filterId, ok := params[0].(string); ok {
		p.FilterID = filterId
		return nil
	}
	return NewInvalidParamsError("invalid filter ID parameter")
}

type EthGetFilterLogsParams struct {
	FilterID string `json:"filterId" validate:"required,hexadecimal"`
}

func (p *EthGetFilterLogsParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) < 1 {
		return NewInvalidParamsError("missing filter ID parameter")
	}
	if filterId, ok := params[0].(string); ok {
		p.FilterID = filterId
		return nil
	}
	return NewInvalidParamsError("invalid filter ID parameter")
}

type EthGetFilterChangesParams struct {
	FilterID string `json:"filterId" validate:"required,hexadecimal"`
}

func (p *EthGetFilterChangesParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) < 1 {
		return NewInvalidParamsError("missing filter ID parameter")
	}
	if filterId, ok := params[0].(string); ok {
		p.FilterID = filterId
		return nil
	}
	return NewInvalidParamsError("invalid filter ID parameter")
}

// EthSubscribeParams represents parameters for eth_subscribe
type EthSubscribeParams struct {
	SubscriptionType string            `json:"subscriptionType" binding:"required"`
	SubscribeOptions *SubscribeOptions `json:"subscribeOptions"`
}

// SubscribeFilterOpts represents filter options for eth_subscribe
type SubscribeOptions struct {
	Address             []string `json:"address,omitempty"`
	Topics              []string `json:"topics,omitempty"`
	IncludeTransactions bool     `json:"includeTransactions,omitempty"`
}

// EthUnsubscribeParams represents parameters for eth_unsubscribe
type EthUnsubscribeParams struct {
	SubscriptionID string `json:"subscriptionId" binding:"required,hexadecimal,startswith=0x"`
}

// FromPositionalParams converts positional parameters to EthSubscribeParams
func (p *EthSubscribeParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) < 1 {
		return NewInvalidParamsError(fmt.Sprintf("Expected at least 1 parameter, got %d", len(params)))
	}

	subscriptionType, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("subscription type must be a string")
	}
	p.SubscriptionType = subscriptionType

	p.SubscribeOptions = &SubscribeOptions{}

	if len(params) > 1 {
		switch filterOptions := params[1].(type) {
		case bool:
			p.SubscribeOptions.IncludeTransactions = filterOptions

		case map[string]interface{}:
			if address, exists := filterOptions["address"]; exists {
				switch addr := address.(type) {
				case string:
					p.SubscribeOptions.Address = []string{addr}
				case []interface{}:
					addresses := make([]string, 0, len(addr))
					for _, a := range addr {
						if addrStr, ok := a.(string); ok {
							addresses = append(addresses, addrStr)
						}
					}
					p.SubscribeOptions.Address = addresses
				}
			}

			if topics, exists := filterOptions["topics"]; exists {
				if topicsArr, ok := topics.([]interface{}); ok {
					topicsStr := make([]string, 0, len(topicsArr))
					for _, t := range topicsArr {
						if topicStr, ok := t.(string); ok {
							topicsStr = append(topicsStr, topicStr)
						}
					}
					p.SubscribeOptions.Topics = topicsStr
				}

			}
		}
	}

	return nil
}

// FromPositionalParams converts positional parameters to EthUnsubscribeParams
func (p *EthUnsubscribeParams) FromPositionalParams(params []interface{}) *RPCError {
	if len(params) != 1 {
		return NewInvalidParamsError(fmt.Sprintf("Expected 1 parameter, got %d", len(params)))
	}

	subscriptionID, ok := params[0].(string)
	if !ok {
		return NewInvalidParamsError("subscription ID must be a string")
	}
	p.SubscriptionID = subscriptionID

	return nil
}

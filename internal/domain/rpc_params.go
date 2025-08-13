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
	FromPositionalParams(params []interface{}) error
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

// FromPositionalParams implements parameter conversion for NoParameters
func (p *NoParameters) FromPositionalParams(params []interface{}) error {
	// No parameters expected
	return nil
}

// FromPositionalParams implements parameter conversion for EthGetBlockByHashParams
func (p *EthGetBlockByHashParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 parameters, got %d", len(params))
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("blockHash must be a string")
	}
	p.BlockHash = blockHash

	showDetails, ok := params[1].(bool)
	if !ok {
		return fmt.Errorf("showDetails must be a boolean")
	}
	p.ShowDetails = showDetails

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetBlockByNumberParams
func (p *EthGetBlockByNumberParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 parameters, got %d", len(params))
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	showDetails, ok := params[1].(bool)
	if !ok {
		return fmt.Errorf("showDetails must be a boolean")
	}
	p.ShowDetails = showDetails

	return nil
}

func (p *EthGetLogsParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("eth_getLogs expects exactly one parameter object")
	}

	filterObj, ok := params[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("eth_getLogs expects a filter object parameter")
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
			return fmt.Errorf("'%s' is not a valid parameter for eth_getLogs", field)
		}
	}

	var filter FilterObject
	filterBytes, err := json.Marshal(filterObj)
	if err != nil {
		return fmt.Errorf("failed to marshal filter object: %v", err)
	}

	if err := json.Unmarshal(filterBytes, &filter); err != nil {
		return fmt.Errorf("failed to unmarshal filter object: %v", err)
	}

	validate := binding.Validator.Engine().(*validator.Validate)
	if err := validate.Struct(&filter); err != nil {
		return fmt.Errorf("invalid filter parameters: %v", err)
	}

	p.Address = filter.Address
	p.Topics = filter.Topics
	p.BlockHash = filter.BlockHash
	p.FromBlock = filter.FromBlock
	p.ToBlock = filter.ToBlock

	if p.BlockHash != "" {
		if p.FromBlock != "" || p.ToBlock != "" {
			return fmt.Errorf("can't use both blockHash and toBlock/fromBlock")
		}
	} else {
		if p.ToBlock == "" {
			p.ToBlock = BlockTagLatest
		}
		if p.FromBlock == "" {
			p.FromBlock = BlockTagLatest
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
func (p *EthGetBalanceParams) FromPositionalParams(params []interface{}) error {
	if len(params) < 1 || len(params) > 2 {
		return fmt.Errorf("expected 1 or 2 parameters, got %d", len(params))
	}

	address, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("address must be a string")
	}
	p.Address = address

	if len(params) > 1 {
		blockNumber, ok := params[1].(string)
		if !ok {
			return fmt.Errorf("blockNumber must be a string")
		}
		p.BlockNumber = blockNumber
	} else {
		p.BlockNumber = BlockTagLatest
	}

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetTransactionCountParams
func (p *EthGetTransactionCountParams) FromPositionalParams(params []interface{}) error {
	if len(params) < 1 || len(params) > 2 {
		return fmt.Errorf("expected 1 or 2 parameters, got %d", len(params))
	}

	address, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("address must be a string")
	}
	p.Address = address

	if len(params) > 1 {
		blockNumber, ok := params[1].(string)
		if !ok {
			return fmt.Errorf("blockNumber must be a string")
		}
		p.BlockNumber = blockNumber
	} else {
		p.BlockNumber = BlockTagLatest
	}

	return nil
}

// FromPositionalParams implements parameter conversion for EthEstimateGasParams
func (p *EthEstimateGasParams) FromPositionalParams(params []interface{}) error {
	if len(params) == 0 || len(params) > 2 {
		return fmt.Errorf("expected 1 or 2 parameters, got %d", len(params))
	}

	callObject, ok := params[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("callObject must be an object")
	}
	p.CallObject = callObject

	if len(params) > 1 {
		blockParam, ok := params[1].(string)
		if !ok {
			return fmt.Errorf("blockParameter must be a string")
		}
		p.BlockParameter = blockParam
	}

	return nil
}

// FromPositionalParams implements parameter conversion for EthCallParams
func (p *EthCallParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 parameters, got %d", len(params))
	}

	callObject, ok := params[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("callObject must be an object")
	}
	p.CallObject = callObject

	block, ok := params[1].(string)
	if !ok {
		return fmt.Errorf("block must be a string")
	}
	p.Block = block

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetTransactionByHashParams
func (p *EthGetTransactionByHashParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 parameter, got %d", len(params))
	}

	txHash, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("transactionHash must be a string")
	}
	p.TransactionHash = txHash

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetTransactionReceiptParams
func (p *EthGetTransactionReceiptParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 parameter, got %d", len(params))
	}

	txHash, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("transactionHash must be a string")
	}
	p.TransactionHash = txHash

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetBlockTransactionCountByHashParams
func (p *EthGetBlockTransactionCountByHashParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 parameter, got %d", len(params))
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("blockHash must be a string")
	}
	p.BlockHash = blockHash

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetBlockTransactionCountByNumberParams
func (p *EthGetBlockTransactionCountByNumberParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 parameter, got %d", len(params))
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetTransactionByBlockHashAndIndexParams
func (p *EthGetTransactionByBlockHashAndIndexParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 parameters, got %d", len(params))
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("blockHash must be a string")
	}
	p.BlockHash = blockHash

	transactionIndex, ok := params[1].(string)
	if !ok {
		return fmt.Errorf("transactionIndex must be a string")
	}
	p.TransactionIndex = transactionIndex

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetTransactionByBlockNumberAndIndexParams
func (p *EthGetTransactionByBlockNumberAndIndexParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 parameters, got %d", len(params))
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	transactionIndex, ok := params[1].(string)
	if !ok {
		return fmt.Errorf("transactionIndex must be a string")
	}
	p.TransactionIndex = transactionIndex

	return nil
}

// FromPositionalParams implements parameter conversion for EthSendRawTransactionParams
func (p *EthSendRawTransactionParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 parameter, got %d", len(params))
	}

	signedTx, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("signedTransaction must be a string")
	}
	p.SignedTransaction = signedTx

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetCodeParams
func (p *EthGetCodeParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 parameters, got %d", len(params))
	}

	address, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("address must be a string")
	}
	p.Address = address

	blockNumber, ok := params[1].(string)
	if !ok {
		return fmt.Errorf("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetStorageAtParams
func (p *EthGetStorageAtParams) FromPositionalParams(params []interface{}) error {
	if len(params) < 2 || len(params) > 3 {
		return fmt.Errorf("expected 2 or 3 parameters, got %d", len(params))
	}

	address, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("address must be a string")
	}
	p.Address = address

	storagePosition, ok := params[1].(string)
	if !ok {
		return fmt.Errorf("storagePosition must be a string")
	}
	p.StoragePosition = storagePosition

	if len(params) > 2 {
		blockNumber, ok := params[2].(string)
		if !ok {
			return fmt.Errorf("blockNumber must be a string")
		}
		p.BlockNumber = blockNumber
	} else {
		p.BlockNumber = BlockTagLatest
	}

	return nil
}

// FromPositionalParams implements parameter conversion for EthFeeHistoryParams
func (p *EthFeeHistoryParams) FromPositionalParams(params []interface{}) error {
	if len(params) < 2 || len(params) > 3 {
		return fmt.Errorf("expected 2 or 3 parameters, got %d", len(params))
	}

	blockCount, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("blockCount must be a string")
	}
	p.BlockCount = blockCount

	newestBlock, ok := params[1].(string)
	if !ok {
		return fmt.Errorf("newestBlock must be a string")
	}
	p.NewestBlock = newestBlock

	if len(params) > 2 {
		rawPercentiles, ok := params[2].([]interface{})
		if !ok {
			return fmt.Errorf("rewardPercentiles must be an array")
		}

		rewardPercentiles := make([]string, 0, len(rawPercentiles))
		for _, rawPercentile := range rawPercentiles {
			percentile, ok := rawPercentile.(string)
			if !ok {
				return fmt.Errorf("each reward percentile must be a string")
			}
			rewardPercentiles = append(rewardPercentiles, percentile)
		}
		p.RewardPercentiles = rewardPercentiles
	}

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetUncleCountByBlockHashParams
func (p *EthGetUncleCountByBlockHashParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 parameter, got %d", len(params))
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("blockHash must be a string")
	}
	p.BlockHash = blockHash

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetUncleCountByBlockNumberParams
func (p *EthGetUncleCountByBlockNumberParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 parameter, got %d", len(params))
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetUncleByBlockHashAndIndexParams
func (p *EthGetUncleByBlockHashAndIndexParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 parameters, got %d", len(params))
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("blockHash must be a string")
	}
	p.BlockHash = blockHash

	index, ok := params[1].(string)
	if !ok {
		return fmt.Errorf("index must be a string")
	}
	p.Index = index

	return nil
}

// FromPositionalParams implements parameter conversion for EthGetUncleByBlockNumberAndIndexParams
func (p *EthGetUncleByBlockNumberAndIndexParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 parameters, got %d", len(params))
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return fmt.Errorf("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	index, ok := params[1].(string)
	if !ok {
		return fmt.Errorf("index must be a string")
	}
	p.Index = index

	return nil
}

type EthNewFilterParams struct {
	FromBlock string   `json:"fromBlock" validate:"omitempty,hexadecimal"`
	ToBlock   string   `json:"toBlock" validate:"omitempty,hexadecimal"`
	Address   Address  `json:"address" validate:"omitempty,dive,eth_address"`
	Topics    []string `json:"topics" validate:"omitempty,dive,hexadecimal"`
}

func (p *EthNewFilterParams) FromPositionalParams(params []interface{}) error {
	if len(params) > 0 {
		filterObj, ok := params[0].(map[string]interface{})
		if !ok {
			p.FromBlock = BlockTagLatest
			p.ToBlock = BlockTagLatest
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
		p.FromBlock = BlockTagLatest
	}
	if p.ToBlock == "" {
		p.ToBlock = BlockTagLatest
	}
	return nil
}

type EthUninstallFilterParams struct {
	FilterID string `json:"filterId" validate:"required,hexadecimal"`
}

func (p *EthUninstallFilterParams) FromPositionalParams(params []interface{}) error {
	if len(params) < 1 {
		return fmt.Errorf("missing filter ID parameter")
	}
	if filterId, ok := params[0].(string); ok {
		p.FilterID = filterId
		return nil
	}
	return fmt.Errorf("invalid filter ID parameter")
}

type EthGetFilterLogsParams struct {
	FilterID string `json:"filterId" validate:"required,hexadecimal"`
}

func (p *EthGetFilterLogsParams) FromPositionalParams(params []interface{}) error {
	if len(params) < 1 {
		return fmt.Errorf("missing filter ID parameter")
	}
	if filterId, ok := params[0].(string); ok {
		p.FilterID = filterId
		return nil
	}
	return fmt.Errorf("invalid filter ID parameter")
}

type EthGetFilterChangesParams struct {
	FilterID string `json:"filterId" validate:"required,hexadecimal"`
}

func (p *EthGetFilterChangesParams) FromPositionalParams(params []interface{}) error {
	if len(params) < 1 {
		return fmt.Errorf("missing filter ID parameter")
	}
	if filterId, ok := params[0].(string); ok {
		p.FilterID = filterId
		return nil
	}
	return fmt.Errorf("invalid filter ID parameter")
}

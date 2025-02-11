package domain

import (
	"fmt"
)

// RPCParams interface defines methods that all RPC parameter structs should implement
type RPCParams interface {
	// FromPositionalParams converts positional parameters (array) to struct fields
	FromPositionalParams(params []interface{}) error
	// FromNamedParams converts named parameters (object) to struct fields
	FromNamedParams(params map[string]interface{}) error
}

// EthGetBlockByHashParams represents parameters for eth_getBlockByHash
type EthGetBlockByHashParams struct {
	BlockHash   string `json:"blockHash" binding:"required,len=66,hexadecimal,startswith=0x"`
	ShowDetails bool   `json:"showDetails" binding:"required"`
}

// EthGetBlockByNumberParams represents parameters for eth_getBlockByNumber
type EthGetBlockByNumberParams struct {
	BlockNumber string `json:"blockNumber" binding:"required,block_number_or_tag"`
	ShowDetails bool   `json:"showDetails" binding:"required"`
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
	RewardPercentiles []string `json:"rewardPercentiles" binding:"omitempty,dive,hexadecimal,startswith=0x"`
}

// EthGetStorageAtParams represents parameters for eth_getStorageAt
type EthGetStorageAtParams struct {
	Address         string `json:"address" binding:"required,eth_address"`
	StoragePosition string `json:"storagePosition" binding:"required,hexadecimal,startswith=0x"`
	BlockNumber     string `json:"blockNumber" binding:"omitempty,block_number_or_tag"`
}

// EthGetLogsParams represents parameters for eth_getLogs
type EthGetLogsParams struct {
	Address   []string `json:"address" binding:"omitempty,dive,eth_address"`
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

func (p *EthGetBlockByHashParams) FromNamedParams(params map[string]interface{}) error {
	blockHash, ok := params["blockHash"].(string)
	if !ok {
		return fmt.Errorf("blockHash must be a string")
	}
	p.BlockHash = blockHash

	showDetails, ok := params["showDetails"].(bool)
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

	if blockNumber, ok := params[0].(string); ok {
		p.BlockNumber = blockNumber
	}
	if showDetails, ok := params[1].(bool); ok {
		p.ShowDetails = showDetails
	}
	return nil
}

func (p *EthGetBlockByNumberParams) FromNamedParams(params map[string]interface{}) error {
	return fmt.Errorf("eth_getBlockByNumber only supports positional parameters")
}

func (p *EthGetLogsParams) FromPositionalParams(params []interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("eth_getLogs expects exactly one parameter object")
	}
	if filterObj, ok := params[0].(map[string]interface{}); ok {
		return p.FromNamedParams(filterObj)
	}

	return fmt.Errorf("eth_getLogs expects a filter object parameter")
}

func (p *EthGetLogsParams) FromNamedParams(params map[string]interface{}) error {
	if address, ok := params["address"]; ok {
		switch addr := address.(type) {
		case string:
			p.Address = []string{addr}
		case []interface{}:
			addresses := make([]string, 0, len(addr))
			for _, a := range addr {
				if strAddr, ok := a.(string); ok {
					addresses = append(addresses, strAddr)
				}
			}
			p.Address = addresses
		}
	}

	if topics, ok := params["topics"].([]interface{}); ok {
		topicsStr := make([]string, 0, len(topics))
		for _, topic := range topics {
			if topicStr, ok := topic.(string); ok {
				topicsStr = append(topicsStr, topicStr)
			}
		}
		p.Topics = topicsStr
	}

	if blockHash, ok := params["blockHash"].(string); ok {
		p.BlockHash = blockHash
	}
	if fromBlock, ok := params["fromBlock"].(string); ok {
		p.FromBlock = fromBlock
	}
	if toBlock, ok := params["toBlock"].(string); ok {
		p.ToBlock = toBlock
	}

	if p.BlockHash == "" && p.FromBlock == "" && p.ToBlock == "" {
		p.FromBlock = "latest"
		p.ToBlock = "latest"
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
		p.BlockNumber = "latest"
	}

	return nil
}

func (p *EthGetBalanceParams) FromNamedParams(params map[string]interface{}) error {
	address, ok := params["address"].(string)
	if !ok {
		return fmt.Errorf("address must be a string")
	}
	p.Address = address

	if blockNumber, ok := params["blockNumber"].(string); ok {
		p.BlockNumber = blockNumber
	} else {
		p.BlockNumber = "latest"
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
		p.BlockNumber = "latest"
	}

	return nil
}

func (p *EthGetTransactionCountParams) FromNamedParams(params map[string]interface{}) error {
	address, ok := params["address"].(string)
	if !ok {
		return fmt.Errorf("address must be a string")
	}
	p.Address = address

	if blockNumber, ok := params["blockNumber"].(string); ok {
		p.BlockNumber = blockNumber
	} else {
		p.BlockNumber = "latest"
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

func (p *EthEstimateGasParams) FromNamedParams(params map[string]interface{}) error {
	callObject, ok := params["callObject"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("callObject must be an object")
	}
	p.CallObject = callObject

	if blockParam, ok := params["blockParameter"].(string); ok {
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

func (p *EthCallParams) FromNamedParams(params map[string]interface{}) error {
	callObject, ok := params["callObject"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("callObject must be an object")
	}
	p.CallObject = callObject

	block, ok := params["block"].(string)
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

func (p *EthGetTransactionByHashParams) FromNamedParams(params map[string]interface{}) error {
	txHash, ok := params["transactionHash"].(string)
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

func (p *EthGetTransactionReceiptParams) FromNamedParams(params map[string]interface{}) error {
	txHash, ok := params["transactionHash"].(string)
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

func (p *EthGetBlockTransactionCountByHashParams) FromNamedParams(params map[string]interface{}) error {
	blockHash, ok := params["blockHash"].(string)
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

func (p *EthGetBlockTransactionCountByNumberParams) FromNamedParams(params map[string]interface{}) error {
	blockNumber, ok := params["blockNumber"].(string)
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

func (p *EthGetTransactionByBlockHashAndIndexParams) FromNamedParams(params map[string]interface{}) error {
	blockHash, ok := params["blockHash"].(string)
	if !ok {
		return fmt.Errorf("blockHash must be a string")
	}
	p.BlockHash = blockHash

	transactionIndex, ok := params["transactionIndex"].(string)
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

func (p *EthGetTransactionByBlockNumberAndIndexParams) FromNamedParams(params map[string]interface{}) error {
	blockNumber, ok := params["blockNumber"].(string)
	if !ok {
		return fmt.Errorf("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	transactionIndex, ok := params["transactionIndex"].(string)
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

func (p *EthSendRawTransactionParams) FromNamedParams(params map[string]interface{}) error {
	signedTx, ok := params["signedTransaction"].(string)
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

func (p *EthGetCodeParams) FromNamedParams(params map[string]interface{}) error {
	address, ok := params["address"].(string)
	if !ok {
		return fmt.Errorf("address must be a string")
	}
	p.Address = address

	blockNumber, ok := params["blockNumber"].(string)
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
		p.BlockNumber = "latest"
	}

	return nil
}

func (p *EthGetStorageAtParams) FromNamedParams(params map[string]interface{}) error {
	address, ok := params["address"].(string)
	if !ok {
		return fmt.Errorf("address must be a string")
	}
	p.Address = address

	storagePosition, ok := params["storagePosition"].(string)
	if !ok {
		return fmt.Errorf("storagePosition must be a string")
	}
	p.StoragePosition = storagePosition

	if blockNumber, ok := params["blockNumber"].(string); ok {
		p.BlockNumber = blockNumber
	} else {
		p.BlockNumber = "latest"
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

func (p *EthFeeHistoryParams) FromNamedParams(params map[string]interface{}) error {
	blockCount, ok := params["blockCount"].(string)
	if !ok {
		return fmt.Errorf("blockCount must be a string")
	}
	p.BlockCount = blockCount

	newestBlock, ok := params["newestBlock"].(string)
	if !ok {
		return fmt.Errorf("newestBlock must be a string")
	}
	p.NewestBlock = newestBlock

	if rawPercentiles, ok := params["rewardPercentiles"].([]interface{}); ok {
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

func (p *EthGetUncleCountByBlockHashParams) FromNamedParams(params map[string]interface{}) error {
	blockHash, ok := params["blockHash"].(string)
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

func (p *EthGetUncleCountByBlockNumberParams) FromNamedParams(params map[string]interface{}) error {
	blockNumber, ok := params["blockNumber"].(string)
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

func (p *EthGetUncleByBlockHashAndIndexParams) FromNamedParams(params map[string]interface{}) error {
	blockHash, ok := params["blockHash"].(string)
	if !ok {
		return fmt.Errorf("blockHash must be a string")
	}
	p.BlockHash = blockHash

	index, ok := params["index"].(string)
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

func (p *EthGetUncleByBlockNumberAndIndexParams) FromNamedParams(params map[string]interface{}) error {
	blockNumber, ok := params["blockNumber"].(string)
	if !ok {
		return fmt.Errorf("blockNumber must be a string")
	}
	p.BlockNumber = blockNumber

	index, ok := params["index"].(string)
	if !ok {
		return fmt.Errorf("index must be a string")
	}
	p.Index = index

	return nil
}

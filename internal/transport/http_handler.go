package transport

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      interface{} `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

func rpcHandler(ctx *gin.Context) {
	var req JSONRPCRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request",
			},
		})
		return
	}
	methodName := req.Method
	logger.Info("JSON-RPC method called", zap.String("method", methodName))

	result, rpcErr := dispatchMethod(ctx, methodName, req.Params)
	resp := JSONRPCResponse{JSONRPC: "2.0", ID: req.ID}
	if rpcErr != nil {
		resp.Error = rpcErr
		ctx.JSON(http.StatusBadRequest, resp)
	} else {
		resp.Result = result
		ctx.JSON(http.StatusOK, resp)
	}
}

func dispatchMethod(ctx *gin.Context, methodName string, params interface{}) (interface{}, map[string]interface{}) {
	switch methodName {
	case "eth_getBlockByHash":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) != 2 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getBlockByHash: expected [blockHash, showDetails]",
			}
		}

		// Type assert and validate blockHash
		blockHash, ok := paramsArray[0].(string)
		if !ok || len(blockHash) != 66 || blockHash[:2] != "0x" {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid blockHash: expected 32 byte hex string with 0x prefix",
			}
		}

		// Type assert showDetails parameter
		showDetails, ok := paramsArray[1].(bool)
		if !ok {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid showDetails: expected boolean",
			}
		}

		return ethService.GetBlockByHash(blockHash, showDetails)
	case "eth_getBlockByNumber":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) != 2 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getBlockByNumber: expected [blockNumber, showDetails]",
			}
		}

		// Type assert and validate block number/tag
		blockNumber, ok := paramsArray[0].(string)
		if !ok {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid blockNumber: expected string",
			}
		}

		// Validate if it's a tag or hex number
		if blockNumber != "earliest" && blockNumber != "latest" && blockNumber != "pending" {
			if !strings.HasPrefix(blockNumber, "0x") {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid blockNumber: expected hex string with 0x prefix or tag (earliest/latest/pending)",
				}
			}
		}

		// Type assert showDetails parameter
		showDetails, ok := paramsArray[1].(bool)
		if !ok {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid showDetails: expected boolean",
			}
		}
		return ethService.GetBlockByNumber(blockNumber, showDetails)
	case "eth_getBalance":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) < 1 || len(paramsArray) > 2 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getBalance: expected [address, blockNumber]",
			}
		}

		// Type assert and validate address parameter
		address, ok := paramsArray[0].(string)
		if !ok {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid address: expected string",
			}
		}

		// Validate address format (0x followed by 40 hex chars)
		if !strings.HasPrefix(address, "0x") || len(address) != 42 || !regexp.MustCompile("^0x[a-fA-F0-9]{40}$").MatchString(address) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid address format: expected 0x followed by 40 hexadecimal characters",
			}
		}

		// Handle optional blockNumber parameter
		var blockNumber string = "latest" // default value
		if len(paramsArray) > 1 {
			blockNumber, ok = paramsArray[1].(string)
			if !ok {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid blockNumber: expected string",
				}
			}

			// Validate if it's a tag or hex number
			if blockNumber != "earliest" && blockNumber != "latest" && blockNumber != "pending" {
				if !strings.HasPrefix(blockNumber, "0x") {
					return nil, map[string]interface{}{
						"code":    -32602,
						"message": "Invalid blockNumber: expected hex string with 0x prefix or tag (earliest/latest/pending)",
					}
				}
			}
		}

		return ethService.GetBalance(address, blockNumber), nil
	case "eth_getTransactionCount":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) < 1 || len(paramsArray) > 2 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getTransactionCount: expected [address, blockNumber]",
			}
		}

		// Type assert and validate address parameter
		address, ok := paramsArray[0].(string)
		if !ok {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid address: expected string",
			}
		}

		// Validate address format (0x followed by 40 hex chars)
		if !strings.HasPrefix(address, "0x") || len(address) != 42 || !regexp.MustCompile("^0x[a-fA-F0-9]{40}$").MatchString(address) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid address format: expected 0x followed by 40 hexadecimal characters",
			}
		}

		// Handle optional blockNumber parameter
		var blockNumber string = "latest" // default value
		if len(paramsArray) > 1 {
			blockNumber, ok = paramsArray[1].(string)
			if !ok {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid blockNumber: expected string",
				}
			}

			// Validate if it's a tag or hex number
			if blockNumber != "earliest" && blockNumber != "latest" && blockNumber != "pending" {
				if !strings.HasPrefix(blockNumber, "0x") {
					return nil, map[string]interface{}{
						"code":    -32602,
						"message": "Invalid blockNumber: expected hex string with 0x prefix or tag (earliest/latest/pending)",
					}
				}
			}
		}

		return ethService.GetTransactionCount(address, blockNumber), nil
	case "eth_estimateGas":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) == 0 || len(paramsArray) > 2 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_estimateGas: expected [callObject] or [callObject, blockParameter]",
			}
		}

		// Second parameter is optional, validate if provided
		var secondParam interface{}
		if len(paramsArray) == 2 {
			blockParam, ok := paramsArray[1].(string)
			if !ok {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid block parameter: expected string",
				}
			}

			validTags := map[string]bool{
				"latest":    true,
				"pending":   true,
				"earliest":  true,
				"safe":      true,
				"finalized": true,
			}

			if !validTags[blockParam] && !strings.HasPrefix(blockParam, "0x") {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid block parameter: must be a tag (latest/pending/earliest/safe/finalized) or hex number",
				}
			}
			secondParam = blockParam
		}

		return ethService.EstimateGas(paramsArray[0], secondParam)
	case "eth_call":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) != 2 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_call: two parameters are required",
			}
		}

		blockParam, ok := paramsArray[1].(string)
		if !ok {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid block parameter: expected string",
			}
		}

		validTags := map[string]bool{
			"latest":    true,
			"pending":   true,
			"earliest":  true,
			"safe":      true,
			"finalized": true,
		}

		if !validTags[blockParam] && !strings.HasPrefix(blockParam, "0x") {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid block parameter: must be a tag (latest/pending/earliest/safe/finalized) or hex number",
			}
		}

		return ethService.Call(paramsArray[0], blockParam)
	case "eth_getTransactionByHash":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) != 1 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getTransactionByHash: expected [transactionHash]",
			}
		}

		txHash, ok := paramsArray[0].(string)
		if !ok || len(txHash) != 66 || !strings.HasPrefix(txHash, "0x") || !regexp.MustCompile("^0x[0-9a-fA-F]{64}$").MatchString(txHash) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid transaction hash: expected 32 byte hex string with 0x prefix",
			}
		}

		return ethService.GetTransactionByHash(txHash), nil
	case "eth_getTransactionReceipt":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) != 1 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getTransactionReceipt: expected [transactionHash]",
			}
		}

		txHash, ok := paramsArray[0].(string)
		if !ok || len(txHash) != 66 || !strings.HasPrefix(txHash, "0x") || !regexp.MustCompile("^0x[0-9a-fA-F]{64}$").MatchString(txHash) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid transaction hash: expected 32 byte hex string with 0x prefix",
			}
		}

		return ethService.GetTransactionReceipt(txHash), nil
	case "eth_feeHistory":
		paramsArray, ok := params.([]interface{})
		if !ok || (len(paramsArray) != 2 && len(paramsArray) != 3) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_feeHistory: expected [blockCount, newestBlock] or [blockCount, newestBlock, rewardPercentiles]",
			}
		}

		blockCount, ok := paramsArray[0].(string)
		if !ok || !regexp.MustCompile("^0x[0-9a-fA-F]+$").MatchString(blockCount) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid blockCount: expected hex string with 0x prefix",
			}
		}

		newestBlock, ok := paramsArray[1].(string)
		if !ok {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid newestBlock: expected hex string with 0x prefix or latest/pending/earliest",
			}
		}

		if newestBlock != "latest" && newestBlock != "pending" && newestBlock != "earliest" {
			if !strings.HasPrefix(newestBlock, "0x") && !regexp.MustCompile("^0x[0-9a-fA-F]$").MatchString(newestBlock) {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid newestBlock: expected hex string with 0x prefix or latest/pending/earliest",
				}
			}
		}

		// We should check if the rewardPercentiles is a list of monotonically increasing integer (maybe)
		var rewardPercentiles []string
		if len(paramsArray) == 3 {
			rawRewardPercentiles, ok := paramsArray[2].([]interface{})
			if !ok {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid rewardPercentiles: expected list of strings",
				}
			}

			for _, rawPercentile := range rawRewardPercentiles {
				percentile, ok := rawPercentile.(string)
				if !ok || !strings.HasPrefix(percentile, "0x") || !regexp.MustCompile("^0x[0-9a-fA-F]$").MatchString(percentile) {
					return nil, map[string]interface{}{
						"code":    -32602,
						"message": "Invalid rewardPercentiles: expected list of strings",
					}
				}
				rewardPercentiles = append(rewardPercentiles, percentile)
			}
		}

		if len(paramsArray) == 2 {
			return ethService.FeeHistory(blockCount, newestBlock, nil)
		}

		return ethService.FeeHistory(blockCount, newestBlock, rewardPercentiles)
	case "eth_getStorageAt":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) < 2 || len(paramsArray) > 3 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getStorageAt: expected [address, slot, blockNumberOrTag] or [address, slot]",
			}
		}

		address, ok := paramsArray[0].(string)
		if !ok || !IsValidAddress(address) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid address: must be a 20-byte hex string starting with 0x",
			}
		}

		slot, ok := paramsArray[1].(string)
		if !ok || !IsValidHexNumber(slot) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid slot: must be a hex string starting with 0x (e.g. 0x0)",
			}
		}

		if len(paramsArray) == 2 {
			return ethService.GetStorageAt(address, slot, "latest")
		}

		blockNumberOrTag, ok := paramsArray[2].(string)
		if !ok || !IsValidBlockNumberOrTag(blockNumberOrTag) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid block parameter: must be a tag (latest/pending/earliest) or hex number starting with 0x",
			}
		}

		return ethService.GetStorageAt(address, slot, blockNumberOrTag)
	case "eth_getLogs":
		paramsArray, ok := params.([]interface{})
		if !ok {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getLogs: expected object",
			}
		}

		paramsMap, ok := paramsArray[0].(map[string]interface{})

		if !ok {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getLogs: expected object",
			}
		}

		logParams := &domain.LogParams{}

		if address, ok := paramsMap["address"].(string); ok {
			if !IsValidAddress(address) {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid address format: expected 0x followed by 40 hexadecimal characters",
				}
			}
			logParams.Address = []string{address}
		} else if addresses, ok := paramsMap["address"].([]interface{}); ok {
			for _, addr := range addresses {
				if addrStr, ok := addr.(string); ok {
					if !IsValidAddress(addrStr) {
						return nil, map[string]interface{}{
							"code":    -32602,
							"message": "Invalid address format: expected 0x followed by 40 hexadecimal characters",
						}
					}
					logParams.Address = append(logParams.Address, addrStr)
				}
			}
		}

		if topics, ok := paramsMap["topics"].([]interface{}); ok {
			for _, topic := range topics {
				if topicStr, ok := topic.(string); ok {
					if !IsValidHexHash(topicStr) {
						return nil, map[string]interface{}{
							"code":    -32602,
							"message": "Invalid topic format: expected 0x followed by 64 hexadecimal characters",
						}
					}
					logParams.Topics = append(logParams.Topics, topicStr)
				}
			}
		}

		if blockHash, ok := paramsMap["blockHash"].(string); ok {
			if !IsValidBlockHashOrTag(blockHash) {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid blockHash format: expected 0x followed by 64 hexadecimal characters",
				}
			}
			logParams.BlockHash = blockHash
		}

		if fromBlock, ok := paramsMap["fromBlock"].(string); ok {
			if !IsValidBlockNumberOrTag(fromBlock) {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid fromBlock: must be a hex string (e.g. '0x1')",
				}
			}
			logParams.FromBlock = fromBlock
		}

		if toBlock, ok := paramsMap["toBlock"].(string); ok {
			if !IsValidBlockNumberOrTag(toBlock) {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid toBlock: must be a hex string (e.g. '0x1')",
				}
			}
			logParams.ToBlock = toBlock
		}

		if logParams.BlockHash != "" {
			if logParams.FromBlock != "" || logParams.ToBlock != "" {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid params: Can't use both blockHash and toBlock/fromBlock",
				}
			}
		} else {
			if logParams.FromBlock != "" && logParams.ToBlock == "" {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid params: Provided fromBlock parameter without specifying toBlock",
				}
			}
			if logParams.ToBlock != "" && logParams.FromBlock == "" {
				return nil, map[string]interface{}{
					"code":    -32602,
					"message": "Invalid params: Provided toBlock parameter without specifying fromBlock",
				}
			}
			if logParams.FromBlock == "" && logParams.ToBlock == "" {
				logParams.FromBlock = "latest"
				logParams.ToBlock = "latest"
			}

		}


		return ethService.GetLogs(*logParams)
	case "eth_getBlockTransactionCountByHash":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) != 1 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getBlockTransactionCountByHash: expected [blockHash]",
			}
		}
		blockHash, ok := paramsArray[0].(string)
		if !ok || !IsValidBlockHash(blockHash) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid blockHash: expected 0x followed by 64 hexadecimal characters",
			}
		}

		return ethService.GetBlockTransactionCountByHash(blockHash)
	case "eth_getBlockTransactionCountByNumber":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) != 1 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getBlockTransactionCountByNumber: expected [blockNumber]",
			}
		}

		blockNumber, ok := paramsArray[0].(string)
		if !ok || !IsValidBlockNumberOrTag(blockNumber) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid blockNumber: must be a hex string (e.g. '0x1') or tag (latest/pending/earliest)",
			}
		}

		return ethService.GetBlockTransactionCountByNumber(blockNumber)
	case "eth_getTransactionByBlockHashAndIndex":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) != 2 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getTransactionByBlockHashAndIndex: expected [blockHash, transactionIndex]",
			}
		}

		blockHash, ok := paramsArray[0].(string)
		if !ok || !IsValidBlockHash(blockHash) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid blockHash: expected 0x followed by 64 hexadecimal characters",
			}
		}

		txIndex, ok := paramsArray[1].(string)
		if !ok || !IsValidHexNumber(txIndex) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid transactionIndex: expected hex string with 0x prefix",
			}
		}

		return ethService.GetTransactionByBlockHashAndIndex(blockHash, txIndex)
	case "eth_getTransactionByBlockNumberAndIndex":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) != 2 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_getTransactionByBlockHashAndIndex: expected [blockHash, transactionIndex]",
			}
		}

		blockNumber, ok := paramsArray[0].(string)
		if !ok || !IsValidBlockNumberOrTag(blockNumber) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid blockNumber: must be a hex string (e.g. '0x1') or tag (latest/pending/earliest)",
			}
		}

		txIndex, ok := paramsArray[1].(string)
		if !ok || !IsValidHexNumber(txIndex) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid transactionIndex: expected hex string with 0x prefix",
			}
		}

		return ethService.GetTransactionByBlockNumberAndIndex(blockNumber, txIndex)
	case "eth_sendRawTransaction":
		paramsArray, ok := params.([]interface{})
		if !ok || len(paramsArray) != 1 {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params for eth_sendRawTransaction: expected [signedTransaction]",
			}
		}

		signedTx, ok := paramsArray[0].(string)
		if !ok || !IsValidHexNumber(signedTx) {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid signedTransaction: expected data with 0x prefix",
			}
		}

		return ethService.SendRawTransaction(signedTx)
	case "eth_blockNumber":
		return ethService.GetBlockNumber()
	case "eth_gasPrice":
		return ethService.GetGasPrice()
	case "eth_chainId":
		return ethService.GetChainId()
	case "eth_accounts":
		return ethService.GetAccounts()
	case "web3_clientVersion":
		return web3Service.ClientVersion(), nil
	case "net_listening":
		return netService.Listening(), nil
	case "net_version":
		return netService.Version(), nil
	case "eth_syncing":
		return ethService.Syncing()
	case "eth_mining":
		return ethService.Mining()
	case "eth_maxPriorityFeePerGas":
		return ethService.MaxPriorityFeePerGas()
	case "eth_hashrate":
		return ethService.Hashrate()
	case "eth_getUncleCountByBlockNumber":
		return ethService.GetUncleCountByBlockNumber()
	case "eth_getUncleByBlockNumberAndIndex":
		return ethService.GetUncleByBlockNumberAndIndex()
	case "eth_getUncleCountByBlockHash":
		return ethService.GetUncleCountByBlockHash()
	case "eth_getUncleByBlockHashAndIndex":
		return ethService.GetUncleByBlockHashAndIndex()
	default:
		return nil, unsupportedMethodError(methodName)
	}
}

func unsupportedMethodError(methodName string) map[string]interface{} {
	return map[string]interface{}{
		"code":    -32601,
		"message": fmt.Sprintf("Unsupported JSON-RPC method: %s", methodName),
		"name":    "Method not found",
	}
}

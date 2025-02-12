package transport

import (
	"fmt"
	"net/http"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
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

var methodParamsMap = map[string]func() domain.RPCParams{
	"eth_blockNumber":                         func() domain.RPCParams { return &domain.EthBlockNumberParams{} },
	"eth_gasPrice":                            func() domain.RPCParams { return &domain.EthGasPriceParams{} },
	"eth_chainId":                             func() domain.RPCParams { return &domain.EthChainIdParams{} },
	"eth_getBlockByHash":                      func() domain.RPCParams { return &domain.EthGetBlockByHashParams{} },
	"eth_getBlockByNumber":                    func() domain.RPCParams { return &domain.EthGetBlockByNumberParams{} },
	"eth_getLogs":                             func() domain.RPCParams { return &domain.EthGetLogsParams{} },
	"eth_getBalance":                          func() domain.RPCParams { return &domain.EthGetBalanceParams{} },
	"eth_getTransactionCount":                 func() domain.RPCParams { return &domain.EthGetTransactionCountParams{} },
	"eth_estimateGas":                         func() domain.RPCParams { return &domain.EthEstimateGasParams{} },
	"eth_call":                                func() domain.RPCParams { return &domain.EthCallParams{} },
	"eth_getTransactionByHash":                func() domain.RPCParams { return &domain.EthGetTransactionByHashParams{} },
	"eth_getTransactionReceipt":               func() domain.RPCParams { return &domain.EthGetTransactionReceiptParams{} },
	"eth_getBlockTransactionCountByHash":      func() domain.RPCParams { return &domain.EthGetBlockTransactionCountByHashParams{} },
	"eth_getBlockTransactionCountByNumber":    func() domain.RPCParams { return &domain.EthGetBlockTransactionCountByNumberParams{} },
	"eth_getTransactionByBlockHashAndIndex":   func() domain.RPCParams { return &domain.EthGetTransactionByBlockHashAndIndexParams{} },
	"eth_getTransactionByBlockNumberAndIndex": func() domain.RPCParams { return &domain.EthGetTransactionByBlockNumberAndIndexParams{} },
	"eth_sendRawTransaction":                  func() domain.RPCParams { return &domain.EthSendRawTransactionParams{} },
	"eth_getCode":                             func() domain.RPCParams { return &domain.EthGetCodeParams{} },
	"eth_getStorageAt":                        func() domain.RPCParams { return &domain.EthGetStorageAtParams{} },
	"eth_feeHistory":                          func() domain.RPCParams { return &domain.EthFeeHistoryParams{} },
	"eth_getUncleCountByBlockHash":            func() domain.RPCParams { return &domain.EthGetUncleCountByBlockHashParams{} },
	"eth_getUncleCountByBlockNumber":          func() domain.RPCParams { return &domain.EthGetUncleCountByBlockNumberParams{} },
	"eth_getUncleByBlockHashAndIndex":         func() domain.RPCParams { return &domain.EthGetUncleByBlockHashAndIndexParams{} },
	"eth_getUncleByBlockNumberAndIndex":       func() domain.RPCParams { return &domain.EthGetUncleByBlockNumberAndIndexParams{} },
}

func init() {
	if err := RegisterCustomValidators(); err != nil {
		panic(fmt.Sprintf("Failed to register custom validators: %v", err))
	}
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
	paramsCreator, ok := methodParamsMap[methodName]
	if !ok {
		return nil, unsupportedMethodError(methodName)
	}

	logger.Debug("Received params", zap.Any("params", params))

	rpcParams := paramsCreator()

	switch p := params.(type) {
	case []interface{}:
		logger.Debug("Processing array params", zap.Any("array_params", p))
		if err := rpcParams.FromPositionalParams(p); err != nil {
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": fmt.Sprintf("Invalid params: %s", err.Error()),
			}
		}
	default:
		logger.Debug("Invalid params type", zap.String("type", fmt.Sprintf("%T", params)))
		return nil, map[string]interface{}{
			"code":    -32602,
			"message": "Invalid params: expected array or object",
		}
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.Struct(rpcParams); err != nil {
			logger.Debug("Validation failed", zap.Error(err))
			return nil, map[string]interface{}{
				"code":    -32602,
				"message": "Invalid params",
			}
		}
	}

	switch methodName {
	case "eth_getBlockByHash":
		params := rpcParams.(*domain.EthGetBlockByHashParams)
		return ethService.GetBlockByHash(params.BlockHash, params.ShowDetails)
	case "eth_getBlockByNumber":
		params := rpcParams.(*domain.EthGetBlockByNumberParams)
		return ethService.GetBlockByNumber(params.BlockNumber, params.ShowDetails)
	case "eth_getLogs":
		params := rpcParams.(*domain.EthGetLogsParams)
		logParams := params.ToLogParams()
		return ethService.GetLogs(logParams)
	case "eth_getBalance":
		params := rpcParams.(*domain.EthGetBalanceParams)
		return ethService.GetBalance(params.Address, params.BlockNumber), nil
	case "eth_getTransactionCount":
		params := rpcParams.(*domain.EthGetTransactionCountParams)
		return ethService.GetTransactionCount(params.Address, params.BlockNumber), nil
	case "eth_estimateGas":
		params := rpcParams.(*domain.EthEstimateGasParams)
		return ethService.EstimateGas(params.CallObject, params.BlockParameter)
	case "eth_call":
		params := rpcParams.(*domain.EthCallParams)
		return ethService.Call(params.CallObject, params.Block)
	case "eth_getTransactionByHash":
		params := rpcParams.(*domain.EthGetTransactionByHashParams)
		return ethService.GetTransactionByHash(params.TransactionHash), nil
	case "eth_getTransactionReceipt":
		params := rpcParams.(*domain.EthGetTransactionReceiptParams)
		return ethService.GetTransactionReceipt(params.TransactionHash)
	case "eth_feeHistory":
		params := rpcParams.(*domain.EthFeeHistoryParams)
		return ethService.FeeHistory(params.BlockCount, params.NewestBlock, params.RewardPercentiles)
	case "eth_getStorageAt":
		params := rpcParams.(*domain.EthGetStorageAtParams)
		return ethService.GetStorageAt(params.Address, params.StoragePosition, params.BlockNumber)
	case "eth_getBlockTransactionCountByHash":
		params := rpcParams.(*domain.EthGetBlockTransactionCountByHashParams)
		return ethService.GetBlockTransactionCountByHash(params.BlockHash)
	case "eth_getBlockTransactionCountByNumber":
		params := rpcParams.(*domain.EthGetBlockTransactionCountByNumberParams)
		return ethService.GetBlockTransactionCountByNumber(params.BlockNumber)
	case "eth_getTransactionByBlockHashAndIndex":
		params := rpcParams.(*domain.EthGetTransactionByBlockHashAndIndexParams)
		return ethService.GetTransactionByBlockHashAndIndex(params.BlockHash, params.TransactionIndex)
	case "eth_getTransactionByBlockNumberAndIndex":
		params := rpcParams.(*domain.EthGetTransactionByBlockNumberAndIndexParams)
		return ethService.GetTransactionByBlockNumberAndIndex(params.BlockNumber, params.TransactionIndex)
	case "eth_sendRawTransaction":
		params := rpcParams.(*domain.EthSendRawTransactionParams)
		return ethService.SendRawTransaction(params.SignedTransaction)
	case "eth_getCode":
		params := rpcParams.(*domain.EthGetCodeParams)
		return ethService.GetCode(params.Address, params.BlockNumber)
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
	case "eth_getUncleCountByBlockHash":
		params := rpcParams.(*domain.EthGetUncleCountByBlockHashParams)
		return ethService.GetUncleCountByBlockHash(params.BlockHash)
	case "eth_getUncleCountByBlockNumber":
		params := rpcParams.(*domain.EthGetUncleCountByBlockNumberParams)
		return ethService.GetUncleCountByBlockNumber(params.BlockNumber)
	case "eth_getUncleByBlockHashAndIndex":
		params := rpcParams.(*domain.EthGetUncleByBlockHashAndIndexParams)
		return ethService.GetUncleByBlockHashAndIndex(params.BlockHash, params.Index)
	case "eth_getUncleByBlockNumberAndIndex":
		params := rpcParams.(*domain.EthGetUncleByBlockNumberAndIndexParams)
		return ethService.GetUncleByBlockNumberAndIndex(params.BlockNumber, params.Index)
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

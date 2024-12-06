package transport

import (
	"fmt"
	"net/http"

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
	Result  interface{} `json:"result,omitempty"`
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
	case "eth_blockNumber":
		return ethService.GetBlockNumber()
	// case "eth_sendRawTransaction":
	// 	paramArr, ok := params.([]interface{})
	// 	if !ok || len(paramArr) < 1 {
	// 		return nil, map[string]interface{}{"code": -32602, "message": "Invalid params"}
	// 	}
	// 	rawTxHex := paramArr[0].(string)
	// 	rawTx, err := hex.DecodeString(rawTxHex[2:])
	// 	if err != nil {
	// 		return nil, map[string]interface{}{"code": -32602, "message": "Invalid hex"}
	// 	}
	// 	return ethService.SendRawTransaction(ctx, rawTx)
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

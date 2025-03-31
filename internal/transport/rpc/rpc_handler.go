package rpc

import (
	"context"
	"fmt"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/service"
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
	JSONRPC string           `json:"jsonrpc"`
	Result  interface{}      `json:"result"`
	Error   *domain.RPCError `json:"error,omitempty"`
	ID      interface{}      `json:"id,omitempty"`
}

type RPCHandler interface {
	HandleRequest(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse
}

type rpcHandler struct {
	logger   *zap.Logger
	registry *Methods
	services service.ServiceProvider
}

func NewHandler(
	logger *zap.Logger,
	services service.ServiceProvider,
) RPCHandler {
	return &rpcHandler{
		logger:   logger,
		registry: NewMethods(),
		services: services,
	}
}

func (h *rpcHandler) HandleRequest(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	methodName := req.Method
	h.logger.Info("JSON-RPC method called", zap.String("method", methodName))

	result, rpcErr := h.dispatchMethod(ctx, methodName, req.Params)
	resp := &JSONRPCResponse{JSONRPC: "2.0", ID: req.ID}
	if rpcErr != nil {
		resp.Error = rpcErr
	} else {
		resp.Result = result
	}
	return resp
}

func (h *rpcHandler) dispatchMethod(ctx context.Context, methodName string, params interface{}) (interface{}, *domain.RPCError) {
	methodInfo, ok := h.registry.GetMethod(methodName)
	if !ok {
		return nil, domain.NewRPCError(domain.MethodNotFound, "Unsupported JSON-RPC method")
	}

	h.logger.Debug("Received params", zap.Any("params", params))

	rpcParams := methodInfo.ParamCreator()

	switch p := params.(type) {
	case []interface{}:
		h.logger.Debug("Processing array params", zap.Any("array_params", p))
		if err := rpcParams.FromPositionalParams(p); err != nil {
			return nil, domain.NewRPCError(domain.InvalidParams, err.Error())
		}
	default:
		h.logger.Debug("Invalid params type", zap.String("type", fmt.Sprintf("%T", params)))
		return nil, domain.NewRPCError(domain.InvalidParams, "Invalid params: expected array or object")
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.Struct(rpcParams); err != nil {
			h.logger.Debug("Validation failed", zap.Error(err))
			return nil, domain.NewRPCError(domain.InvalidParams, err.Error())
		}
	}

	return methodInfo.Handler(ctx, rpcParams, h.services)
}

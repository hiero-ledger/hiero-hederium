package http_server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hiero-ledger/hiero-hederium/internal/domain"
	"github.com/hiero-ledger/hiero-hederium/internal/transport/rpc"
	"go.uber.org/zap"
)

// JSONRPCRecovery recovers from a panic anywhere in the HTTP request chain and
// answers with a JSON-RPC internal error (-32603) instead of gin's empty
// HTTP 500. Panics inside a JSON-RPC method never reach it: HandleRequest
// recovers those per request, so one bad batch item does not fail the others.
func JSONRPCRecovery(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
		logger.Error("Recovered from panic while serving HTTP request",
			zap.Any("panic", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.Stack("stack"),
		)
		c.AbortWithStatusJSON(http.StatusInternalServerError, rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   domain.NewInternalError("Internal error"),
		})
	})
}

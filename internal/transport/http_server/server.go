package http_server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	"github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/LimeChain/Hederium/internal/infrastructure/limiter"
	"github.com/LimeChain/Hederium/internal/service"
	"github.com/LimeChain/Hederium/internal/transport/rpc"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server interface {
	Start() error
}

type server struct {
	router          *gin.Engine
	logger          *zap.Logger
	port            string
	serviceProvider service.ServiceProvider
	apiKeyStore     *limiter.APIKeyStore
	tieredLimiter   *limiter.TieredLimiter
	enforceAPIKey   bool
	rpcHandler      rpc.RPCHandler
}

func NewServer(
	hClient *hedera.HederaClient,
	mClient *hedera.MirrorClient,
	logger *zap.Logger,
	applicationVersion string,
	chainId string,
	apiKeyStore *limiter.APIKeyStore,
	tieredLimiter *limiter.TieredLimiter,
	enforceAPIKey bool,
	cacheService cache.CacheService,
	port string,
) Server {
	serviceProvider := service.NewServiceProvider(hClient, mClient, logger, applicationVersion, chainId, apiKeyStore, tieredLimiter, cacheService)

	router := gin.Default()

	rpcHandler := rpc.NewHandler(
		logger,
		serviceProvider,
	)

	s := &server{
		router:          router,
		logger:          logger,
		port:            port,
		serviceProvider: serviceProvider,
		apiKeyStore:     apiKeyStore,
		tieredLimiter:   tieredLimiter,
		enforceAPIKey:   enforceAPIKey,
		rpcHandler:      rpcHandler,
	}

	router.Use(s.LoggingMiddleware())

	if enforceAPIKey {
		router.POST("/", s.authAndRateLimitMiddleware(), s.handleRPCRequest)
	} else {
		router.POST("/", s.handleRPCRequest)
	}

	return s
}

func (s *server) Start() error {
	srv := &http.Server{
		Handler:      s.router,
		Addr:         fmt.Sprintf(":%s", s.port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	errChan := make(chan error, 1)

	go func() {
		s.logger.Info("Starting server on port", zap.String("port", s.port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case <-c:
		s.logger.Info("Shutting down the server...")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	case err := <-errChan:
		return err
	}
}

func (s *server) authAndRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			return
		}

		tier, exists := s.apiKeyStore.GetTierForKey(apiKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid API key"})
			return
		}

		if !s.tieredLimiter.CheckLimits(apiKey, tier) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		c.Set("apiKey", apiKey)
		c.Set("tier", tier)

		c.Next()
	}
}

func (s *server) handleRPCRequest(ctx *gin.Context) {
	// Read the raw request body
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		s.logger.Error("Failed to read request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   domain.NewRPCError(domain.InvalidRequest, "Failed to read request body"),
		})
		return
	}

	// Restore the request body for later use
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// First, try to unmarshal as a batch request
	var batchReq []rpc.JSONRPCRequest
	batchErr := json.Unmarshal(body, &batchReq)

	// Verify this is actually a batch request by checking if the first byte is '['
	isBatchRequest := len(body) > 0 && body[0] == '['

	s.logger.Debug("Request parsing",
		zap.Bool("isBatchRequest", isBatchRequest),
		zap.Error(batchErr),
		zap.Int("batchLength", len(batchReq)))

	if batchErr == nil && isBatchRequest {
		// Check if the batch is empty
		if len(batchReq) == 0 {
			s.logger.Debug("Empty batch request")
			ctx.JSON(http.StatusBadRequest, rpc.JSONRPCResponse{
				JSONRPC: "2.0",
				Error:   domain.NewRPCError(domain.InvalidRequest, "Empty batch request"),
			})
			return
		}

		s.logger.Debug("Processing batch request", zap.Int("count", len(batchReq)))

		// Handle batch request
		responses := make([]rpc.JSONRPCResponse, 0, len(batchReq))
		hasErrors := false

		for i, req := range batchReq {
			s.logger.Debug("Processing batch item",
				zap.Int("index", i),
				zap.String("method", req.Method))

			resp := s.rpcHandler.HandleRequest(ctx.Request.Context(), &req)
			responses = append(responses, *resp)
			if resp.Error != nil {
				hasErrors = true
			}
		}

		// If any response has an error, return 400, otherwise 200
		if hasErrors {
			ctx.JSON(http.StatusBadRequest, responses)
		} else {
			ctx.JSON(http.StatusOK, responses)
		}
		return
	}

	// If not a batch, try as a single request
	var req rpc.JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.logger.Debug("Invalid single request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   domain.NewRPCError(domain.InvalidRequest, "Invalid Request"),
		})
		return
	}

	s.logger.Debug("Processing single request", zap.String("method", req.Method))
	resp := s.rpcHandler.HandleRequest(ctx.Request.Context(), &req)

	if resp.Error != nil {
		ctx.JSON(http.StatusBadRequest, resp)
	} else {
		ctx.JSON(http.StatusOK, resp)
	}
}

func (s *server) LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			s.logger.Error("Error reading request body:", zap.Error(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Restore the request body so it can be read again later
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Log the request details
		s.logger.Info("Request", zap.String("method", c.Request.Method), zap.String("url", c.Request.URL.String()), zap.Any("headers", c.Request.Header), zap.String("body", string(body)))

		// Process the request
		c.Next()
	}
}

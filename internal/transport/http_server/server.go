package http_server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	router              *gin.Engine
	logger              *zap.Logger
	port                string
	serviceProvider     service.ServiceProvider
	apiKeyStore         *limiter.APIKeyStore
	tieredLimiter       *limiter.TieredLimiter
	enforceAPIKey       bool
	enableBatchRequests bool
	rpcHandler          rpc.RPCHandler
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
	enableBatchRequests bool,
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
		router:              router,
		logger:              logger,
		port:                port,
		serviceProvider:     serviceProvider,
		apiKeyStore:         apiKeyStore,
		tieredLimiter:       tieredLimiter,
		enforceAPIKey:       enforceAPIKey,
		enableBatchRequests: enableBatchRequests,
		rpcHandler:          rpcHandler,
	}

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

type batchResponse struct {
	index    int
	response rpc.JSONRPCResponse
}

func (s *server) handleBatchRequest(ctx *gin.Context, requests []rpc.JSONRPCRequest) {
	// Create a context with timeout for the entire batch
	batchCtx, cancel := context.WithTimeout(ctx.Request.Context(), 30*time.Second)
	defer cancel()

	// Create a worker pool with a reasonable size
	workerCount := 10
	if len(requests) < workerCount {
		workerCount = len(requests)
	}

	// Create channels for work distribution and results
	workChan := make(chan int, len(requests))
	resultsChan := make(chan batchResponse, len(requests))
	errorChan := make(chan error, 1)

	// Start workers
	for i := 0; i < workerCount; i++ {
		go func() {
			for index := range workChan {
				select {
				case <-batchCtx.Done():
					// Context was cancelled, stop processing
					return
				default:
					// Process the request
					req := requests[index]
					resp := s.rpcHandler.HandleRequest(batchCtx, &req)
					resultsChan <- batchResponse{
						index:    index,
						response: *resp,
					}
				}
			}
		}()
	}

	// Send work to workers
	go func() {
		defer close(workChan)
		for i := range requests {
			select {
			case <-batchCtx.Done():
				return
			case workChan <- i:
			}
		}
	}()

	// Collect results
	responses := make([]rpc.JSONRPCResponse, len(requests))
	completed := 0

	for completed < len(requests) {
		select {
		case <-batchCtx.Done():
			// Timeout or cancellation occurred
			ctx.JSON(http.StatusRequestTimeout, rpc.JSONRPCResponse{
				JSONRPC: "2.0",
				Error:   domain.NewRPCError(domain.ServerError, "Batch request timeout"),
			})
			return
		case result := <-resultsChan:
			responses[result.index] = result.response
			completed++
		case err := <-errorChan:
			// An error occurred in one of the workers
			ctx.JSON(http.StatusInternalServerError, rpc.JSONRPCResponse{
				JSONRPC: "2.0",
				Error:   domain.NewRPCError(domain.ServerError, err.Error()),
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, responses)
}

func (s *server) handleRPCRequest(ctx *gin.Context) {
	// Read the request body once
	body, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   domain.NewRPCError(domain.InvalidRequest, "Failed to read request body"),
		})
		return
	}

	// Try to parse as a batch request
	var batchReq []rpc.JSONRPCRequest
	if err := json.Unmarshal(body, &batchReq); err == nil {
		// It's a batch request
		if len(batchReq) > 1 && !s.enableBatchRequests {
			ctx.JSON(http.StatusBadRequest, rpc.JSONRPCResponse{
				JSONRPC: "2.0",
				Error:   domain.NewRPCError(domain.InvalidRequest, "Batch requests are disabled"),
			})
			return
		}

		// Handle single request in batch format
		if len(batchReq) == 1 {
			resp := s.rpcHandler.HandleRequest(ctx.Request.Context(), &batchReq[0])
			if resp.Error != nil {
				ctx.JSON(http.StatusBadRequest, resp)
			} else {
				ctx.JSON(http.StatusOK, resp)
			}
			return
		}

		// Handle multiple requests in parallel
		s.handleBatchRequest(ctx, batchReq)
		return
	}

	// Try to parse as a single request
	var singleReq rpc.JSONRPCRequest
	if err := json.Unmarshal(body, &singleReq); err != nil {
		ctx.JSON(http.StatusBadRequest, rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   domain.NewRPCError(domain.InvalidRequest, "Invalid Request"),
		})
		return
	}

	resp := s.rpcHandler.HandleRequest(ctx.Request.Context(), &singleReq)
	if resp.Error != nil {
		ctx.JSON(http.StatusBadRequest, resp)
	} else {
		ctx.JSON(http.StatusOK, resp)
	}
}

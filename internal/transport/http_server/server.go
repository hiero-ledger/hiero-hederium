package http_server

import (
	"context"
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
	var req rpc.JSONRPCRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, rpc.JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   domain.NewRPCError(domain.InvalidRequest, "Invalid Request"),
		})
		return
	}

	resp := s.rpcHandler.HandleRequest(ctx.Request.Context(), &req)

	if resp.Error != nil {
		ctx.JSON(http.StatusBadRequest, resp)
	} else {
		ctx.JSON(http.StatusOK, resp)
	}
}

package ws_server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	"github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/LimeChain/Hederium/internal/infrastructure/limiter"
	"github.com/LimeChain/Hederium/internal/service"
	"github.com/LimeChain/Hederium/internal/transport/rpc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/thanhpk/randstr"
	"go.uber.org/zap"
)

type WSServer interface {
	Start() error
}

type wsServer struct {
	router          *gin.Engine
	logger          *zap.Logger
	port            string
	serviceProvider service.ServiceProvider
	apiKeyStore     *limiter.APIKeyStore
	tieredLimiter   *limiter.TieredLimiter
	enforceAPIKey   bool
	rpcHandler      rpc.RPCHandler
	upgrader        websocket.Upgrader
	connectionCount int
	connectionMutex sync.Mutex
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
) WSServer {
	serviceProvider := service.NewServiceProvider(hClient, mClient, logger, applicationVersion, chainId, apiKeyStore, tieredLimiter, cacheService)

	router := gin.Default()

	rpcHandler := rpc.NewHandler(
		logger,
		serviceProvider,
	)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	s := &wsServer{
		router:          router,
		logger:          logger,
		port:            port,
		serviceProvider: serviceProvider,
		apiKeyStore:     apiKeyStore,
		tieredLimiter:   tieredLimiter,
		enforceAPIKey:   enforceAPIKey,
		rpcHandler:      rpcHandler,
		upgrader:        upgrader,
		connectionCount: 0,
		connectionMutex: sync.Mutex{},
	}

	if enforceAPIKey {
		router.GET("/", s.AuthAndRateLimitMiddleware(), s.handleWebSocket)
	} else {
		router.GET("/", s.handleWebSocket)
	}

	return s
}

func (s *wsServer) incrementConnectionCount() int {
	s.connectionMutex.Lock()
	defer s.connectionMutex.Unlock()
	s.connectionCount++
	return s.connectionCount
}

func (s *wsServer) decrementConnectionCount() int {
	s.connectionMutex.Lock()
	defer s.connectionMutex.Unlock()
	s.connectionCount--
	return s.connectionCount
}

func (s *wsServer) Start() error {
	srv := &http.Server{
		Handler:      s.router,
		Addr:         fmt.Sprintf(":%s", s.port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	errChan := make(chan error, 1)

	go func() {
		s.logger.Info("Starting WebSocket server on port", zap.String("port", s.port))
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

func (s *wsServer) AuthAndRateLimitMiddleware() gin.HandlerFunc {
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

func (s *wsServer) handleWebSocket(c *gin.Context) {
	requestID := uuid.New().String()
	ID := fmt.Sprintf("0x%s", randstr.Hex(32))

	c.Set("ID", ID)
	c.Set("requestID", requestID)

	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	currentConnections := s.incrementConnectionCount()
	s.logger.Info("New WebSocket connection established", zap.String("Connection ID", c.MustGet("ID").(string)), zap.String("Request ID", c.MustGet("requestID").(string)),
		zap.Int("active_connections", currentConnections))

	defer func() {
		conn.Close()
		remainingConnections := s.decrementConnectionCount()
		s.logger.Info("WebSocket connection closed", zap.Int("active_connections", remainingConnections))
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				s.logger.Info("Closing connection", zap.String("Connection ID", c.MustGet("ID").(string)), zap.String("code", err.Error()))
			} else if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("WebSocket connection closed unexpectedly", zap.String("Connection ID", c.MustGet("ID").(string)), zap.Error(err))
			}
			break
		}

		if messageType != websocket.TextMessage {
			s.logger.Warn("Received non-text message, ignoring")
			continue
		}

		var req rpc.JSONRPCRequest
		if err := json.Unmarshal(message, &req); err != nil {
			errResp := &rpc.JSONRPCResponse{
				JSONRPC: "2.0",
				Error:   domain.NewRPCError(domain.InvalidRequest, "Invalid Request"),
				ID:      nil,
			}
			s.sendResponse(conn, errResp)
			continue
		}

		resp := s.rpcHandler.HandleRequest(ctx, &req)

		s.sendResponse(conn, resp)
	}
}

func (s *wsServer) sendResponse(conn *websocket.Conn, resp *rpc.JSONRPCResponse) {
	respBytes, err := json.Marshal(resp)
	if err != nil {
		s.logger.Error("Failed to marshal response", zap.Error(err))
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
		s.logger.Error("Failed to write response", zap.Error(err))
	}
}

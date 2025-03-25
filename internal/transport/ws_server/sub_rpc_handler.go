package ws_server

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	"github.com/LimeChain/Hederium/internal/service"
	"github.com/LimeChain/Hederium/internal/transport/rpc"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type JSONRPCNotification struct {
	JSONRPC string             `json:"jsonrpc"`
	Method  string             `json:"method"`
	Params  SubscriptionParams `json:"params"`
}

type SubscriptionParams struct {
	Result       interface{} `json:"result"`
	Subscription string      `json:"subscription"`
}

type SubscriptionHandler struct {
	logger           *zap.Logger
	ethService       *service.EthService
	pollerService    service.PollerService
	subscribeService service.SubscribeServicer
	connections      map[*websocket.Conn]map[string]bool
	connectionMutex  sync.RWMutex
}

func NewSubscriptionHandler(logger *zap.Logger, ethService *service.EthService, cacheService cache.CacheService) *SubscriptionHandler {
	pollerService := service.NewPollerService(ethService, logger, service.DefaultPollingInterval)
	subscribeService := service.NewSubscribeService(pollerService, logger, cacheService)

	return &SubscriptionHandler{
		logger:           logger,
		ethService:       ethService,
		pollerService:    pollerService,
		subscribeService: subscribeService,
		connections:      make(map[*websocket.Conn]map[string]bool),
		connectionMutex:  sync.RWMutex{},
	}
}

var subscriptionMethods = map[string]func() domain.RPCParams{
	"eth_subscribe": func() domain.RPCParams {
		return &domain.EthSubscribeParams{}
	},
	"eth_unsubscribe": func() domain.RPCParams {
		return &domain.EthUnsubscribeParams{}
	},
}

func (h *SubscriptionHandler) HandleRequest(conn *websocket.Conn, req *rpc.JSONRPCRequest) *rpc.JSONRPCResponse {
	methodName := req.Method
	h.logger.Info("Subscription method called", zap.String("method", methodName))

	var result interface{}
	var rpcErr *domain.RPCError

	switch methodName {
	case "eth_subscribe":
		result, rpcErr = h.handleSubscribeMethod(conn, req)
	case "eth_unsubscribe":
		result, rpcErr = h.handleUnsubscribeMethod(conn, req)
	default:
		rpcErr = domain.NewRPCError(domain.MethodNotFound, fmt.Sprintf("Unsupported subscription method: %s", methodName))
	}

	resp := &rpc.JSONRPCResponse{JSONRPC: "2.0", ID: req.ID}
	if rpcErr != nil {
		resp.Error = rpcErr
	} else {
		resp.Result = result
	}
	return resp
}

func (h *SubscriptionHandler) dispatchSubscriptionMethod(req *rpc.JSONRPCRequest) (domain.RPCParams, *domain.RPCError) {
	methodInfo, ok := subscriptionMethods[req.Method]
	if !ok {
		return nil, domain.NewRPCError(domain.InvalidParams, "Invalid subscription method")
	}

	params := methodInfo()

	switch p := req.Params.(type) {
	case []interface{}:
		h.logger.Debug("Processing array params", zap.Any("array_params", p))
		if err := params.FromPositionalParams(p); err != nil {
			h.logger.Error("Failed to parse positional params", zap.Error(err))
			return nil, domain.NewRPCError(domain.InvalidParams, err.Error())
		}
	default:
		h.logger.Debug("Invalid params type", zap.String("type", fmt.Sprintf("%T", params)))
		return nil, domain.NewRPCError(domain.InvalidParams, "Invalid params: expected array or object")
	}

	return params, nil
}

func (h *SubscriptionHandler) handleSubscribeMethod(conn *websocket.Conn, req *rpc.JSONRPCRequest) (interface{}, *domain.RPCError) {
	params, rpcErr := h.dispatchSubscriptionMethod(req)
	if rpcErr != nil {
		return nil, rpcErr
	}

	subscribeParams, ok := params.(*domain.EthSubscribeParams)
	if !ok {
		return nil, domain.NewRPCError(domain.InvalidParams, "Invalid parameters for eth_subscribe")
	}

	subscriptionType := subscribeParams.SubscriptionType
	var subscribeOptions domain.SubscribeOptions
	if subscribeParams.SubscribeOptions != nil {
		subscribeOptions = *subscribeParams.SubscribeOptions
	}

	// Check if this connection already has a subscription of the same type
	h.connectionMutex.RLock()
	if connSubs, exists := h.connections[conn]; exists {
		for existingSubID := range connSubs {
			if h.subscribeService.HasSubscription(existingSubID) {
				existingTag, found := h.subscribeService.GetSubscriptionTag(existingSubID)
				if found {
					var tagData struct {
						Event string `json:"event"`
					}
					if err := json.Unmarshal([]byte(existingTag), &tagData); err == nil {
						if tagData.Event == subscriptionType {
							h.connectionMutex.RUnlock()
							h.logger.Info("Returning existing subscription for same type", zap.String("subscription", existingSubID), zap.String("type", subscriptionType))
							return existingSubID, nil
						} else {
							h.connectionMutex.RUnlock()
							h.logger.Warn("Rejecting subscription request for different type", zap.String("subscription", existingSubID), zap.String("type", subscriptionType))
							return nil, domain.NewRPCError(domain.InvalidParams, fmt.Sprintf("Connection already has a subscription of type '%s'. Only one subscription type per connection is allowed.", tagData.Event))
						}
					}
				}
			}
		}
	}
	h.connectionMutex.RUnlock()

	tag, err := h.createSubscriptionTag(subscriptionType, subscribeOptions)
	if err != nil {
		return nil, domain.NewRPCError(domain.InvalidParams, err.Error())
	}

	callback := func(subscriptionID string, result interface{}) {
		notification := &JSONRPCNotification{
			JSONRPC: "2.0",
			Method:  "eth_subscription",
			Params: SubscriptionParams{
				Subscription: subscriptionID,
				Result:       result,
			},
		}

		notificationBytes, err := json.Marshal(notification)
		if err != nil {
			h.logger.Error("Failed to marshal notification", zap.Error(err))
			return
		}

		h.connectionMutex.RLock()
		defer h.connectionMutex.RUnlock()

		if _, exists := h.connections[conn]; exists {
			if err := conn.WriteMessage(websocket.TextMessage, notificationBytes); err != nil {
				h.logger.Error("Failed to write notification", zap.Error(err))
			}
		}
	}

	subscriptionID, err := h.subscribeService.Subscribe(subscriptionType, subscribeOptions, callback)
	if err != nil {
		return nil, domain.NewRPCError(domain.InvalidParams, err.Error())
	}

	h.connectionMutex.Lock()
	if _, exists := h.connections[conn]; !exists {
		h.connections[conn] = make(map[string]bool)
	}
	h.connections[conn][subscriptionID] = true
	h.connectionMutex.Unlock()

	h.logger.Info("New subscription created", zap.String("subscription", subscriptionID), zap.String("tag", tag))

	return subscriptionID, nil
}

func (h *SubscriptionHandler) handleUnsubscribeMethod(conn *websocket.Conn, req *rpc.JSONRPCRequest) (interface{}, *domain.RPCError) {
	params, rpcErr := h.dispatchSubscriptionMethod(req)
	if rpcErr != nil {
		return nil, rpcErr
	}

	unsubscribeParams, ok := params.(*domain.EthUnsubscribeParams)
	if !ok {
		return nil, domain.NewRPCError(domain.InvalidParams, "Invalid parameters for eth_unsubscribe")
	}

	subscriptionID := unsubscribeParams.SubscriptionID

	h.connectionMutex.Lock()
	defer h.connectionMutex.Unlock()

	if connSubs, exists := h.connections[conn]; exists {
		if _, hasSub := connSubs[subscriptionID]; hasSub {
			success, err := h.subscribeService.Unsubscribe(subscriptionID)
			if err != nil {
				return nil, domain.NewRPCError(domain.InvalidParams, err.Error())
			}

			delete(connSubs, subscriptionID)

			h.logger.Info("Subscription removed", zap.String("subscription", subscriptionID))

			return success, nil
		}
	}

	return false, nil
}

func (h *SubscriptionHandler) createSubscriptionTag(eventType string, filterOptions domain.SubscribeOptions) (string, error) {
	tagData := struct {
		Event   string   `json:"event"`
		Address []string `json:"address,omitempty"`
		Topics  []string `json:"topics,omitempty"`
	}{
		Event:   eventType,
		Address: filterOptions.Address,
		Topics:  filterOptions.Topics,
	}

	tagBytes, err := json.Marshal(tagData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal subscription tag: %v", err)
	}

	return string(tagBytes), nil
}

func (h *SubscriptionHandler) CleanupConnection(conn *websocket.Conn) {
	h.connectionMutex.Lock()
	defer h.connectionMutex.Unlock()

	if connSubs, exists := h.connections[conn]; exists {
		h.logger.Info("Starting connection cleanup process", zap.Int("subscription_count", len(connSubs)))

		successCount := 0
		failureCount := 0

		for subID := range connSubs {
			h.logger.Info("Unsubscribing from subscription during connection cleanup", zap.String("subscription_id", subID))

			success, err := h.subscribeService.Unsubscribe(subID)
			if err != nil {
				failureCount++
				h.logger.Error("Failed to unsubscribe during connection cleanup", zap.String("subscription_id", subID), zap.Error(err))
			} else {
				successCount++
				h.logger.Info("Successfully unsubscribed during connection cleanup", zap.String("subscription_id", subID), zap.Bool("success", success))
			}
		}
		delete(h.connections, conn)
		h.logger.Info("Connection cleanup completed", zap.Int("subscriptions_removed", len(connSubs)), zap.Int("successful_unsubscribes", successCount),
			zap.Int("failed_unsubscribes", failureCount))
	}
}

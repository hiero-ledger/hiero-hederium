package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	"github.com/thanhpk/randstr"
	"go.uber.org/zap"
)

type SubscriptionCallback func(subscriptionID string, result interface{})

type SubscriptionData struct {
	ID       string
	Type     string
	Callback SubscriptionCallback
	Filters  *PollFilters
	Tag      string
}

type SubscribeServicer interface {
	Subscribe(subscriptionType string, subscribeOptions domain.SubscribeOptions, callback SubscriptionCallback) (string, error)
	Unsubscribe(subscriptionID string) (bool, error)
	HasSubscription(subscriptionID string) bool
	GetSubscriptionTag(subscriptionID string) (string, bool)
	NotifySubscribers(tag string, data interface{})
}

type subscribeService struct {
	poller             PollerService
	logger             *zap.Logger
	subscriptions      map[string]*SubscriptionData
	subMutex           sync.RWMutex
	tagToSubscriptions map[string]map[string]bool
	tagMutex           sync.RWMutex
	cacheService       cache.CacheService
}

func NewSubscribeService(poller PollerService, logger *zap.Logger, cacheService cache.CacheService) SubscribeServicer {
	return &subscribeService{
		poller:             poller,
		logger:             logger,
		subscriptions:      make(map[string]*SubscriptionData),
		subMutex:           sync.RWMutex{},
		tagToSubscriptions: make(map[string]map[string]bool),
		tagMutex:           sync.RWMutex{},
		cacheService:       cacheService,
	}
}

func (s *subscribeService) Subscribe(subscriptionType string, subscribeOptions domain.SubscribeOptions, callback SubscriptionCallback) (string, error) {
	if subscriptionType != EventLogs && subscriptionType != EventNewHeads {
		return "", fmt.Errorf("unsupported subscription type: %s", subscriptionType)
	}

	subscriptionID := fmt.Sprintf("0x%s", randstr.Hex(32))

	tag, err := CreateSubscriptionTag(subscriptionType, subscribeOptions)
	if err != nil {
		return "", err
	}

	subscription := &SubscriptionData{
		ID:       subscriptionID,
		Type:     subscriptionType,
		Callback: callback,
		Filters: &PollFilters{
			Address:             subscribeOptions.Address,
			Topics:              subscribeOptions.Topics,
			IncludeTransactions: subscribeOptions.IncludeTransactions,
		},
		Tag: tag,
	}

	s.tagMutex.Lock()
	isFirstSubscription := false
	if _, exists := s.tagToSubscriptions[tag]; !exists {
		s.tagToSubscriptions[tag] = make(map[string]bool)
		isFirstSubscription = true
	}
	s.tagToSubscriptions[tag][subscriptionID] = true
	s.tagMutex.Unlock()

	if isFirstSubscription {
		// Create a callback for the poller that will notify all subscribers for this tag
		pollCallback := func(result interface{}) {
			s.NotifySubscribers(tag, result)
		}

		if err := s.poller.AddPoll(tag, pollCallback, subscription.Filters); err != nil {
			s.tagMutex.Lock()
			delete(s.tagToSubscriptions[tag], subscriptionID)
			if len(s.tagToSubscriptions[tag]) == 0 {
				delete(s.tagToSubscriptions, tag)
			}
			s.tagMutex.Unlock()

			return "", err
		}
	} else {
		if err := s.poller.AddPoll(tag, nil, nil); err != nil {
			s.logger.Warn("Failed to increment subscriber count for existing poll", zap.String("tag", tag), zap.Error(err))
		}
	}

	// Store the subscription
	s.subMutex.Lock()
	s.subscriptions[subscriptionID] = subscription
	s.subMutex.Unlock()

	s.logger.Info("New subscription created",
		zap.String("id", subscriptionID),
		zap.String("type", subscriptionType),
		zap.String("tag", tag),
		zap.Int("tag_subscribers", len(s.tagToSubscriptions[tag])))

	return subscriptionID, nil
}

func (s *subscribeService) Unsubscribe(subscriptionID string) (bool, error) {
	s.subMutex.Lock()
	subscription, exists := s.subscriptions[subscriptionID]
	if !exists {
		s.subMutex.Unlock()
		s.logger.Warn("Subscription not found during unsubscribe", zap.String("subscription_id", subscriptionID))
		return false, fmt.Errorf("subscription not found: %s", subscriptionID)
	}

	tag := subscription.Tag
	s.logger.Info("Unsubscribing from subscription", zap.String("subscription_id", subscriptionID), zap.String("tag", tag))

	delete(s.subscriptions, subscriptionID)
	s.subMutex.Unlock()

	s.tagMutex.Lock()
	if subs, exists := s.tagToSubscriptions[tag]; exists {
		delete(subs, subscriptionID)
		remainingSubscriptions := len(subs)
		s.logger.Info("Removed subscription from tag mapping",
			zap.String("subscription_id", subscriptionID),
			zap.String("tag", tag),
			zap.Int("remaining_subscriptions", remainingSubscriptions))

		// If there are no more subscriptions for this tag, remove the tag and poll
		if remainingSubscriptions == 0 {
			s.logger.Info("No more subscriptions for tag, removing tag and poll", zap.String("tag", tag))
			delete(s.tagToSubscriptions, tag)
			s.poller.RemoveSubscriptionFromPoll(tag)
		} else {
			// If there are still subscriptions, just decrement the subscriber count
			s.poller.RemoveSubscriptionFromPoll(tag)
		}
	} else {
		s.logger.Warn("Tag not found in tag mapping during unsubscribe",
			zap.String("subscription_id", subscriptionID),
			zap.String("tag", tag))
	}
	s.tagMutex.Unlock()

	s.logger.Info("Subscription successfully unsubscribed",
		zap.String("subscription_id", subscriptionID))
	return true, nil
}

func (s *subscribeService) HasSubscription(subscriptionID string) bool {
	s.subMutex.RLock()
	defer s.subMutex.RUnlock()

	_, exists := s.subscriptions[subscriptionID]
	return exists
}

func (s *subscribeService) GetSubscriptionTag(subscriptionID string) (string, bool) {
	s.subMutex.RLock()
	defer s.subMutex.RUnlock()

	subscription, exists := s.subscriptions[subscriptionID]
	if !exists {
		return "", false
	}

	return subscription.Tag, true
}

func (s *subscribeService) NotifySubscribers(tag string, data interface{}) {
	s.tagMutex.RLock()
	subscriptionIDs, existsSubscriptions := s.tagToSubscriptions[tag]
	s.tagMutex.RUnlock()

	if !existsSubscriptions || len(subscriptionIDs) == 0 {
		return
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("Failed to marshal data", zap.Error(err))
		return
	}

	// For block notifications, extract the block hash or number for deduplication
	var blockIdentifier string
	if blockData, ok := data.(map[string]interface{}); ok {
		if hash, exists := blockData["hash"].(string); exists && hash != "" {
			blockIdentifier = hash
		} else if number, exists := blockData["number"].(string); exists && number != "" {
			blockIdentifier = number
		}
	}

	var cacheKey string
	if blockIdentifier != "" {
		cacheKey = fmt.Sprintf("block_notification:%s:%s", tag, blockIdentifier)
	} else {
		dataHash := createHash(string(dataBytes))
		cacheKey = fmt.Sprintf("notification:%s:%s", tag, dataHash)
	}

	var cached bool
	if err := s.cacheService.Get(context.Background(), cacheKey, &cached); err == nil && cached {
		s.logger.Debug("Skipping duplicate notification", zap.String("tag", tag), zap.String("cache_key", cacheKey))
		return
	}

	if err := s.cacheService.Set(context.Background(), cacheKey, true, ShortExpiration); err != nil {
		s.logger.Warn("Failed to cache notification", zap.Error(err))
	}

	s.subMutex.RLock()
	defer s.subMutex.RUnlock()

	for subID := range subscriptionIDs {
		subscription, exists := s.subscriptions[subID]
		if !exists {
			continue
		}
		s.logger.Debug("Sending notification to subscriber", zap.String("subscription", subID), zap.String("tag", tag))

		subscription.Callback(subID, data)
	}
}

func CreateSubscriptionTag(eventType string, subscribeOptions domain.SubscribeOptions) (string, error) {
	tagData := struct {
		Event               string   `json:"event"`
		Address             []string `json:"address,omitempty"`
		Topics              []string `json:"topics,omitempty"`
		IncludeTransactions bool     `json:"includeTransactions,omitempty"`
	}{
		Event:               eventType,
		Address:             subscribeOptions.Address,
		Topics:              subscribeOptions.Topics,
		IncludeTransactions: subscribeOptions.IncludeTransactions,
	}

	tagBytes, err := json.Marshal(tagData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal subscription tag: %v", err)
	}

	return string(tagBytes), nil
}

func createHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

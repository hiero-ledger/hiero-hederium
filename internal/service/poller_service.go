package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/LimeChain/Hederium/internal/domain"
	"go.uber.org/zap"
)

type PollCallback func(interface{})

type Poll struct {
	Tag             string
	Callback        PollCallback
	LastPolled      string
	SubscriberCount int
}

type PollFilters struct {
	IncludeTransactions bool     `json:"includeTransactions,omitempty"`
	Address             []string `json:"address,omitempty"`
	Topics              []string `json:"topics,omitempty"`
}

type PollerService interface {
	Start()
	Stop()
	AddPoll(tag string, callback PollCallback, filters *PollFilters) error
	RemoveSubscriptionFromPoll(tag string)
	IsPolling() bool
	HasPoll(tag string) bool
	GetPoll(tag string) *Poll
}

type pollerService struct {
	ethService      *EthService
	logger          *zap.Logger
	polls           []*Poll
	pollsMutex      sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	latestBlock     string
	newHeadsEnabled bool
	pollingEnabled  bool
	interval        time.Duration
}

func NewPollerService(ethService *EthService, logger *zap.Logger, interval int) PollerService {
	if interval <= 0 {
		interval = DefaultPollingInterval
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &pollerService{
		ethService:      ethService,
		logger:          logger,
		polls:           make([]*Poll, 0),
		ctx:             ctx,
		cancel:          cancel,
		interval:        time.Duration(interval) * time.Millisecond,
		newHeadsEnabled: true, // TODO: This should be set in config file
	}
}

func (p *pollerService) Start() {
	p.logger.Info(fmt.Sprintf("Poller: Starting polling with interval=%d", p.interval.Milliseconds()))
	p.pollingEnabled = true

	go func() {
		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		for {
			select {
			case <-p.ctx.Done():
				return
			case <-ticker.C:
				if err := p.updateLatestBlock(); err != nil {
					p.logger.Error("Failed to update latest block", zap.Error(err))
					continue
				}
				p.doPoll()
			}
		}
	}()
}

func (p *pollerService) Stop() {
	p.logger.Info("Stopping poller service")
	if p.IsPolling() {
		p.cancel()
		p.pollingEnabled = false
		p.logger.Info("Poller service stopped successfully")
	} else {
		p.logger.Warn("Attempted to stop poller service, but it was not running")
	}

	// Create a new context for future use
	ctx, cancel := context.WithCancel(context.Background())
	p.ctx = ctx
	p.cancel = cancel
}

func (p *pollerService) AddPoll(tag string, callback PollCallback, filters *PollFilters) error {
	p.pollsMutex.Lock()
	defer p.pollsMutex.Unlock()

	for _, poll := range p.polls {
		if poll.Tag == tag {
			poll.SubscriberCount++
			p.logger.Info("Added subscriber to existing poll", zap.String("tag", tag), zap.Int("total_subscribers", poll.SubscriberCount))
			return nil
		}
	}

	// Only add a new poll if we have a callback (first subscription)
	if callback != nil {
		p.logger.Info("Adding new poll to polling list", zap.String("tag", tag))
		p.polls = append(p.polls, &Poll{
			Tag:             tag,
			Callback:        callback,
			SubscriberCount: 1,
		})

		if !p.IsPolling() {
			p.Start()
		}
	} else {
		p.logger.Warn("Attempted to add poll without callback", zap.String("tag", tag))
		return fmt.Errorf("cannot add poll without callback")
	}
	return nil
}

func (p *pollerService) RemoveSubscriptionFromPoll(tag string) {
	p.pollsMutex.Lock()
	defer p.pollsMutex.Unlock()

	found := false
	for i, poll := range p.polls {
		if poll.Tag == tag {
			found = true
			poll.SubscriberCount--
			p.logger.Info("Removed subscriber from poll", zap.String("tag", tag), zap.Int("remaining_subscribers", poll.SubscriberCount))

			if poll.SubscriberCount <= 0 {
				p.logger.Info("Removing poll completely as no subscribers remain", zap.String("tag", tag))
				p.polls = append(p.polls[:i], p.polls[i+1:]...)
			}
			break
		}
	}

	if !found {
		p.logger.Warn("Attempted to remove non-existent poll", zap.String("tag", tag))
	}

	p.logger.Info("Poll removal status", zap.Int("remaining_polls", len(p.polls)))

	if len(p.polls) == 0 {
		p.logger.Info("No active polls, stopping poller service")
		p.Stop()
	}
}

func (p *pollerService) HasPoll(tag string) bool {
	p.pollsMutex.RLock()
	defer p.pollsMutex.RUnlock()

	for _, poll := range p.polls {
		if poll.Tag == tag {
			return true
		}
	}
	return false
}

func (p *pollerService) GetPoll(tag string) *Poll {
	p.pollsMutex.RLock()
	defer p.pollsMutex.RUnlock()

	for _, poll := range p.polls {
		if poll.Tag == tag {
			return poll
		}
	}
	return nil
}

func (p *pollerService) IsPolling() bool {
	return p.pollingEnabled
}

func (p *pollerService) updateLatestBlock() error {
	blockNumber, err := p.ethService.GetBlockNumber()
	if err != nil {
		return fmt.Errorf("failed to get block number: %v", err)
	}
	p.latestBlock = blockNumber.(string)
	return nil
}

func (p *pollerService) doPoll() {
	p.pollsMutex.RLock()
	defer p.pollsMutex.RUnlock()

	for _, poll := range p.polls {
		go func(poll *Poll) {
			p.logger.Debug(fmt.Sprintf("Poller: Fetching data for tag: %s", poll.Tag))

			var tagData struct {
				Event   string       `json:"event"`
				Filters *PollFilters `json:"filters,omitempty"`
			}

			if err := json.Unmarshal([]byte(poll.Tag), &tagData); err != nil {
				p.logger.Error("Failed to parse poll tag", zap.Error(err))
				return
			}

			// Skip if we've already processed this block for this poll
			if poll.LastPolled == p.latestBlock {
				return
			}

			var result interface{}
			var errRpc *domain.RPCError

			switch tagData.Event {
			case EventLogs:
				logParams := domain.LogParams{
					FromBlock: poll.LastPolled,
					ToBlock:   p.latestBlock,
				}

				if tagData.Filters != nil {
					logParams.Address = tagData.Filters.Address
					logParams.Topics = tagData.Filters.Topics
				}

				result, errRpc = p.ethService.GetLogs(logParams)
				poll.LastPolled = p.latestBlock

			case EventNewHeads:
				if p.newHeadsEnabled {
					includeTransactions := false
					if tagData.Filters != nil {
						includeTransactions = tagData.Filters.IncludeTransactions
					}
					result, errRpc = p.ethService.GetBlockByNumber(p.latestBlock, includeTransactions)
					poll.LastPolled = p.latestBlock
				} else {
					p.logger.Warn("NewHeads event is disabled")
					return
				}

			default:
				p.logger.Error("Unsupported event type", zap.String("event", tagData.Event))
				return
			}

			if errRpc != nil {
				p.logger.Error("Failed to fetch data", zap.String("event", tagData.Event), zap.Error(errRpc))
				return
			}

			if result != nil {
				if results, ok := result.([]interface{}); ok && len(results) > 0 {
					p.logger.Debug(fmt.Sprintf("Poller: Received %d results from tag: %s", len(results), poll.Tag))
					for _, item := range results {
						poll.Callback(item)
					}
				} else {
					poll.Callback(result)
				}
			}
		}(poll)
	}
}

package service

import (
	"context"
	"fmt"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	infrahedera "github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/thanhpk/randstr"
	"go.uber.org/zap"
)

type FilterServicer interface {
	NewFilter(fromBlock, toBlock string, address, topics []string) (*string, *domain.RPCError)
	NewBlockFilter() (*string, *domain.RPCError)
	UninstallFilter(filterID string) (interface{}, *domain.RPCError)
	NewPendingTransactionFilter() (interface{}, *domain.RPCError)
	GetFilterLogs(filterID string) ([]domain.Log, *domain.RPCError)
	GetFilterChanges(filterID string) (interface{}, *domain.RPCError)
}

type filterService struct {
	mirrorClient  infrahedera.MirrorNodeClient
	cacheService  cache.CacheService
	logger        *zap.Logger
	commonService CommonService
}

func NewFilterService(mirrorClient infrahedera.MirrorNodeClient, cacheService cache.CacheService, logger *zap.Logger, commonService CommonService) FilterServicer {
	return &filterService{
		mirrorClient:  mirrorClient,
		cacheService:  cacheService,
		logger:        logger,
		commonService: commonService,
	}
}

func (s *filterService) createFilter(filterType, fromBlock, toBlock, blockAtCreation string, address, topics []string) *string {
	ctx := context.Background()

	filterId := fmt.Sprintf("0x%s", randstr.Hex(32))

	filter := &domain.Filter{
		ID:              filterId,
		Type:            filterType,
		FromBlock:       fromBlock,
		ToBlock:         toBlock,
		Address:         address,
		Topics:          topics,
		BlockAtCreation: blockAtCreation,
		LastQueried:     "",
	}

	s.logger.Info("Saving:", zap.Any("filter", filter))

	cacheKey := fmt.Sprintf("filterId_%s", filterId)
	if err := s.cacheService.Set(ctx, cacheKey, filter, DefaultExpiration); err != nil {
		s.logger.Error("failed to set filter id to cache", zap.Error(err))
	}

	s.logger.Info("created filter with id and type", zap.String("id", filterId), zap.String("type", filterType))

	return &filterId
}

// TODO: Check it in config file
func (s *filterService) requireFilterEnabled() error {
	return nil
}

func (s *filterService) NewFilter(fromBlock, toBlock string, address, topics []string) (*string, *domain.RPCError) {
	s.logger.Info("creating new filter", zap.String("fromBlock", fromBlock), zap.String("toBlock", toBlock), zap.Any("address", address), zap.Strings("topics", topics))

	if err := s.requireFilterEnabled(); err != nil {
		return nil, domain.NewUnsupportedMethodError("eth_newFilter")
	}

	if err := s.commonService.ValidateBlockRange(fromBlock, toBlock); err != nil {
		return nil, domain.NewInvalidBlockRangeError()
	}

	if fromBlock == "latest" {
		fromBlockNum, errRpc := s.commonService.GetBlockNumberByNumberOrTag(fromBlock)
		if errRpc != nil {
			return nil, errRpc
		}

		fromBlock = fmt.Sprintf("0x%x", fromBlockNum)
	}

	filterId := s.createFilter("log", fromBlock, toBlock, "", address, topics)

	return filterId, nil
}

func (s *filterService) NewBlockFilter() (*string, *domain.RPCError) {
	if err := s.requireFilterEnabled(); err != nil {
		return nil, domain.NewUnsupportedMethodError("eth_newFilter")
	}

	blockAtCreation, errRpc := s.commonService.GetBlockNumberByNumberOrTag("latest")
	if errRpc != nil {
		return nil, errRpc
	}

	filterId := s.createFilter("new_block", "", "", fmt.Sprintf("0x%x", blockAtCreation), nil, nil)

	return filterId, nil
}

func (s *filterService) UninstallFilter(filterID string) (interface{}, *domain.RPCError) {
	ctx := context.Background()

	if err := s.requireFilterEnabled(); err != nil {
		return false, domain.NewUnsupportedMethodError("eth_newFilter")
	}

	cacheKey := fmt.Sprintf("filterId_%s", filterID)

	var filter domain.Filter
	if err := s.cacheService.Get(ctx, cacheKey, &filter); err != nil {
		return false, domain.NewFilterNotFoundError()
	}

	if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
		s.logger.Error("failed to delete filter id from cache", zap.Error(err))
		return false, domain.NewInternalError("failed to delete filter id from cache")
	}

	return true, nil
}

func (s *filterService) NewPendingTransactionFilter() (interface{}, *domain.RPCError) {
	s.logger.Info("creating new pending transaction filter")
	return nil, domain.NewUnsupportedJSONRPCMethodError()
}

func (s *filterService) GetFilterLogs(filterID string) ([]domain.Log, *domain.RPCError) {
	s.logger.Info("getting filter logs", zap.String("filterID", filterID))
	ctx := context.Background()

	cacheKey := fmt.Sprintf("filterId_%s", filterID)
	var filter domain.Filter
	if err := s.cacheService.Get(ctx, cacheKey, &filter); err != nil {
		return nil, domain.NewFilterNotFoundError()
	}

	if filter.Type != "log" {
		return nil, domain.NewFilterNotFoundError()
	}

	s.logger.Info("getting logs for filter", zap.String("filterID", filterID), zap.Any("filter", filter))

	logParams := domain.LogParams{
		FromBlock: filter.FromBlock,
		ToBlock:   filter.ToBlock,
		Address:   filter.Address,
		Topics:    filter.Topics,
	}

	logs, errRpc := s.commonService.GetLogs(logParams)
	if errRpc != nil {
		return nil, errRpc
	}

	if err := s.cacheService.Set(ctx, cacheKey, filter, DefaultExpiration); err != nil {
		s.logger.Error("failed to set filter id to cache", zap.Error(err))
	}

	return logs, nil
}

func (s *filterService) GetFilterChanges(filterID string) (interface{}, *domain.RPCError) {
	s.logger.Info("getting filter changes", zap.String("filterID", filterID))

	if err := s.requireFilterEnabled(); err != nil {
		return nil, domain.NewUnsupportedMethodError("eth_getFiltetChanges")
	}

	ctx := context.Background()

	cacheKey := fmt.Sprintf("filterId_%s", filterID)
	var filter domain.Filter
	if err := s.cacheService.Get(ctx, cacheKey, &filter); err != nil {
		return nil, domain.NewFilterNotFoundError()
	}

	var blockResult []string
	var result interface{}

	if filter.Type == "log" {
		logParams := domain.LogParams{
			FromBlock: filter.FromBlock,
			ToBlock:   filter.ToBlock,
			Address:   filter.Address,
			Topics:    filter.Topics,
		}

		logResult, errRpc := s.commonService.GetLogs(logParams)
		if errRpc != nil {
			return nil, errRpc
		}

		var latestBlock int64
		var err error
		if len(logResult) > 0 {
			latestBlock, err = HexToDec(logResult[len(logResult)-1].BlockNumber)
			if err != nil {
				s.logger.Error("failed to convert block number to int64", zap.Error(err))
				return nil, domain.NewInternalError("unexpected error")
			}
		} else {
			latestBlock, errRpc = s.commonService.GetBlockNumberByNumberOrTag("latest")
			if errRpc != nil {
				return nil, errRpc
			}
		}
		latestBlock++
		filter.LastQueried = fmt.Sprintf("0x%x", latestBlock)

		result = logResult
	} else if filter.Type == "new_block" {

		var blockNum string
		if filter.LastQueried != "" {
			blockNum = filter.LastQueried
		} else if filter.BlockAtCreation != "" {
			blockNum = filter.BlockAtCreation
		}

		blocks, err := s.mirrorClient.GetBlocks(blockNum)
		if err != nil {
			s.logger.Error("failed to get blocks from mirror node", zap.Error(err))
			return nil, domain.NewInternalError("unexpected error")
		}

		var latestBlock int64
		var errRpc *domain.RPCError
		if len(blocks) > 0 {

			if blockNumFloat, ok := blocks[len(blocks)-1]["number"].(float64); ok {
				latestBlock = int64(blockNumFloat)
			} else {
				s.logger.Error("failed to convert block number to int64")
				return nil, domain.NewInternalError("unexpected error")
			}
		} else {
			latestBlock, errRpc = s.commonService.GetBlockNumberByNumberOrTag("latest")
			if errRpc != nil {
				return nil, errRpc
			}
		}
		filter.LastQueried = fmt.Sprintf("0x%x", latestBlock)
		for _, b := range blocks {
			blockResult = append(blockResult, b["hash"].(string))
		}

		result = blockResult

	} else {
		return nil, domain.NewUnsupportedMethodError("eth_getFilterChanges")
	}

	if err := s.cacheService.Set(ctx, cacheKey, filter, DefaultExpiration); err != nil {
		s.logger.Error("failed to set filter id to cache", zap.Error(err))
	}

	return result, nil
}

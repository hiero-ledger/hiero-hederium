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

type FilterService interface {
	NewFilter(fromBlock, toBlock string, address, topics []string) (*string, *domain.RPCError)
}

type filterService struct {
	mirrorClient  infrahedera.MirrorNodeClient
	cacheService  cache.CacheService
	logger        *zap.Logger
	commonService CommonService
}

func NewFilterService(mirrorClient infrahedera.MirrorNodeClient, cacheService cache.CacheService, logger *zap.Logger, commonService CommonService) FilterService {
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
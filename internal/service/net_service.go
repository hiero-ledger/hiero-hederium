package service

import (
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type NetServicer interface {
	Listening() string
	Version() string
}

type netService struct {
	log     *zap.Logger
	chainId string
}

func NewNetService(log *zap.Logger, chainId string) NetServicer {
	return &netService{
		log:     log,
		chainId: chainId,
	}
}

// Listening returns false because the Hedera network does not support listening.
func (n *netService) Listening() string {
	return "false"
}

// Version returns the chain ID.
func (n *netService) Version() string {
	chainId, err := strconv.ParseInt(strings.TrimPrefix(n.chainId, "0x"), 16, 64)
	if err != nil {
		n.log.Error("Failed to convert chain ID to number", zap.Error(err))
		return "0"
	}
	return fmt.Sprintf("%d", chainId)
}

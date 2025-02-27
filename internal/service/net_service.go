package service

import "go.uber.org/zap"

type NetServicer interface {
	Listening() bool
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
func (n *netService) Listening() bool {
	return false
}

// Version returns the chain ID.
func (n *netService) Version() string {
	return n.chainId
}

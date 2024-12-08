package service

import "go.uber.org/zap"

type Net interface {
	Listening() bool
	Version() string
}

type NetService struct {
	log     *zap.Logger
	chainId string
}

func NewNetService(log *zap.Logger, chainId string) *NetService {
	return &NetService{
		log:     log,
		chainId: chainId,
	}
}

// Listening returns false because the Hedera network does not support listening.
func (n *NetService) Listening() bool {
	return false
}

// Version returns the chain ID.
func (n *NetService) Version() string {
	return n.chainId
}

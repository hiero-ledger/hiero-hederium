package service

import (
	"math/big"
	"strconv"

	infrahedera "github.com/georgi-l95/Hederium/internal/infrastructure/hedera"
	"github.com/georgi-l95/Hederium/internal/infrastructure/limiter"
	"go.uber.org/zap"
)

type EthService struct {
	hClient       infrahedera.HederaNodeClient
	mClient       infrahedera.MirrorNodeClient
	logger        *zap.Logger
	tieredLimiter *limiter.TieredLimiter
	chainId       string
}

func NewEthService(
	hClient infrahedera.HederaNodeClient,
	mClient infrahedera.MirrorNodeClient,
	log *zap.Logger,
	l *limiter.TieredLimiter,
	chainId string,
) *EthService {
	return &EthService{
		hClient:       hClient,
		mClient:       mClient,
		logger:        log,
		tieredLimiter: l,
		chainId:       chainId,
	}
}

// GetBlockNumber retrieves the latest block number from the Hedera network and returns it
// in hexadecimal format, compatible with Ethereum JSON-RPC specifications.
// It returns two values:
//   - interface{}: A hex string representing the block number (e.g., "0x1234") on success,
//     or nil on failure
//   - map[string]interface{}: Error details if the operation fails, nil on success.
//     Error format follows Ethereum JSON-RPC error specifications.
func (s *EthService) GetBlockNumber() (interface{}, map[string]interface{}) {
	s.logger.Info("Getting block number")
	block, err := s.mClient.GetLatestBlock()
	if err != nil {
		s.logger.Error("Failed to fetch latest block", zap.Error(err))
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to fetch block data",
		}
	}

	s.logger.Debug("Received block data", zap.Any("block", block))

	if blockNumber, ok := block["number"].(float64); ok {
		s.logger.Debug("Found block number", zap.Float64("blockNumber", blockNumber))

		blockNum := uint64(blockNumber)
		hexBlockNum := "0x" + strconv.FormatUint(blockNum, 16)
		s.logger.Debug("Successfully converted to hex", zap.String("hexBlockNum", hexBlockNum))
		s.logger.Info("Successfully returned block number", zap.String("blockNumber", hexBlockNum))
		return hexBlockNum, nil
	}

	s.logger.Error("Block number not found or invalid type", zap.Any("block", block))
	return nil, map[string]interface{}{
		"code":    -32000,
		"message": "Invalid block data",
	}
}

// GetGasPrice retrieves the gas price from the Hedera network and returns it
// in hexadecimal format, compatible with Ethereum JSON-RPC specifications.
func (s *EthService) GetGasPrice() (interface{}, map[string]interface{}) {
	s.logger.Info("Getting gas price")
	weibars, err := getFeeWeibars(s)
	if err != nil {
		errMsg := "Failed to fetch gas price"
		s.logger.Error(errMsg)
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": errMsg,
		}
	}
	// Add 10% buffer to the gas price
	buffer := new(big.Int).Div(weibars, big.NewInt(10))
	weibars.Add(weibars, buffer)
	gasPrice := "0x" + strconv.FormatUint(weibars.Uint64(), 16)

	s.logger.Info("Successfully returned gas price", zap.String("gasPrice", gasPrice))
	return gasPrice, nil
}

func (s *EthService) GetChainId() (interface{}, map[string]interface{}) {
	s.logger.Info("Getting chain ID")
	s.logger.Info("Returning chain ID", zap.String("chainId", s.chainId))
	return s.chainId, nil
}

// Methods that return false values, because the Hedera network does not support them

// GetAccounts returns an empty array of accounts, similar to Infura's implementation
func (s *EthService) GetAccounts() (interface{}, map[string]interface{}) {
	s.logger.Info("Getting accounts")
	s.logger.Debug("Returning empty accounts array as per specification")
	return []string{}, nil
}

// Syncing returns false, because the Hedera network does not support syncing
func (s *EthService) Syncing() (interface{}, map[string]interface{}) {
	s.logger.Info("Syncing")
	s.logger.Debug("Returning false as per specification")
	return false, nil
}

// Mining returns false, because the Hedera network does not support mining
func (s *EthService) Mining() (interface{}, map[string]interface{}) {
	s.logger.Info("Mining")
	s.logger.Debug("Returning false as per specification")
	return false, nil
}

// MaxPriorityFeePerGas returns 0x0, because the Hedera network does not support it
func (s *EthService) MaxPriorityFeePerGas() (interface{}, map[string]interface{}) {
	s.logger.Info("MaxPriorityFeePerGas")
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// Hashrate returns 0x0, because the Hedera network does not support it
func (s *EthService) Hashrate() (interface{}, map[string]interface{}) {
	s.logger.Info("Hashrate")
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// GetUncleCountByBlockNumber returns 0x0, because the Hedera network does not support it
func (s *EthService) GetUncleCountByBlockNumber() (interface{}, map[string]interface{}) {
	s.logger.Info("GetUncleCountByBlockNumber")
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// GetUncleByBlockNumberAndIndex returns nil, because the Hedera network does not support it
func (s *EthService) GetUncleByBlockNumberAndIndex() (interface{}, map[string]interface{}) {
	s.logger.Info("GetUncleByBlockNumberAndIndex")
	s.logger.Debug("Returning nil as per specification")
	return nil, nil
}

// GetUncleCountByBlockHash returns 0x0, because the Hedera network does not support it
func (s *EthService) GetUncleCountByBlockHash() (interface{}, map[string]interface{}) {
	s.logger.Info("GetUncleCountByBlockHash")
	s.logger.Debug("Returning 0x0 as per specification")
	return "0x0", nil
}

// GetUncleByBlockHashAndIndex returns nil, because the Hedera network does not support it
func (s *EthService) GetUncleByBlockHashAndIndex() (interface{}, map[string]interface{}) {
	s.logger.Info("GetUncleByBlockHashAndIndex")
	s.logger.Debug("Returning nil as per specification")
	return nil, nil
}

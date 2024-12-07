package service

import (
	"strconv"

	infrahedera "github.com/georgi-l95/Hederium/internal/infrastructure/hedera"
	"github.com/georgi-l95/Hederium/internal/infrastructure/limiter"
	sdkhedera "github.com/hashgraph/hedera-sdk-go/v2"
	"go.uber.org/zap"
)

type EthService struct {
	hClient       *sdkhedera.Client
	mClient       infrahedera.MirrorNodeClient // use the interface here
	logger        *zap.Logger
	tieredLimiter *limiter.TieredLimiter
}

func NewEthService(
	hClient *sdkhedera.Client,
	mClient infrahedera.MirrorNodeClient, // also change the constructor to accept the interface
	log *zap.Logger,
	l *limiter.TieredLimiter,
) *EthService {
	return &EthService{
		hClient:       hClient,
		mClient:       mClient,
		logger:        log,
		tieredLimiter: l,
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
		return hexBlockNum, nil
	}

	s.logger.Error("Block number not found or invalid type", zap.Any("block", block))
	return nil, map[string]interface{}{
		"code":    -32000,
		"message": "Invalid block data",
	}
}

// func (s *EthService) SendRawTransaction(c *gin.Context, rawTx []byte) (interface{}, map[string]interface{}) {
// 	apiKeyVal, _ := c.Get("apiKey")
// 	tierVal, _ := c.Get("tier")
// 	apiKey := apiKeyVal.(string)
// 	tier := tierVal.(string)

// 	hbarCost := 1 // Example cost estimation

// 	if !s.tieredLimiter.DeductHbarUsage(apiKey, tier, hbarCost) {
// 		return nil, map[string]interface{}{
// 			"code":    -32000,
// 			"message": "Hbar budget exceeded",
// 		}
// 	}

// 	// Integrate with Hedera SDK to submit transaction (placeholder)
// 	// Example:
// 	// tx, err := hedera.TransactionFromBytes(rawTx)
// 	// if err != nil {
// 	//   return nil, map[string]interface{}{"code": -32602, "message": "Invalid raw transaction"}
// 	// }
// 	// receipt, err := tx.Execute(s.hClient)
// 	// if err != nil {
// 	//   return nil, map[string]interface{}{"code": -32000, "message": "Transaction execution failed"}
// 	// }
// 	// hash := receipt.TransactionID.String()

// 	return "0x123abc", nil
// }

// GetAccounts returns an empty array of accounts, similar to Infura's implementation
func (s *EthService) GetAccounts() (interface{}, map[string]interface{}) {
	s.logger.Info("Getting accounts")
	s.logger.Debug("Returning empty accounts array as per specification")
	return []string{}, nil
}

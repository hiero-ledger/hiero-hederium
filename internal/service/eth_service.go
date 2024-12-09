package service

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/georgi-l95/Hederium/internal/domain"
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

func (s *EthService) GetBlockByHash(hash string, showDetails bool) (interface{}, map[string]interface{}) {
	s.logger.Info("Getting block by hash", zap.String("hash", hash), zap.Bool("showDetails", showDetails))
	block := s.mClient.GetBlockByHashOrNumber(hash)
	if block == nil {
		return nil, nil
	}

	// Create a new Block instance with default values
	ethBlock := domain.NewBlock()

	hexNumber := "0x" + strconv.FormatUint(uint64(block.Number), 16)
	hexGasUsed := "0x" + strconv.FormatUint(uint64(block.GasUsed), 16)
	hexSize := "0x" + strconv.FormatUint(uint64(block.Size), 16)
	timestampStr := strings.Split(block.Timestamp.From, ".")[0]
	timestampInt, _ := strconv.ParseUint(timestampStr, 10, 64)
	hexTimestamp := "0x" + strconv.FormatUint(timestampInt, 16)
	trimmedHash := block.Hash
	if len(trimmedHash) > 66 {
		trimmedHash = trimmedHash[:66]
	}
	trimmedParentHash := block.PreviousHash
	if len(trimmedParentHash) > 66 {
		trimmedParentHash = trimmedParentHash[:66]
	}

	ethBlock.Number = &hexNumber
	ethBlock.GasUsed = hexGasUsed
	ethBlock.GasLimit = "0x" + strconv.FormatUint(15000000, 16) // Hedera's default gas limit
	ethBlock.Hash = &trimmedHash
	ethBlock.LogsBloom = block.LogsBloom
	ethBlock.TransactionsRoot = &trimmedHash
	ethBlock.ParentHash = trimmedParentHash
	ethBlock.Timestamp = hexTimestamp
	ethBlock.Size = hexSize

	// Handle transactions based on showDetails parameter
	// if txs, ok := block["transactions"].([]interface{}); ok {
	// 	if showDetails {
	// 		// Convert each transaction to full Transaction object
	// 		for _, tx := range txs {
	// 			if txMap, ok := tx.(map[string]interface{}); ok {
	// 				transaction := &domain.Transaction{}
	// 				// Fill transaction details here based on txMap
	// 				if txHash, ok := txMap["hash"].(string); ok {
	// 					transaction.Hash = txHash
	// 				}
	// 				// Add more transaction field mappings as needed
	// 				ethBlock.Transactions = append(ethBlock.Transactions, transaction)
	// 			}
	// 		}
	// 	} else {
	// 		// Only include transaction hashes
	// 		for _, tx := range txs {
	// 			if txMap, ok := tx.(map[string]interface{}); ok {
	// 				if txHash, ok := txMap["hash"].(string); ok {
	// 					ethBlock.Transactions = append(ethBlock.Transactions, txHash)
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	s.logger.Debug("Returning block data", zap.Any("block", ethBlock))
	return ethBlock, nil
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

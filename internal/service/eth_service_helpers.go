package service

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/georgi-l95/Hederium/internal/domain"
	"go.uber.org/zap"
)

func GetFeeWeibars(s *EthService) (*big.Int, map[string]interface{}) {
	gasTinybars, err := s.mClient.GetNetworkFees()
	if err != nil {
		return nil, map[string]interface{}{
			"code":    -32000,
			"message": "Failed to fetch gas price",
		}
	}

	// Convert tinybars to weibars
	weibars := big.NewInt(gasTinybars).
		Mul(big.NewInt(gasTinybars), big.NewInt(100000000)) // 10^8 conversion factor

	return weibars, nil
}

func ProcessBlock(s *EthService, block *domain.BlockResponse, showDetails bool) (interface{}, map[string]interface{}) {
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

	contractResults := s.mClient.GetContractResults(block.Timestamp)
	for _, contractResult := range contractResults {
		if contractResult.Result == "WRONG_NONCE" || contractResult.Result == "INVALID_ACCOUNT_ID" {
			continue
		}

		// TODO: Resolve evm addresses
		if showDetails {
			// TODO: Handle detailed transaction view
		} else {
			ethBlock.Transactions = append(ethBlock.Transactions, contractResult.Hash)
		}
	}

	s.logger.Debug("Returning block data", zap.Any("block", ethBlock))
	s.logger.Info("Successfully returned block data block: %s with %d transactions",
		zap.String("blockHash", *ethBlock.Hash),
		zap.Int("txCount", len(ethBlock.Transactions)))
	return ethBlock, nil
}

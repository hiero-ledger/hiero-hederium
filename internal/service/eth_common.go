package service

import (
	"fmt"
	"strconv"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	infrahedera "github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"go.uber.org/zap"
)

type CommonService interface {
	GetLogs(logParams domain.LogParams) ([]domain.Log, *domain.RPCError)
	ValidateBlockHashAndAddTimestampToParams(params map[string]interface{}, blockHash string) error
	ValidateBlockRangeAndAddTimestampToParams(params map[string]interface{}, fromBlock, toBlock string, address []string) (bool, *domain.RPCError)
	GetLogsWithParams(address []string, params map[string]interface{}) ([]domain.Log, error)
	GetBlockNumberByNumberOrTag(blockNumberOrTag string) (int64, *domain.RPCError)
	ValidateBlockRange(fromBlock, toBlock string) *domain.RPCError
	GetBlockNumber() (interface{}, *domain.RPCError)
}

type commonService struct {
	mClient infrahedera.MirrorNodeClient
	logger  *zap.Logger
	cache   cache.CacheService
}

func NewCommonService(mClient infrahedera.MirrorNodeClient, logger *zap.Logger, cache cache.CacheService) CommonService {
	return &commonService{
		mClient: mClient,
		logger:  logger,
		cache:   cache,
	}
}

func (s *commonService) GetLogs(logParams domain.LogParams) ([]domain.Log, *domain.RPCError) {
	params := make(map[string]interface{})

	if logParams.BlockHash != "" {
		if err := s.ValidateBlockHashAndAddTimestampToParams(params, logParams.BlockHash); err != nil {
			return []domain.Log{}, nil
		}
	} else {
		if ok, errRpc := s.ValidateBlockRangeAndAddTimestampToParams(params, logParams.FromBlock, logParams.ToBlock, logParams.Address); errRpc != nil {
			return nil, errRpc
		} else if !ok {
			return []domain.Log{}, nil
		}
	}

	if logParams.Topics != nil {
		for i, topic := range logParams.Topics {
			if topic != "" {
				params[fmt.Sprintf("topic%d", i)] = topic
			}
		}
	}

	logs, err := s.GetLogsWithParams(logParams.Address, params)
	if err != nil {
		s.logger.Error("Failed to get logs", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to get logs")
	}

	return logs, nil
}

func (s *commonService) ValidateBlockHashAndAddTimestampToParams(params map[string]interface{}, blockHash string) error {
	block := s.mClient.GetBlockByHashOrNumber(blockHash)
	if block == nil {
		s.logger.Debug("Failed to get block data")
		return fmt.Errorf("block not found")
	}
	s.logger.Debug("Received block data", zap.Any("block", block))

	params["timestamp"] = fmt.Sprintf("gte:%s&timestamp=lte:%s", block.Timestamp.From, block.Timestamp.To)

	s.logger.Debug("Returning timestamp", zap.Any("timestamp", params["timestamp"]))

	return nil
}

func (s *commonService) ValidateBlockRangeAndAddTimestampToParams(params map[string]interface{}, fromBlock, toBlock string, address []string) (bool, *domain.RPCError) {

	// We get the latestBlockNum only once to avoid multiple calls
	latestBlockNum, errRpc := s.GetBlockNumberByNumberOrTag("latest")
	if errRpc != nil {
		return false, errRpc
	}

	var toBlockNum int64

	if blockTagIsLatestOrPending(&toBlock) {
		toBlock = "latest"
		toBlockNum = latestBlockNum
	} else {
		toBlockNum, errRpc = s.GetBlockNumberByNumberOrTag(toBlock)
		if errRpc != nil {
			return false, errRpc
		}

		// - When `fromBlock` is not explicitly provided, it defaults to `latest`.
		// - Then if `toBlock` equals `latestBlockNumber`, it means both `toBlock` and `fromBlock` essentially refer to the latest block, so the `MISSING_FROM_BLOCK_PARAM` error is not necessary.
		// - If `toBlock` is explicitly provided and does not equals to `latestBlockNumber`, it establishes a solid upper bound.
		// - If `fromBlock` is missing, indicating the absence of a lower bound, throw the `MISSING_FROM_BLOCK_PARAM` error.
		if toBlockNum != latestBlockNum && fromBlock == "" {
			return false, domain.NewRPCError(domain.MissingFromBlockParam, "Provided toBlock parameter without specifying fromBlock")
		}
	}

	var fromBlockNum int64

	if blockTagIsLatestOrPending(&fromBlock) {
		fromBlock = "latest"
		fromBlockNum = latestBlockNum
	} else {
		fromBlockNum, errRpc = s.GetBlockNumberByNumberOrTag(fromBlock)
		if errRpc != nil {
			return false, errRpc
		}
	}

	fromBlockResponse := s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(fromBlockNum, 10))
	if fromBlockResponse == nil {
		s.logger.Debug("Failed to get from block data")
		return false, nil
	}

	var timestamp string

	timestamp = fmt.Sprintf("gte:%s", fromBlockResponse.Timestamp.From)

	if fromBlock == toBlock {
		timestamp += fmt.Sprintf("&timestamp=lte:%s", fromBlockResponse.Timestamp.To)

	} else {
		fromBlockNum := fromBlockResponse.Number
		toBlockResponse := s.mClient.GetBlockByHashOrNumber(strconv.FormatInt(toBlockNum, 10))

		/**
		 * If `toBlock` is not provided, the `lte` field cannot be set,
		 * resulting in a request to the Mirror Node that includes only the `gte` parameter.
		 * Such requests will be rejected, hence causing the whole request to fail.
		 * Return false to handle this gracefully and return an empty response to end client.
		 */
		if toBlockResponse == nil {
			s.logger.Debug("failed to get to block data")
			return false, nil
		}

		timestamp = fmt.Sprintf("%s&timestamp=lte:%s", timestamp, toBlockResponse.Timestamp.To)
		toBlockNum := toBlockResponse.Number

		toBlockTo, err := strconv.ParseFloat(toBlockResponse.Timestamp.To, 64)
		if err != nil {
			return false, domain.NewRPCError(domain.InvalidParams, "Invalid timestamp")
		}

		fromBlockFrom, err := strconv.ParseFloat(fromBlockResponse.Timestamp.From, 64)
		if err != nil {
			return false, domain.NewRPCError(domain.InvalidParams, "Invalid timestamp")
		}

		// Validate timestamp range for Mirror Node requests (maximum: 7 days or 604,800 seconds) to prevent exceeding the limit,
		// as requests with timestamp parameters beyond 7 days are rejected by the Mirror Node.
		timestampDiff := toBlockTo - fromBlockFrom
		if timestampDiff > 604800 {
			s.logger.Debug("Timestamp range is too large")
			return false, domain.NewTimeStampRangeTooLargeError(fmt.Sprintf("0x%x", fromBlockNum), fmt.Sprintf("0x%x", toBlockNum), toBlockTo, fromBlockFrom)
		}

		if fromBlockNum > toBlockNum {
			return false, domain.NewInvalidBlockRangeError()
		}

		// Increasing it to more then one address may degrade mirror node performance
		// when addresses contains many log events.
		isSingleAddress := len(address) == 1
		if !isSingleAddress && toBlockNum-fromBlockNum > blockRangeLimit {
			return false, domain.NewRangeTooLarge(blockRangeLimit)
		}
	}

	s.logger.Debug("Returning timestamp", zap.String("timestamp", timestamp))
	params["timestamp"] = timestamp

	return true, nil
}

func (s *commonService) GetLogsWithParams(address []string, params map[string]interface{}) ([]domain.Log, error) {
	addresses := address

	var logs []domain.Log

	if address == nil {
		logResults, err := s.mClient.GetContractResultsLogsWithRetry(params)
		if err != nil {
			s.logger.Error("Failed to get logs", zap.Error(err))
			return nil, err
		}

		s.logger.Debug("Received logs", zap.Any("logs", logResults))

		for _, logResult := range logResults {
			if len(logResult.BlockHash) > 66 {
				logResult.BlockHash = logResult.BlockHash[:66]
			}
			if len(logResult.TransactionHash) > 66 {
				logResult.TransactionHash = logResult.TransactionHash[:66]
			}

			logs = append(logs, domain.Log{
				Address:          logResult.Address,
				BlockHash:        logResult.BlockHash,
				BlockNumber:      fmt.Sprintf("0x%x", *logResult.BlockNumber),
				Data:             logResult.Data,
				LogIndex:         fmt.Sprintf("0x%x", *logResult.Index),
				Removed:          false,
				Topics:           logResult.Topics,
				TransactionHash:  logResult.TransactionHash,
				TransactionIndex: fmt.Sprintf("0x%x", *logResult.TransactionIndex),
			})
		}
	}

	for _, addr := range addresses {
		logResults, err := s.mClient.GetContractResultsLogsByAddress(addr, params)
		if err != nil {
			s.logger.Error("Failed to get logs", zap.Error(err))
			return nil, err
		}
		for _, logResult := range logResults {
			if len(logResult.BlockHash) > 66 {
				logResult.BlockHash = logResult.BlockHash[:66]
			}
			if len(logResult.TransactionHash) > 66 {
				logResult.TransactionHash = logResult.TransactionHash[:66]
			}
			logs = append(logs, domain.Log{
				Address:          logResult.Address,
				BlockHash:        logResult.BlockHash,
				BlockNumber:      fmt.Sprintf("0x%x", *logResult.BlockNumber),
				Data:             logResult.Data,
				LogIndex:         fmt.Sprintf("0x%x", *logResult.Index),
				Removed:          false,
				Topics:           logResult.Topics,
				TransactionHash:  logResult.TransactionHash,
				TransactionIndex: fmt.Sprintf("0x%x", *logResult.TransactionIndex),
			})
		}
	}

	if logs == nil {
		return []domain.Log{}, nil
	}

	return logs, nil
}

func (s *commonService) GetBlockNumberByNumberOrTag(blockNumberOrTag string) (int64, *domain.RPCError) {
	s.logger.Debug("Getting block number by hash or tag", zap.String("blockHashOrTag", blockNumberOrTag))

	if blockTagIsLatestOrPending(&blockNumberOrTag) {
		blockNumberOrTag = "latest"
	}

	switch blockNumberOrTag {
	case "latest", "pending":
		latestBlock, errMap := s.GetBlockNumber()
		if errMap != nil {
			s.logger.Error("Failed to get latest block number", zap.Error(errMap))
			return 0, errMap
		}

		latestBlockStr, ok := latestBlock.(string)
		if !ok {
			s.logger.Error("Invalid block number format", zap.Error(errMap))
			return 0, errMap
		}

		// Convert hex string to int, remove "0x" prefix
		latestBlockNum, err := HexToDec(latestBlockStr)
		if err != nil {
			s.logger.Error("Failed to parse latest block number", zap.Error(err))
			return 0, domain.NewRPCError(domain.ServerError, "Invalid block number")
		}
		return latestBlockNum, nil

	case "earliest":
		return int64(0), nil
	default:
		// Convert hex string to int, remove "0x" prefix
		latestBlockNum, err := HexToDec(blockNumberOrTag)
		if err != nil {
			s.logger.Error("Failed to parse latest block number", zap.Error(err))
			return 0, domain.NewRPCError(domain.ServerError, "Invalid block number")
		}

		return latestBlockNum, nil
	}
}

func (s *commonService) GetBlockNumber() (interface{}, *domain.RPCError) {
	s.logger.Info("Getting block number")
	block, err := s.mClient.GetLatestBlock()
	if err != nil {
		s.logger.Error("Failed to fetch latest block", zap.Error(err))
		return nil, domain.NewRPCError(domain.ServerError, "Failed to fetch block data: "+err.Error())
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
	return nil, domain.NewRPCError(domain.ServerError, "Invalid block data")
}

func (s *commonService) ValidateBlockRange(fromBlock, toBlock string) *domain.RPCError {
	var fromBlockNum, toBlockNum int64

	latestBlockNum, errRpc := s.GetBlockNumberByNumberOrTag("latest")
	if errRpc != nil {
		return errRpc
	}

	if blockTagIsLatestOrPending(&toBlock) {
		toBlockNum = latestBlockNum
	} else {
		toBlockNum, errRpc = s.GetBlockNumberByNumberOrTag(toBlock)
		if errRpc != nil {
			return errRpc
		}

		// - When `fromBlock` is not explicitly provided, it defaults to `latest`.
		// - Then if `toBlock` equals `latestBlockNumber`, it means both `toBlock` and `fromBlock` essentially refer to the latest block, so the `MISSING_FROM_BLOCK_PARAM` error is not necessary.
		// - If `toBlock` is explicitly provided and does not equals to `latestBlockNumber`, it establishes a solid upper bound.
		// - If `fromBlock` is missing, indicating the absence of a lower bound, throw the `MISSING_FROM_BLOCK_PARAM` error
		if toBlockNum != latestBlockNum && fromBlock == "" {
			return domain.NewRPCError(domain.InvalidParams, "Provided toBlock parameter without specifying fromBlock")
		}
	}

	if blockTagIsLatestOrPending(&fromBlock) {
		fromBlockNum = latestBlockNum
	} else {
		fromBlockNum, errRpc = s.GetBlockNumberByNumberOrTag(fromBlock)
		if errRpc != nil {
			return errRpc
		}
	}

	if fromBlockNum > toBlockNum {
		return domain.NewInvalidBlockRangeError()
	}

	return nil
}

func blockTagIsLatestOrPending(tag *string) bool {
	return tag == nil ||
		*tag == "latest" ||
		*tag == "pending" ||
		*tag == "safe" ||
		*tag == "finalized"
}

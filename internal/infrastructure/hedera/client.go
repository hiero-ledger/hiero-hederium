package hedera

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/ethereum/go-ethereum/common"
	hedera "github.com/hiero-ledger/hiero-sdk-go/v2/sdk"
	"go.uber.org/zap"
)

type HederaNodeClient interface {
	GetNetworkFees() (int64, error)
	SendRawTransaction(transactionData []byte, networkGasPriceInTinyBars int64, callerId *common.Address) (*domain.TransactionResponse, error)
	DeleteFile(fileID *hedera.FileID) error
	GetContractByteCode(shard, realm int64, address string) ([]byte, error)
	GetOperatorPublicKey() string
}

type HederaClient struct {
	*hedera.Client
	operatorKeyFormat string
	logger            *zap.Logger
}

func NewHederaClient(network, operatorId, operatorKey, operatorKeyFormat string, networkConfig map[string]string, logger *zap.Logger) (*HederaClient, error) {
	var client *hedera.Client
	switch network {
	case "mainnet":
		client = hedera.ClientForMainnet()
	case "testnet":
		client = hedera.ClientForTestnet()
	case "previewnet":
		client = hedera.ClientForPreviewnet()
	case "local":
		var err error
		data, err := json.Marshal(networkConfig)
		if err != nil {
			return nil, err
		}
		jsonBytes := []byte(fmt.Sprintf(`{"network":%s}`, string(data)))
		client, err = hedera.ClientFromConfig(jsonBytes)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid network: %s", network)
	}

	accID, err := hedera.AccountIDFromString(operatorId)
	if err != nil {
		return nil, err
	}
	opKey, err := hedera.PrivateKeyFromString(operatorKey)
	if err != nil {
		return nil, err
	}
	client.SetOperator(accID, opKey)
	return &HederaClient{Client: client, operatorKeyFormat: operatorKeyFormat, logger: logger}, nil
}

func (h *HederaClient) GetNetworkFees() (int64, error) {
	// var feeScheduleBytes []byte
	// feeScheduleBytes, err := hedera.NewFileContentsQuery().
	// 	SetFileID(hedera.FileID{
	// 		Shard: 0,
	// 		Realm: 0,
	// 		File:  111,
	// 	}).
	// 	Execute(c)
	// if err != nil {
	// 	return 0, err
	// }

	// feeSchedule, err := hedera.FeeScheduleFromBytes(feeScheduleBytes)
	// if err != nil {
	// 	return 0, err
	// }

	// for _, txFeeSchedule := range feeSchedule.TransactionFeeSchedules {
	// 	txFeeSchedule.RequestType.
	// }

	// Hardcode for now, for simplicity
	return 72, nil
}

// SendRawTransaction submits an Ethereum transaction to the Hedera network.
// It handles large call data by creating a file if needed and validates gas prices.
func (h *HederaClient) SendRawTransaction(transactionData []byte, networkGasPriceInTinyBars int64, callerId *common.Address) (*domain.TransactionResponse, error) {
	ethereumTx := hedera.NewEthereumTransaction()

	ethereumData, err := hedera.EthereumTransactionDataFromBytes(transactionData)
	if err != nil {
		return nil, fmt.Errorf("failed to create ethereum transaction data: %v", err)
	}

	data, err := ethereumData.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to convert ethereum transaction data to bytes: %v", err)
	}

	h.logger.Info("Sending raw transaction", zap.Int("data length", len(data)))

	var fileID *hedera.FileID
	if len(data) <= fileAppendChunkSize {
		ethereumTx.SetEthereumData(data)
	} else {
		fileID, err = h.createFileForCallData(data)
		if err != nil && fileID == nil {
			h.logger.Error("Failed to create file for call data", zap.Error(err))
			return nil, fmt.Errorf("failed to create file for call data: %v", err)
		}

		ethereumTx.SetEthereumData(data).SetCallDataFileID(*fileID)
	}

	maxFee := hedera.HbarFromTinybar(networkGasPriceInTinyBars * maxGasPerSec)
	ethereumTx.SetMaxTransactionFee(maxFee)

	h.logger.Info("Executing transaction", zap.Int("data length", len(transactionData)), zap.Bool("using file", fileID != nil))

	response, err := ethereumTx.Execute(h.Client)
	if err != nil {
		h.logger.Error("Failed to execute transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to execute transaction: %v", err)
	}

	_, err = response.GetReceipt(h.Client)
	if err != nil {
		h.logger.Error("Failed to get transaction receipt", zap.Error(err))
	}

	return &domain.TransactionResponse{
		TransactionID: response.TransactionID.String(),
		FileID:        fileID,
	}, nil
}

// createFileForCallData creates a file to store large call data
func (h *HederaClient) createFileForCallData(data []byte) (*hedera.FileID, error) {
	// TODO: EstimateTxFee
	// TODO: hbarLimitService - check if the limit is reached

	h.logger.Info("Creating file for call data", zap.Int("data length", len(data)))

	fileCreateTx := hedera.NewFileCreateTransaction().
		SetContents(data[:fileAppendChunkSize]).
		SetKeys(h.Client.GetOperatorPublicKey())

	response, err := fileCreateTx.Execute(h.Client)
	if err != nil {
		h.logger.Error("Failed to execute file create transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to execute transaction: %v", err)
	}

	h.logger.Info("File create transaction executed successfully", zap.Any("response", response))

	receipt, err := response.GetReceipt(h.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to get file creation receipt: %v", err)
	}

	fileID := receipt.FileID
	if fileID == nil {
		return nil, fmt.Errorf("file creation did not return a file ID")
	}

	if len(data) > fileAppendChunkSize {
		remaining := data[fileAppendChunkSize:]
		appendTx := hedera.NewFileAppendTransaction().
			SetFileID(*fileID).
			SetContents(remaining).
			SetMaxChunkSize(fileAppendChunkSize).
			SetMaxChunks(maxChunks)
		transactionResponses, err := appendTx.ExecuteAll(h.Client)

		if err != nil {
			h.logger.Error("Failed to execute file append transaction", zap.Error(err))
			return nil, fmt.Errorf("failed to execute transaction: %v", err)
		}

		h.logger.Info(fmt.Sprintf("Successfully execute all %d file append transactions", len(transactionResponses)))
	}

	// Make query to see if the file is created successfully
	query := hedera.NewFileInfoQuery().SetFileID(*fileID)
	queryResponse, err := query.Execute(h.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}

	if queryResponse.Size == 0 {
		return nil, fmt.Errorf("created file is empty")
	}

	return fileID, nil
}

func (h *HederaClient) DeleteFile(fileID *hedera.FileID) error {
	h.logger.Info("Deleting file", zap.String("fileID", fileID.String()))

	deleteTx, err := hedera.NewFileDeleteTransaction().
		SetFileID(*fileID).SetMaxTransactionFee(hedera.NewHbar(2)).FreezeWith(h.Client)
	if err != nil {
		return fmt.Errorf("failed to freeze delete transaction: %v", err)
	}

	response, err := deleteTx.Execute(h.Client)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	_, err = response.GetReceipt(h.Client)
	if err != nil {
		return fmt.Errorf("failed to get delete receipt: %v", err)
	}

	query := hedera.NewFileInfoQuery().SetFileID(*fileID)
	queryResponse, err := query.Execute(h.Client)
	if err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}

	if queryResponse.Size == 0 {
		return fmt.Errorf("file was not deleted")
	}

	return nil
}

func (h *HederaClient) GetContractByteCode(shard, realm int64, address string) ([]byte, error) {
	h.logger.Info("Getting contract bytecode", zap.String("address", address))

	address = strings.TrimPrefix(address, "0x")
	contractID, err := hedera.ContractIDFromEvmAddress(uint64(shard), uint64(realm), address)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract ID from EVM address: %w", err)
	}

	query := hedera.NewContractBytecodeQuery().SetContractID(contractID)

	cost, err := query.GetCost(h.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to get query cost: %w", err)
	}

	query.SetQueryPayment(cost)

	response, err := query.Execute(h.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return response, nil
}

func (h *HederaClient) GetOperatorPublicKey() string {
	if h.operatorKeyFormat == "HEX_ECDSA" {
		return h.Client.GetOperatorPublicKey().ToEvmAddress()
	}
	accountId := h.Client.GetOperatorAccountID().String()
	return accountId
}

package hedera

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashgraph/hedera-sdk-go/v2"
)

type HederaNodeClient interface {
	GetNetworkFees() (int64, error)
	SendRawTransaction(transactionData []byte, networkGasPriceInWeiBars int64, callerId *common.Address) (*TransactionResponse, error)
	GetContractByteCode(shard, realm int64, address string) ([]byte, error)
}

type HederaClient struct {
	*hedera.Client
}

func NewHederaClient(network, operatorId, operatorKey string) (*HederaClient, error) {
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
		client, err = hedera.ClientForName("local")
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported Hedera network: %s", network)
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
	return &HederaClient{Client: client}, nil
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

const (
	// Maximum gas that can be used per second
	maxGasPerSec = 15000000
	// Transaction size limit in bytes (128KB)
	transactionSizeLimit = 128 * 1024
	// Default file append chunk size
	fileAppendChunkSize = 5120
	// Maximum number of chunks for file append
	maxChunks = 20
)

type TransactionResponse struct {
	TransactionID string
	FileID        *hedera.FileID
}

// SendRawTransaction submits an Ethereum transaction to the Hedera network.
// It handles large call data by creating a file if needed and validates gas prices.
func (h *HederaClient) SendRawTransaction(transactionData []byte, networkGasPriceInWeiBars int64, callerId *common.Address) (*TransactionResponse, error) {
	ethereumTx := hedera.NewEthereumTransaction()

	var fileID *hedera.FileID
	var err error

	if len(transactionData) <= fileAppendChunkSize {
		ethereumTx.SetEthereumData(transactionData)
	} else {
		fileID, err = h.createFileForCallData(transactionData)
		if err != nil {
			return nil, fmt.Errorf("failed to create file for call data: %v", err)
		}

		ethereumTx.SetEthereumData([]byte{})
		ethereumTx.SetCallDataFileID(*fileID)
	}

	// TODO: Make this in separate function
	networkGasPriceInTinyBars := networkGasPriceInWeiBars / 10000000000
	maxFee := hedera.NewHbar(float64(networkGasPriceInTinyBars*maxGasPerSec) / 100000000.0)
	ethereumTx.SetMaxTransactionFee(maxFee)

	response, err := ethereumTx.Execute(h.Client)
	if err != nil {
		if fileID != nil {
			_ = h.deleteFile(*fileID)
		}
		return nil, fmt.Errorf("failed to execute transaction: %v", err)
	}

	return &TransactionResponse{
		TransactionID: response.TransactionID.String(),
		FileID:        fileID,
	}, nil
}

// createFileForCallData creates a file to store large call data
func (h *HederaClient) createFileForCallData(data []byte) (*hedera.FileID, error) {
	// TODO: EstimateTxFee
	// TODO: hbarLimitService - check if the limit is reached

	// Create initial file with first chunk
	fileCreateTx := hedera.NewFileCreateTransaction().
		SetContents(data[:fileAppendChunkSize]).
		SetKeys(h.Client.GetOperatorPublicKey())

	resp, err := fileCreateTx.Execute(h.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}

	receipt, err := resp.GetReceipt(h.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to get file creation receipt: %v", err)
	}

	fileID := receipt.FileID
	if fileID == nil {
		return nil, fmt.Errorf("file creation did not return a file ID")
	}

	if len(data) > fileAppendChunkSize {
		remaining := data[fileAppendChunkSize:]
		for i := 0; i < len(remaining); i += fileAppendChunkSize {
			end := i + fileAppendChunkSize
			if end > len(remaining) {
				end = len(remaining)
			}

			chunk := remaining[i:end]
			appendTx := hedera.NewFileAppendTransaction().
				SetFileID(*fileID).
				SetContents(chunk)

			_, err = appendTx.Execute(h.Client)
			if err != nil {
				_ = h.deleteFile(*fileID)
				return nil, fmt.Errorf("failed to append chunk %d: %v", i/fileAppendChunkSize+1, err)
			}
		}
	}

	return fileID, nil
}

func (h *HederaClient) deleteFile(fileID hedera.FileID) error {
	deleteTx, err := hedera.NewFileDeleteTransaction().
		SetFileID(fileID).SetMaxTransactionFee(hedera.NewHbar(2)).FreezeWith(h.Client)
	if err != nil {
		return fmt.Errorf("failed to freeze delete transaction: %v", err)
	}

	_, err = deleteTx.Execute(h.Client)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

func (h *HederaClient) GetContractByteCode(shard, realm int64, address string) ([]byte, error) {
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

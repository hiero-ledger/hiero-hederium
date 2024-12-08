package hedera

import (
	"fmt"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

type HederaNodeClient interface {
	GetNetworkFees() (int64, error)
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

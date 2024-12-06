package hedera

import (
	"fmt"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

func NewHederaClient(network, operatorId, operatorKey string) (*hedera.Client, error) {
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
	return client, nil
}

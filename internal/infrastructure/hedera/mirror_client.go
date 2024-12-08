package hedera

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/georgi-l95/Hederium/internal/domain"
	"go.uber.org/zap"
)

type MirrorNodeClient interface {
	GetLatestBlock() (map[string]interface{}, error)
	GetNetworkFees() (int64, error)
}

type MirrorClient struct {
	BaseURL string
	Timeout time.Duration
	logger  *zap.Logger
}

func NewMirrorClient(baseURL string, timeoutSeconds int, logger *zap.Logger) *MirrorClient {
	return &MirrorClient{
		BaseURL: baseURL,
		Timeout: time.Duration(timeoutSeconds) * time.Second,
		logger:  logger,
	}
}

func (m *MirrorClient) GetLatestBlock() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.BaseURL+"/api/v1/blocks?order=desc&limit=1", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mirror node returned status %d", resp.StatusCode)
	}

	var result struct {
		Blocks []map[string]interface{} `json:"blocks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Blocks) == 0 {
		return nil, fmt.Errorf("no blocks returned by mirror node")
	}

	return result.Blocks[0], nil
}

func (m *MirrorClient) GetNetworkFees() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.BaseURL+"/api/v1/network/fees", nil)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("mirror node returned status %d", resp.StatusCode)
	}

	var feeResponse domain.FeeResponse

	if err := json.NewDecoder(resp.Body).Decode(&feeResponse); err != nil {
		return 0, err
	}
	if len(feeResponse.Fees) == 0 {
		return 0, fmt.Errorf("no fees returned by mirror node")
	}

	var gasTinybars int64
	for _, fee := range feeResponse.Fees {
		if fee.TransactionType == "EthereumTransaction" {
			gasTinybars = fee.Gas
			break
		}
	}
	return gasTinybars, nil
}

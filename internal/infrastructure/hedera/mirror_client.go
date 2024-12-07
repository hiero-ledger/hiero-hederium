package hedera

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type MirrorNodeClient interface {
	GetLatestBlock() (map[string]interface{}, error)
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

// Example: fetching the latest block information
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

// Additional methods for balance, transaction info, logs etc. can be added similarly.

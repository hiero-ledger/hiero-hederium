package hedera

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/georgi-l95/Hederium/internal/domain"
	"go.uber.org/zap"
)

type MirrorNodeClient interface {
	GetLatestBlock() (map[string]interface{}, error)
	GetBlockByHashOrNumber(hashOrNumber string) *domain.BlockResponse
	GetNetworkFees() (int64, error)
	GetContractResults(timestamp domain.Timestamp) []domain.ContractResults
	GetBalance(address string, timestampTo string) string
	GetAccount(address string, timestampTo string) interface{}
	GetContractResult(transactionId string) interface{}
	PostCall(callObject map[string]interface{}) interface{}
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

func (m *MirrorClient) GetBlockByHashOrNumber(hashOrNumber string) *domain.BlockResponse {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.BaseURL+"/api/v1/blocks/"+hashOrNumber, nil)
	if err != nil {
		m.logger.Error("Error creating request to get block by hash or number", zap.Error(err))
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Error("Error getting block by hash or number", zap.Error(err))
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
		return nil
	}

	var result domain.BlockResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.logger.Error("Error decoding response body", zap.Error(err))
		return nil
	}

	m.logger.Debug("Block", zap.Any("block", result))
	return &result
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

func (m *MirrorClient) GetContractResults(timestamp domain.Timestamp) []domain.ContractResults {
	var allResults []domain.ContractResults
	currentURL := fmt.Sprintf("%s/api/v1/contracts/results?timestamp=gte:%s&timestamp=lte:%s&limit=100&order=asc",
		m.BaseURL, timestamp.From, timestamp.To)

	for currentURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, currentURL, nil)
		if err != nil {
			m.logger.Error("Error creating request", zap.Error(err))
			return []domain.ContractResults{} // Return empty array instead of nil
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			m.logger.Error("Error making request", zap.Error(err))
			return []domain.ContractResults{} // Return empty array instead of nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
			return []domain.ContractResults{} // Return empty array instead of nil
		}

		var result struct {
			Results []domain.ContractResults `json:"results"`
			Links   struct {
				Next *string `json:"next"`
			} `json:"links"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			m.logger.Error("Error decoding response body", zap.Error(err))
			return []domain.ContractResults{} // Return empty array instead of nil
		}

		// It's okay if there are no results, just continue with the empty array
		allResults = append(allResults, result.Results...)

		// Update URL for next iteration or break the loop
		if result.Links.Next != nil {
			currentURL = m.BaseURL + *result.Links.Next
		} else {
			currentURL = ""
		}
	}

	return allResults
}

func (m *MirrorClient) GetBalance(address string, timestampTo string) string {
	m.logger.Debug("Getting balance", zap.String("address", address), zap.String("timestampTo", timestampTo))
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.BaseURL+"/api/v1/balances?account.id="+address+"&timestamp=lte:"+timestampTo, nil)
	if err != nil {
		m.logger.Error("Error creating request to get balance", zap.Error(err))
		return "0x0"
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Error("Error getting balance", zap.Error(err))
		return "0x0"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
		return "0x0"
	}

	var result struct {
		Timestamp string `json:"timestamp"`
		Balances  []struct {
			Account string        `json:"account"`
			Balance *big.Int      `json:"balance"`
			Tokens  []interface{} `json:"tokens"`
		} `json:"balances"`
		Links struct {
			Next *string `json:"next"`
		} `json:"links"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.logger.Error("Error decoding response body", zap.Error(err))
		return "0x0"
	}

	m.logger.Debug("Balance", zap.Any("balance", result))
	if len(result.Balances) == 0 {
		m.logger.Debug("No balances found")
		return "0x0"
	}

	// Convert tinybars to weibars
	balance := result.Balances[0].Balance.Mul(result.Balances[0].Balance, big.NewInt(10000000000))
	return "0x" + fmt.Sprintf("%x", balance)
}

func (m *MirrorClient) GetAccount(address string, timestampTo string) interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.BaseURL+"/api/v1/accounts/"+address+"?limit=1&order=desc&timestamp=lte:"+timestampTo+"&transactiontype=ETHEREUMTRANSACTION&transactions=true", nil)
	if err != nil {
		m.logger.Error("Error creating request to get account", zap.Error(err))
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Error("Error getting account", zap.Error(err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
		return nil
	}

	var result domain.AccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.logger.Error("Error decoding response body", zap.Error(err))
		return nil
	}

	return result
}

func (m *MirrorClient) GetContractResult(transactionId string) interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.BaseURL+"/api/v1/contracts/results/"+transactionId, nil)
	if err != nil {
		m.logger.Error("Error creating request to get contract result", zap.Error(err))
		return nil
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Error("Error getting contract result", zap.Error(err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
		return nil
	}

	var result domain.ContractResultResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.logger.Error("Error decoding response body", zap.Error(err))
		return nil
	}

	return result
}

func (m *MirrorClient) PostCall(callObject map[string]interface{}) interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	jsonBody, err := json.Marshal(callObject)
	if err != nil {
		m.logger.Error("Error marshaling call object", zap.Error(err))
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.BaseURL+"/api/v1/contracts/call", bytes.NewBuffer(jsonBody))
	if err != nil {
		m.logger.Error("Error creating request for contract call", zap.Error(err))
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Error("Error making contract call", zap.Error(err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Mirror node returned non-OK status", zap.Int("status", resp.StatusCode))
		return nil
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.logger.Error("Error decoding response body", zap.Error(err))
		return nil
	}

	return result.Result
}

package hedera

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/infrastructure/cache"
	"go.uber.org/zap"
)

type MirrorNodeClient interface {
	GetLatestBlock() (map[string]interface{}, error)
	GetBlocks(blockNumber string) ([]map[string]interface{}, error)
	GetBlockByHashOrNumber(hashOrNumber string) *domain.BlockResponse
	GetNetworkFees(timestampTo, order string) (int64, error)
	GetContractResults(timestamp domain.Timestamp) []domain.ContractResults
	GetBalance(address string, timestampTo string) string
	GetAccount(address string, timestampTo string) interface{}
	GetContractResult(transactionId string) interface{}
	PostCall(callObject map[string]interface{}) interface{}
	GetContractStateByAddressAndSlot(address string, slot string, timestampTo string) (*domain.ContractStateResponse, error)
	GetContractResultsLogsByAddress(address string, queryParams map[string]interface{}) ([]domain.LogEntry, error)
	GetContractResultsLogsWithRetry(queryParams map[string]interface{}) ([]domain.LogEntry, error)
	GetContractResultWithRetry(queryParams map[string]interface{}) (*domain.ContractResults, error)
	GetContractById(contractIdOrAddress string) (*domain.ContractResponse, error)
	GetAccountById(idOrAliasOrEvmAddress string) (*domain.AccountResponse, error)
	GetTokenById(tokenId string) (*domain.TokenResponse, error)
	RepeatGetContractResult(transactionIdOrHash string, retries int) *domain.ContractResultResponse
}

type MirrorClient struct {
	BaseURL      string
	Web3URL      string
	Timeout      time.Duration
	logger       *zap.Logger
	cacheService cache.CacheService
}

func NewMirrorClient(baseURL string, web3Url string, timeoutSeconds int, logger *zap.Logger, cacheService cache.CacheService) *MirrorClient {
	return &MirrorClient{
		BaseURL:      baseURL,
		Web3URL:      web3Url,
		Timeout:      time.Duration(timeoutSeconds) * time.Second,
		logger:       logger,
		cacheService: cacheService,
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

func (m *MirrorClient) GetBlocks(blockNumber string) ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	str := fmt.Sprintf("block.number=gt:%s&order=asc", blockNumber)

	url := fmt.Sprintf("%s/api/v1/blocks?%s", m.BaseURL, str)

	m.logger.Info("Gettting blocks", zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting blocks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mirror node returned status %d", resp.StatusCode)
	}

	var result struct {
		Blocks []map[string]interface{} `json:"blocks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("no blocks returned by mirror node")
	}

	return result.Blocks, nil
}

func (m *MirrorClient) GetBlockByHashOrNumber(hashOrNumber string) *domain.BlockResponse {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	cachedKey := fmt.Sprintf("%s_%s", GetBlockByHashOrNumber, hashOrNumber)

	var cachedBlock domain.BlockResponse
	if err := m.cacheService.Get(ctx, cachedKey, &cachedBlock); err == nil && cachedBlock.Hash != "" {
		return &cachedBlock
	}

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

	if err := m.cacheService.Set(ctx, cachedKey, &result, DefaultExpiration); err != nil {
		m.logger.Error("Error caching block", zap.Error(err))
	}

	m.logger.Debug("Block", zap.Any("block", result))
	return &result
}

func (m *MirrorClient) GetNetworkFees(timestampTo, order string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	queryParams := ""
	if order == "" {
		order = "desc"
	}

	if timestampTo != "" {
		queryParams += fmt.Sprintf("?order=%s", order)
		queryParams += "&timestamp=lte:" + timestampTo
	}

	m.logger.Debug("Asking this endpoint:", zap.String("url", m.BaseURL+"/api/v1/network/fees"+queryParams))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.BaseURL+"/api/v1/network/fees"+queryParams, nil)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	// TODO: If the mirror node does not return fee then ask the SDK for the fee
	var checkSDK bool
	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
		//return 0, fmt.Errorf("mirror node returned status %d", resp.StatusCode)
		checkSDK = true
	}
	// For now the default fee is 23
	if checkSDK {
		return 23, nil
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

	var reqUrl string
	if timestampTo == "0" {
		reqUrl = m.BaseURL + "/api/v1/balances?account.id=" + address
	} else {
		reqUrl = m.BaseURL + "/api/v1/balances?account.id=" + address + "&timestamp=lte:" + timestampTo
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
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

func (m *MirrorClient) GetContractResult(transactionIdOrHash string) interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	cachedKey := fmt.Sprintf("%s_%s", GetContractResult, transactionIdOrHash)

	var cachedResult domain.ContractResultResponse
	if err := m.cacheService.Get(ctx, cachedKey, &cachedResult); err == nil && cachedResult.BlockHash != "" {
		m.logger.Info("Contract result found in cache", zap.Any("result", cachedResult))
		return cachedResult
	}

	url := fmt.Sprintf("%s/api/v1/contracts/results/%s", m.BaseURL, transactionIdOrHash)

	m.logger.Info("Getting contract result", zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	if err := m.cacheService.Set(ctx, cachedKey, &result, DefaultExpiration); err != nil {
		m.logger.Error("Error caching contract result", zap.Error(err))
	}

	cachedKeyHash := fmt.Sprintf("%s_%s", GetContractResult, result.Hash[:66])
	if err := m.cacheService.Set(ctx, cachedKeyHash, &result, DefaultExpiration); err != nil {
		m.logger.Error("Error caching contract result", zap.Error(err))
	}

	m.logger.Info("Contract result", zap.Any("result", result))

	return result
}

func (m *MirrorClient) RepeatGetContractResult(transactionIdOrHash string, retries int) *domain.ContractResultResponse {
	for i := 0; i < retries; i++ {
		result := m.GetContractResult(transactionIdOrHash)
		if result, ok := result.(domain.ContractResultResponse); ok {
			return &result
		}

		time.Sleep(1 * time.Second)
	}
	return nil
}

func (m *MirrorClient) PostCall(callObject map[string]interface{}) interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()
	jsonBody, err := json.Marshal(callObject)
	if err != nil {
		m.logger.Error("Error marshaling call object", zap.Error(err))
		return nil
	}
	url := fmt.Sprintf("%s/api/v1/contracts/call", m.Web3URL)
	m.logger.Info("Posting contract call", zap.String("url", url))
	m.logger.Info("Body", zap.String("body", string(jsonBody)))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
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

func (m *MirrorClient) GetContractStateByAddressAndSlot(address string, slot string, timestampTo string) (*domain.ContractStateResponse, error) {
	queryParams := make([]string, 0, 3)

	// Hardcode limit and order
	queryParams = append(queryParams, "limit=100", "order=desc")

	// If we have a blockEndTimestamp, add it
	if timestampTo != "" {
		queryParams = append(queryParams, "timestamp="+timestampTo)
	}

	queryParams = append(queryParams, "slot="+fmt.Sprint(slot))

	url := fmt.Sprintf("%s/api/v1/contracts/%s/state?%s", m.BaseURL, address, strings.Join(queryParams, "&"))

	m.logger.Info("Getting contract state", zap.String("url", url))

	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		m.logger.Error("Error creating request to get contract state", zap.Error(err))
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Error("Error getting contract state", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
		return nil, nil // Here we return nil, nil to tell the service that the mirror node did not return a state
	}

	var result domain.ContractStateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.logger.Error("Error decoding response body", zap.Error(err))
		return nil, err
	}

	return &result, nil
}

func (m *MirrorClient) GetContractResultsLogsWithRetry(queryParams map[string]interface{}) ([]domain.LogEntry, error) {
	queryParamsStr := formatQueryParams(queryParams)

	url := fmt.Sprintf("%s/api/v1/contracts/results/logs?%s&limit=%d", m.BaseURL, queryParamsStr, Limit)

	logs, err := m.getPaginatedResults(url)
	if err != nil {
		return nil, err
	}

	for i := 0; i < maxRetries; i++ {
		isLastAttempt := i == maxRetries-1

		m.logger.Debug("Contract results logs", zap.Any("logs", logs))

		foundImmatureRecord := false
		for _, log := range logs {
			if log.TransactionIndex == nil || log.BlockNumber == nil || log.BlockHash == "0x" || log.Index == nil {
				m.logger.Debug("Contract results log contains nullable transaction_index or block_number, or block_hash is an empty hex (0x)",
					zap.String("contract_result", fmt.Sprintf("%+v", log)),
					zap.Duration("retry_delay", retryDelay))

				if isLastAttempt {
					return nil, fmt.Errorf("dependent service returned immature records")
				}

				foundImmatureRecord = true
				break
			}
		}

		if !foundImmatureRecord {
			return logs, nil
		}

		m.logger.Debug("Found immature record, retrying", zap.Duration("retry_delay", retryDelay))

		time.Sleep(retryDelay)

		logs, err = m.getPaginatedResults(url)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (m *MirrorClient) GetContractResultsLogsByAddress(address string, queryParams map[string]interface{}) ([]domain.LogEntry, error) {
	queryParamsStr := formatQueryParams(queryParams)

	url := fmt.Sprintf("%s/api/v1/contracts/%s/results/logs?%s&limit=%d", m.BaseURL, address, queryParamsStr, Limit)

	logs, err := m.getPaginatedResults(url)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (m *MirrorClient) fetchLogsPages(url string) (*domain.ContractResultsLogResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		m.logger.Error("Error creating request", zap.Error(err))
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Error("Error making request", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("mirror node returned status %d", resp.StatusCode)
	}

	var result domain.ContractResultsLogResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.logger.Error("Error decoding response", zap.Error(err))
		return nil, err
	}

	return &result, nil
}

func (m *MirrorClient) getPaginatedResults(url string) ([]domain.LogEntry, error) {
	var logs []domain.LogEntry
	for page := 1; page <= MaxPages; page++ {
		m.logger.Info("", zap.String("url", url))
		result, err := m.fetchLogsPages(url)
		if err != nil {
			return nil, err
		}

		if len(result.Logs) == 0 {
			break
		}

		logs = append(logs, result.Logs...)

		if result.Links.Next == nil {
			break
		}

		url = fmt.Sprintf("%s%s", m.BaseURL, *result.Links.Next)
	}

	return logs, nil
}

func (m *MirrorClient) GetContractResultWithRetry(queryParams map[string]interface{}) (*domain.ContractResults, error) {
	queryParamsStr := formatQueryParams(queryParams)

	url := fmt.Sprintf("%s/api/v1/contracts/results?%s", m.BaseURL, queryParamsStr)

	m.logger.Info("Getting contract result with retry", zap.String("url", url))

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			m.logger.Error("Error creating request", zap.Error(err))
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			m.logger.Error("Error making request", zap.Error(err))
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
			return nil, fmt.Errorf("mirror node returned status %d", resp.StatusCode)
		}

		// Should make struct for this
		var result struct {
			Results []domain.ContractResults `json:"results"`
			Links   struct {
				Next *string `json:"next"`
			} `json:"links"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			m.logger.Error("Error decoding response", zap.Error(err))
			return nil, err
		}

		// Check if results are empty and links.next is null
		if len(result.Results) == 0 && result.Links.Next == nil {
			m.logger.Info("Empty results and no next link, returning")
			return nil, nil
		}

		foundImmatureRecord := false
		for _, res := range result.Results {
			if res.TransactionIndex == 0 || res.BlockNumber == 0 || res.BlockHash == "0x" {
				m.logger.Debug("Contract result contains nullable transaction_index or block_number, or block_hash is an empty hex (0x)",
					zap.String("contract_result", fmt.Sprintf("%+v", res)),
					zap.Duration("retry_delay", retryDelay))
				foundImmatureRecord = true
				break
			}
		}

		if !foundImmatureRecord && len(result.Results) > 0 {
			return &result.Results[0], nil
		}

		m.logger.Debug("Found immature record, retrying")

		time.Sleep(retryDelay)
	}

	return nil, nil
}

// Util function to format query params
func formatQueryParams(params map[string]interface{}) string {
	var queryParams []string
	for key, value := range params {
		queryParams = append(queryParams, fmt.Sprintf("%s=%v", key, value))
	}
	queryParamsStr := strings.Join(queryParams, "&")
	if queryParamsStr != "" {
		queryParamsStr += "&order=desc" // Hardcoded order for now
	}
	return queryParamsStr
}

func (m *MirrorClient) GetContractById(contractIdOrAddress string) (*domain.ContractResponse, error) {
	url := fmt.Sprintf("%s/api/v1/contracts/%s", m.BaseURL, contractIdOrAddress)

	m.logger.Info("Getting contract by id", zap.String("url", url))

	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	cachedKey := fmt.Sprintf("%s_%s", GetContractById, contractIdOrAddress)

	var cachedContract domain.ContractResponse
	if err := m.cacheService.Get(ctx, cachedKey, &cachedContract); err == nil && cachedContract.EvmAddress != "" {
		return &cachedContract, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		m.logger.Error("Error creating request", zap.Error(err))
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Error("Error making request", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("mirror node returned status %d", resp.StatusCode)
	}

	var result domain.ContractResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.logger.Error("Error decoding response", zap.Error(err))
		return nil, err
	}

	if err := m.cacheService.Set(ctx, cachedKey, &result, DefaultExpiration); err != nil {
		m.logger.Error("Error caching contract", zap.Error(err))
	}

	return &result, nil
}

func (m *MirrorClient) GetAccountById(idOrAliasOrEvmAddress string) (*domain.AccountResponse, error) {
	url := fmt.Sprintf("%s/api/v1/accounts/%s?transactions=false", m.BaseURL, idOrAliasOrEvmAddress)

	m.logger.Info("Getting account by id", zap.String("url", url))

	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	cachedKey := fmt.Sprintf("%s_%s", GetAccountById, idOrAliasOrEvmAddress)

	var cachedAccount domain.AccountResponse
	if err := m.cacheService.Get(ctx, cachedKey, &cachedAccount); err == nil && cachedAccount.EvmAddress != "" {
		return &cachedAccount, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		m.logger.Error("Error creating request", zap.Error(err))
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Error("Error making request", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("mirror node returned status %d", resp.StatusCode)
	}

	var result domain.AccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.logger.Error("Error decoding response", zap.Error(err))
		return nil, err
	}

	if err := m.cacheService.Set(ctx, cachedKey, &result, DefaultExpiration); err != nil {
		m.logger.Error("Error caching account", zap.Error(err))
	}

	return &result, nil
}

func (m *MirrorClient) GetTokenById(tokenId string) (*domain.TokenResponse, error) {
	url := fmt.Sprintf("%s/api/v1/tokens/%s", m.BaseURL, tokenId)

	m.logger.Info("Getting token by id", zap.String("url", url))

	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
	defer cancel()

	cachedKey := fmt.Sprintf("%s_%s", GetTokenById, tokenId)

	var cachedToken domain.TokenResponse
	if err := m.cacheService.Get(ctx, cachedKey, &cachedToken); err == nil && cachedToken.TokenId != "" {
		return &cachedToken, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		m.logger.Error("Error creating request", zap.Error(err))
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.logger.Error("Error making request", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.logger.Error("Mirror node returned status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("mirror node returned status %d", resp.StatusCode)
	}

	var result domain.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		m.logger.Error("Error decoding response", zap.Error(err))
		return nil, err
	}

	if err := m.cacheService.Set(ctx, cachedKey, &result, DefaultExpiration); err != nil {
		m.logger.Error("Error caching token", zap.Error(err))
	}

	return &result, nil
}

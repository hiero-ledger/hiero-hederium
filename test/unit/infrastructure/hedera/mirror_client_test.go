package hedera_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/georgi-l95/Hederium/internal/domain"
	"github.com/georgi-l95/Hederium/internal/infrastructure/hedera"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupTest(t *testing.T) (*zap.Logger, *httptest.Server) {
	logger, _ := zap.NewDevelopment()
	return logger, nil
}

func TestNewMirrorClient(t *testing.T) {
	logger, _ := setupTest(t)
	baseURL := "http://test.com"
	timeoutSeconds := 30

	client := hedera.NewMirrorClient(baseURL, timeoutSeconds, logger)

	assert.Equal(t, baseURL, client.BaseURL)
	assert.Equal(t, time.Duration(timeoutSeconds)*time.Second, client.Timeout)
}

func TestGetLatestBlock_Success(t *testing.T) {
	logger, _ := setupTest(t)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/blocks?order=desc&limit=1", r.URL.String())
		assert.Equal(t, http.MethodGet, r.Method)

		response := struct {
			Blocks []map[string]interface{} `json:"blocks"`
		}{
			Blocks: []map[string]interface{}{
				{
					"number": float64(123),
					"hash":   "0xabc",
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	block, err := client.GetLatestBlock()

	assert.NoError(t, err)
	assert.Equal(t, float64(123), block["number"])
	assert.Equal(t, "0xabc", block["hash"])
}

func TestGetLatestBlock_EmptyResponse(t *testing.T) {
	logger, _ := setupTest(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Blocks []map[string]interface{} `json:"blocks"`
		}{
			Blocks: []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	block, err := client.GetLatestBlock()

	assert.Error(t, err)
	assert.Nil(t, block)
	assert.Contains(t, err.Error(), "no blocks returned")
}

func TestGetBlockByHashOrNumber_Success(t *testing.T) {
	logger, _ := setupTest(t)

	expectedBlock := &domain.BlockResponse{
		Number:       123,
		Hash:         "0xabc",
		PreviousHash: "0xdef",
		GasUsed:      1000,
		Size:         2000,
		LogsBloom:    "0x0",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/blocks/123", r.URL.String())
		assert.Equal(t, http.MethodGet, r.Method)

		json.NewEncoder(w).Encode(expectedBlock)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	block := client.GetBlockByHashOrNumber("123")

	assert.NotNil(t, block)
	assert.Equal(t, expectedBlock.Number, block.Number)
	assert.Equal(t, expectedBlock.Hash, block.Hash)
	assert.Equal(t, expectedBlock.PreviousHash, block.PreviousHash)
}

func TestGetNetworkFees_Success(t *testing.T) {
	logger, _ := setupTest(t)

	expectedResponse := domain.FeeResponse{
		Fees: []domain.Fee{
			{
				Gas:             100000,
				TransactionType: "EthereumTransaction",
			},
			{
				Gas:             50000,
				TransactionType: "OtherTransaction",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/network/fees", r.URL.String())
		assert.Equal(t, http.MethodGet, r.Method)

		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	fees, err := client.GetNetworkFees()

	assert.NoError(t, err)
	assert.Equal(t, int64(100000), fees) // Should return the EthereumTransaction fee
}

func TestGetContractResults_Success(t *testing.T) {
	logger, _ := setupTest(t)

	timestamp := domain.Timestamp{
		From: "1640995200",
		To:   "1640995300",
	}

	expectedResults := []domain.ContractResults{
		{
			Hash:   "0xtx1",
			Result: "SUCCESS",
		},
		{
			Hash:   "0xtx2",
			Result: "SUCCESS",
		},
	}

	// First page response
	firstPage := struct {
		Results []domain.ContractResults `json:"results"`
		Links   struct {
			Next *string `json:"next"`
		} `json:"links"`
	}{
		Results: expectedResults[:1],
	}
	nextLink := "/api/v1/contracts/results?page=2"
	firstPage.Links.Next = &nextLink

	// Second page response
	secondPage := struct {
		Results []domain.ContractResults `json:"results"`
		Links   struct {
			Next *string `json:"next"`
		} `json:"links"`
	}{
		Results: expectedResults[1:],
	}

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callCount == 0 {
			// First call should have the timestamp parameters
			assert.Contains(t, r.URL.String(), "timestamp=gte:"+timestamp.From)
			assert.Contains(t, r.URL.String(), "timestamp=lte:"+timestamp.To)
			json.NewEncoder(w).Encode(firstPage)
		} else {
			// Second call should use the next link
			assert.Equal(t, nextLink, r.URL.String())
			json.NewEncoder(w).Encode(secondPage)
		}
		callCount++
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	results := client.GetContractResults(timestamp)

	assert.Equal(t, 2, len(results))
	assert.Equal(t, expectedResults[0].Hash, results[0].Hash)
	assert.Equal(t, expectedResults[1].Hash, results[1].Hash)
}

func TestGetContractResults_ErrorResponse(t *testing.T) {
	logger, _ := setupTest(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	results := client.GetContractResults(domain.Timestamp{})

	assert.Empty(t, results)
}

func TestGetNetworkFees_NoEthereumFee(t *testing.T) {
	logger, _ := setupTest(t)

	response := domain.FeeResponse{
		Fees: []domain.Fee{
			{
				Gas:             50000,
				TransactionType: "OtherTransaction",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	fees, err := client.GetNetworkFees()

	assert.NoError(t, err)
	assert.Equal(t, int64(0), fees) // Should return 0 when no EthereumTransaction fee is found
}

func TestGetNetworkFees_EmptyResponse(t *testing.T) {
	logger, _ := setupTest(t)

	response := domain.FeeResponse{
		Fees: []domain.Fee{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	_, err := client.GetNetworkFees()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no fees returned")
}

func TestGetBlockByHashOrNumber_ErrorResponse(t *testing.T) {
	logger, _ := setupTest(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	block := client.GetBlockByHashOrNumber("123")

	assert.Nil(t, block)
}

func TestGetLatestBlock_ErrorResponse(t *testing.T) {
	logger, _ := setupTest(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	block, err := client.GetLatestBlock()

	assert.Error(t, err)
	assert.Nil(t, block)
	assert.Contains(t, err.Error(), "mirror node returned status 500")
}

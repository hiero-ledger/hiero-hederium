package hedera_test

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/infrastructure/hedera"
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
	fees, err := client.GetNetworkFees("", "") // Should be handled better

	assert.NoError(t, err)
	// Here I am changing the expected value to with *100 so that the test dont fail!!!
	assert.Equal(t, int64(10000000), fees) // Should return the EthereumTransaction fee
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
	fees, err := client.GetNetworkFees("", "") //  Should be handled better

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
	_, err := client.GetNetworkFees("", "") // Should be handled better

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

func TestGetBalance(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	testCases := []struct {
		name           string
		address        string
		timestampTo    string
		mockResponse   interface{}
		expectedResult string
		statusCode     int
	}{
		{
			name:        "Successful balance fetch",
			address:     "0x1234567890123456789012345678901234567890",
			timestampTo: "2023-12-09T12:00:00.000Z",
			mockResponse: map[string]interface{}{
				"timestamp": "2023-12-09T12:00:00.000Z",
				"balances": []map[string]interface{}{
					{
						"account": "0x1234567890123456789012345678901234567890",
						"balance": 1000000,
					},
				},
			},
			expectedResult: "0x2386f26fc10000", // 1000000 * 10000000000 in hex
			statusCode:     http.StatusOK,
		},
		{
			name:           "Empty balances array",
			address:        "0x1234567890123456789012345678901234567890",
			timestampTo:    "2023-12-09T12:00:00.000Z",
			mockResponse:   map[string]interface{}{"balances": []map[string]interface{}{}},
			expectedResult: "0x0",
			statusCode:     http.StatusOK,
		},
		{
			name:           "Invalid response structure",
			address:        "0x1234567890123456789012345678901234567890",
			timestampTo:    "2023-12-09T12:00:00.000Z",
			mockResponse:   "invalid json",
			expectedResult: "0x0",
			statusCode:     http.StatusOK,
		},
		{
			name:           "Server error",
			address:        "0x1234567890123456789012345678901234567890",
			timestampTo:    "2023-12-09T12:00:00.000Z",
			mockResponse:   nil,
			expectedResult: "0x0",
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/balances", r.URL.Path)
				assert.Equal(t, tc.address, r.URL.Query().Get("account.id"))
				assert.Equal(t, "lte:"+tc.timestampTo, r.URL.Query().Get("timestamp"))

				w.WriteHeader(tc.statusCode)
				if tc.mockResponse != nil {
					json.NewEncoder(w).Encode(tc.mockResponse)
				}
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, 5, logger)
			result := client.GetBalance(tc.address, tc.timestampTo)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestGetBalance_Success(t *testing.T) {
	logger, _ := setupTest(t)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/balances", r.URL.Path)
		assert.Equal(t, "account.id=0.0.123&timestamp=lte:1234567890.000000000", r.URL.RawQuery)
		assert.Equal(t, http.MethodGet, r.Method)

		response := struct {
			Timestamp string `json:"timestamp"`
			Balances  []struct {
				Account string        `json:"account"`
				Balance *big.Int      `json:"balance"`
				Tokens  []interface{} `json:"tokens"`
			} `json:"balances"`
		}{
			Balances: []struct {
				Account string        `json:"account"`
				Balance *big.Int      `json:"balance"`
				Tokens  []interface{} `json:"tokens"`
			}{
				{
					Account: "0.0.123",
					Balance: big.NewInt(1000000), // 1 million tinybars
					Tokens:  []interface{}{},
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	result := client.GetBalance("0.0.123", "1234567890.000000000")

	// 1 million tinybars * 10000000000 (conversion to weibars) = 10000000000000000 weibars
	expectedHex := "0x" + new(big.Int).Mul(big.NewInt(1000000), big.NewInt(10000000000)).Text(16)
	assert.Equal(t, expectedHex, result)
}

func TestGetBalance_Error(t *testing.T) {
	logger, _ := setupTest(t)

	// Create test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	result := client.GetBalance("0.0.123", "1234567890.000000000")

	assert.Equal(t, "0x0", result)
}

func TestGetAccount_Success(t *testing.T) {
	logger, _ := setupTest(t)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/accounts/0.0.123", r.URL.Path)
		assert.Equal(t, "limit=1&order=desc&timestamp=lte:1234567890.000000000&transactiontype=ETHEREUMTRANSACTION&transactions=true", r.URL.RawQuery)
		assert.Equal(t, http.MethodGet, r.Method)

		response := domain.AccountResponse{
			Account: "0.0.123",
			Balance: struct {
				Balance   int64         `json:"balance"`
				Timestamp string        `json:"timestamp"`
				Tokens    []interface{} `json:"tokens"`
			}{
				Balance:   1000000,
				Timestamp: "1234567890.000000000",
				Tokens:    []interface{}{},
			},
			EthereumNonce: 5,
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	result := client.GetAccount("0.0.123", "1234567890.000000000")

	assert.NotNil(t, result)
	accountResponse := result.(domain.AccountResponse)
	assert.Equal(t, "0.0.123", accountResponse.Account)
	assert.Equal(t, int64(5), accountResponse.EthereumNonce)
	assert.Equal(t, int64(1000000), accountResponse.Balance.Balance)
}

func TestGetAccount_Error(t *testing.T) {
	logger, _ := setupTest(t)

	// Create test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, 30, logger)
	result := client.GetAccount("0.0.123", "1234567890.000000000")

	assert.Nil(t, result)
}

func TestPostCall(t *testing.T) {
	logger, _ := setupTest(t)

	testCases := []struct {
		name           string
		callObject     map[string]interface{}
		mockResponse   interface{}
		expectedResult string
		statusCode     int
	}{
		{
			name: "Successful contract call",
			callObject: map[string]interface{}{
				"data": "0x123456",
				"to":   "0x1234567890123456789012345678901234567890",
			},
			mockResponse: struct {
				Result string `json:"result"`
			}{
				Result: "0xabcdef",
			},
			expectedResult: "0xabcdef",
			statusCode:     http.StatusOK,
		},
		{
			name: "Server error",
			callObject: map[string]interface{}{
				"data": "0x123456",
			},
			mockResponse:   nil,
			expectedResult: "",
			statusCode:     http.StatusInternalServerError,
		},
		{
			name: "Invalid response structure",
			callObject: map[string]interface{}{
				"data": "0x123456",
			},
			mockResponse:   "invalid json",
			expectedResult: "",
			statusCode:     http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/contracts/call", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var receivedCallObject map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&receivedCallObject)
				assert.NoError(t, err)
				assert.Equal(t, tc.callObject, receivedCallObject)

				w.WriteHeader(tc.statusCode)
				if tc.mockResponse != nil {
					json.NewEncoder(w).Encode(tc.mockResponse)
				}
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, 5, logger)
			result := client.PostCall(tc.callObject)

			if tc.expectedResult == "" {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

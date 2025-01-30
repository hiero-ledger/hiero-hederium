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
	fees, err := client.GetNetworkFees("", "")
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

func TestGetContractStateByAddressAndSlot(t *testing.T) {
	logger, _ := setupTest(t)

	testCases := []struct {
		name           string
		address        string
		slot           string
		timestampTo    string
		mockResponse   *domain.ContractStateResponse
		expectedResult *domain.ContractStateResponse
		expectedError  bool
		statusCode     int
	}{
		{
			name:        "Successful contract state fetch",
			address:     "0x1234567890123456789012345678901234567890",
			slot:        "0x0000000000000000000000000000000000000000000000000000000000000001",
			timestampTo: "2023-12-09T12:00:00.000Z",
			mockResponse: &domain.ContractStateResponse{
				State: []domain.ContractState{
					{
						Value: "0x0000000000000000000000000000000000000000000000000000000000000123",
					},
				},
			},
			expectedResult: &domain.ContractStateResponse{
				State: []domain.ContractState{
					{
						Value: "0x0000000000000000000000000000000000000000000000000000000000000123",
					},
				},
			},
			expectedError: false,
			statusCode:    http.StatusOK,
		},
		{
			name:           "Not found error (404)",
			address:        "0x1234567890123456789012345678901234567890",
			slot:           "0x0000000000000000000000000000000000000000000000000000000000000001",
			timestampTo:    "2023-12-09T12:00:00.000Z",
			mockResponse:   nil,
			expectedResult: nil,
			expectedError:  false,
			statusCode:     http.StatusNotFound,
		},
		{
			name:           "Server error (500)",
			address:        "0x1234567890123456789012345678901234567890",
			slot:           "0x0000000000000000000000000000000000000000000000000000000000000001",
			timestampTo:    "2023-12-09T12:00:00.000Z",
			mockResponse:   nil,
			expectedResult: nil,
			expectedError:  false,
			statusCode:     http.StatusInternalServerError,
		},
		{
			name:           "Invalid response structure",
			address:        "0x1234567890123456789012345678901234567890",
			slot:           "0x0000000000000000000000000000000000000000000000000000000000000001",
			timestampTo:    "2023-12-09T12:00:00.000Z",
			mockResponse:   nil,
			expectedResult: nil,
			expectedError:  true,
			statusCode:     http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/contracts/"+tc.address+"/state", r.URL.Path)
				assert.Contains(t, r.URL.RawQuery, "limit=100")
				assert.Contains(t, r.URL.RawQuery, "order=desc")
				assert.Contains(t, r.URL.RawQuery, "slot="+tc.slot)
				if tc.timestampTo != "" {
					assert.Contains(t, r.URL.RawQuery, "timestamp="+tc.timestampTo)
				}

				w.WriteHeader(tc.statusCode)
				if tc.mockResponse != nil {
					json.NewEncoder(w).Encode(tc.mockResponse)
				}
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, 5, logger)
			result, err := client.GetContractStateByAddressAndSlot(tc.address, tc.slot, tc.timestampTo)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestGetContractResultsLogsByAddress(t *testing.T) {
	logger, _ := setupTest(t)

	testCases := []struct {
		name           string
		address        string
		queryParams    map[string]interface{}
		mockResponse   interface{}
		expectedResult []domain.ContractResults
		expectError    bool
		statusCode     int
	}{
		{
			name:    "Successful logs fetch",
			address: "0x1234567890123456789012345678901234567890",
			queryParams: map[string]interface{}{
				"timestamp.gte": "1640995200.000000000",
				"timestamp.lte": "1640995300.000000000",
				"order":         "desc",
			},
			mockResponse: map[string]interface{}{
				"logs": []map[string]interface{}{
					{
						"address": "0x1234567890123456789012345678901234567890",
						"hash":    "0xtx1",
						"result":  "SUCCESS",
					},
				},
			},
			expectedResult: []domain.ContractResults{
				{
					Address: "0x1234567890123456789012345678901234567890",
					Hash:    "0xtx1",
					Result:  "SUCCESS",
				},
			},
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:    "Server error",
			address: "0x1234567890123456789012345678901234567890",
			queryParams: map[string]interface{}{
				"timestamp.gte": "1640995200.000000000",
				"order":         "desc",
			},
			mockResponse:   map[string]interface{}{"error": "Internal server error"},
			expectedResult: nil,
			expectError:    true,
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/contracts/"+tc.address+"/results/logs", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				queryValues := r.URL.Query()
				for key, value := range tc.queryParams {
					assert.Contains(t, queryValues.Get(key), value)
				}

				w.Header().Set("Content-Type", "application/json")
				if tc.statusCode != http.StatusOK {
					w.WriteHeader(tc.statusCode)
					return
				}

				w.WriteHeader(tc.statusCode)
				if tc.mockResponse != nil {
					json.NewEncoder(w).Encode(tc.mockResponse)
				}
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, 5, logger)
			results, err := client.GetContractResultsLogsByAddress(tc.address, tc.queryParams)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, results)
			}
		})
	}
}

func TestGetContractResultsLogsWithRetry(t *testing.T) {
	logger, _ := setupTest(t)

	testCases := []struct {
		name           string
		queryParams    map[string]interface{}
		mockResponse   interface{}
		expectedResult []domain.ContractResults
		expectError    bool
		statusCode     int
	}{
		{
			name: "Successful logs fetch",
			queryParams: map[string]interface{}{
				"timestamp.gte": "1640995200.000000000",
				"timestamp.lte": "1640995300.000000000",
				"order":         "desc",
			},
			mockResponse: map[string]interface{}{
				"logs": []map[string]interface{}{
					{
						"address":           "0x1234567890123456789012345678901234567890",
						"hash":              "0xtx1",
						"result":            "SUCCESS",
						"transaction_index": 1,
						"block_number":      100,
						"block_hash":        "0xblock1",
					},
				},
			},
			expectedResult: []domain.ContractResults{
				{
					Address:          "0x1234567890123456789012345678901234567890",
					Hash:             "0xtx1",
					Result:           "SUCCESS",
					TransactionIndex: 1,
					BlockNumber:      100,
					BlockHash:        "0xblock1",
				},
			},
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name: "Server error",
			queryParams: map[string]interface{}{
				"timestamp.gte": "1640995200.000000000",
				"order":         "desc",
			},
			mockResponse:   map[string]interface{}{"error": "Internal server error"},
			expectedResult: nil,
			expectError:    true,
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/contracts/results/logs", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				queryValues := r.URL.Query()
				for key, value := range tc.queryParams {
					assert.Contains(t, queryValues.Get(key), value)
				}

				w.Header().Set("Content-Type", "application/json")
				if tc.statusCode != http.StatusOK {
					w.WriteHeader(tc.statusCode)
					return
				}

				w.WriteHeader(tc.statusCode)
				if tc.mockResponse != nil {
					json.NewEncoder(w).Encode(tc.mockResponse)
				}
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, 5, logger)
			results, err := client.GetContractResultsLogsWithRetry(tc.queryParams)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, results)
			}
		})
	}
}

func TestGetAccountById(t *testing.T) {
	logger, _ := setupTest(t)

	testCases := []struct {
		name           string
		accountId      string
		mockResponse   interface{}
		expectedResult *domain.AccountResponse
		expectError    bool
		statusCode     int
	}{
		{
			name:      "Successful account fetch",
			accountId: "0.0.123",
			mockResponse: &domain.AccountResponse{
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
			},
			expectedResult: &domain.AccountResponse{
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
			},
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:           "Server error",
			accountId:      "0.0.123",
			mockResponse:   map[string]interface{}{"error": "Internal server error"},
			expectedResult: nil,
			expectError:    true,
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/accounts/"+tc.accountId+"?transactions=false", r.URL.String())
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(tc.statusCode)
				if tc.mockResponse != nil {
					json.NewEncoder(w).Encode(tc.mockResponse)
				}
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, 5, logger)
			result, err := client.GetAccountById(tc.accountId)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "mirror node returned status 500")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult.Account, result.Account)
				assert.Equal(t, tc.expectedResult.EthereumNonce, result.EthereumNonce)
				assert.Equal(t, tc.expectedResult.Balance.Balance, result.Balance.Balance)
			}
		})
	}
}

func TestGetContractById(t *testing.T) {
	logger, _ := setupTest(t)

	testCases := []struct {
		name           string
		contractId     string
		mockResponse   interface{}
		expectedResult *domain.ContractResponse
		expectError    bool
		statusCode     int
	}{
		{
			name:       "Successful contract fetch",
			contractId: "0.0.123",
			mockResponse: func() *domain.ContractResponse {
				bytecode := "0x123456"
				adminKey := "admin_key"
				return &domain.ContractResponse{
					ContractID: "0.0.123",
					AdminKey:   &adminKey,
					Bytecode:   &bytecode,
					Timestamp:  domain.Timestamp{From: "1234567890.000000000", To: "1234567890.000000000"},
					EvmAddress: "0x1234567890123456789012345678901234567890",
					Nonce:      5,
				}
			}(),
			expectedResult: func() *domain.ContractResponse {
				bytecode := "0x123456"
				adminKey := "admin_key"
				return &domain.ContractResponse{
					ContractID: "0.0.123",
					AdminKey:   &adminKey,
					Bytecode:   &bytecode,
					Timestamp:  domain.Timestamp{From: "1234567890.000000000", To: "1234567890.000000000"},
					EvmAddress: "0x1234567890123456789012345678901234567890",
					Nonce:      5,
				}
			}(),
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:           "Server error",
			contractId:     "0.0.123",
			mockResponse:   map[string]interface{}{"error": "Internal server error"},
			expectedResult: nil,
			expectError:    true,
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/contracts/"+tc.contractId, r.URL.String())
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(tc.statusCode)
				if tc.mockResponse != nil {
					json.NewEncoder(w).Encode(tc.mockResponse)
				}
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, 5, logger)
			result, err := client.GetContractById(tc.contractId)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "mirror node returned status 500")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult.ContractID, result.ContractID)
				assert.Equal(t, tc.expectedResult.AdminKey, result.AdminKey)
				assert.Equal(t, tc.expectedResult.Bytecode, result.Bytecode)
				assert.Equal(t, tc.expectedResult.EvmAddress, result.EvmAddress)
				assert.Equal(t, tc.expectedResult.Nonce, result.Nonce)
			}
		})
	}
}

func TestGetContractResultWithRetry(t *testing.T) {
	logger, _ := setupTest(t)

	testCases := []struct {
		name           string
		queryParams    map[string]interface{}
		mockResponses  []interface{}
		expectedResult *domain.ContractResults
		expectError    bool
		statusCode     int
		expectedCalls  int
	}{
		{
			name: "Successful result fetch",
			queryParams: map[string]interface{}{
				"timestamp": "1234567890",
			},
			mockResponses: []interface{}{
				struct {
					Results []domain.ContractResults `json:"results"`
					Links   struct {
						Next *string `json:"next"`
					} `json:"links"`
				}{
					Results: []domain.ContractResults{
						{
							Address:          "0x1234567890123456789012345678901234567890",
							Hash:             "0xtx1",
							Result:           "SUCCESS",
							TransactionIndex: 1,
							BlockNumber:      100,
							BlockHash:        "0xblock1",
						},
					},
				},
			},
			expectedResult: &domain.ContractResults{
				Address:          "0x1234567890123456789012345678901234567890",
				Hash:             "0xtx1",
				Result:           "SUCCESS",
				TransactionIndex: 1,
				BlockNumber:      100,
				BlockHash:        "0xblock1",
			},
			expectError:   false,
			statusCode:    http.StatusOK,
			expectedCalls: 1,
		},
		{
			name: "Immature record with retry",
			queryParams: map[string]interface{}{
				"timestamp": "1234567890",
			},
			mockResponses: []interface{}{
				struct {
					Results []domain.ContractResults `json:"results"`
					Links   struct {
						Next *string `json:"next"`
					} `json:"links"`
				}{
					Results: []domain.ContractResults{
						{
							Hash:             "0xtx1",
							Result:           "SUCCESS",
							TransactionIndex: 0,
							BlockNumber:      0,
							BlockHash:        "0x",
						},
					},
				},
				struct {
					Results []domain.ContractResults `json:"results"`
					Links   struct {
						Next *string `json:"next"`
					} `json:"links"`
				}{
					Results: []domain.ContractResults{
						{
							Hash:             "0xtx1",
							Result:           "SUCCESS",
							TransactionIndex: 0,
							BlockNumber:      0,
							BlockHash:        "0x",
						},
					},
				},
			},
			expectedResult: nil,
			expectError:    false,
			statusCode:     http.StatusOK,
			expectedCalls:  2,
		},
		{
			name: "Server error",
			queryParams: map[string]interface{}{
				"timestamp": "1234567890",
			},
			mockResponses:  []interface{}{map[string]interface{}{"error": "Internal server error"}},
			expectedResult: nil,
			expectError:    true,
			statusCode:     http.StatusInternalServerError,
			expectedCalls:  1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/contracts/results?timestamp=1234567890&order=desc", r.URL.String())
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(tc.statusCode)
				if callCount < len(tc.mockResponses) {
					json.NewEncoder(w).Encode(tc.mockResponses[callCount])
				}
				callCount++
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, 5, logger)
			result, err := client.GetContractResultWithRetry(tc.queryParams)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "mirror node returned status 500")
			} else {
				assert.NoError(t, err)
				if tc.expectedResult != nil {
					assert.Equal(t, tc.expectedResult.Hash, result.Hash)
					assert.Equal(t, tc.expectedResult.Result, result.Result)
					assert.Equal(t, tc.expectedResult.TransactionIndex, result.TransactionIndex)
					assert.Equal(t, tc.expectedResult.BlockNumber, result.BlockNumber)
					assert.Equal(t, tc.expectedResult.BlockHash, result.BlockHash)
				} else {
					assert.Nil(t, result)
				}
			}
			assert.Equal(t, tc.expectedCalls, callCount)
		})
	}
}

func TestGetTokenById(t *testing.T) {
	logger, _ := setupTest(t)

	testCases := []struct {
		name           string
		tokenId        string
		mockResponse   interface{}
		expectedResult *domain.TokenResponse
		expectError    bool
		statusCode     int
	}{
		{
			name:    "Successful token fetch",
			tokenId: "0.0.123",
			mockResponse: &domain.TokenResponse{
				TokenId:     "0.0.123",
				Name:        "Test Token",
				Symbol:      "TST",
				Decimals:    18,
				TotalSupply: 1000000,
				Type:        "FUNGIBLE_COMMON",
			},
			expectedResult: &domain.TokenResponse{
				TokenId:     "0.0.123",
				Name:        "Test Token",
				Symbol:      "TST",
				Decimals:    18,
				TotalSupply: 1000000,
				Type:        "FUNGIBLE_COMMON",
			},
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:           "Server error",
			tokenId:        "0.0.123",
			mockResponse:   map[string]interface{}{"error": "Internal server error"},
			expectedResult: nil,
			expectError:    true,
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/tokens/"+tc.tokenId, r.URL.String())
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(tc.statusCode)
				if tc.mockResponse != nil {
					json.NewEncoder(w).Encode(tc.mockResponse)
				}
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, 5, logger)
			result, err := client.GetTokenById(tc.tokenId)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), "mirror node returned status 500")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult.TokenId, result.TokenId)
				assert.Equal(t, tc.expectedResult.Name, result.Name)
				assert.Equal(t, tc.expectedResult.Symbol, result.Symbol)
				assert.Equal(t, tc.expectedResult.Decimals, result.Decimals)
				assert.Equal(t, tc.expectedResult.TotalSupply, result.TotalSupply)
				assert.Equal(t, tc.expectedResult.Type, result.Type)
			}
		})
	}
}

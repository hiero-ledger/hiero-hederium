package hedera_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/LimeChain/Hederium/internal/domain"
	"github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/LimeChain/Hederium/test/unit/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var ErrCacheMiss = errors.New("cache miss")

type testSetup struct {
	logger       *zap.Logger
	cacheService *mocks.MockCacheService
	ctrl         *gomock.Controller
}

func setupTest(t *testing.T) *testSetup {
	logger, _ := zap.NewDevelopment()
	ctrl := gomock.NewController(t)
	cacheService := mocks.NewMockCacheService(ctrl)
	return &testSetup{
		logger:       logger,
		cacheService: cacheService,
		ctrl:         ctrl,
	}
}

func TestNewMirrorClient(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()
	baseURL := "http://test.com"
	timeoutSeconds := 30

	client := hedera.NewMirrorClient(baseURL, baseURL, timeoutSeconds, setup.logger, setup.cacheService)

	assert.Equal(t, baseURL, client.BaseURL)
	assert.Equal(t, time.Duration(timeoutSeconds)*time.Second, client.Timeout)
}

func TestGetLatestBlock_Success(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

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

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	block, err := client.GetLatestBlock()

	assert.NoError(t, err)
	assert.Equal(t, float64(123), block["number"])
	assert.Equal(t, "0xabc", block["hash"])
}

func TestGetLatestBlock_EmptyResponse(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Blocks []map[string]interface{} `json:"blocks"`
		}{
			Blocks: []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	block, err := client.GetLatestBlock()

	assert.Error(t, err)
	assert.Nil(t, block)
	assert.Contains(t, err.Error(), "no blocks returned")
}

func TestGetBlockByHashOrNumber_Success(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

	expectedBlock := &domain.BlockResponse{
		Number:       123,
		Hash:         "0xabc",
		PreviousHash: "0xdef",
		GasUsed:      1000,
		Size:         2000,
		LogsBloom:    "0x0",
	}

	// Add cache expectations
	setup.cacheService.EXPECT().
		Get(gomock.Any(), "getBlockByHashOrNumber_123", gomock.Any()).
		Return(ErrCacheMiss)
	setup.cacheService.EXPECT().
		Set(gomock.Any(), "getBlockByHashOrNumber_123", expectedBlock, gomock.Any()).
		Return(nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/blocks/123", r.URL.String())
		assert.Equal(t, http.MethodGet, r.Method)

		json.NewEncoder(w).Encode(expectedBlock)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	block := client.GetBlockByHashOrNumber("123")

	assert.NotNil(t, block)
	assert.Equal(t, expectedBlock.Number, block.Number)
	assert.Equal(t, expectedBlock.Hash, block.Hash)
	assert.Equal(t, expectedBlock.PreviousHash, block.PreviousHash)
}

func TestGetBlockByHashOrNumber_ErrorResponse(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

	// Add cache expectations
	setup.cacheService.EXPECT().
		Get(gomock.Any(), "getBlockByHashOrNumber_123", gomock.Any()).
		Return(ErrCacheMiss)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	block := client.GetBlockByHashOrNumber("123")

	assert.Nil(t, block)
}

func TestGetNetworkFees_Success(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

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

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	fees, err := client.GetNetworkFees("", "")
	assert.NoError(t, err)

	assert.Equal(t, int64(100000), fees) // Should return the EthereumTransaction fee
}

func TestGetContractResults_Success(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

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

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	results := client.GetContractResults(timestamp)

	assert.Equal(t, 2, len(results))
	assert.Equal(t, expectedResults[0].Hash, results[0].Hash)
	assert.Equal(t, expectedResults[1].Hash, results[1].Hash)
}

func TestGetContractResults_ErrorResponse(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	results := client.GetContractResults(domain.Timestamp{})

	assert.Empty(t, results)
}

func TestGetNetworkFees_NoEthereumFee(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

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

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	fees, err := client.GetNetworkFees("", "") //  Should be handled better

	assert.NoError(t, err)
	assert.Equal(t, int64(0), fees) // Should return 0 when no EthereumTransaction fee is found
}

func TestGetNetworkFees_EmptyResponse(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

	response := domain.FeeResponse{
		Fees: []domain.Fee{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	_, err := client.GetNetworkFees("", "") // Should be handled better

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no fees returned")
}

func TestGetBalance(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

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

			client := hedera.NewMirrorClient(server.URL, server.URL, 5, setup.logger, setup.cacheService)
			result := client.GetBalance(tc.address, tc.timestampTo)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestGetBalance_Success(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

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

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	result := client.GetBalance("0.0.123", "1234567890.000000000")

	// 1 million tinybars * 10000000000 (conversion to weibars) = 10000000000000000 weibars
	expectedHex := "0x" + new(big.Int).Mul(big.NewInt(1000000), big.NewInt(10000000000)).Text(16)
	assert.Equal(t, expectedHex, result)
}

func TestGetBalance_Error(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

	// Create test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	result := client.GetBalance("0.0.123", "1234567890.000000000")

	assert.Equal(t, "0x0", result)
}

func TestGetAccount_Success(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

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

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	result := client.GetAccount("0.0.123", "1234567890.000000000")

	assert.NotNil(t, result)
	accountResponse := result.(domain.AccountResponse)
	assert.Equal(t, "0.0.123", accountResponse.Account)
	assert.Equal(t, int64(5), accountResponse.EthereumNonce)
	assert.Equal(t, int64(1000000), accountResponse.Balance.Balance)
}

func TestGetAccount_Error(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

	// Create test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := hedera.NewMirrorClient(server.URL, server.URL, 30, setup.logger, setup.cacheService)
	result := client.GetAccount("0.0.123", "1234567890.000000000")

	assert.Nil(t, result)
}

func TestPostCall(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

	testCases := []struct {
		name           string
		callObject     map[string]interface{}
		mockResponse   interface{}
		expectedResult string
		expectError    bool
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
			expectError:    false,
			statusCode:     http.StatusOK,
		},
		{
			name: "Server error",
			callObject: map[string]interface{}{
				"data": "0x123456",
			},
			mockResponse:   nil,
			expectedResult: "",
			expectError:    true,
			statusCode:     http.StatusInternalServerError,
		},
		{
			name: "Invalid response structure",
			callObject: map[string]interface{}{
				"data": "0x123456",
			},
			mockResponse:   "invalid json",
			expectedResult: "",
			expectError:    true,
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

			client := hedera.NewMirrorClient(server.URL, server.URL, 5, setup.logger, setup.cacheService)
			result, err := client.PostCall(tc.callObject)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestGetContractStateByAddressAndSlot(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

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

			client := hedera.NewMirrorClient(server.URL, server.URL, 5, setup.logger, setup.cacheService)
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
	setup := setupTest(t)
	defer setup.ctrl.Finish()

	testCases := []struct {
		name           string
		address        string
		queryParams    map[string]interface{}
		mockResponse   interface{}
		expectedResult []domain.LogEntry
		expectEmpty    bool
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
			expectedResult: []domain.LogEntry{
				{
					Address: "0x1234567890123456789012345678901234567890",
				},
			},
			expectEmpty: false,
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
			expectEmpty:    true,
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

				w.WriteHeader(tc.statusCode)
				if tc.mockResponse != nil {
					json.NewEncoder(w).Encode(tc.mockResponse)
				}
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, server.URL, 5, setup.logger, setup.cacheService)
			results, err := client.GetContractResultsLogsByAddress(tc.address, tc.queryParams)

			assert.NoError(t, err)

			if tc.expectEmpty {
				assert.Empty(t, results)
			} else {
				assert.Equal(t, tc.expectedResult, results)
			}
		})
	}
}

func TestGetContractResultWithRetry(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

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
			name:        "Success on first try",
			queryParams: map[string]interface{}{"timestamp": "1234567890"},
			mockResponses: []interface{}{
				// For the GetContractResults endpoint, we need to use the structure it expects
				struct {
					Results []domain.ContractResults `json:"results"`
					Links   struct {
						Next *string `json:"next"`
					} `json:"links"`
				}{
					Results: []domain.ContractResults{
						{
							Hash:             "0xtx1",
							BlockHash:        "0xblock123",
							BlockNumber:      100,
							Result:           "SUCCESS",
							TransactionIndex: 1,
						},
					},
				},
			},
			expectedResult: &domain.ContractResults{
				Hash:             "0xtx1",
				BlockHash:        "0xblock123",
				BlockNumber:      100,
				Result:           "SUCCESS",
				TransactionIndex: 1,
			},
			expectError:   false,
			statusCode:    http.StatusOK,
			expectedCalls: 1,
		},
		{
			name:        "Success after retries",
			queryParams: map[string]interface{}{"timestamp": "1234567890"},
			mockResponses: []interface{}{
				// First response has immature records (BlockHash is "0x")
				struct {
					Results []domain.ContractResults `json:"results"`
					Links   struct {
						Next *string `json:"next"`
					} `json:"links"`
				}{
					Results: []domain.ContractResults{
						{
							Hash:             "0xtx2",
							BlockHash:        "0x", // Immature
							BlockNumber:      0,    // Immature
							Result:           "SUCCESS",
							TransactionIndex: 0, // Immature
						},
					},
				},
				// Second response has a mature record
				struct {
					Results []domain.ContractResults `json:"results"`
					Links   struct {
						Next *string `json:"next"`
					} `json:"links"`
				}{
					Results: []domain.ContractResults{
						{
							Hash:             "0xtx2",
							BlockHash:        "0xblock456",
							BlockNumber:      200,
							Result:           "SUCCESS",
							TransactionIndex: 2,
						},
					},
				},
			},
			expectedResult: &domain.ContractResults{
				Hash:             "0xtx2",
				BlockHash:        "0xblock456",
				BlockNumber:      200,
				Result:           "SUCCESS",
				TransactionIndex: 2,
			},
			expectError:   false,
			statusCode:    http.StatusOK,
			expectedCalls: 2,
		},
		{
			name:        "Empty results",
			queryParams: map[string]interface{}{"timestamp": "1234567890"},
			mockResponses: []interface{}{
				struct {
					Results []domain.ContractResults `json:"results"`
					Links   struct {
						Next *string `json:"next"`
					} `json:"links"`
				}{
					Results: []domain.ContractResults{},
				},
			},
			expectedResult: nil,
			expectError:    false, // Empty results returns nil, nil
			statusCode:     http.StatusOK,
			expectedCalls:  1,
		},
		{
			name:        "Fail after all retries with immature records",
			queryParams: map[string]interface{}{"timestamp": "1234567890"},
			mockResponses: []interface{}{
				// First response with immature record
				struct {
					Results []domain.ContractResults `json:"results"`
					Links   struct {
						Next *string `json:"next"`
					} `json:"links"`
				}{
					Results: []domain.ContractResults{
						{
							Hash:             "0xtx3",
							BlockHash:        "0x", // Immature
							BlockNumber:      0,
							Result:           "SUCCESS",
							TransactionIndex: 0,
						},
					},
				},
				// Second response still immature - package will only retry once by default
				struct {
					Results []domain.ContractResults `json:"results"`
					Links   struct {
						Next *string `json:"next"`
					} `json:"links"`
				}{
					Results: []domain.ContractResults{
						{
							Hash:             "0xtx3",
							BlockHash:        "0x", // Still immature
							BlockNumber:      0,
							Result:           "SUCCESS",
							TransactionIndex: 0,
						},
					},
				},
			},
			expectedResult: nil,
			expectError:    false, // Returns nil, nil after max retries
			statusCode:     http.StatusOK,
			expectedCalls:  2, // Default maxRetries is 2
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedQuery := fmt.Sprintf("/api/v1/contracts/results?%s&order=desc", strings.Join(mapToQueryParams(tc.queryParams), "&"))
				assert.Equal(t, expectedQuery, r.URL.String())
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(tc.statusCode)
				if callCount < len(tc.mockResponses) {
					json.NewEncoder(w).Encode(tc.mockResponses[callCount])
				}
				callCount++
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, server.URL, 5, setup.logger, setup.cacheService)

			result, err := client.GetContractResultWithRetry(tc.queryParams)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tc.expectedResult != nil {
					assert.NotNil(t, result)
					assert.Equal(t, tc.expectedResult.BlockHash, result.BlockHash)
					assert.Equal(t, tc.expectedResult.BlockNumber, result.BlockNumber)
					assert.Equal(t, tc.expectedResult.Result, result.Result)
					assert.Equal(t, tc.expectedResult.TransactionIndex, result.TransactionIndex)
					assert.Equal(t, tc.expectedResult.Hash, result.Hash)
				} else {
					assert.Nil(t, result)
				}
			}
			assert.Equal(t, tc.expectedCalls, callCount, "Expected %d calls but got %d", tc.expectedCalls, callCount)
		})
	}
}

// Helper to convert map to query params
func mapToQueryParams(params map[string]interface{}) []string {
	queryParams := []string{}
	for key, value := range params {
		queryParams = append(queryParams, fmt.Sprintf("%s=%v", key, value))
	}
	sort.Strings(queryParams) // Sort for consistent order in tests
	return queryParams
}

func TestGetAccountById(t *testing.T) {
	setup := setupTest(t)
	defer setup.ctrl.Finish()

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
				},
				EthereumNonce: 5,
				EvmAddress:    "0x1234567890123456789012345678901234567890",
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
				},
				EthereumNonce: 5,
				EvmAddress:    "0x1234567890123456789012345678901234567890",
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
			// Add cache expectations
			setup.cacheService.EXPECT().
				Get(gomock.Any(), "getAccountById_"+tc.accountId, gomock.Any()).
				Return(ErrCacheMiss)

			if !tc.expectError {
				setup.cacheService.EXPECT().
					Set(gomock.Any(), "getAccountById_"+tc.accountId, tc.mockResponse, gomock.Any()).
					Return(nil)
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/accounts/"+tc.accountId+"?transactions=false", r.URL.String())
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(tc.statusCode)
				if tc.mockResponse != nil {
					json.NewEncoder(w).Encode(tc.mockResponse)
				}
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, server.URL, 5, setup.logger, setup.cacheService)
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
	setup := setupTest(t)
	defer setup.ctrl.Finish()

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
				return &domain.ContractResponse{
					ContractID: "0.0.123",
					Bytecode:   &bytecode,
					Timestamp:  domain.Timestamp{From: "1234567890.000000000", To: "1234567890.000000000"},
					EvmAddress: "0x1234567890123456789012345678901234567890",
					Nonce:      5,
				}
			}(),
			expectedResult: func() *domain.ContractResponse {
				bytecode := "0x123456"
				return &domain.ContractResponse{
					ContractID: "0.0.123",
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
			mockResponse:   nil,
			expectedResult: nil,
			expectError:    true,
			statusCode:     http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Add cache expectations
			setup.cacheService.EXPECT().
				Get(gomock.Any(), "getContractById_"+tc.contractId, gomock.Any()).
				Return(ErrCacheMiss)

			if !tc.expectError {
				setup.cacheService.EXPECT().
					Set(gomock.Any(), "getContractById_"+tc.contractId, tc.mockResponse, gomock.Any()).
					Return(nil)
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/contracts/"+tc.contractId, r.URL.String())
				assert.Equal(t, http.MethodGet, r.Method)

				w.WriteHeader(tc.statusCode)
				if tc.mockResponse != nil {
					json.NewEncoder(w).Encode(tc.mockResponse)
				}
			}))
			defer server.Close()

			client := hedera.NewMirrorClient(server.URL, server.URL, 5, setup.logger, setup.cacheService)
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

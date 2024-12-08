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

func TestGetLatestBlock(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/blocks?order=desc&limit=1", r.URL.String())
		assert.Equal(t, http.MethodGet, r.Method)

		response := map[string]interface{}{
			"blocks": []map[string]interface{}{
				{
					"number": 12345,
					"hash":   "0x123",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := hedera.NewMirrorClient(server.URL, 5, logger)

	block, err := client.GetLatestBlock()
	assert.NoError(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, float64(12345), block["number"])
	assert.Equal(t, "0x123", block["hash"])
}

func TestGetLatestBlock_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := hedera.NewMirrorClient(server.URL, 5, logger)

	block, err := client.GetLatestBlock()
	assert.Error(t, err)
	assert.Nil(t, block)
}

func TestGetLatestBlock_Timeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := hedera.NewMirrorClient(server.URL, 1, logger) // 1 second timeout

	block, err := client.GetLatestBlock()
	assert.Error(t, err)
	assert.Nil(t, block)
}

func TestGetNetworkFees(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/network/fees", r.URL.String())
		assert.Equal(t, http.MethodGet, r.Method)

		response := domain.FeeResponse{
			Fees: []domain.Fee{
				{
					TransactionType: "EthereumTransaction",
					Gas:             100,
				},
				{
					TransactionType: "OtherTransaction",
					Gas:             200,
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := hedera.NewMirrorClient(server.URL, 5, logger)

	fees, err := client.GetNetworkFees()
	assert.NoError(t, err)
	assert.Equal(t, int64(100), fees) // Should return gas for EthereumTransaction
}

func TestGetNetworkFees_NoEthereumFees(t *testing.T) {
	// Create a test server that returns fees but no Ethereum fees
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := domain.FeeResponse{
			Fees: []domain.Fee{
				{
					TransactionType: "OtherTransaction",
					Gas:             200,
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := hedera.NewMirrorClient(server.URL, 5, logger)

	fees, err := client.GetNetworkFees()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), fees) // Should return 0 when no Ethereum fees found
}

func TestGetNetworkFees_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := hedera.NewMirrorClient(server.URL, 5, logger)

	fees, err := client.GetNetworkFees()
	assert.Error(t, err)
	assert.Equal(t, int64(0), fees)
}

package service_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hiero-ledger/hiero-hederium/internal/domain"
	"github.com/hiero-ledger/hiero-hederium/internal/service"
	"github.com/hiero-ledger/hiero-hederium/test/unit/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// These tests cover the checks that replaced unchecked type assertions: an
// upstream answer of an unexpected shape is an internal error, not a panic.

func newGuardedEthService(t *testing.T) (*service.EthService, *mocks.MockMirrorClient, *mocks.MockCacheService, *mocks.MockCommonService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockClient := mocks.NewMockMirrorClient(ctrl)
	cacheService := mocks.NewMockCacheService(ctrl)
	commonService := mocks.NewMockCommonService(ctrl)
	s := service.NewEthService(nil, mockClient, commonService, zap.NewNop(), nil, defaultChainId, cacheService)
	return s, mockClient, cacheService, commonService
}

// The process crash in the gap report (E-1): the latest block could not be
// read, and GetBalance asserted a field of the nil map it got back.
func TestGetBalance_LatestBlockUnavailable(t *testing.T) {
	testCases := []struct {
		name        string
		latestBlock map[string]interface{}
		latestErr   error
	}{
		{name: "mirror node error", latestBlock: nil, latestErr: fmt.Errorf("mirror node returned status 500")},
		{name: "block without a number", latestBlock: map[string]interface{}{"hash": "0x1"}, latestErr: nil},
		{name: "number of the wrong type", latestBlock: map[string]interface{}{"number": "200"}, latestErr: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, mockClient, _, _ := newGuardedEthService(t)
			mockClient.EXPECT().GetBlockByHashOrNumber("100").Return(&domain.BlockResponse{
				Number:    100,
				Timestamp: domain.Timestamp{To: "1234567890.000000000"},
			})
			mockClient.EXPECT().GetLatestBlock().Return(tc.latestBlock, tc.latestErr)

			var result string
			var errRpc *domain.RPCError
			require.NotPanics(t, func() { result, errRpc = s.GetBalance("0x123", "100") })

			require.NotNil(t, errRpc)
			assert.Equal(t, domain.InternalError, errRpc.Code)
			assert.Equal(t, "Failed to get latest block", errRpc.Message)
			assert.Empty(t, result, "no balance is reported when it cannot be determined")
		})
	}
}

func TestEthService_UnexpectedMirrorResponseTypes(t *testing.T) {
	hash := "0x" + strings.Repeat("ab", 32)

	t.Run("eth_getTransactionByHash", func(t *testing.T) {
		s, mockClient, cacheService, _ := newGuardedEthService(t)
		cacheService.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("cache miss"))
		mockClient.EXPECT().GetContractResult(hash).Return("not a contract result")

		var result interface{}
		var errRpc *domain.RPCError
		require.NotPanics(t, func() { result, errRpc = s.GetTransactionByHash(hash) })
		assert.Nil(t, result)
		require.NotNil(t, errRpc)
		assert.Equal(t, domain.InternalError, errRpc.Code)
	})

	t.Run("eth_getTransactionReceipt", func(t *testing.T) {
		s, mockClient, cacheService, _ := newGuardedEthService(t)
		cacheService.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("cache miss"))
		mockClient.EXPECT().GetContractResult(hash).Return("not a contract result")

		var result interface{}
		var errRpc *domain.RPCError
		require.NotPanics(t, func() { result, errRpc = s.GetTransactionReceipt(hash) })
		assert.Nil(t, result)
		require.NotNil(t, errRpc)
		assert.Equal(t, domain.InternalError, errRpc.Code)
	})

	t.Run("eth_estimateGas", func(t *testing.T) {
		s, mockClient, _, _ := newGuardedEthService(t)
		mockClient.EXPECT().PostCall(gomock.Any()).Return(map[string]interface{}{"result": "0x5208"})

		var errRpc *domain.RPCError
		require.NotPanics(t, func() {
			_, errRpc = s.EstimateGas(map[string]interface{}{"to": "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"}, "latest")
		})
		require.NotNil(t, errRpc)
		assert.Equal(t, domain.InternalError, errRpc.Code)
	})

	// eth_getTransactionCount still reports "0x0" on every upstream failure;
	// returning errors there instead is item 1.9 (B-42).
	t.Run("eth_getTransactionCount account", func(t *testing.T) {
		s, mockClient, _, commonService := newGuardedEthService(t)
		commonService.EXPECT().GetBlockNumberByNumberOrTag("earliest").Return(int64(0), nil)
		mockClient.EXPECT().GetBlockByHashOrNumber("0").Return(&domain.BlockResponse{Timestamp: domain.Timestamp{To: "1.0"}})
		mockClient.EXPECT().GetAccount("0x123", "1.0").Return("not an account")

		var result string
		require.NotPanics(t, func() { result = s.GetTransactionCount("0x123", "earliest") })
		assert.Equal(t, "0x0", result)
	})

	t.Run("eth_getTransactionCount contract result", func(t *testing.T) {
		s, mockClient, _, commonService := newGuardedEthService(t)
		commonService.EXPECT().GetBlockNumberByNumberOrTag("0x64").Return(int64(100), nil)
		commonService.EXPECT().GetBlockNumberByNumberOrTag("latest").Return(int64(500), nil)
		mockClient.EXPECT().GetBlockByHashOrNumber("100").Return(&domain.BlockResponse{Timestamp: domain.Timestamp{To: "1.0"}})
		var account domain.AccountResponse
		require.NoError(t, json.Unmarshal([]byte(`{"transactions":[{"transaction_id":"0.0.2-1-1"}]}`), &account))
		mockClient.EXPECT().GetAccount("0x123", "1.0").Return(account)
		mockClient.EXPECT().GetContractResult("0.0.2-1-1").Return("not a contract result")

		var result string
		require.NotPanics(t, func() { result = s.GetTransactionCount("0x123", "0x64") })
		assert.Equal(t, "0x0", result)
	})
}

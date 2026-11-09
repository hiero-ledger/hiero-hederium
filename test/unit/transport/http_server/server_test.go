package http_server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hiero-ledger/hiero-hederium/internal/domain"
	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/cache"
	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/hedera"
	"github.com/hiero-ledger/hiero-hederium/internal/infrastructure/limiter"
	"github.com/hiero-ledger/hiero-hederium/internal/transport/http_server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const chainID = "0x12a"

// receiptHash is answered by the stub mirror node with a contract result whose
// block_hash is shorter than 66 characters. GetTransactionReceipt slices
// BlockHash[:66], so this request panics inside the service (B-25).
var receiptHash = "0x" + strings.Repeat("ab", 32)

type rpcResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	Result  interface{}      `json:"result"`
	Error   *domain.RPCError `json:"error"`
	ID      interface{}      `json:"id"`
}

// newStubMirror answers the given paths (with their query) and returns 404 for
// everything else.
func newStubMirror(t *testing.T, routes map[string]func(http.ResponseWriter)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path
		if r.URL.RawQuery != "" {
			key += "?" + r.URL.RawQuery
		}
		if handle, ok := routes[key]; ok {
			handle(w)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"_status":{"messages":[{"message":"Not found"}]}}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func jsonBody(body string) func(http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}
}

func shortBlockHashMirror(t *testing.T) *httptest.Server {
	return newStubMirror(t, map[string]func(http.ResponseWriter){
		"/api/v1/contracts/results/" + receiptHash: jsonBody(`{"block_hash":"0x12","block_number":1,"hash":"` + receiptHash + `"}`),
	})
}

func newTestServer(t *testing.T, mirrorURL string) http.Handler {
	t.Helper()
	return newTestServerWithKeys(t, mirrorURL, limiter.NewAPIKeyStore(nil), false)
}

func newTestServerWithKeys(t *testing.T, mirrorURL string, keys *limiter.APIKeyStore, enforceAPIKey bool) http.Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	cacheService := cache.NewMemoryCache(time.Minute, time.Minute)
	mClient := hedera.NewMirrorClient(mirrorURL, 5, logger, cacheService)
	srv := http_server.NewServer(
		nil, // no consensus node client: none of these tests submits transactions
		mClient,
		logger,
		"test",
		chainID,
		keys,
		limiter.NewTieredLimiter(nil, 0),
		enforceAPIKey,
		true,
		cacheService,
		"0",
	)
	return srv.Handler()
}

func post(t *testing.T, handler http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// Before the fix, the panicking batch item ran in a worker goroutine with no
// recover, so this test took the whole test binary down (E-1).
func TestBatch_PanicInOneItemOnlyFailsThatItem(t *testing.T) {
	handler := newTestServer(t, shortBlockHashMirror(t).URL)

	rec := post(t, handler, `[
		{"jsonrpc":"2.0","id":1,"method":"eth_chainId","params":[]},
		{"jsonrpc":"2.0","id":2,"method":"eth_getTransactionReceipt","params":["`+receiptHash+`"]},
		{"jsonrpc":"2.0","id":3,"method":"web3_clientVersion","params":[]}
	]`)

	require.Equal(t, http.StatusOK, rec.Code)
	var responses []rpcResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &responses), rec.Body.String())
	require.Len(t, responses, 3)

	assert.Equal(t, float64(1), responses[0].ID)
	assert.Nil(t, responses[0].Error)
	assert.Equal(t, chainID, responses[0].Result)

	assert.Equal(t, float64(2), responses[1].ID)
	require.NotNil(t, responses[1].Error)
	assert.Equal(t, domain.InternalError, responses[1].Error.Code)
	assert.Equal(t, "Internal error", responses[1].Error.Message)

	assert.Equal(t, float64(3), responses[2].ID)
	assert.Nil(t, responses[2].Error)
	assert.Equal(t, "relay/test", responses[2].Result)
}

// A panic in a single request used to reach gin's default recovery, which
// answers HTTP 500 with an empty body.
func TestSingleRequest_PanicReturnsJSONRPCInternalError(t *testing.T) {
	handler := newTestServer(t, shortBlockHashMirror(t).URL)

	rec := post(t, handler, `{"jsonrpc":"2.0","id":"r1","method":"eth_getTransactionReceipt","params":["`+receiptHash+`"]}`)

	// Every error reply is HTTP 400 today; matching the relay's status codes is item 1.4.
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp rpcResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp), rec.Body.String())
	assert.Equal(t, "r1", resp.ID)
	require.NotNil(t, resp.Error)
	assert.Equal(t, domain.InternalError, resp.Error.Code)
	assert.Equal(t, "Internal error", resp.Error.Message)
}

// The server's router uses JSONRPCRecovery: a nil API-key store makes the
// API-key middleware panic before the JSON-RPC handler runs.
func TestServer_PanicInMiddlewareReturnsJSONRPCInternalError(t *testing.T) {
	handler := newTestServerWithKeys(t, newStubMirror(t, nil).URL, nil, true)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"eth_chainId","params":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", "any")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp rpcResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp), rec.Body.String())
	require.NotNil(t, resp.Error)
	assert.Equal(t, domain.InternalError, resp.Error.Code)
}

// The router's own recovery covers panics outside the JSON-RPC handler (e.g. in
// middleware or while writing the response) with a JSON-RPC body instead of
// gin's empty HTTP 500.
func TestJSONRPCRecovery_PanicOutsideHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(http_server.JSONRPCRecovery(zap.NewNop()))
	router.POST("/", func(*gin.Context) { panic("boom") })

	rec := post(t, router, `{"jsonrpc":"2.0","id":1,"method":"eth_chainId","params":[]}`)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp rpcResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp), rec.Body.String())
	assert.Equal(t, "2.0", resp.JSONRPC)
	require.NotNil(t, resp.Error)
	assert.Equal(t, domain.InternalError, resp.Error.Code)
	assert.Equal(t, "Internal error", resp.Error.Message)
}

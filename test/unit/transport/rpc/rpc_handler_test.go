package rpc_test

import (
	"context"
	"testing"

	"github.com/hiero-ledger/hiero-hederium/internal/domain"
	"github.com/hiero-ledger/hiero-hederium/internal/service"
	"github.com/hiero-ledger/hiero-hederium/internal/transport/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// panickingNetService panics on every call, standing in for any method that
// hits a programming error (nil map, bad slice, failed type assertion).
type panickingNetService struct{}

func (panickingNetService) Listening() bool { panic("boom: listening") }
func (panickingNetService) Version() string { panic("boom: version") }

type stubWeb3Service struct{}

func (stubWeb3Service) ClientVersion() string { return "relay/test" }

type fakeServiceProvider struct {
	net service.NetServicer
}

func (f fakeServiceProvider) EthService() *service.EthService       { return nil }
func (f fakeServiceProvider) Web3Service() service.Web3Servicer     { return stubWeb3Service{} }
func (f fakeServiceProvider) NetService() service.NetServicer       { return f.net }
func (f fakeServiceProvider) FilterService() service.FilterServicer { return nil }

func newHandler() rpc.RPCHandler {
	return rpc.NewHandler(zap.NewNop(), fakeServiceProvider{net: panickingNetService{}})
}

func TestHandleRequest_PanicBecomesInternalError(t *testing.T) {
	handler := newHandler()

	var resp *rpc.JSONRPCResponse
	require.NotPanics(t, func() {
		resp = handler.HandleRequest(context.Background(), &rpc.JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "net_version",
			Params:  []interface{}{},
			ID:      float64(7),
		})
	})

	require.NotNil(t, resp)
	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, float64(7), resp.ID, "the response keeps the request id")
	assert.Nil(t, resp.Result)
	require.NotNil(t, resp.Error)
	assert.Equal(t, domain.InternalError, resp.Error.Code)
	assert.Equal(t, "Internal error", resp.Error.Message, "the panic value is not exposed to the client")
}

func TestHandleRequest_NilRequestIsInvalidRequest(t *testing.T) {
	handler := newHandler()

	var resp *rpc.JSONRPCResponse
	require.NotPanics(t, func() {
		resp = handler.HandleRequest(context.Background(), nil)
	})

	require.NotNil(t, resp)
	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Nil(t, resp.ID)
	assert.Nil(t, resp.Result)
	require.NotNil(t, resp.Error)
	assert.Equal(t, domain.InvalidRequest, resp.Error.Code)
}

// The recovery is registered before the first log line, so a panic while
// logging the call is answered like a panic in the method.
func TestHandleRequest_PanicWhileLoggingTheCallBecomesInternalError(t *testing.T) {
	logger := zap.NewNop().WithOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return panicOnMessageCore{message: "JSON-RPC method called"}
	}))
	handler := rpc.NewHandler(logger, fakeServiceProvider{net: panickingNetService{}})

	var resp *rpc.JSONRPCResponse
	require.NotPanics(t, func() {
		resp = handler.HandleRequest(context.Background(), &rpc.JSONRPCRequest{
			JSONRPC: "2.0", Method: "web3_clientVersion", Params: []interface{}{}, ID: "c",
		})
	})

	require.NotNil(t, resp)
	assert.Equal(t, "c", resp.ID)
	require.NotNil(t, resp.Error)
	assert.Equal(t, domain.InternalError, resp.Error.Code)
}

func TestNewHandler_NilLoggerStillRecoversPanics(t *testing.T) {
	handler := rpc.NewHandler(nil, fakeServiceProvider{net: panickingNetService{}})

	var panicked, ok *rpc.JSONRPCResponse
	require.NotPanics(t, func() {
		panicked = handler.HandleRequest(context.Background(), &rpc.JSONRPCRequest{
			JSONRPC: "2.0", Method: "net_version", Params: []interface{}{}, ID: 1,
		})
		ok = handler.HandleRequest(context.Background(), &rpc.JSONRPCRequest{
			JSONRPC: "2.0", Method: "web3_clientVersion", Params: []interface{}{}, ID: 2,
		})
	})

	require.NotNil(t, panicked.Error)
	assert.Equal(t, domain.InternalError, panicked.Error.Code)
	require.Nil(t, ok.Error)
	assert.Equal(t, "relay/test", ok.Result)
}

// panicOnMessageCore is a zap core that panics when it writes the given
// message and drops every other entry.
type panicOnMessageCore struct{ message string }

func (c panicOnMessageCore) Enabled(zapcore.Level) bool        { return true }
func (c panicOnMessageCore) With([]zapcore.Field) zapcore.Core { return c }
func (c panicOnMessageCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(e, c)
}
func (c panicOnMessageCore) Write(e zapcore.Entry, _ []zapcore.Field) error {
	if e.Message == c.message {
		panic("boom: logging " + e.Message)
	}
	return nil
}
func (c panicOnMessageCore) Sync() error { return nil }

func TestHandleRequest_OtherMethodsUnaffectedByAPanic(t *testing.T) {
	handler := newHandler()

	_ = handler.HandleRequest(context.Background(), &rpc.JSONRPCRequest{
		JSONRPC: "2.0", Method: "net_listening", Params: []interface{}{}, ID: "a",
	})
	resp := handler.HandleRequest(context.Background(), &rpc.JSONRPCRequest{
		JSONRPC: "2.0", Method: "web3_clientVersion", Params: []interface{}{}, ID: "b",
	})

	require.Nil(t, resp.Error)
	assert.Equal(t, "relay/test", resp.Result)
	assert.Equal(t, "b", resp.ID)
}

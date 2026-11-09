package http_server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hiero-ledger/hiero-hederium/internal/transport/rpc"
)

// FuzzRPCRequest sends arbitrary bodies (single requests, batches, malformed
// JSON) through the real router. The stub mirror node answers 404 except for a
// contract result that makes eth_getTransactionReceipt panic, and there is no
// consensus node client, so send paths fail too. The properties are Gate 0's:
//   - the process survives (a panic in a batch worker would end the fuzzer);
//   - every answer is a JSON object or array, and never a transport-level 500;
//   - a batch of N > 1 requests gets HTTP 200 and exactly N reply objects, so a
//     failure in one item cannot turn the batch into a single error.
//
// The server and its cache are shared across inputs, as in production, so an
// input may depend on state (e.g. filters) left by earlier ones.
//
// `go test` runs the seeds below; CI also fuzzes for a fixed time with
// `go test -run '^$' -fuzz '^FuzzRPCRequest$' -fuzztime 60s -fuzzminimizetime 1s`
// (minimizing each new interesting input for the default 60s stalls the run).
func FuzzRPCRequest(f *testing.F) {
	seeds := []string{
		`{"jsonrpc":"2.0","id":1,"method":"eth_chainId","params":[]}`,
		`{"jsonrpc":"2.0","id":1,"method":"eth_chainId"}`,
		`{"jsonrpc":"2.0","id":"a","method":"eth_getBalance","params":["0x1111111111111111111111111111111111111111","0x1"]}`,
		`{"jsonrpc":"2.0","id":2,"method":"eth_getTransactionReceipt","params":["` + receiptHash + `"]}`,
		`{"jsonrpc":"2.0","id":3,"method":"eth_getTransactionByHash","params":["` + receiptHash + `"]}`,
		`{"jsonrpc":"2.0","id":4,"method":"eth_getLogs","params":[{"fromBlock":"0x1","toBlock":"0x2","topics":["0x1",null]}]}`,
		`{"jsonrpc":"2.0","id":5,"method":"eth_newFilter","params":[{"address":"0x1111111111111111111111111111111111111111"}]}`,
		`{"jsonrpc":"2.0","id":6,"method":"eth_getFilterChanges","params":["0x1"]}`,
		`{"jsonrpc":"2.0","id":7,"method":"eth_feeHistory","params":["0x2","latest",[25,75]]}`,
		`{"jsonrpc":"2.0","id":8,"method":"eth_call","params":[{"to":"0x1111111111111111111111111111111111111111","data":"0x"},"latest"]}`,
		`{"jsonrpc":"2.0","id":9,"method":"eth_estimateGas","params":[{"to":"0x1111111111111111111111111111111111111111"}]}`,
		`{"jsonrpc":"2.0","id":10,"method":"eth_sendRawTransaction","params":["0xf86c"]}`,
		`{"jsonrpc":"2.0","id":11,"method":"eth_getCode","params":["0x1111111111111111111111111111111111111111","latest"]}`,
		`{"jsonrpc":"2.0","id":12,"method":"eth_getTransactionCount","params":["0x1111111111111111111111111111111111111111","0x1"]}`,
		`[{"jsonrpc":"2.0","id":1,"method":"eth_chainId","params":[]},{"jsonrpc":"2.0","id":2,"method":"eth_getTransactionReceipt","params":["` + receiptHash + `"]},{"jsonrpc":"2.0","id":3,"method":"net_version","params":[]}]`,
		`[{"jsonrpc":"2.0","id":1,"method":"eth_getBlockByNumber","params":["0x1",true]},{"jsonrpc":"2.0","id":2,"method":"eth_getStorageAt","params":["0x1111111111111111111111111111111111111111","0x0","latest"]}]`,
		`[]`,
		`[1,"x",null]`,
		`{"jsonrpc":"2.0","id":1,"method":"eth_getBalance","params":{"address":"0x1"}}`,
		`{}`,
		`null`,
		`not json`,
		``,
	}
	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	handler := newTestServer(f, shortBlockHashMirror(f).URL)

	f.Fuzz(func(t *testing.T, body []byte) {
		req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		out := bytes.TrimSpace(rec.Body.Bytes())
		if !json.Valid(out) || len(out) == 0 || (out[0] != '{' && out[0] != '[') {
			t.Fatalf("status %d, body is not a JSON object or array: %q (request %q)", rec.Code, out, body)
		}
		if rec.Code == http.StatusInternalServerError {
			t.Fatalf("transport-level 500 (a panic outside the per-request recover): %q (request %q)", out, body)
		}

		// The server treats a body as a batch exactly when it decodes into a
		// slice of requests.
		var batch []rpc.JSONRPCRequest
		if json.Unmarshal(body, &batch) != nil || len(batch) < 2 {
			return
		}
		var replies []json.RawMessage
		if rec.Code != http.StatusOK || json.Unmarshal(out, &replies) != nil || len(replies) != len(batch) {
			t.Fatalf("batch of %d: status %d, want 200 and %d replies: %q (request %q)", len(batch), rec.Code, len(batch), out, body)
		}
		for i, reply := range replies {
			if trimmed := bytes.TrimSpace(reply); len(trimmed) == 0 || trimmed[0] != '{' {
				t.Fatalf("batch reply %d is not an object: %q (request %q)", i, reply, body)
			}
		}
	})
}

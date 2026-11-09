package http_server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// FuzzRPCRequest sends arbitrary bodies (single requests, batches, malformed
// JSON) through the real router. The stub mirror node answers 404 except for a
// contract result that makes eth_getTransactionReceipt panic, and there is no
// consensus node client, so send paths fail too. The property is Gate 0's: the
// process survives (a panic in a batch worker would end the fuzzer) and every
// answer is a JSON object or array.
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
	})
}

import http from "k6/http";
import { check } from "k6";
import { config, validateJsonRpcResponse } from "../common.js";

export function getBlockTransactionCountByHash() {
  const validBlockHash = "0x52c41586e9e9e517a9c2c0a5a36925299e398e0f69bbf2c1b740a97456788d06";
  
  const payload = JSON.stringify({
    jsonrpc: "2.0",
    method: "eth_getBlockTransactionCountByHash",
    params: [validBlockHash],
    id: 1,
  });

  const headers = { "Content-Type": "application/json" };
  if (config.apiKey) {
    headers["X-API-KEY"] = config.apiKey;
  }

  const res = http.post(config.endpoint, payload, { headers: headers });

  // Validate common JSON-RPC structure
  const passed = validateJsonRpcResponse(res, "eth_getBlockTransactionCountByHash");

  // Additional checks specific to block transaction count response
  check(res, {
    "result is a hex string": (r) => {
      const body = r.json();
      return typeof body.result === "string" && body.result.startsWith("0x");
    },
    "result is a valid hex number": (r) => {
      const body = r.json();
      const hexPattern = /^0x[0-9a-fA-F]+$/;
      return hexPattern.test(body.result);
    }
  });
}

// For running the test script directly
export default function () {
  // Test cases
  getBlockTransactionCountByHash();
  runErrorTest("0x0000000000000000000000000000000000000000000000000000000000000000", "non-existent block");
  runErrorTest("invalid_hash", "invalid hash format");
}

function runErrorTest(blockHash, testCase) {
  const payload = JSON.stringify({
    jsonrpc: "2.0",
    method: "eth_getBlockTransactionCountByHash",
    params: [blockHash],
    id: 1,
  });

  const headers = { "Content-Type": "application/json" };
  if (config.apiKey) {
    headers["X-API-KEY"] = config.apiKey;
  }

  const res = http.post(config.endpoint, payload, { headers: headers });

  switch (testCase) {
    case "non-existent block":
      check(res, {
        "non-existent block: result is null": (r) => {
          const body = r.json();
          return body.result === null;
        }
      });
      break;

    case "invalid hash format":
      check(res, {
        "invalid hash: has error field": (r) => {
          const body = r.json();
          return body.error !== undefined;
        },
        "invalid hash: error code is present": (r) => {
          const body = r.json();
          return body.error && typeof body.error.code === "number";
        },
        "invalid hash: error message is present": (r) => {
          const body = r.json();
          return body.error && typeof body.error.message === "string";
        }
      });
      break;
  }
} 
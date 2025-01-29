import http from "k6/http";
import { check } from "k6";
import { config, validateJsonRpcResponse } from "../common.js";

export function getBlockTransactionCountByNumber() {
  const validBlockNumber = "0x1"; // Using block number 1 as a valid block
  
  const payload = JSON.stringify({
    jsonrpc: "2.0",
    method: "eth_getBlockTransactionCountByNumber",
    params: [validBlockNumber],
    id: 1,
  });

  const headers = { "Content-Type": "application/json" };
  if (config.apiKey) {
    headers["X-API-KEY"] = config.apiKey;
  }

  const res = http.post(config.endpoint, payload, { headers: headers });

  // Validate common JSON-RPC structure
  const passed = validateJsonRpcResponse(res, "eth_getBlockTransactionCountByNumber");

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
  getBlockTransactionCountByNumber();
  runErrorTest("0x999999999", "non-existent block");
  runErrorTest("latest", "latest block");
  runErrorTest("earliest", "earliest block");
  runErrorTest("invalid_number", "invalid block number");
}

function runErrorTest(blockNumber, testCase) {
  const payload = JSON.stringify({
    jsonrpc: "2.0",
    method: "eth_getBlockTransactionCountByNumber",
    params: [blockNumber],
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

    case "latest block":
    case "earliest block":
      check(res, {
        "block tag: result is hex": (r) => {
          const body = r.json();
          const hexPattern = /^0x[0-9a-fA-F]+$/;
          return typeof body.result === "string" && hexPattern.test(body.result);
        }
      });
      break;

    case "invalid block number":
      check(res, {
        "invalid number: has error field": (r) => {
          const body = r.json();
          return body.error !== undefined;
        },
        "invalid number: error code is present": (r) => {
          const body = r.json();
          return body.error && typeof body.error.code === "number";
        },
        "invalid number: error message is present": (r) => {
          const body = r.json();
          return body.error && typeof body.error.message === "string";
        }
      });
      break;
  }
}
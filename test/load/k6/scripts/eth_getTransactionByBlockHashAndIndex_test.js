import http from "k6/http";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";
import { getBaseUrl, getDefaultHeaders } from "../common.js";

const errors = new Counter("errors");

export function getTransactionByBlockHashAndIndex() {
  const payload = JSON.stringify({
    method: "eth_getTransactionByBlockHashAndIndex",
    params: [
      "0x52c41586e9e9e517a9c2c0a5a36925299e398e0f69bbf2c1b740a97456788d06",
      "0x0"
    ],
    id: 1,
    jsonrpc: "2.0",
  });

  const response = http.post(getBaseUrl(), payload, {
    headers: getDefaultHeaders(),
  });

  // Verify the response
  const success = check(response, {
    "is status 200": (r) => r.status === 200,
    "has valid response": (r) => {
      const body = JSON.parse(r.body);
      return (
        body.jsonrpc === "2.0" && 
        body.id === 1 && 
        body.result !== undefined
      );
    },
    "has valid transaction fields": (r) => {
      const body = JSON.parse(r.body);
      const result = body.result;
      return (
        result === null || // null is valid if transaction not found
        (
          result.hash !== undefined &&
          result.blockHash !== undefined &&
          result.blockNumber !== undefined &&
          result.transactionIndex !== undefined &&
          result.from !== undefined &&
          result.to !== undefined &&
          result.value !== undefined &&
          result.gas !== undefined &&
          result.gasPrice !== undefined &&
          result.input !== undefined &&
          result.nonce !== undefined &&
          result.type !== undefined
        )
      );
    },
    "transaction fields are properly formatted": (r) => {
      const body = JSON.parse(r.body);
      const result = body.result;
      if (result === null) return true;
      
      return (
        result.hash.startsWith("0x") &&
        result.blockHash.startsWith("0x") &&
        result.blockNumber.startsWith("0x") &&
        result.transactionIndex.startsWith("0x") &&
        result.from.startsWith("0x") &&
        (result.to === null || result.to.startsWith("0x")) &&
        result.value.startsWith("0x") &&
        result.gas.startsWith("0x") &&
        result.gasPrice.startsWith("0x") &&
        result.input.startsWith("0x") &&
        result.nonce.startsWith("0x") &&
        result.type.startsWith("0x")
      );
    }
  });

  if (!success) {
    errors.add(1);
  }

  sleep(1);
}

// For running the test script directly
export default function () {
  getTransactionByBlockHashAndIndex();
  runErrorTest("0x0000000000000000000000000000000000000000000000000000000000000000", "0x0", "non-existent block");
  runErrorTest("invalid_hash", "0x0", "invalid hash format");
  runErrorTest("0x52c41586e9e9e517a9c2c0a5a36925299e398e0f69bbf2c1b740a97456788d06", "0xinvalid", "invalid index format");
}

function runErrorTest(blockHash, index, testCase) {
  const payload = JSON.stringify({
    jsonrpc: "2.0",
    method: "eth_getTransactionByBlockHashAndIndex",
    params: [blockHash, index],
    id: 1,
  });

  const response = http.post(getBaseUrl(), payload, {
    headers: getDefaultHeaders(),
  });

  switch (testCase) {
    case "non-existent block":
      check(response, {
        "non-existent block: result is null": (r) => {
          const body = JSON.parse(r.body);
          return body.result === null;
        }
      });
      break;

    case "invalid hash format":
    case "invalid index format":
      check(response, {
        "invalid format: has error field": (r) => {
          const body = JSON.parse(r.body);
          return body.error !== undefined;
        },
        "invalid format: error code is present": (r) => {
          const body = JSON.parse(r.body);
          return body.error && typeof body.error.code === "number";
        },
        "invalid format: error message is present": (r) => {
          const body = JSON.parse(r.body);
          return body.error && typeof body.error.message === "string";
        }
      });
      break;
  }
} 
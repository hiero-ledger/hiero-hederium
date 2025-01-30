import http from "k6/http";
import { check } from "k6";
import { Counter } from "k6/metrics";
import { config, validateJsonRpcResponse } from "../common.js";

const errors = new Counter("errors");

export function getTransactionByBlockNumberAndIndex() {
  const payload = JSON.stringify({
    jsonrpc: "2.0",
    method: "eth_getTransactionByBlockNumberAndIndex",
    params: ["0xc47ac3", "0x0"],
    id: 1,
  });

  const headers = { "Content-Type": "application/json" };
  if (config.apiKey) {
    headers["X-API-KEY"] = config.apiKey;
  }

  const res = http.post(config.endpoint, payload, { headers: headers });

  // Validate common JSON-RPC structure
  const passed = validateJsonRpcResponse(res, "eth_getTransactionByBlockNumberAndIndex");
  if (!passed) {
    errors.add(1);
    return;
  }

  // Additional checks specific to transaction response
  const checksPassed = check(res, {
    "result is an object or null": (r) => {
      const body = r.json();
      return typeof body.result === "object" || body.result === null;
    },
    "transaction has required fields if not null": (r) => {
      const body = r.json();
      const result = body.result;
      return (
        result === null ||
        (result &&
          result.hash &&
          result.blockHash &&
          result.blockNumber &&
          result.transactionIndex !== undefined &&
          result.from &&
          result.type !== undefined &&
          result.chainId)
      );
    },
    "transaction fields are properly formatted if not null": (r) => {
      const body = r.json();
      const result = body.result;
      return (
        result === null ||
        (result &&
          result.hash.startsWith("0x") &&
          result.blockHash.startsWith("0x") &&
          result.blockNumber.startsWith("0x") &&
          result.transactionIndex.startsWith("0x") &&
          result.from.startsWith("0x") &&
          result.type.startsWith("0x") &&
          result.chainId.startsWith("0x"))
      );
    },
  });

  if (!checksPassed) {
    errors.add(1);
  }
} 
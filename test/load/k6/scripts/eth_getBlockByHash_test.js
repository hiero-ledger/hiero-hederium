import http from "k6/http";
import { check } from "k6";
import { config, validateJsonRpcResponse } from "../common.js";

export let options = {
  vus: config.vus,
  duration: config.duration,
};

export default function () {
  const payload = JSON.stringify({
    jsonrpc: "2.0",
    method: "eth_getBlockByHash",
    params: [
      "0x52c41586e9e9e517a9c2c0a5a36925299e398e0f69bbf2c1b740a97456788d06",
      false,
    ],
    id: 1,
  });

  const headers = { "Content-Type": "application/json" };
  if (config.apiKey) {
    headers["X-API-KEY"] = config.apiKey;
  }

  const res = http.post(config.endpoint, payload, { headers: headers });

  // Validate common JSON-RPC structure
  const passed = validateJsonRpcResponse(res, "eth_getBlockByHash");

  // Additional checks specific to block response
  check(res, {
    "result is an object": (r) => {
      const body = r.json();
      return typeof body.result === "object";
    },
    "block has correct hash": (r) => {
      const body = r.json();
      return (
        body.result &&
        body.result.hash ===
          "0x52c41586e9e9e517a9c2c0a5a36925299e398e0f69bbf2c1b740a97456788d06"
      );
    },
    "block has required fields": (r) => {
      const body = r.json();
      const result = body.result;
      return (
        result &&
        result.hash &&
        result.parentHash &&
        result.number &&
        result.timestamp &&
        result.transactions !== undefined &&
        Array.isArray(result.transactions)
      );
    },
    "block fields are properly formatted": (r) => {
      const body = r.json();
      const result = body.result;
      return (
        result &&
        result.hash.startsWith("0x") &&
        result.parentHash.startsWith("0x") &&
        result.number.startsWith("0x") &&
        result.timestamp.startsWith("0x")
      );
    },
  });
}

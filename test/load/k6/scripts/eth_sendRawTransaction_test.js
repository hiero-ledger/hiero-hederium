import http from "k6/http";
import { check } from "k6";
import { config, validateJsonRpcResponse } from "../common.js";

export let options = {
  vus: config.vus,
  duration: config.duration,
};

export default function () {
  // This is a sample transaction that will likely fail with nonce too low
  // but is useful for load testing the endpoint's error handling
  const rawTransaction = '0xf8cc1e854f29944800832dc6c0940a56fd9e0c4f67df549e7f375a9451c0086482ec80b864a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000b757064617465645f6d7367000000000000000000000000000000000000000000820274a0cd6095ae91ea5d609b32923a9f73572e2d031fde0b7e38de44d3eda187474140a03028ecf5eb61070cba8e927ad5e11eac116da441307f2d54dae8be90f4476c59';

  const payload = JSON.stringify({
    jsonrpc: "2.0",
    method: "eth_sendRawTransaction",
    params: [rawTransaction],
    id: 1,
  });

  const headers = { "Content-Type": "application/json" };
  if (config.apiKey) {
    headers["X-API-KEY"] = config.apiKey;
  }

  const res = http.post(config.endpoint, payload, { headers: headers });

  // Validate common JSON-RPC structure
  const passed = validateJsonRpcResponse(res, "eth_sendRawTransaction");

  // Additional checks specific to eth_sendRawTransaction
  check(res, {
    "response has transaction hash or expected error": (r) => {
      const body = r.json();
      // Either we get a transaction hash (success) or an error about nonce/gas (expected)
      return (
        (body.result && body.result.startsWith("0x")) || // Success case
        (body.error && ( // Expected error cases
          body.error.message.includes("nonce") ||
          body.error.message.includes("gas")
        ))
      );
    },
  });
} 
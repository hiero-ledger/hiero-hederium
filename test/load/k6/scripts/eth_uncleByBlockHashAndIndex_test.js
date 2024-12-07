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
    method: "eth_getUncleByBlockHashAndIndex",
    params: [],
    id: 1,
  });

  const headers = { "Content-Type": "application/json" };
  if (config.apiKey) {
    headers["X-API-KEY"] = config.apiKey;
  }

  const res = http.post(config.endpoint, payload, { headers: headers });

  // Validate common JSON-RPC structure
  const passed = validateJsonRpcResponse(
    res,
    "eth_getUncleByBlockHashAndIndex"
  );

  // Additional checks
  check(res, {
    "result is null": (r) => {
      const body = r.json();
      return body.result === null;
    },
  });
}

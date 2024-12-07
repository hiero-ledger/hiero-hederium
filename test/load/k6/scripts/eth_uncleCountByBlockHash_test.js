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
    method: "eth_getUncleCountByBlockHash",
    params: [],
    id: 1,
  });

  const headers = { "Content-Type": "application/json" };
  if (config.apiKey) {
    headers["X-API-KEY"] = config.apiKey;
  }

  const res = http.post(config.endpoint, payload, { headers: headers });

  // Validate common JSON-RPC structure
  const passed = validateJsonRpcResponse(res, "eth_getUncleCountByBlockHash");

  // Additional checks
  check(res, {
    "result is 0x0": (r) => {
      const body = r.json();
      return body.result === "0x0";
    },
  });
}

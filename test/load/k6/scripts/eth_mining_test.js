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
    method: "eth_mining",
    params: [],
    id: 1,
  });

  const headers = { "Content-Type": "application/json" };
  if (config.apiKey) {
    headers["X-API-KEY"] = config.apiKey;
  }

  const res = http.post(config.endpoint, payload, { headers: headers });

  // Validate common JSON-RPC structure
  const passed = validateJsonRpcResponse(res, "eth_mining");

  // Additional checks specific to eth_mining
  check(res, {
    "result is false": (r) => {
      const body = r.json();
      return body.result === false;
    },
  });
}

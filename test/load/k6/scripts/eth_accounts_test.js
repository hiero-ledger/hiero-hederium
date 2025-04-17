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
    method: "eth_accounts",
    params: [],
    id: 1,
  });

  const headers = { "Content-Type": "application/json" };
  if (config.apiKey) {
    headers["X-API-KEY"] = config.apiKey;
  }

  const res = http.post(config.endpoint, payload, { headers: headers });

  // First check if the request was successful
  check(res, {
    "status is 200": (r) => r.status === 200,
  });

  // Only proceed with JSON validation if we got a successful response
  if (res.status === 200) {
    try {
      const body = res.json();
      
      // Validate the response structure
      check(res, {
        "jsonrpc is 2.0": (r) => body.jsonrpc === "2.0",
        "id matches request": (r) => body.id === 1,
        "result is an array": (r) => Array.isArray(body.result),
        "result is empty": (r) => body.result.length === 0,
      });
    } catch (e) {
      console.error("Failed to parse JSON response:", e);
    }
  }
}

import http from "k6/http";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";
import { getBaseUrl, getDefaultHeaders } from "../common.js";

const errors = new Counter("errors");

export function getTransactionByHash() {
  const payload = JSON.stringify({
    method: "eth_getTransactionByHash",
    params: [
      "0x5d019848d6dad96bc3a9e947350975cd16cf1c51efd4d5b9a273803446fbbb43",
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
        body.jsonrpc === "2.0" && body.id === 1 && body.result !== undefined
      );
    },
  });

  if (!success) {
    errors.add(1);
  }

  sleep(1);
}

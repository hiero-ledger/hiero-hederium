import { check } from "k6";
import http from "k6/http";
import { Counter } from "k6/metrics";
import { getBaseUrl, getDefaultHeaders } from "../common.js";

const errors = new Counter("errors");

export function call() {
  const baseUrl = getBaseUrl();
  const headers = getDefaultHeaders();

  const payload = {
    method: "eth_call",
    params: [
      {
        from: "0x17b2b8c63fa35402088640e426c6709a254c7ffb",
        to: "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
        data: "0x70a08231000000000000000000000000b1d6b01b94d854f521665696ea17fcf87c160d97",
      },
      "latest",
    ],
    id: 1,
    jsonrpc: "2.0",
  };

  const response = http.post(baseUrl, JSON.stringify(payload), { headers });

  // Check if response is successful
  const success = check(response, {
    "status is 200": (r) => r.status === 200,
    "has result": (r) => r.json().result !== undefined,
    "no error": (r) => r.json().error === undefined,
    "valid hex": (r) => /^0x[0-9a-fA-F]+$/.test(r.json().result),
  });

  if (!success) {
    console.log(`Error in eth_call: ${response.body}`);
    errors.add(1);
  }

  return response;
}

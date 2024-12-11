import { check } from "k6";
import http from "k6/http";
import { Counter } from "k6/metrics";
import { randomIntBetween } from "https://jslib.k6.io/k6-utils/1.2.0/index.js";
import { getBaseUrl, getDefaultHeaders } from "../common.js";

const errors = new Counter("errors");

export function getTransactionCount() {
  const baseUrl = getBaseUrl();
  const headers = getDefaultHeaders();
  const address = "0x117EE00a3191499e816068b10b4BE78DEEd74087";
  const blockParams = ["latest", "earliest", "pending", "0x1", "0x100"];
  const blockParam = blockParams[randomIntBetween(0, blockParams.length - 1)];

  const payload = {
    jsonrpc: "2.0",
    method: "eth_getTransactionCount",
    params: [address, blockParam],
    id: 1,
  };

  const response = http.post(baseUrl, JSON.stringify(payload), { headers });

  const success = check(response, {
    "status is 200": (r) => r.status === 200,
    "has result": (r) => r.json().result !== undefined,
    "no error": (r) => r.json().error === undefined,
    "valid hex": (r) => /^0x[0-9a-fA-F]+$/.test(r.json().result),
    "valid nonce format": (r) => parseInt(r.json().result, 16) >= 0,
  });

  if (!success) {
    console.log(`Error in getTransactionCount: ${response.body}`);
    errors.add(1);
  }

  return response;
}

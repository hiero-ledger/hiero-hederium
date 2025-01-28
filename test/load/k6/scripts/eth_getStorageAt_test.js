import { check } from "k6";
import http from "k6/http";
import { Counter } from "k6/metrics";
import { randomIntBetween } from "https://jslib.k6.io/k6-utils/1.2.0/index.js";
import { getBaseUrl, getDefaultHeaders } from "../common.js";

const errors = new Counter("errors");

export default function getStorageAt() {
  const baseUrl = getBaseUrl();
  const headers = getDefaultHeaders();
  const address = "0x117EE00a3191499e816068b10b4BE78DEEd74087";
  const slot = "0x0000000000000000000000000000000000000000000000000000000000000000";
  const blockParams = ["latest", "earliest", "0x1", "0x100"];
  const blockParam = blockParams[randomIntBetween(0, blockParams.length - 1)];

  const payload = {
    jsonrpc: "2.0",
    method: "eth_getStorageAt",
    params: [address, slot, blockParam],
    id: 1,
  };

  const response = http.post(baseUrl, JSON.stringify(payload), { headers });

  const success = check(response, {
    "status is 200": (r) => r.status === 200,
    "has result": (r) => r.json().result !== undefined,
    "no error": (r) => r.json().error === undefined,
    "valid hex": (r) => /^0x[0-9a-fA-F]{64}$/.test(r.json().result),
  });

  if (!success) {
    console.log(`Error in getStorageAt: ${response.body}`);
    errors.add(1);
  }

  return response;
} 
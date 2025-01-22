import http from "k6/http";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";
import { getBaseUrl, getDefaultHeaders, config } from "../common.js";

const errors = new Counter("errors");

export default function getFeeHistory() {
  const payload = JSON.stringify({
    method: "eth_feeHistory",
    params: [
      "0x5", // blockCount
      "latest", // newestBlock
      [] // rewardPercentiles
    ],
    id: 0,
    jsonrpc: "2.0"
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
        body.jsonrpc === "2.0" && 
        body.id === 0 && 
        body.result !== undefined
      );
    },
    "has valid feeHistory fields": (r) => {
      const body = JSON.parse(r.body);
      const result = body.result;
      
      return (
        result === null || // null is valid if there's an error
        (
          Array.isArray(result.base_fee_per_gas) &&
          Array.isArray(result.gas_used_ratio) &&
          typeof result.oldest_block === "string"
        )
      );
    }
  });

  if (!success) {
    errors.add(1);
  }

  sleep(1);
} 

export const options = {
  vus: config.vus,
  duration: config.duration,
}; 
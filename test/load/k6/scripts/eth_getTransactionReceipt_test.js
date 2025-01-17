import http from "k6/http";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";
import { getBaseUrl, getDefaultHeaders } from "../common.js";

const errors = new Counter("errors");

export function getTransactionReceipt() {
  const payload = JSON.stringify({
    method: "eth_getTransactionReceipt",
    params: [
      "0xa2773c69ed43bec8057e40151f651d6195c4351f4b9b49d20a92567d50042290"
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
    "has valid receipt fields": (r) => {
      const body = JSON.parse(r.body);
      const result = body.result;
      return (
        result === null || // null is valid if transaction not found
        (
          result.transactionHash !== undefined &&
          result.blockHash !== undefined &&
          result.blockNumber !== undefined &&
          result.transactionIndex !== undefined &&
          result.from !== undefined &&
          result.to !== undefined &&
          result.cumulativeGasUsed !== undefined &&
          result.gasUsed !== undefined &&
          result.logs !== undefined &&
          result.logsBloom !== undefined &&
          result.status !== undefined
        )
      );
    }
  });

  if (!success) {
    errors.add(1);
  }

  sleep(1);
} 
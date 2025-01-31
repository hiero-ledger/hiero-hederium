import http from "k6/http";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";
import { getBaseUrl, getDefaultHeaders, config } from "../common.js";

const errors = new Counter("errors");

// Helper function to verify fee history response
function verifyFeeHistory(response, blockCount, hasRewardPercentiles) {
  return check(response, {
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
      
      if (result === null) return true; // null is valid if there's an error
      
      // Check array lengths
      const expectedLength = parseInt(blockCount, 16);
      if (result.base_fee_per_gas.length !== expectedLength + 1) return false;
      if (result.gas_used_ratio.length !== expectedLength) return false;
      
      // Check that all base fees are the same (fixed fee)
      const firstFee = result.base_fee_per_gas[0];
      if (!result.base_fee_per_gas.every(fee => fee === firstFee)) return false;
      
      // Check that all gas used ratios are 0.5
      if (!result.gas_used_ratio.every(ratio => ratio === 0.5)) return false;
      
      // Check reward array if percentiles were provided
      if (hasRewardPercentiles) {
        if (!Array.isArray(result.reward)) return false;
        if (result.reward.length !== expectedLength) return false;
      }
      
      return (
        typeof result.oldest_block === "string" &&
        result.oldest_block.startsWith("0x")
      );
    }
  });
}

export function getFeeHistory() {
  // Test case 1: Basic fee history with no reward percentiles
  const basicPayload = JSON.stringify({
    method: "eth_feeHistory",
    params: ["0x5", "latest", []],
    id: 0,
    jsonrpc: "2.0"
  });
  
  let response = http.post(getBaseUrl(), basicPayload, {
    headers: getDefaultHeaders(),
  });
  
  let success = verifyFeeHistory(response, "0x5", false);
  if (!success) errors.add(1);
  
  sleep(1);
  
  // Test case 2: Fee history with reward percentiles
  const rewardPayload = JSON.stringify({
    method: "eth_feeHistory",
    params: ["0x3", "latest", ["0x5", "0xa", "0xf"]],
    id: 0,
    jsonrpc: "2.0"
  });
  
  response = http.post(getBaseUrl(), rewardPayload, {
    headers: getDefaultHeaders(),
  });
  
  success = verifyFeeHistory(response, "0x3", true);
  if (!success) errors.add(1);
  
  sleep(1);
  
  // Test case 3: Maximum block count (should be capped)
  const maxBlocksPayload = JSON.stringify({
    method: "eth_feeHistory",
    params: ["0x14", "latest", []],
    id: 0,
    jsonrpc: "2.0"
  });
  
  response = http.post(getBaseUrl(), maxBlocksPayload, {
    headers: getDefaultHeaders(),
  });
  
  success = verifyFeeHistory(response, "0xa", false); // Should be capped at 10 blocks
  if (!success) errors.add(1);
  
  sleep(1);
}

export const options = {
  vus: config.vus,
  duration: config.duration,
}; 
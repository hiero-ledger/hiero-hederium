import { check } from "k6";
import http from "k6/http";
import { Counter } from "k6/metrics";
import { getBaseUrl, getDefaultHeaders } from "../common.js";

const errors = new Counter("errors");

export function testGetCode() {
  const baseUrl = getBaseUrl();
  const headers = getDefaultHeaders();

  // Test iHTS precompile address
  const iHTSPayload = {
    jsonrpc: "2.0",
    method: "eth_getCode",
    params: ["0x0000000000000000000000000000000000000167", "latest"],
    id: 1,
  };

  const iHTSResponse = http.post(baseUrl, JSON.stringify(iHTSPayload), { headers });

  // Check if response is successful
  const iHTSSuccess = check(iHTSResponse, {
    "status is 200": (r) => r.status === 200,
    "has result": (r) => r.json().result !== undefined,
    "no error": (r) => r.json().error === undefined,
    "iHTS precompile returns 0xfe": (r) => r.json().result === "0xfe",
  });

  if (!iHTSSuccess) {
    console.log(`Error in eth_getCode iHTS test: ${iHTSResponse.body}`);
    errors.add(1);
  }

  // Test existing contract address
  const existingContractPayload = {
    jsonrpc: "2.0",
    method: "eth_getCode",
    params: ["0xc528c46d0e37ea111c63e306c548bb909ced7efa", "latest"],
    id: 2,
  };

  const existingContractResponse = http.post(baseUrl, JSON.stringify(existingContractPayload), { headers });

  // Check if response is successful
  const existingContractSuccess = check(existingContractResponse, {
    "status is 200": (r) => r.status === 200,
    "has result": (r) => r.json().result !== undefined,
    "no error": (r) => r.json().error === undefined,
    "existing contract returns bytecode": (r) => r.json().result.length > 2, // Should be longer than "0x"
  });

  if (!existingContractSuccess) {
    console.log(`Error in eth_getCode existing contract test: ${existingContractResponse.body}`);
    errors.add(1);
  }

  // Test non-existent address
  const nonExistentPayload = {
    jsonrpc: "2.0",
    method: "eth_getCode",
    params: ["0x1234567890123456789012345678901234567890", "latest"],
    id: 3,
  };

  const nonExistentResponse = http.post(baseUrl, JSON.stringify(nonExistentPayload), { headers });

  // Check if response is successful
  const nonExistentSuccess = check(nonExistentResponse, {
    "status is 200": (r) => r.status === 200,
    "has result": (r) => r.json().result !== undefined,
    "no error": (r) => r.json().error === undefined,
    "non-existent address returns 0x": (r) => r.json().result === "0x",
  });

  if (!nonExistentSuccess) {
    console.log(`Error in eth_getCode non-existent address test: ${nonExistentResponse.body}`);
    errors.add(1);
  }

  return { iHTSResponse, existingContractResponse, nonExistentResponse };
}

export default function () {
  testGetCode();
} 
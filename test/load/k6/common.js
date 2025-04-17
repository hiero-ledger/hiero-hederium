import { check } from "k6";

export function getEnvVar(name, defaultValue) {
  const value = __ENV[name];
  return value === undefined || value === "" ? defaultValue : value;
}

// Load configuration from environment variables with defaults
export const config = {
  endpoint: getEnvVar("ENDPOINT_URL", "http://localhost:7546"),
  apiKey: getEnvVar("API_KEY", ""), // empty means no key required
  vus: parseInt(getEnvVar("VUS", "5000")),
  duration: getEnvVar("DURATION", "3m"),
};

export function getBaseUrl() {
  return config.endpoint;
}

export function getDefaultHeaders() {
  const headers = {
    "Content-Type": "application/json",
  };

  if (config.apiKey) {
    headers["X-API-Key"] = config.apiKey;
  }

  return headers;
}

// A basic check function for JSON-RPC responses
export function validateJsonRpcResponse(response, methodExpected) {
  return check(response, {
    "status is 200": (r) => r.status === 200,
    "jsonrpc is 2.0": (r) => {
      const body = r.json();
      return body.jsonrpc === "2.0";
    },
    [`method ${methodExpected} result`]: (r) => {
      const body = r.json();
      return body.result !== undefined;
    },
  });
}

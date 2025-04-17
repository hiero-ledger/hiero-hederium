import { default as test } from "../scripts/eth_accounts_test.js";

export let options = {
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95th percentile under 500ms
    http_req_failed: ["rate<0.01"], // less than 1% of requests should fail
  },
  stages: [
    { duration: "10s", target: 50 }, // Ramp up to 50 users
    { duration: "30s", target: 1000 }, // Ramp up to 1000 users
    { duration: "30s", target: 2000 }, // Ramp up to 2000 users
    { duration: "30s", target: 5000 }, // Stay at 5000 users
    { duration: "10s", target: 0 }, // Ramp down to 0 users
  ],
};

export default test;
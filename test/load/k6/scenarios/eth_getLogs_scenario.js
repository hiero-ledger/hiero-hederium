import { getLogs } from "../scripts/eth_getLogs_test.js";

export const options = {
  scenarios: {
    constant_load: {
      executor: "constant-vus",
      vus: 10,
      duration: "30s",
      gracefulStop: "5s",
    },
    stress_test: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "10s", target: 5 },
        { duration: "20s", target: 10 },
        { duration: "20s", target: 20 },
        { duration: "10s", target: 0 },
      ],
      gracefulRampDown: "5s",
    },
    spike_test: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "10s", target: 10 },
        { duration: "10s", target: 20 },
        { duration: "10s", target: 5 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<500"],
    http_req_failed: ["rate<0.01"],
    errors: ["count<100"],
  },
};

export default function () {
  getLogs();
} 
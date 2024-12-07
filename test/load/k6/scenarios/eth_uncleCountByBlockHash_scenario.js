import { config } from "../common.js";

export let options = {
  vus: config.vus,
  duration: config.duration,
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95th percentile under 500ms
    http_req_failed: ["rate<0.01"], // less than 1% of requests should fail
  },
  stages: [
    { duration: "10s", target: 5 }, // Ramp up to 5 users
    { duration: "20s", target: 10 }, // Ramp up to 10 users
    { duration: "30s", target: 10 }, // Stay at 10 users
    { duration: "10s", target: 0 }, // Ramp down to 0 users
  ],
};

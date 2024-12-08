import { config } from "../common.js";

export let options = {
  vus: config.vus,
  duration: config.duration,
  thresholds: {
    http_req_duration: ["p(95)<100"], // 95th percentile under 500ms
  },
  stages: [
    { duration: "10s", target: 5 },
    { duration: "20s", target: 10 },
    { duration: "10s", target: 5 },
  ],
};

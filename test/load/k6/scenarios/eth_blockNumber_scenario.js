import { default as test } from "../scripts/eth_blockNumber_test.js";
import { config } from "../common.js";

export let options = {
  vus: config.vus,
  thresholds: {
    http_req_duration: ["p(99)<200"], // 99th percentile under 200ms
  },
  stages: [
    { duration: "20s", target: 50 },
    { duration: "30s", target: 100 },
    { duration: "20s", target: 50 },
  ],
};

export default test;

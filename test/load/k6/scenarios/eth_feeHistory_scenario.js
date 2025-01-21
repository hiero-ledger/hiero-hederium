import { getFeeHistory } from "../scripts/eth_feeHistory_test.js";
import { config } from "../common.js";

export const options = {
  scenarios: {
    constant_load: {
      executor: "constant-vus",
      vus: 5,
      duration: "30s",
      gracefulStop: "5s",
    }
  },
  thresholds: {
    http_req_duration: ["p(95)<1000"],
    http_req_failed: ["rate<0.01"],
    errors: ["count<10"],
  },
};

export default function () {
  getFeeHistory();
} 
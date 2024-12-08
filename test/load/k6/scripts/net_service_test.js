import http from "k6/http";
import { check, sleep } from "k6";
import {
  netServiceScenario,
  netServiceThresholds,
} from "../scenarios/net_service_scenario.js";

export const options = {
  scenarios: netServiceScenario,
  thresholds: netServiceThresholds,
};

export default function () {
  // Test listening endpoint
  const listeningRes = http.get("http://localhost:8545/net/listening");
  check(listeningRes, {
    "listening status is 200": (r) => r.status === 200,
    "listening returns false": (r) => r.json().result === false,
  });

  // Test version endpoint
  const versionRes = http.get("http://localhost:8545/net/version");
  check(versionRes, {
    "version status is 200": (r) => r.status === 200,
    "version is not empty": (r) => r.json().result.length > 0,
  });

  sleep(1);
}

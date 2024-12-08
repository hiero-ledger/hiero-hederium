export const netServiceScenario = {
  net_service_smoke: {
    executor: "constant-vus",
    vus: 1,
    duration: "30s",
  },
  net_service_load: {
    executor: "ramping-vus",
    startVUs: 1,
    stages: [
      { duration: "30s", target: 20 },
      { duration: "1m", target: 20 },
      { duration: "30s", target: 0 },
    ],
    gracefulRampDown: "30s",
  },
  net_service_stress: {
    executor: "ramping-vus",
    startVUs: 20,
    stages: [
      { duration: "2m", target: 50 },
      { duration: "5m", target: 50 },
      { duration: "2m", target: 0 },
    ],
    gracefulRampDown: "30s",
  },
  net_service_soak: {
    executor: "constant-vus",
    vus: 10,
    duration: "30m",
  },
};

export const netServiceThresholds = {
  http_req_duration: ["p(95)<500"], // 95% of requests should be below 500ms
  http_req_failed: ["rate<0.01"], // Less than 1% of requests should fail
  http_reqs: ["rate>100"], // Should maintain at least 100 RPS
};

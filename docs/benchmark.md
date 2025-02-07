# Benchmark Results (Work in progress)

This document summarizes the performance benchmarking results between the **Go** and other implementations of the json-rpc-relay. Tests were performed using [k6](https://k6.io/) for load testing and custom end-to-end (E2E) integration tests. The goal is to provide an objective comparison and highlight areas for optimization in each implementation.

## Table of Contents
1. [Overview](#overview)  
2. [Environment & Setup](#environment--setup)  
3. [Methodology](#methodology)  
4. [Results](#results)  
   1. [k6 Load Test](#k6-load-test)  
   2. [E2E Tests](#e2e-tests)  
5. [Analysis & Observations](#analysis--observations)  

---

## Overview
We developed two implementations of the same product:
- **Other Implementation**  
- **Go Implementation**

This document aims to:
- Compare response times, throughput, and resource usage under various loads.
- Provide insights that can guide optimization and further development efforts.

---

## Environment & Setup
Below are details about the testing environment and how each test was set up.

**Infrastructure:**
- **Operating System:** MacOs 15.1
- **CPU:** Apple M4 Pro
- **Memory:** 24 GB
- **Network:** Local environment or Docker container-based setup

**Software Versions:**
- **Node.js:** v20.x
- **Go:** 1.22
- **k6:** 0.56.0 (for load testing)
- **hardhat:** 2.19.4 (for E2E tests)

> **Note**: Each implementation ran in identical container configurations and on the same host machine to ensure a fair comparison.

---

## Methodology

### k6 Load Test
1. **Scenario**: Testing CRUD operations on the core API endpoints.  
2. **k6 Script**:
   - Virtual users: 50, 100, 200, and 500.  
   - Duration: 5 minutes per load step, 1 minute ramp-up between steps.  
   - Endpoints tested are the same as the ones in the [RPC API](docs/rpc-api.md)
   - Metrics collected:
     - **Response time (p95, p99)**  
     - **Throughput (requests per second)**  
     - **Error rate**  
     - **CPU & Memory usage** (collected via system monitoring or container metrics)

### E2E Tests
1. **Scenario**: Using the same test suite to validate the end-to-end flow, from client to database.  
2. **Tooling**: Custom test scripts using **Jest** (Node.js) and **Go testing**.  
3. **Metrics**:
   - **Execution time** for the entire test suite.  
   - **Pass/Fail** count of test cases.  
   - **Resource usage** (if applicable).

---

## Results

### k6 Load Test
| Metric                  | Other       | Go            | Notes                             |
|-------------------------|---------------|---------------|------------------------------------|
| **Requests/sec (avg)** | ?          | ?          | Go has higher throughput           |
| **p95 Latency**         | ?         | ?         | p95: Go is ~33% faster             |
| **p99 Latency**         | ?         | ?         | p99: Go is ~27% faster             |
| **Error Rate**          | ?          | ?          | Other had slightly more errors   |
| **CPU Usage (avg)**     | ?           | ?           | Both implementations scaled similarly |
| **Memory (avg)**        | ?        | ?        | Other used slightly more memory  |

### E2E Tests
| Criterion                | Other        | Go             |
|--------------------------|----------------|----------------|
| **Full Test Suite Time** | ?         | ?         |
| **Test Cases Passed**    | ?   | ?   |
| **Test Cases Failed**    | ?         | ?         |
| **Resource Usage**       | ?        | ?        |

---

## Analysis & Observations
1. **Performance**:
   - 

2. **Resource Utilization**:
   - 

3. **Stability**:
   - 

---

**_If you have any questions or would like more details on the benchmarking methodology, feel free to open an issue or reach out to the team._**

> **Disclaimer**: The results are based on our specific environment and configurations, which can significantly affect performance metrics. Actual performance may vary under different conditions.

---

**Last Updated**: 2025-02-07  
**Maintainers**: [@LimeChain](https://github.com/LimeChain)

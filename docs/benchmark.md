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
We compared two implementations of the same product:
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
- **Network:** Docker container-based setup
- **Docker Resources:**
  - **CPU:** 2 cores
  - **Memory:** 1 GB RAM


**Software Versions:**
- **Node.js:** v20.x
- **Go:** 1.22
- **k6:** 0.56.0 (for load testing)
- **hardhat:** 2.19.4 (for E2E tests)

> **Note**: Each implementation ran in identical container configurations and on the same host machine to ensure a fair comparison.

---

## Methodology

### k6 Load Test
1. **Scenario**: Testing operations on the core API endpoints.  
2. **k6 Script**:
   - Virtual users: 50, 100, 200, and 500.  
   - Duration: 5 minutes per load step, 1 minute ramp-up between steps.  
   - Endpoints tested are the same as the ones in the [RPC API](docs/rpc-api.md)
   - Metrics collected:
     - **Response time (p95, p99)**  
     - **Throughput (requests per second)**  
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
### Resource Usage
| Resource                | Other       | Go            | Notes                             |
|------------------------|-------------|---------------|-----------------------------------|
| **CPU Usage (avg)**    | 1%         | 0.05%          | Go uses 20x less CPU on average     |
| **CPU Peak**           | 135.55%         | 95%          | Go has ~30% lower CPU peaks     |
| **RAM Usage (avg)**    | 340MB       | 30MB        | Go uses ~11x less memory on average         |
| **RAM Peak**           | 1360MB       | 40MB        | Go has significantly lower memory spikes (~34x less) |
| **Disk Read (avg)**    | 133 MB   | 246 KB    | Go performs ~550x less disk reads   |
| **Disk Write (avg)**    | 152 KB   | 0 B    | Go performs no disk writes vs Other's minimal writes   |


### k6 Load Test Results by Endpoint

#### eth_blockNumber
| Metric                  | Other       | Go (Problem)           | Notes                             |
|------------------------|-------------|---------------|------------------------------------|
| **Requests/sec (avg)** | 9257.982843/s          | 29943.711392/s             | Go handles ~3.2x more requests per second           |
| **All Requests**        | 2591776          | 8384404             | Go processed ~3.2x more total requests           |
| **Request Duration(avg)**        | 8.1ms          | 2.5ms             | Go is ~3.2x faster per request            |


#### eth_call (10 VUs)
| Metric                  | Other       | Go            | Notes                             |
|------------------------|-------------|---------------|------------------------------------|
| **Requests/sec (avg)** | 88.964511/s          | 124.823159/s             | Go handles ~40% more requests per second           |
| **All Requests**        | 2676          | 3762             | Go processed ~40% more total requests           |
| **Request Duration(avg)**        | 112.27ms         | 79.5ms             | Go is ~29% faster per request            |


#### eth_estimateGas (10 VUs)
| Metric                  | Other       | Go            | Notes                             |
|------------------------|-------------|---------------|------------------------------------|
| **Requests/sec (avg)** | 69.75052/s          | 109.292417/s             | Go handles ~57% more requests per second           |
| **All Requests**        | 2098          | 3289             | Go processed ~57% more total requests           |
| **Request Duration(avg)**        | 143.2ms          | 90.86ms             | Go is ~37% faster per request            |

#### eth_getBalance (50 VUs)
| Metric                  | Other       | Go            | Notes                             |
|------------------------|-------------|---------------|------------------------------------|
| **Requests/sec (avg)** | 406.830025/s          | 1616.015625/s             | Go handles ~4x more requests per second           |
| **All Requests**        | 14240          | 48526             | Go processed ~3.4x more total requests           |
| **Request Duration(avg)**        | 110.66ms          | 30.91ms             | Go is ~3.6x faster per request            |

#### eth_getTransactionByHash (50 VUs)
| Metric                  | Other       | Go            | Notes                             |
|------------------------|-------------|---------------|------------------------------------|
| **Requests/sec (avg)** | 313.29197/s          | 18817.782432/s             | Go handles ~60x more requests per second           |
| **All Requests**        | 10967          | 564626             | Go processed ~51x more total requests           |
| **Request Duration(avg)**        | 131.58ms          | 2.65ms             | Go is ~50x faster per request            |

#### eth_getBlockByHash (50 VUs)
| Metric                  | Other       | Go            | Notes                             |
|------------------------|-------------|---------------|------------------------------------|
| **Requests/sec (avg)** | 7431.685721/s          | 20121.079659/s             | Go handles ~2.7x more requests per second           |
| **All Requests**        | 222991          | 603664             | Go processed ~2.7x more total requests           |
| **Request Duration(avg)**        | 6.72ms          | 2.48ms             | Go is ~2.7x faster per request            |

### E2E Tests 
| Criterion                | Other        | Go             |
|--------------------------|----------------|----------------|
| **Full Test Suite Time** | 28.71 s         | 26.55 s         |

---

## Analysis & Observations
1. **Performance**:
   - The Go implementation consistently outperforms the Other implementation across all tested endpoints
   - Most significant performance difference is in eth_getTransactionByHash, where Go is ~50x faster
   - Lowest performance gain is in eth_call, where Go is still ~29% faster
   - Request throughput is substantially higher in Go, ranging from 40% to 60x improvement
   - Average response times are consistently lower in Go implementation

2. **Resource Utilization**:
   - Go implementation shows dramatically better resource efficiency:
     - CPU usage is 20x lower on average (0.05% vs 1%)
     - Memory consumption is 11x lower on average (30MB vs 340MB)
     - Peak memory usage shows even greater efficiency (40MB vs 1360MB)
     - Disk I/O is significantly reduced, with 550x less disk reads and no disk writes
   - Lower resource utilization suggests better scalability and cost-effectiveness in production

3. **Stability**:
   - Go implementation shows more consistent performance with lower CPU peaks (95% vs 135.55%)
   - Memory usage in Go remains stable with minimal spikes (40MB peak vs 1360MB)
   - E2E test execution times are comparable, with Go slightly faster (26.55s vs 28.71s)

---

**_If you have any questions or would like more details on the benchmarking methodology, feel free to open an issue or reach out to the team._**

> **Disclaimer**: The results are based on our specific environment and configurations, which can significantly affect performance metrics. Actual performance may vary under different conditions.

---

**Last Updated**: 2025-02-24  
**Maintainers**: [@LimeChain](https://github.com/LimeChain)

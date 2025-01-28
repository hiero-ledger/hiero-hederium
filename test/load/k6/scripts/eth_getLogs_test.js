import { check } from "k6";
import http from "k6/http";
import { Counter } from "k6/metrics";
import { randomIntBetween } from "https://jslib.k6.io/k6-utils/1.2.0/index.js";
import { getBaseUrl, getDefaultHeaders } from "../common.js";

const errors = new Counter("errors");

export function getLogs() {
  // First get the latest block number
  const blockNumberPayload = {
    jsonrpc: '2.0',
    id: 1,
    method: 'eth_blockNumber',
    params: []
  };

  const blockNumberResponse = http.post(getBaseUrl(), JSON.stringify(blockNumberPayload), {
    headers: getDefaultHeaders(),
  });

  console.log('Block number response:', blockNumberResponse.body);
  
  const latestBlock = blockNumberResponse.json().result;
  // Query the last 5 blocks
  const fromBlock = '0x' + (parseInt(latestBlock, 16) - 5).toString(16);
  const toBlock = latestBlock;

  const payload = {
    jsonrpc: '2.0',
    id: 1,
    method: 'eth_getLogs',
    params: [{
      fromBlock: fromBlock,
      toBlock: toBlock,
      address: [],    // Empty array to get all addresses
      topics: []      // Empty array to get all topics
    }]
  };

  console.log('Request payload:', JSON.stringify(payload));
  
  const response = http.post(getBaseUrl(), JSON.stringify(payload), {
    headers: getDefaultHeaders(),
  });

  console.log('Response status:', response.status);
  console.log('Response body:', response.body);

  check(response, {
    'Status is 200': (r) => r.status === 200,
    'Has result': (r) => r.json().result !== undefined,
    'No error in response': (r) => r.json().error === undefined,
    'Result is array': (r) => Array.isArray(r.json().result),
    'Response structure is valid': (r) => {
      const logs = r.json().result;
      if (!Array.isArray(logs)) return false;
      
      // If we have logs, verify their structure
      return logs.every(log => (
        typeof log === 'object' &&
        typeof log.address === 'string' && log.address.startsWith('0x') &&
        typeof log.blockHash === 'string' && log.blockHash.startsWith('0x') &&
        typeof log.blockNumber === 'string' && log.blockNumber.startsWith('0x') &&
        typeof log.data === 'string' &&
        typeof log.logIndex === 'string' &&
        typeof log.removed === 'boolean' &&
        (log.topics === null || Array.isArray(log.topics)) &&
        typeof log.transactionHash === 'string' &&
        typeof log.transactionIndex === 'string'
      )) || logs.length === 0; // Empty array is also valid
    },
    'Log fields are properly formatted': (r) => {
      const logs = r.json().result;
      if (!Array.isArray(logs) || logs.length === 0) return true;
      
      return logs.every(log => {
        // Required fields must be hex strings
        const requiredHexFields = ['address', 'blockHash', 'blockNumber'];
        const validRequiredFields = requiredHexFields.every(field => 
          log[field] && log[field].startsWith('0x')
        );

        // Optional fields can be empty strings or hex strings if present
        const optionalHexFields = ['logIndex', 'transactionHash', 'transactionIndex'];
        const validOptionalFields = optionalHexFields.every(field => 
          !log[field] || log[field] === '' || log[field].startsWith('0x')
        );

        // Data field must be a hex string or empty string
        const validData = !log.data || log.data === '' || log.data.startsWith('0x');

        // Topics must be null or an array of hex strings
        const validTopics = log.topics === null || 
          (Array.isArray(log.topics) && log.topics.every(topic => 
            typeof topic === 'string' && topic.startsWith('0x')
          ));

        return validRequiredFields && validOptionalFields && validData && validTopics;
      });
    }
  });

  if (response.json().error) {
    console.log('Error details:', JSON.stringify(response.json().error));
  }
}

export default function() {
  getLogs();
} 
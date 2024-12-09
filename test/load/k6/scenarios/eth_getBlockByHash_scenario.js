import { config } from "../common.js";

export const options = {
  scenarios: {
    eth_getBlockByHash_constant: {
      executor: "constant-vus",
      vus: config.vus,
      duration: config.duration,
      exec: "eth_getBlockByHash",
    },
  },
};

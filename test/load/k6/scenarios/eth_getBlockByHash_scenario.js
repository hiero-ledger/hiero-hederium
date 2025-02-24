import { getBlockByHash } from "../scripts/eth_getBlockByHash_test.js";

export const options = {
  scenarios: {
    constant_load: {
      executor: "constant-vus",
      vus: 50,
      duration: "30s",
      gracefulStop: "5s",
    },
  },
};

export default function () {
  getBlockByHash();
}

import { expect } from "chai";
import { ethers } from "hardhat";
import { SimpleStorage } from "../typechain-types";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";

describe("Hedera JSON RPC Integration Tests", function () {
  let simpleStorage: SimpleStorage;
  let owner: SignerWithAddress;
  let addr1: SignerWithAddress;

  beforeEach(async function () {
    console.log("🔄 Setting up test environment...");
    // Get signers
    [owner, addr1] = await ethers.getSigners();
    
    // Deploy the contract
    const SimpleStorageFactory = await ethers.getContractFactory("SimpleStorage");
    simpleStorage = await SimpleStorageFactory.deploy(42);
    await simpleStorage.waitForDeployment();
    console.log("✨ Test contract deployed successfully");
  });

  describe("Basic Operations", function () {
    it("Should execute a native token transfer", async function () {
      console.log("💸 Testing native token transfer...");
      // Get initial balances
      const initialBalance = await ethers.provider.getBalance(addr1.address);
      
      // Send 0.1 native token
      const tx = await owner.sendTransaction({
        to: addr1.address,
        value: ethers.parseEther("0.1")
      });
      await tx.wait();

      // Check new balance
      const newBalance = await ethers.provider.getBalance(addr1.address);
      expect(newBalance).to.be.gt(initialBalance);
    });

    it("Should deploy contract and verify initial value", async function () {
      console.log("🔍 Verifying initial contract value...");
      const value = await simpleStorage.getValue();
      expect(value).to.equal(42n);
    });

    it("Should make contract calls and emit events", async function () {
      console.log("📡 Testing contract calls and events...");
      // Test setting a new value
      const tx = await simpleStorage.setValue(100);
      await tx.wait();

      // Verify the new value
      const value = await simpleStorage.getValue();
      expect(value).to.equal(100n);

      // Verify event was emitted
      await expect(tx)
        .to.emit(simpleStorage, "ValueChanged")
        .withArgs(100n);
    });

    it("Should estimate gas for contract calls", async function () {
      console.log("⛽ Estimating gas for operations...");
      // Estimate gas for a simple operation
      const gasEstimateSimple = await simpleStorage.setValue.estimateGas(200);
      expect(gasEstimateSimple).to.be.gt(0);

      // Estimate gas for a more complex operation
      const gasEstimateComplex = await simpleStorage.expensiveOperation.estimateGas(100);
      expect(gasEstimateComplex).to.be.gt(gasEstimateSimple);
    });

    it("Should get transaction details and receipt", async function () {
      console.log("📝 Getting transaction details...");
      // Make a contract call
      const tx = await simpleStorage.setValue(150);
      const receipt = await tx.wait();

      if (!receipt) throw new Error("No receipt received");

      // Verify transaction receipt properties
      expect(receipt.status).to.equal(1); // Success
      expect(receipt.gasUsed).to.be.gt(0);
      
      // Get transaction details
      const txDetails = await ethers.provider.getTransaction(receipt.hash);
      if (!txDetails) throw new Error("No transaction details found");

      expect(txDetails.from).to.equal(owner.address);
      expect(txDetails.to).to.equal(await simpleStorage.getAddress());
    });
  });

  describe("Account Operations", function () {
    it("Should get account balance and transaction count", async function () {
      const balance = await ethers.provider.getBalance(owner.address);
      expect(balance).to.be.gt(0);

      const nonce = await ethers.provider.getTransactionCount(owner.address);
      expect(nonce).to.be.gte(0);
    });
  });
});

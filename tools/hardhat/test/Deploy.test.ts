import { expect } from "chai";
import { ethers } from "hardhat";
import { SimpleStorage } from "../typechain-types";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";

describe("Contract Deployment", function () {
  let deployer: SignerWithAddress;
  let simpleStorage: SimpleStorage;

  it("Should deploy SimpleStorage contract successfully", async function () {
    // Get the deployer account
    [deployer] = await ethers.getSigners();

    // Log deployment details
    console.log("Deploying contracts with the account:", deployer.address);
    const initialBalance = await ethers.provider.getBalance(deployer.address);
    console.log("Account balance:", initialBalance.toString());

    // Deploy the contract
    const SimpleStorageFactory = await ethers.getContractFactory("SimpleStorage");
    simpleStorage = await SimpleStorageFactory.deploy(42);
    await simpleStorage.waitForDeployment();

    // Get the deployed contract address
    const deployedAddress = await simpleStorage.getAddress();
    console.log("SimpleStorage deployed to:", deployedAddress);

    // Add some assertions to verify the deployment
    expect(deployedAddress).to.be.properAddress;
    expect(await simpleStorage.getValue()).to.equal(42n);
    
    // Verify deployer's balance changed (spent some gas)
    const finalBalance = await ethers.provider.getBalance(deployer.address);
    expect(finalBalance).to.be.lt(initialBalance);
  });
}); 
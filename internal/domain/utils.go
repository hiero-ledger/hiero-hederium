package domain

import (
	"regexp"
)

func IsValidAddress(address string) bool {
	return regexp.MustCompile("^0x[a-fA-F0-9]{40}$").MatchString(address)
}

func IsValidBlockNumberOrTag(blockNumber string) bool {
	return blockNumber == "latest" || blockNumber == "earliest" || blockNumber == "pending" || IsValidBlockNumber(blockNumber) || blockNumber == "finalized" || blockNumber == "safe"
}

func IsValidHexNumber(hexNumber string) bool {
	return regexp.MustCompile("^0x[a-fA-F0-9]*$").MatchString(hexNumber)
}

func IsValidBlockHashOrTag(blockHash string) bool {
	return regexp.MustCompile("^0x[a-fA-F0-9]{64}$").MatchString(blockHash) || blockHash == "latest" || blockHash == "earliest" || blockHash == "pending" || blockHash == "finalized" || blockHash == "safe"
}

func IsValidHexHash(hexHash string) bool {
	return regexp.MustCompile("^0x[a-fA-F0-9]{64}$").MatchString(hexHash)
}

func IsValidBlockHash(blockHash string) bool {
	return regexp.MustCompile("^0x[a-fA-F0-9]{64}$").MatchString(blockHash)
}

func IsValidBlock(block string) bool {
	return IsValidBlockNumberOrTag(block) || IsValidBlockHash(block)
}

func IsValidBlockNumber(blockNumber string) bool {
	return regexp.MustCompile("^0x[a-fA-F0-9]+$").MatchString(blockNumber)
}

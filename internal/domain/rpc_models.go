package domain

// Block represents an Ethereum-compatible block structure
type Block struct {
	Number           *string       `json:"number"`           // The block number (hex)
	Hash             *string       `json:"hash"`             // The block hash
	ParentHash       string        `json:"parentHash"`       // Hash of parent block
	Nonce            string        `json:"nonce"`            // Nonce used in the block
	Sha3Uncles       string        `json:"sha3Uncles"`       // Keccak hash of uncles data
	LogsBloom        string        `json:"logsBloom"`        // Bloom filter for the logs
	TransactionsRoot *string       `json:"transactionsRoot"` // Root of transaction trie
	StateRoot        string        `json:"stateRoot"`        // Root of final state trie
	ReceiptsRoot     string        `json:"receiptsRoot"`     // Root of receipts trie
	Miner            string        `json:"miner"`            // The address of the beneficiary
	Difficulty       string        `json:"difficulty"`       // Integer of the difficulty
	TotalDifficulty  string        `json:"totalDifficulty"`  // Integer of total difficulty
	ExtraData        string        `json:"extraData"`        // Extra data field
	Size             string        `json:"size"`             // Size of block in bytes
	GasLimit         string        `json:"gasLimit"`         // Maximum gas allowed
	GasUsed          string        `json:"gasUsed"`          // Total gas used
	Timestamp        string        `json:"timestamp"`        // Unix timestamp
	Transactions     []interface{} `json:"transactions"`     // Array of transaction objects or hashes
	Uncles           []string      `json:"uncles"`           // Array of uncle hashes
	Withdrawals      []string      `json:"withdrawals"`      // Array of withdrawal objects
	WithdrawalsRoot  string        `json:"withdrawalsRoot"`  // Root of withdrawals trie
	BaseFeePerGas    string        `json:"baseFeePerGas"`    // Base fee per gas
	MixHash          string        `json:"mixHash"`          // Mix hash
}

// Transaction represents an Ethereum-compatible transaction structure
type Transaction struct {
	BlockHash        *string `json:"blockHash"`
	BlockNumber      *string `json:"blockNumber"`
	From             string  `json:"from"`
	Gas              string  `json:"gas"`
	GasPrice         string  `json:"gasPrice"`
	Hash             string  `json:"hash"`
	Input            string  `json:"input"`
	Nonce            string  `json:"nonce"`
	To               *string `json:"to"`
	TransactionIndex *string `json:"transactionIndex"`
	Value            string  `json:"value"`
	V                string  `json:"v"`
	R                string  `json:"r"`
	S                string  `json:"s"`
	ChainId          *string `json:"chainId,omitempty"`
	Type             string  `json:"type"`
}

// Transaction2930 represents an EIP-2930 transaction
type Transaction2930 struct {
	Transaction
	AccessList []AccessListEntry `json:"accessList"`
}

// Transaction1559 represents an EIP-1559 transaction
type Transaction1559 struct {
	Transaction
	AccessList           []AccessListEntry `json:"accessList"`
	MaxPriorityFeePerGas string            `json:"maxPriorityFeePerGas"`
	MaxFeePerGas         string            `json:"maxFeePerGas"`
}

// AccessListEntry represents an entry in the access list
type AccessListEntry struct {
	Address     string   `json:"address"`
	StorageKeys []string `json:"storageKeys"`
}

type TransactionCallObject struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Value    string `json:"value"`
	Data     string `json:"data"`
	Input    string `json:"input"`
	Nonce    string `json:"nonce"`
	Estimate bool   `json:"estimate"`
}

// NewBlock creates a new Block instance with default values for non-nullable fields
func NewBlock() *Block {
	return &Block{
		ReceiptsRoot: "0x41f639e5f179099843a6b73fdf71f0fc8b4fb7de9dba6a98e902c082236e13f3",
		Miner:        "0x0000000000000000000000000000000000000000",
		Nonce:        "0x0000000000000000",
		StateRoot:    "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
		Sha3Uncles:   "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347", // Keccak-256 hash for empty array
		LogsBloom:    "0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		Difficulty:   "0x0",
		ExtraData:    "0x",
		Transactions: make([]interface{}, 0),
		Uncles:       make([]string, 0),
	}
}

type Log struct {
	Address          string   `json:"address"`
	BlockHash        string   `json:"blockHash"`
	BlockNumber      string   `json:"blockNumber"`
	Data             string   `json:"data"`
	LogIndex         string   `json:"logIndex"`
	Removed          bool     `json:"removed"`
	Topics           []string `json:"topics"`
	TransactionHash  string   `json:"transactionHash"`
	TransactionIndex string   `json:"transactionIndex"`
}

type TransactionReceipt struct {
	BlockHash         string  `json:"blockHash"`
	BlockNumber       string  `json:"blockNumber"`
	ContractAddress   string  `json:"contractAddress"`
	CumulativeGasUsed string  `json:"cumulativeGasUsed"`
	EffectiveGasPrice string  `json:"effectiveGasPrice"`
	From              string  `json:"from"`
	GasUsed           string  `json:"gasUsed"`
	Logs              []Log   `json:"logs"`
	LogsBloom         string  `json:"logsBloom"`
	Root              string  `json:"root"`
	Status            string  `json:"status"`
	To                string  `json:"to"`
	TransactionHash   string  `json:"transactionHash"`
	TransactionIndex  string  `json:"transactionIndex"`
	Type              *string `json:"type"`
	RevertReason      string  `json:"revertReason,omitempty"`
}

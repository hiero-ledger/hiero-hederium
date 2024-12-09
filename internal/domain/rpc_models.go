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
	TotalDifficulty  *string       `json:"totalDifficulty"`  // Integer of total difficulty
	ExtraData        string        `json:"extraData"`        // Extra data field
	Size             string        `json:"size"`             // Size of block in bytes
	GasLimit         string        `json:"gasLimit"`         // Maximum gas allowed
	GasUsed          string        `json:"gasUsed"`          // Total gas used
	Timestamp        string        `json:"timestamp"`        // Unix timestamp
	Transactions     []interface{} `json:"transactions"`     // Array of transaction objects or hashes
	Uncles           []string      `json:"uncles"`           // Array of uncle hashes
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
}

// NewBlock creates a new Block instance with default values for non-nullable fields
func NewBlock() *Block {
	return &Block{
		ReceiptsRoot: "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
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

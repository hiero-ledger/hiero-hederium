package domain

type FeeResponse struct {
	Fees      []Fee  `json:"fees"`
	Timestamp string `json:"timestamp"`
}

type Fee struct {
	Gas             int64  `json:"gas"`
	TransactionType string `json:"transaction_type"`
}

type BlockResponse struct {
	Count        int       `json:"count"`
	HapiVersion  string    `json:"hapi_version"`
	Hash         string    `json:"hash"`
	Name         string    `json:"name"`
	Number       int       `json:"number"`
	PreviousHash string    `json:"previous_hash"`
	Size         int       `json:"size"`
	Timestamp    Timestamp `json:"timestamp"`
	GasUsed      int       `json:"gas_used"`
	LogsBloom    string    `json:"logs_bloom"`
}

type Timestamp struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ContractResults struct {
	Address              string   `json:"address"`
	Amount               int      `json:"amount"`
	Bloom                string   `json:"bloom"`
	CallResult           string   `json:"call_result"`
	ContractID           string   `json:"contract_id"`
	CreatedContractIDs   []string `json:"created_contract_ids"`
	ErrorMessage         *string  `json:"error_message"`
	From                 string   `json:"from"`
	FunctionParameters   string   `json:"function_parameters"`
	GasConsumed          int64    `json:"gas_consumed"`
	GasLimit             int64    `json:"gas_limit"`
	GasUsed              int64    `json:"gas_used"`
	Timestamp            string   `json:"timestamp"`
	To                   string   `json:"to"`
	Hash                 string   `json:"hash"`
	BlockHash            string   `json:"block_hash"`
	BlockNumber          int64    `json:"block_number"`
	Result               string   `json:"result"`
	TransactionIndex     int      `json:"transaction_index"`
	Status               string   `json:"status"`
	FailedInitcode       *string  `json:"failed_initcode"`
	AccessList           string   `json:"access_list"`
	BlockGasUsed         int64    `json:"block_gas_used"`
	ChainID              string   `json:"chain_id"`
	GasPrice             string   `json:"gas_price"`
	MaxFeePerGas         string   `json:"max_fee_per_gas"`
	MaxPriorityFeePerGas string   `json:"max_priority_fee_per_gas"`
	R                    string   `json:"r"`
	S                    string   `json:"s"`
	Type                 int      `json:"type"`
	V                    int      `json:"v"`
	Nonce                int64    `json:"nonce"`
}

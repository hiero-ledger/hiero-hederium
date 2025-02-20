package domain

type FeeResponse struct {
	Fees      []Fee  `json:"fees"`
	Timestamp string `json:"timestamp"`
}

type Fee struct {
	Gas             int64  `json:"gas"`
	TransactionType string `json:"transaction_type"`
}

type FeeHistory struct {
	BaseFeePerGas []string   `json:"base_fee_per_gas"`
	GasUsedRatio  []float64  `json:"gas_used_ratio"`
	OldestBlock   string     `json:"oldest_block"`
	Reward        [][]string `json:"reward,omitempty"`
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

type ContractResultResponse struct {
	Address              string          `json:"address"`
	Amount               int             `json:"amount"`
	Bloom                string          `json:"bloom"`
	CallResult           string          `json:"call_result"`
	ContractID           string          `json:"contract_id"`
	CreatedContractIDs   []string        `json:"created_contract_ids"`
	ErrorMessage         *string         `json:"error_message"`
	From                 string          `json:"from"`
	FunctionParameters   string          `json:"function_parameters"`
	GasConsumed          int64           `json:"gas_consumed"`
	GasLimit             int64           `json:"gas_limit"`
	GasUsed              int64           `json:"gas_used"`
	Timestamp            string          `json:"timestamp"`
	To                   string          `json:"to"`
	Hash                 string          `json:"hash"`
	BlockHash            string          `json:"block_hash"`
	BlockNumber          int64           `json:"block_number"`
	Logs                 []MirroNodeLogs `json:"logs"`
	Result               string          `json:"result"`
	TransactionIndex     int             `json:"transaction_index"`
	Status               string          `json:"status"`
	FailedInitcode       *string         `json:"failed_initcode"`
	AccessList           string          `json:"access_list"`
	BlockGasUsed         int64           `json:"block_gas_used"`
	ChainID              string          `json:"chain_id"`
	GasPrice             string          `json:"gas_price"`
	MaxFeePerGas         string          `json:"max_fee_per_gas"`
	MaxPriorityFeePerGas string          `json:"max_priority_fee_per_gas"`
	R                    string          `json:"r"`
	S                    string          `json:"s"`
	Type                 *int            `json:"type"`
	V                    int             `json:"v"`
	Nonce                int64           `json:"nonce"`
	StateChanges         []struct {
		Address      string `json:"address"`
		ContractID   string `json:"contract_id"`
		Slot         string `json:"slot"`
		ValueRead    string `json:"value_read"`
		ValueWritten string `json:"value_written"`
	} `json:"state_changes"`
}

type MirroNodeLogs struct {
	Address    string   `json:"address"`
	Bloom      string   `json:"bloom"`
	ContractID string   `json:"contract_id"`
	Data       string   `json:"data"`
	Index      int      `json:"index"`
	Topics     []string `json:"topics"`
}

type AccountResponse struct {
	Account         string `json:"account"`
	Alias           string `json:"alias"`
	AutoRenewPeriod int64  `json:"auto_renew_period"`
	Balance         struct {
		Balance   int64         `json:"balance"`
		Timestamp string        `json:"timestamp"`
		Tokens    []interface{} `json:"tokens"`
	} `json:"balance"`
	CreatedTimestamp string `json:"created_timestamp"`
	DeclineReward    bool   `json:"decline_reward"`
	Deleted          bool   `json:"deleted"`
	EthereumNonce    int64  `json:"ethereum_nonce"`
	EvmAddress       string `json:"evm_address"`
	ExpiryTimestamp  string `json:"expiry_timestamp"`
	Key              struct {
		Type string `json:"_type"`
		Key  string `json:"key"`
	} `json:"key"`
	MaxAutomaticTokenAssociations int         `json:"max_automatic_token_associations"`
	Memo                          string      `json:"memo"`
	PendingReward                 int64       `json:"pending_reward"`
	ReceiverSigRequired           bool        `json:"receiver_sig_required"`
	StakedAccountId               interface{} `json:"staked_account_id"`
	StakedNodeId                  interface{} `json:"staked_node_id"`
	StakePeriodStart              interface{} `json:"stake_period_start"`
	Transactions                  []struct {
		Bytes                    interface{}   `json:"bytes"`
		ChargedTxFee             int64         `json:"charged_tx_fee"`
		ConsensusTimestamp       string        `json:"consensus_timestamp"`
		EntityId                 string        `json:"entity_id"`
		MaxFee                   string        `json:"max_fee"`
		MemoBase64               string        `json:"memo_base64"`
		Name                     string        `json:"name"`
		NftTransfers             []interface{} `json:"nft_transfers"`
		Node                     string        `json:"node"`
		Nonce                    int           `json:"nonce"`
		ParentConsensusTimestamp interface{}   `json:"parent_consensus_timestamp"`
		Result                   string        `json:"result"`
		Scheduled                bool          `json:"scheduled"`
		StakingRewardTransfers   []interface{} `json:"staking_reward_transfers"`
		TokenTransfers           []interface{} `json:"token_transfers"`
		TransactionHash          string        `json:"transaction_hash"`
		TransactionId            string        `json:"transaction_id"`
		Transfers                []struct {
			Account    string `json:"account"`
			Amount     int64  `json:"amount"`
			IsApproval bool   `json:"is_approval"`
		} `json:"transfers"`
		ValidDurationSeconds string `json:"valid_duration_seconds"`
		ValidStartTimestamp  string `json:"valid_start_timestamp"`
	} `json:"transactions"`
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
}

type ContractState struct {
	Value string `json:"value"`
}

type ContractStateResponse struct {
	State []ContractState `json:"state"`
}

type LogParams struct {
	BlockHash string
	FromBlock string
	ToBlock   string
	Address   []string
	Topics    []string
}

type ContractResultsLogResponse struct {
	Logs  []LogEntry `json:"logs"`
	Links struct {
		Next *string `json:"next"`
	} `json:"links"`
}

type LogEntry struct {
	Address          string   `json:"address"`
	Bloom            string   `json:"bloom"`
	ContractID       string   `json:"contract_id"`
	Data             string   `json:"data"`
	Index            *int     `json:"index"`
	Topics           []string `json:"topics"`
	BlockHash        string   `json:"block_hash"`
	BlockNumber      *int64   `json:"block_number"`
	RootContractID   string   `json:"root_contract_id"`
	Timestamp        string   `json:"timestamp"`
	TransactionHash  string   `json:"transaction_hash"`
	TransactionIndex *int     `json:"transaction_index"`
}

type ContractResponse struct {
	AdminKey                      ProtobufEncodedKey `json:"admin_key"`
	AutoRenewAccount              *string            `json:"auto_renew_account"`
	AutoRenewPeriod               int64              `json:"auto_renew_period"`
	ContractID                    string             `json:"contract_id"`
	CreatedTimestamp              string             `json:"created_timestamp"`
	Deleted                       bool               `json:"deleted"`
	EvmAddress                    string             `json:"evm_address"`
	ExpirationTimestamp           string             `json:"expiration_timestamp"`
	FileID                        *string            `json:"file_id"`
	MaxAutomaticTokenAssociations int                `json:"max_automatic_token_associations"`
	Memo                          string             `json:"memo"`
	Nonce                         int64              `json:"nonce"`
	ObtainerID                    *string            `json:"obtainer_id"`
	PermanentRemoval              *string            `json:"permanent_removal"`
	ProxyAccountID                *string            `json:"proxy_account_id"`
	Timestamp                     Timestamp          `json:"timestamp"`
	Bytecode                      *string            `json:"bytecode"`
	RuntimeBytecode               *string            `json:"runtime_bytecode"`
}

type TokenResponse struct {
	AdminKey          ProtobufEncodedKey `json:"admin_key"`
	AutoRenewAccount  string             `json:"auto_renew_account"`
	AutoRenewPeriod   *string            `json:"auto_renew_period"`
	CreatedTimestamp  string             `json:"created_timestamp"`
	Deleted           bool               `json:"deleted"`
	Decimals          int                `json:"decimals"`
	ExpiryTimestamp   *string            `json:"expiry_timestamp"`
	FreezeDefault     bool               `json:"freeze_default"`
	FreezeKey         ProtobufEncodedKey `json:"freeze_key"`
	InitialSupply     int                `json:"initial_supply"`
	KycKey            ProtobufEncodedKey `json:"kyc_key"`
	MaxSupply         int64              `json:"max_supply"`
	Memo              string             `json:"memo"`
	ModifiedTimestamp string             `json:"modified_timestamp"`
	Name              string             `json:"name"`
	PauseKey          ProtobufEncodedKey `json:"pause_key"`
	PauseStatus       string             `json:"pause_status"`
	SupplyKey         ProtobufEncodedKey `json:"supply_key"`
	SupplyType        string             `json:"supply_type"`
	Symbol            string             `json:"symbol"`
	TokenId           string             `json:"token_id"`
	TotalSupply       int                `json:"total_supply"`
	TreasuryAccountId string             `json:"treasury_account_id"`
	Type              string             `json:"type"`
	WipeKey           ProtobufEncodedKey `json:"wipe_key"`
	CustomFees        CustomFees         `json:"custom_fees"`
}

type ProtobufEncodedKey struct {
	Type string `json:"_type"`
	Key  string `json:"key"`
}

type CustomFees struct {
	CreatedTimestamp string          `json:"created_timestamp"`
	FixedFees        []FixedFee      `json:"fixed_fees"`
	FractionalFees   []FractionalFee `json:"fractional_fees"`
}

type FixedFee struct {
	Amount              int    `json:"amount"`
	CollectorAccountId  string `json:"collector_account_id"`
	DenominatingTokenId string `json:"denominating_token_id"`
}

type FractionalFee struct {
	Amount              `json:"amount"`
	CollectorAccountId  string `json:"collector_account_id"`
	DenominatingTokenId string `json:"denominating_token_id"`
	Maximum             int    `json:"maximum"`
	Minimum             int    `json:"minimum"`
	NetOfTransfers      bool   `json:"net_of_transfers"`
}

type Amount struct {
	Numerator   int `json:"numerator"`
	Denominator int `json:"denominator"`
}

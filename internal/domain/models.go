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
	Count        int    `json:"count"`
	HapiVersion  string `json:"hapi_version"`
	Hash         string `json:"hash"`
	Name         string `json:"name"`
	Number       int    `json:"number"`
	PreviousHash string `json:"previous_hash"`
	Size         int    `json:"size"`
	Timestamp    struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"timestamp"`
	GasUsed   int    `json:"gas_used"`
	LogsBloom string `json:"logs_bloom"`
}

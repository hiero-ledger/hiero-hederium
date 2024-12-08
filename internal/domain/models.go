package domain

type FeeResponse struct {
	Fees      []Fee  `json:"fees"`
	Timestamp string `json:"timestamp"`
}

type Fee struct {
	Gas             int64  `json:"gas"`
	TransactionType string `json:"transaction_type"`
}

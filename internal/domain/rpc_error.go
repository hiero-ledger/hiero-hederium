package domain

import "fmt"

// Standard JSON-RPC 2.0 error codes
const (
	// Parse error (-32700): Invalid JSON was received by the server.
	ParseError = -32700

	// Invalid Request (-32600): The JSON sent is not a valid Request object.
	InvalidRequest = -32600

	// Method not found (-32601): The method does not exist / is not available.
	MethodNotFound = -32601

	// Invalid params (-32602): Invalid method parameter(s).
	InvalidParams = -32602

	// Internal error (-32603): Internal JSON-RPC error.
	InternalError = -32603

	// Server error (-32000 to -32099): Implementation-defined server errors.
	ServerError = -32000

	// Not found (-32001): Not found
	NotFound = -32001

	// Gas price too low (-32009): Gas price is too low
	GasPriceTooLow = -32009

	// Execution error (-32015): Transaction execution error
	ExecutionError = -32015

	// Nonce too low (32001): Nonce is too low
	NonceTooLow = 32001

	// Nonce too high (32002): Nonce is too high
	NonceTooHigh = 32002

	// Insufficient funds (-32018): Insufficient funds for transfer
	InsufficientFunds = -32018

	// Invalid block range (-39013): Invalid block range
	InvalidBlockRange = -39013

	// Timestamp range too large (-32004): The provided fromBlock and toBlock contain timestamps that exceed the maximum allowed duration of 7 days (604800 seconds)
	InvalidTimestampRange = -32004

	// Unsupported transaction type (-32611): Unsupported transaction type
	UnsupportedTransactionType = -32611

	// Missing from block param (-32011): Missing from block param
	MissingFromBlockParam = -32011

	// Oversized data (-32201): Oversized data
	OversizedData = -32201

	// Gas limit too high (-32005): Gas limit is too high
	GasLimitTooHigh = -32005

	// Gas limit too low (-32003): Gas limit is too low
	GasLimitTooLow = -32003

	// ContractRevertError (3): Contract reverted
	ContractRevertError = 3
)

// RPCError represents a JSON-RPC 2.0 error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *RPCError) Error() string {
	return e.Message
}

// NewRPCError creates a new RPCError without additional data
func NewRPCError(code int, message string) *RPCError {
	return &RPCError{
		Code:    code,
		Message: message,
	}
}

// Common error constructors
func NewParseError(msg string) *RPCError {
	return NewRPCError(ParseError, msg)
}

func NewInvalidRequestError(msg string) *RPCError {
	return NewRPCError(InvalidRequest, msg)
}

func NewMethodNotFoundError(method string) *RPCError {
	return NewRPCError(MethodNotFound, fmt.Sprintf("Method not found: %s", method))
}

func NewInvalidParamsError(msg string) *RPCError {
	return NewRPCError(InvalidParams, msg)
}

func NewInternalError(msg string) *RPCError {
	return NewRPCError(InternalError, msg)
}

func NewServerError(msg string) *RPCError {
	return NewRPCError(ServerError, msg)
}

func NewExecutionError(msg string) *RPCError {
	return NewRPCError(ExecutionError, msg)
}

func NewNonceTooLowError() *RPCError {
	return NewRPCError(NonceTooLow, "nonce too low")
}

func NewGasPriceTooLowError() *RPCError {
	return NewRPCError(GasPriceTooLow, "gas price too low")
}

func NewInsufficientFundsError() *RPCError {
	return NewRPCError(InsufficientFunds, "insufficient funds for transfer")
}

func NewUnsupportedMethodError(method string) *RPCError {
	return NewRPCError(MethodNotFound, fmt.Sprintf("Method not supported: %s", method))
}

func NewInvalidBlockRangeError() *RPCError {
	return NewRPCError(InvalidBlockRange, "Invalid block range")
}

func NewNotFoundError() *RPCError {
	return NewRPCError(NotFound, "Requested resource not found")
}

func NewTimeStampRangeTooLargeError(fromBlock, toBlock string, fromTimestamp, toTimestamp float64) *RPCError {
	return NewRPCError(InvalidTimestampRange, fmt.Sprintf("The provided fromBlock and toBlock contain timestamps that exceed the maximum allowed duration of 7 days (604800 seconds): fromBlock: %s (%f), toBlock: %s (%f)", fromBlock, fromTimestamp, toBlock, toTimestamp))
}

func NewRangeTooLarge(blockRange int) *RPCError {
	return NewRPCError(ServerError, fmt.Sprintf("Exceeded maximum block range: %d", blockRange))
}

func NewUnsupportedJSONRPCMethodError() *RPCError {
	return NewRPCError(MethodNotFound, "Unsupported JSON-RPC method")
}

func NewContractRevertError(message string) *RPCError {
	return NewRPCError(ContractRevertError, fmt.Sprintf("execution reverted: %s", message))
}

func NewFilterNotFoundError() *RPCError {
	return NewRPCError(NotFound, "Filter not found")
}

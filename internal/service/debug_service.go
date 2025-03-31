package service

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/LimeChain/Hederium/internal/domain"
	infrahedera "github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"go.uber.org/zap"
)

type DebugServicer interface {
	DebugTraceTransaction(transactionIDOrHash string, tracer string, tracerConfig interface{}) (interface{}, *domain.RPCError)
}

const (
	// CallTracer tracks all the call frames executed during a transaction
	CallTracerType string = "callTracer"
	// OpcodeLogger executes a transaction and emits the opcodes and context at every step
	OpcodeLoggerType string = "opcodeLogger"
)

// DebugService provides functionality for tracing and debugging transactions
type DebugService struct {
	mClient          infrahedera.MirrorNodeClient
	logger           *zap.Logger
	isServiceEnabled bool
	ethService       *EthService
}

// NewDebugService creates a new instance of DebugService
func NewDebugService(mClient infrahedera.MirrorNodeClient, logger *zap.Logger, isServiceEnabled bool, ethService *EthService) *DebugService {
	return &DebugService{mClient: mClient, logger: logger, isServiceEnabled: isServiceEnabled, ethService: ethService}
}

// DebugTraceTransaction traces a transaction for debugging purposes
func (d *DebugService) DebugTraceTransaction(transactionIDOrHash string, tracer string, tracerConfig interface{}) (interface{}, *domain.RPCError) {
	d.logger.Debug("Calling DebugTraceTransaction", zap.String("transactionIDOrHash", transactionIDOrHash), zap.String("tracer", tracer), zap.Any("tracerConfig", tracerConfig))

	if !d.isServiceEnabled {
		return nil, domain.NewUnsupportedJSONRPCMethodError()
	}

	var result interface{}
	var err error

	switch tracer {
	case CallTracerType:
		config, ok := tracerConfig.(domain.CallTracerConfig)
		if !ok {
			return nil, domain.NewInternalError("Invalid tracer configuration for CallTracer")
		}
		result, err = d.CallTracer(transactionIDOrHash, &config)
	case OpcodeLoggerType:
		config, ok := tracerConfig.(domain.OpcodeLoggerConfig)
		if !ok {
			return nil, domain.NewInternalError("Invalid tracer configuration for OpcodeLogger")
		}
		result, err = d.CallOpcodeLogger(transactionIDOrHash, &config)
	default:
		return nil, domain.NewUnsupportedJSONRPCMethodError()
	}

	if err != nil {
		switch err := err.(type) {
		case *domain.RPCError:
			return nil, err
		default:
			return nil, domain.NewInternalError(fmt.Sprintf("Failed to trace transaction: %v", err))
		}
	}

	return result, nil
}

// FormatActionsResult formats the result from the actions endpoint
func (d *DebugService) FormatActionsResult(actions []domain.Action) []domain.ContractAction {
	formattedResults := make([]domain.ContractAction, 0, len(actions))
	d.logger.Info("Formatting actions result", zap.Any("actions", actions))
	for i, action := range actions {
		d.logger.Info("Formatting action", zap.String("from", action.From), zap.String("to", action.To))

		// We do not care if the address is empty
		from, _ := d.ethService.resolveEvmAddress(action.From)
		to, _ := d.ethService.resolveEvmAddress(action.To)

		var input, output string

		input = action.Input
		output = action.ResultData

		if i != 0 && action.CallOperationType == "CREATE" {
			if contractResp, err := d.mClient.GetContractById(action.To); err == nil && contractResp != nil {
				input = *contractResp.Bytecode
				output = *contractResp.RuntimeBytecode
			}
		}

		contractAction := domain.ContractAction{
			Type:    action.CallOperationType,
			From:    *from,
			To:      *to,
			Gas:     fmt.Sprintf("0x%x", action.Gas),
			GasUsed: fmt.Sprintf("0x%x", action.GasUsed),
			Value:   fmt.Sprintf("0x%x", action.Value),
			Input:   input,
			Output:  output,
		}

		formattedResults = append(formattedResults, contractAction)
	}

	return formattedResults
}

// FormatOpcodesResult formats the result from the opcodes endpoint
func (d *DebugService) FormatOpcodesResult(result *domain.OpcodesResponse, options *domain.OpcodeLoggerConfig) *domain.OpcodesResponse {
	if result == nil {
		return &domain.OpcodesResponse{
			Gas:         0,
			Failed:      true,
			ReturnValue: "",
			Opcodes:     []domain.Opcode{},
		}
	}

	result.ReturnValue = strings.TrimPrefix(result.ReturnValue, "0x")

	for _, opcode := range result.Opcodes {
		for _, stackItem := range opcode.Stack {
			opcode.Stack = append(opcode.Stack, strings.TrimPrefix(stackItem, "0x"))
		}

		for _, memoryItem := range opcode.Memory {
			opcode.Memory = append(opcode.Memory, strings.TrimPrefix(memoryItem, "0x"))
		}

		for key, value := range opcode.Storage {
			opcode.Storage[key] = strings.TrimPrefix(value, "0x")
		}

		opcode.Reason = strings.TrimPrefix(opcode.Reason, "0x")
	}

	return result
}

// CallOpcodeLogger implements the OpcodeLogger tracer
func (d *DebugService) CallOpcodeLogger(transactionIdOrHash string, tracerConfig *domain.OpcodeLoggerConfig) (*domain.OpcodesResponse, error) {
	d.logger.Info("Calling CallOpcodeLogger", zap.Any("tracerConfig", tracerConfig))

	options := map[string]interface{}{
		"memory":  tracerConfig.EnableMemory,
		"stack":   !tracerConfig.DisableStack,
		"storage": !tracerConfig.DisableStorage,
	}

	response, err := d.mClient.GetContractsResultsOpcodes(transactionIdOrHash, options)
	if err != nil {
		return nil, err
	}

	return d.FormatOpcodesResult(response, tracerConfig), nil
}

func (d *DebugService) CallTracer(transactionHash string, tracerConfig *domain.CallTracerConfig) (*domain.CallTracerResult, error) {
	d.logger.Info("Calling CallTracer", zap.Any("tracerConfig", tracerConfig))
	actionsResponse, err := d.mClient.GetContractsResultsActions(transactionHash)
	if err != nil {
		return nil, err
	}

	response := d.mClient.GetContractResult(transactionHash)

	if actionsResponse == nil || response == nil {
		return nil, domain.NewRPCError(domain.NotFound, fmt.Sprintf("Requested resource not found. Failed to retrieve contract results for transaction %s", transactionHash))
	}

	transactionsResponse := response.(domain.ContractResultResponse)

	actions := d.FormatActionsResult(actionsResponse.Actions)

	from, _ := d.ethService.resolveEvmAddress(transactionsResponse.From)
	to, _ := d.ethService.resolveEvmAddress(transactionsResponse.To)

	value := zeroHex

	if transactionsResponse.Amount != 0 {
		value = fmt.Sprintf("0x%x", transactionsResponse.Amount)
	}

	var revertReason, errResult string
	output := transactionsResponse.CallResult

	if transactionsResponse.Result != "SUCCESS" && transactionsResponse.ErrorMessage != nil {
		errResult = transactionsResponse.Result
		output = *transactionsResponse.ErrorMessage
		revertReason, _ = decodeRevertReason(*transactionsResponse.ErrorMessage)
	}

	// If we have more than one call executed during the transactions we would return all calls
	// except the first one in the sub-calls array,
	// therefore we need to exclude the first one from the actions response
	if (tracerConfig.OnlyTopCall || len(actionsResponse.Actions) == 1) && len(actionsResponse.Actions) > 1 {
		actions = []domain.ContractAction{}
	} else {
		actions = actions[1:]
	}

	return &domain.CallTracerResult{
		Type:         actionsResponse.Actions[0].CallType,
		From:         *from,
		To:           *to,
		Value:        value,
		Gas:          fmt.Sprintf("0x%x", transactionsResponse.GasLimit),
		GasUsed:      fmt.Sprintf("0x%x", transactionsResponse.GasUsed),
		Input:        transactionsResponse.FunctionParameters,
		Output:       output,
		Error:        errResult,
		RevertReason: revertReason,
		Calls:        actions,
	}, nil
}

func decodeRevertReason(str string) (string, error) {
	if !strings.HasPrefix(str, "0x") {
		return str, nil
	}

	str = strings.TrimPrefix(str, "0x")

	bytes, err := hex.DecodeString(str)
	if err != nil {
		return "", err
	}

	// Extract the string part from the decoded bytes
	offset := 4 + 32 // Skip first 4 bytes (method signature) + next 32 bytes (offset)
	if len(bytes) < offset+32 {
		return "", fmt.Errorf("invalid revert data length")
	}

	length := int(bytes[offset+31]) // Length is stored in the last byte of the next 32-byte chunk
	start := offset + 32
	end := start + length

	if len(bytes) < end {
		return "", fmt.Errorf("invalid string length")
	}

	return string(bytes[start:end]), nil
}

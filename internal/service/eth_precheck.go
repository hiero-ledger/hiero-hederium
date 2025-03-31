package service

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/LimeChain/Hederium/internal/domain"
	infrahedera "github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

type Precheck interface {
	ParseTxIfNeeded(transaction interface{}) *types.Transaction
	Value(tx *types.Transaction) error
	SendRawTransactionCheck(parsedTx *types.Transaction, networkGasPriceInWeiBars int64) *domain.RPCError
	VerifyAccount(tx *types.Transaction) (*domain.AccountResponse, error)
	Nonce(tx *types.Transaction, accountInfoNonce int64) *domain.RPCError
	ChainID(tx *types.Transaction) error
	GasPrice(tx *types.Transaction, networkGasPriceInWeiBars int64) *domain.RPCError
	Balance(tx *types.Transaction, account *domain.AccountResponse) *domain.RPCError
	GasLimit(tx *types.Transaction) *domain.RPCError
	CheckSize(transaction string) *domain.RPCError
	TransactionType(tx *types.Transaction) error
	ReceiverAccount(tx *types.Transaction) *domain.RPCError
}

type precheck struct {
	mClient infrahedera.MirrorNodeClient
	logger  *zap.Logger
	chainID string
}

func NewPrecheck(mClient infrahedera.MirrorNodeClient, logger *zap.Logger, chainID string) Precheck {
	return &precheck{
		mClient: mClient,
		logger:  logger,
		chainID: chainID,
	}
}

func (p *precheck) ParseTxIfNeeded(transaction interface{}) *types.Transaction {
	if txStr, ok := transaction.(string); ok {
		tx, _, err := ParseTransaction(txStr)
		if err != nil {
			return nil
		}
		return tx
	}
	if tx, ok := transaction.(*types.Transaction); ok {
		return tx
	}
	return nil
}

func (p *precheck) Value(tx *types.Transaction) error {
	value := tx.Value()
	if (value.Cmp(big.NewInt(0)) > 0 && value.Cmp(big.NewInt(TinybarToWeibarCoef)) < 0) || value.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("value can't be non-zero and less than 10_000_000_000 wei which is 1 tinybar")
	}
	return nil
}

func (p *precheck) SendRawTransactionCheck(parsedTx *types.Transaction, networkGasPriceInWeiBars int64) *domain.RPCError {

	if err := p.TransactionType(parsedTx); err != nil {
		return domain.NewRPCError(domain.UnsupportedTransactionType, err.Error())
	}
	if errRpc := p.GasLimit(parsedTx); errRpc != nil {
		return errRpc
	}

	mirrorAccountInfo, err := p.VerifyAccount(parsedTx)
	if err != nil {
		return domain.NewRPCError(domain.NotFound, err.Error())
	}

	p.logger.Info("Account info", zap.Any("accountInfo", mirrorAccountInfo))

	if errRpc := p.Nonce(parsedTx, mirrorAccountInfo.EthereumNonce); errRpc != nil {
		return errRpc
	}
	if err := p.ChainID(parsedTx); err != nil {
		return domain.NewRPCError(domain.ServerError, err.Error())
	}
	if err := p.Value(parsedTx); err != nil {
		return domain.NewRPCError(domain.InvalidParams, err.Error())
	}
	if errRpc := p.GasPrice(parsedTx, networkGasPriceInWeiBars); errRpc != nil {
		return errRpc
	}
	if errRpc := p.Balance(parsedTx, mirrorAccountInfo); errRpc != nil {
		return errRpc
	}
	if errRpc := p.ReceiverAccount(parsedTx); errRpc != nil {
		return errRpc
	}

	return nil
}

func (p *precheck) VerifyAccount(tx *types.Transaction) (*domain.AccountResponse, error) {
	signer := types.LatestSignerForChainID(tx.ChainId())

	from, err := types.Sender(signer, tx)
	if err != nil {
		p.logger.Info("Verify account precheck failed", zap.Error(err))
		return nil, err
	}

	accountInfo, err := p.mClient.GetAccountById(from.Hex())
	if err != nil {
		p.logger.Debug("Failed to retrieve address account details", zap.String("address", from.Hex()), zap.Error(err))
		return nil, err
	}

	if accountInfo == nil {
		p.logger.Debug("Failed to retrieve address account details", zap.String("address", from.Hex()))
		return nil, fmt.Errorf("requested resource not found. address '%s'", from.Hex())
	}

	return accountInfo, nil
}

func (p *precheck) Nonce(tx *types.Transaction, accountInfoNonce int64) *domain.RPCError {
	txNonce := tx.Nonce()
	p.logger.Debug("Nonce precheck", zap.Uint64("tx.nonce", txNonce), zap.Int64("accountInfoNonce", accountInfoNonce))

	if uint64(accountInfoNonce) > txNonce {
		p.logger.Debug("Nonce too low", zap.Uint64("tx.nonce", txNonce), zap.Any("accountInfoNonce", uint64(accountInfoNonce)))
		return domain.NewRPCError(domain.NonceTooLow, fmt.Sprintf("Nonce too low. Provided nonce: %d, current nonce: %d", txNonce, accountInfoNonce))
	}

	return nil
}

func (p *precheck) isLegacyUnprotectedEtx(tx *types.Transaction) bool {
	v, _, _ := tx.RawSignatureValues()

	p.logger.Debug("isLegacyUnprotectedEtx", zap.Int64("chainId", tx.ChainId().Int64()), zap.Bool("result", tx.ChainId().Int64() == 0), zap.Uint64("v", v.Uint64()))
	return tx.ChainId() != nil && tx.ChainId().Int64() == 0 && (v.Uint64() == 27 || v.Uint64() == 1 || v.Uint64() == 0 || v.Uint64() == 28)
}

func (p *precheck) ChainID(tx *types.Transaction) error {
	txChainID := fmt.Sprintf("0x%x", tx.ChainId())
	passes := p.isLegacyUnprotectedEtx(tx) || txChainID == p.chainID

	if !passes {
		p.logger.Debug("Failed chainId precheck",
			zap.String("transaction", tx.Hash().Hex()),
			zap.String("chainId", txChainID),
			zap.String("expectedChainId", p.chainID))

		return fmt.Errorf("ChainId (%s) not supported. The correct chainId is %s", txChainID, p.chainID)
	}

	return nil
}

// Add a function to check if a transaction matches the deterministic deployer
func isDeterministicDeployerTx(tx *types.Transaction) bool {
	// Check basic properties that would identify this special transaction
	if tx.Type() != 0 || // Must be legacy transaction
		tx.To() != nil || // Must be contract creation
		tx.Gas() != 100000 || // Must have specific gas limit
		tx.GasPrice().Cmp(big.NewInt(100000000000)) != 0 { // Must have specific gas price
		return false
	}

	// Check data matches expected pattern
	expectedData := "0x604580600e600039806000f350fe7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe03601600081602082378035828234f58015156039578182fd5b8082525050506014600cf3"
	if hex.EncodeToString(tx.Data()) != strings.TrimPrefix(expectedData, "0x") {
		return false
	}

	// Check signature values
	_, r, s := tx.RawSignatureValues()
	expectedR, _ := new(big.Int).SetString("2222222222222222222222222222222222222222222222222222222222222222", 16)
	expectedS, _ := new(big.Int).SetString("2222222222222222222222222222222222222222222222222222222222222222", 16)

	if r.Cmp(expectedR) != 0 || s.Cmp(expectedS) != 0 {
		return false
	}

	return true
}

// Update the GasPrice function to use this check
func (p *precheck) GasPrice(tx *types.Transaction, networkGasPriceInWeiBars int64) *domain.RPCError {
	// **notice: Pass gasPrice precheck if txGasPrice is greater than the minimum network's gas price value,
	//          OR if the transaction is the deterministic deployment transaction (a special case).
	// **explanation: The deterministic deployment transaction is pre-signed with a gasPrice value of only 10 hbars,
	//                which is lower than the minimum gas price value in all Hedera network environments. Therefore,
	//                this special case is exempt from the precheck in the Relay, and the gas price logic will be resolved at the Services level.
	if isDeterministicDeployerTx(tx) {
		p.logger.Info("Detected deterministic deployer transaction, bypassing gas price check")
		return nil
	}

	networkGasPrice := big.NewInt(networkGasPriceInWeiBars)
	var txGasPrice *big.Int

	p.logger.Info("gasPrice precheck", zap.String("tx.gasPrice", tx.GasPrice().String()), zap.String("tx.gasFeeCap", tx.GasFeeCap().String()), zap.String("tx.gasTipCap", tx.GasTipCap().String()))

	if tx.GasPrice() != nil {
		txGasPrice = tx.GasPrice()
	} else {
		maxFeePerGas := tx.GasFeeCap()
		maxPriorityFeePerGas := tx.GasTipCap()
		txGasPrice = new(big.Int).Add(maxFeePerGas, maxPriorityFeePerGas)
	}

	if txGasPrice.Cmp(networkGasPrice) < 0 {

		// Check if there is a buffer from config
		isBuffer := false
		if isBuffer {
			txGasPriceWithBuffer := new(big.Int).Add(txGasPrice, big.NewInt(GasPriceTinyBarBuffer))
			if txGasPriceWithBuffer.Cmp(networkGasPrice) >= 0 {
				return nil
			}
		}

		p.logger.Debug("Failed gas price precheck",
			zap.String("transaction", tx.Hash().Hex()),
			zap.String("gasPrice", txGasPrice.String()),
			zap.String("requiredGasPrice", networkGasPrice.String()))

		return domain.NewRPCError(domain.GasPriceTooLow, fmt.Sprintf("Gas price %s is below configured minimum gas price %s", txGasPrice.String(), networkGasPrice.String()))
	}

	return nil
}

func (p *precheck) Balance(tx *types.Transaction, account *domain.AccountResponse) *domain.RPCError {
	if account == nil {
		return domain.NewRPCError(domain.NotFound, fmt.Sprintf("Resource not found: tx.from '%s'", tx.Hash().Hex()))
	}

	var txGasPrice *big.Int
	if tx.GasPrice() != nil {
		txGasPrice = tx.GasPrice()
	} else {
		maxFeePerGas := tx.GasFeeCap()
		maxPriorityFeePerGas := tx.GasTipCap()
		txGasPrice = new(big.Int).Add(maxFeePerGas, maxPriorityFeePerGas)
	}

	gasLimit := big.NewInt(int64(tx.Gas()))
	gasCost := new(big.Int).Mul(txGasPrice, gasLimit)
	totalValue := new(big.Int).Add(tx.Value(), gasCost)

	balance := new(big.Int).Mul(big.NewInt(account.Balance.Balance), big.NewInt(TinybarToWeibarCoef))

	if balance.Cmp(totalValue) < 0 {
		p.logger.Debug("Failed balance precheck",
			zap.String("transaction", tx.Hash().Hex()),
			zap.String("totalValue", totalValue.String()),
			zap.String("accountBalance", balance.String()))
		return domain.NewRPCError(domain.ServerError, "Insufficient funds for transfer")
	}

	return nil
}

func (p *precheck) transactionIntrinsicGasCost(data []byte) uint64 {
	var zeros, nonZeros uint64

	for _, b := range data {
		if b == 0 {
			zeros++
		} else {
			nonZeros++
		}
	}

	return TxBaseCost + TxDataZeroCost*zeros + IstanbulTxDataNonZeroCost*nonZeros
}

func (p *precheck) GasLimit(tx *types.Transaction) *domain.RPCError {
	gasLimit := tx.Gas()
	intrinsicGasCost := p.transactionIntrinsicGasCost(tx.Data())

	if gasLimit > uint64(MaxGasPerSec) {
		p.logger.Debug("Gas limit too high", zap.String("transaction", tx.Hash().Hex()), zap.Uint64("gasLimit", gasLimit), zap.Int("maxGasPerSec", MaxGasPerSec))

		return domain.NewRPCError(domain.GasLimitTooHigh, fmt.Sprintf("transaction gas limit %d exceeds gas per sec limit %d", gasLimit, MaxGasPerSec))
	} else if gasLimit < intrinsicGasCost {
		p.logger.Debug("Gas limit too low", zap.String("transaction", tx.Hash().Hex()), zap.Uint64("gasLimit", gasLimit), zap.Uint64("intrinsicGasCost", intrinsicGasCost))

		return domain.NewRPCError(domain.GasLimitTooLow, fmt.Sprintf("Transaction gas limit provided %d is insufficient of intrinsic gas required %d", gasLimit, intrinsicGasCost))
	}

	return nil
}

func (p *precheck) CheckSize(transaction string) *domain.RPCError {
	transaction = strings.TrimPrefix(transaction, "0x")

	transactionBytes, err := hex.DecodeString(transaction)
	if err != nil {
		return domain.NewRPCError(domain.InvalidParams, fmt.Sprintf("invalid transaction hex: %v", err))
	}

	const transactionSizeLimit = 128 * 1024 // 128KB
	if len(transactionBytes) > transactionSizeLimit {
		return domain.NewRPCError(domain.OversizedData, fmt.Sprintf("Oversized data: transaction size %d, transaction limit %d", len(transactionBytes), transactionSizeLimit))
	}

	return nil
}

func (p *precheck) TransactionType(tx *types.Transaction) error {
	if tx.Type() == 3 {
		p.logger.Debug("Unsupported transaction type", zap.String("transaction", tx.Hash().Hex()), zap.Uint8("type", tx.Type()))

		return fmt.Errorf("unsupported transaction type")
	}
	return nil
}

func (p *precheck) ReceiverAccount(tx *types.Transaction) *domain.RPCError {
	p.logger.Debug("Receiver account precheck", zap.String("transaction", tx.Hash().Hex()), zap.Any("tx.to", tx.To()))
	if tx.To() != nil {
		verifyAccount, err := p.mClient.GetAccountById(tx.To().Hex())
		if err != nil || verifyAccount == nil {
			return nil
		}

		if verifyAccount.ReceiverSigRequired {
			return domain.NewRPCError(domain.ServerError, "Operation is not supported when receiver's signature is enabled.")
		}
	}
	return nil
}

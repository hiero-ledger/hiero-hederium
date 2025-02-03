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

const (
	TxBaseCost                = 21000
	TxDataZeroCost            = 4
	IstanbulTxDataNonZeroCost = 16
	MaxGasPerSec              = 15000000
	TinybarToWeibarCoef       = 10000000000
	GasPriceTinyBarBuffer     = 1
)

type Precheck interface {
	ParseTxIfNeeded(transaction interface{}) *types.Transaction
	Value(tx *types.Transaction) error
	SendRawTransactionCheck(parsedTx *types.Transaction, networkGasPriceInWeiBars int64) error
	VerifyAccount(tx *types.Transaction) (*domain.AccountResponse, error)
	Nonce(tx *types.Transaction, accountInfoNonce int64) error
	ChainID(tx *types.Transaction) error
	GasPrice(tx *types.Transaction, networkGasPriceInWeiBars int64) error
	Balance(tx *types.Transaction, account *domain.AccountResponse) error
	GasLimit(tx *types.Transaction) error
	CheckSize(transaction string) error
	TransactionType(tx *types.Transaction) error
	ReceiverAccount(tx *types.Transaction) error
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
		tx, err := ParseTransaction(p.logger, txStr)
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
		return fmt.Errorf("value too low")
	}
	return nil
}

func (p *precheck) SendRawTransactionCheck(parsedTx *types.Transaction, networkGasPriceInWeiBars int64) error {

	if err := p.TransactionType(parsedTx); err != nil {
		return err
	}
	if err := p.GasLimit(parsedTx); err != nil {
		return err
	}

	mirrorAccountInfo, err := p.VerifyAccount(parsedTx)
	if err != nil {
		return err
	}

	if err := p.Nonce(parsedTx, mirrorAccountInfo.EthereumNonce); err != nil {
		return err
	}
	if err := p.ChainID(parsedTx); err != nil {
		return err
	}
	if err := p.Value(parsedTx); err != nil {
		return err
	}
	if err := p.GasPrice(parsedTx, networkGasPriceInWeiBars); err != nil {
		return err
	}
	if err := p.Balance(parsedTx, mirrorAccountInfo); err != nil {
		return err
	}
	if err := p.ReceiverAccount(parsedTx); err != nil {
		return err
	}

	return nil
}

func (p *precheck) VerifyAccount(tx *types.Transaction) (*domain.AccountResponse, error) {
	from, err := types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	if err != nil {
		return nil, err
	}

	accountInfo, err := p.mClient.GetAccountById(from.Hex())
	if err != nil {
		p.logger.Debug("Failed to retrieve address account details",
			zap.String("address", from.Hex()),
			zap.Error(err))
		return nil, err
	}

	if accountInfo == nil {
		p.logger.Debug("Failed to retrieve address account details",
			zap.String("address", from.Hex()))
		return nil, fmt.Errorf("resource not found: address '%s'", from.Hex())
	}

	return accountInfo, nil
}

func (p *precheck) Nonce(tx *types.Transaction, accountInfoNonce int64) error {

	p.logger.Debug("Nonce precheck", zap.Uint64("tx.nonce", tx.Nonce()), zap.Int64("accountInfoNonce", accountInfoNonce))

	if uint64(accountInfoNonce) > tx.Nonce() {
		return fmt.Errorf("nonce too low: provided nonce: %d, current nonce: %d", tx.Nonce(), accountInfoNonce)
	}

	return nil
}

func (p *precheck) isLegacyUnprotectedEtx(tx *types.Transaction) bool {
	v, _, _ := tx.RawSignatureValues()

	return tx.ChainId().Int64() == 0 && (v.Int64() == 27 || v.Int64() == 28)
}

func (p *precheck) ChainID(tx *types.Transaction) error {
	txChainID := fmt.Sprintf("0x%x", tx.ChainId())

	passes := p.isLegacyUnprotectedEtx(tx) || tx.ChainId().String() == p.chainID

	if !passes {
		p.logger.Debug("Failed chainId precheck",
			zap.String("transaction", tx.Hash().Hex()),
			zap.String("chainId", txChainID))

		return fmt.Errorf("unsupported chain id: got %s, want %s", tx.ChainId().String(), p.chainID)
	}

	return nil
}

func (p *precheck) GasPrice(tx *types.Transaction, networkGasPriceInWeiBars int64) error {
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
		txGasPriceWithBuffer := new(big.Int).Add(txGasPrice, big.NewInt(GasPriceTinyBarBuffer))
		if txGasPriceWithBuffer.Cmp(networkGasPrice) >= 0 {
			return nil
		}

		if p.logger.Core().Enabled(zap.DebugLevel) {
			p.logger.Debug("Failed gas price precheck",
				zap.String("transaction", tx.Hash().Hex()),
				zap.String("gasPrice", txGasPrice.String()),
				zap.String("requiredGasPrice", networkGasPrice.String()))
		}
		return fmt.Errorf("gas price too low: got %s, required %s", txGasPrice.String(), networkGasPrice.String())
	}

	return nil
}

func (p *precheck) Balance(tx *types.Transaction, account *domain.AccountResponse) error {
	if account == nil {
		return fmt.Errorf("resource not found: tx.from '%s'", tx.Hash().Hex())
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
		if p.logger.Core().Enabled(zap.DebugLevel) {
			p.logger.Debug("Failed balance precheck",
				zap.String("transaction", tx.Hash().Hex()),
				zap.String("totalValue", totalValue.String()),
				zap.String("accountBalance", balance.String()))
		}
		return fmt.Errorf("insufficient account balance")
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

func (p *precheck) GasLimit(tx *types.Transaction) error {
	gasLimit := tx.Gas()
	intrinsicGasCost := p.transactionIntrinsicGasCost(tx.Data())

	if gasLimit > uint64(MaxGasPerSec) {
		if p.logger.Core().Enabled(zap.DebugLevel) {
			p.logger.Debug("Gas limit too high",
				zap.String("transaction", tx.Hash().Hex()),
				zap.Uint64("gasLimit", gasLimit),
				zap.Int("maxGasPerSec", MaxGasPerSec))
		}
		return fmt.Errorf("gas limit too high: got %d, max %d", gasLimit, MaxGasPerSec)
	} else if gasLimit < intrinsicGasCost {
		if p.logger.Core().Enabled(zap.DebugLevel) {
			p.logger.Debug("Gas limit too low",
				zap.String("transaction", tx.Hash().Hex()),
				zap.Uint64("gasLimit", gasLimit),
				zap.Uint64("intrinsicGasCost", intrinsicGasCost))
		}
		return fmt.Errorf("gas limit too low: got %d, required %d", gasLimit, intrinsicGasCost)
	}

	return nil
}

func (p *precheck) CheckSize(transaction string) error {
	transaction = strings.TrimPrefix(transaction, "0x")

	transactionBytes, err := hex.DecodeString(transaction)
	if err != nil {
		return fmt.Errorf("invalid transaction hex: %v", err)
	}

	const transactionSizeLimit = 128 * 1024 // 128KB
	if len(transactionBytes) > transactionSizeLimit {
		return fmt.Errorf("transaction size too big: got %d, max %d", len(transactionBytes), transactionSizeLimit)
	}

	return nil
}

func (p *precheck) TransactionType(tx *types.Transaction) error {
	if tx.Type() == 3 {
		if p.logger.Core().Enabled(zap.DebugLevel) {
			p.logger.Debug("Unsupported transaction type",
				zap.String("transaction", tx.Hash().Hex()),
				zap.Uint8("type", tx.Type()))
		}
		return fmt.Errorf("unsupported transaction type: %d", tx.Type())
	}
	return nil
}

func (p *precheck) ReceiverAccount(tx *types.Transaction) error {
	if tx.To() != nil {
		verifyAccount := p.mClient.GetAccount(tx.To().Hex(), "")
		if verifyAccount == nil {
			return nil
		}

		if account, ok := verifyAccount.(*domain.AccountResponse); ok && account.ReceiverSigRequired {
			return fmt.Errorf("receiver signature required")
		}
	}
	return nil
}

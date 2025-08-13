package service

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/LimeChain/Hederium/internal/domain"
	infrahedera "github.com/LimeChain/Hederium/internal/infrastructure/hedera"
	"github.com/LimeChain/Hederium/internal/util"
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
	ParseTxIfNeeded(transaction interface{}) *util.Tx
	Value(tx *util.Tx) error
	SendRawTransactionCheck(parsedTx *util.Tx, networkGasPriceInWeiBars int64) error
	VerifyAccount(tx *util.Tx) (*domain.AccountResponse, error)
	Nonce(tx *util.Tx, accountInfoNonce int64) error
	ChainID(tx *util.Tx) error
	GasPrice(tx *util.Tx, networkGasPriceInWeiBars int64) error
	Balance(tx *util.Tx, account *domain.AccountResponse) error
	GasLimit(tx *util.Tx) error
	CheckSize(transaction string) error
	TransactionType(tx *util.Tx) error
	ReceiverAccount(tx *util.Tx) error
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

func (p *precheck) ParseTxIfNeeded(transaction interface{}) *util.Tx {
	if txStr, ok := transaction.(string); ok {
		tx, err := ParseTransaction(txStr)
		if err != nil {
			return nil
		}
		return tx
	}
	if tx, ok := transaction.(*util.Tx); ok {
		return tx
	}
	return nil
}

func (p *precheck) Value(tx *util.Tx) error {
	value := tx.Value
	if (value.Cmp(big.NewInt(0)) > 0 && value.Cmp(big.NewInt(TinybarToWeibarCoef)) < 0) || value.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("value too low")
	}
	return nil
}

func (p *precheck) SendRawTransactionCheck(parsedTx *util.Tx, networkGasPriceInWeiBars int64) error {

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

func (p *precheck) VerifyAccount(tx *util.Tx) (*domain.AccountResponse, error) {
	from, err := tx.Sender()
	if err != nil {
		return nil, err
	}

	accountInfo, err := p.mClient.GetAccountById(from)
	if err != nil {
		p.logger.Debug("Failed to retrieve address account details",
			zap.String("address", from),
			zap.Error(err))
		return nil, err
	}

	if accountInfo == nil {
		p.logger.Debug("Failed to retrieve address account details",
			zap.String("address", from))
		return nil, fmt.Errorf("resource not found: address '%s'", from)
	}

	return accountInfo, nil
}

func (p *precheck) Nonce(tx *util.Tx, accountInfoNonce int64) error {

	p.logger.Debug("Nonce precheck", zap.Uint64("tx.nonce", tx.Nonce), zap.Int64("accountInfoNonce", accountInfoNonce))

	if accountInfoNonce < 0 || uint64(accountInfoNonce) > tx.Nonce {
		return fmt.Errorf("nonce too low: provided nonce: %d, current nonce: %d", tx.Nonce, accountInfoNonce)
	}

	return nil
}

func (p *precheck) isLegacyUnprotectedEtx(tx *util.Tx) bool {
	return tx.ChainID.Int64() == 0 && (tx.V.Int64() == 27 || tx.V.Int64() == 28)
}

func (p *precheck) ChainID(tx *util.Tx) error {
	txChainID := fmt.Sprintf("0x%x", tx.ChainID)
	passes := p.isLegacyUnprotectedEtx(tx) || txChainID == p.chainID

	if !passes {
		p.logger.Debug("Failed chainId precheck",
			zap.String("transaction", tx.Hash),
			zap.String("chainId", txChainID),
			zap.String("expectedChainId", p.chainID))

		return fmt.Errorf("unsupported chain id: got %s, want %s", tx.ChainID.String(), p.chainID)
	}

	return nil
}

func (p *precheck) GasPrice(tx *util.Tx, networkGasPriceInWeiBars int64) error {
	networkGasPrice := big.NewInt(networkGasPriceInWeiBars)
	var txGasPrice *big.Int

	p.logger.Info("gasPrice precheck", zap.String("tx.gasPrice", tx.GasPrice.String()), zap.String("tx.gasFeeCap", tx.GasFeeCap.String()), zap.String("tx.gasTipCap", tx.GasTipCap.String()))

	if tx.GasPrice != nil {
		txGasPrice = tx.GasPrice
	} else {
		maxFeePerGas := tx.GasFeeCap
		maxPriorityFeePerGas := tx.GasTipCap
		txGasPrice = new(big.Int).Add(maxFeePerGas, maxPriorityFeePerGas)
	}

	if txGasPrice.Cmp(networkGasPrice) < 0 {
		txGasPriceWithBuffer := new(big.Int).Add(txGasPrice, big.NewInt(GasPriceTinyBarBuffer))
		if txGasPriceWithBuffer.Cmp(networkGasPrice) >= 0 {
			return nil
		}

		if p.logger.Core().Enabled(zap.DebugLevel) {
			p.logger.Debug("Failed gas price precheck",
				zap.String("transaction", tx.Hash),
				zap.String("gasPrice", txGasPrice.String()),
				zap.String("requiredGasPrice", networkGasPrice.String()))
		}
		return fmt.Errorf("gas price too low: got %s, required %s", txGasPrice.String(), networkGasPrice.String())
	}

	return nil
}

func (p *precheck) Balance(tx *util.Tx, account *domain.AccountResponse) error {
	if account == nil {
		return fmt.Errorf("resource not found: tx.from '%s'", tx.Hash)
	}

	var txGasPrice *big.Int
	if tx.GasPrice != nil {
		txGasPrice = tx.GasPrice
	} else {
		maxFeePerGas := tx.GasFeeCap
		maxPriorityFeePerGas := tx.GasTipCap
		txGasPrice = new(big.Int).Add(maxFeePerGas, maxPriorityFeePerGas)
	}

	gasLimit := new(big.Int).SetUint64(tx.GasLimit)
	gasCost := new(big.Int).Mul(txGasPrice, gasLimit)
	totalValue := new(big.Int).Add(tx.Value, gasCost)

	balance := new(big.Int).Mul(big.NewInt(account.Balance.Balance), big.NewInt(TinybarToWeibarCoef))

	if balance.Cmp(totalValue) < 0 {
		if p.logger.Core().Enabled(zap.DebugLevel) {
			p.logger.Debug("Failed balance precheck",
				zap.String("transaction", tx.Hash),
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

func (p *precheck) GasLimit(tx *util.Tx) error {
	p.logger.Info("gasLimit precheck", zap.Any("tx", tx))
	gasLimit := tx.GasLimit

	// Convert hex-encoded data string back to bytes for gas calculation
	dataBytes, err := hex.DecodeString(tx.Data)
	if err != nil {
		// If decoding fails, treat as empty data
		dataBytes = []byte{}
	}

	intrinsicGasCost := p.transactionIntrinsicGasCost(dataBytes)

	if gasLimit > uint64(MaxGasPerSec) {
		if p.logger.Core().Enabled(zap.DebugLevel) {
			p.logger.Debug("Gas limit too high",
				zap.String("transaction", tx.Hash),
				zap.Uint64("gasLimit", gasLimit),
				zap.Int("maxGasPerSec", MaxGasPerSec))
		}
		return fmt.Errorf("gas limit too high: got %d, max %d", gasLimit, MaxGasPerSec)
	} else if gasLimit < intrinsicGasCost {
		if p.logger.Core().Enabled(zap.DebugLevel) {
			p.logger.Debug("Gas limit too low",
				zap.String("transaction", tx.Hash),
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

func (p *precheck) TransactionType(tx *util.Tx) error {
	if tx.Type == 3 {
		if p.logger.Core().Enabled(zap.DebugLevel) {
			p.logger.Debug("Unsupported transaction type",
				zap.String("transaction", tx.Hash),
				zap.Uint8("type", tx.Type))
		}
		return fmt.Errorf("unsupported transaction type: %d", tx.Type)
	}
	return nil
}

func (p *precheck) ReceiverAccount(tx *util.Tx) error {
	if tx.To != "" {
		verifyAccount := p.mClient.GetAccount(tx.To, "")
		if verifyAccount == nil {
			return nil
		}

		if account, ok := verifyAccount.(*domain.AccountResponse); ok && account.ReceiverSigRequired {
			return fmt.Errorf("receiver signature required")
		}
	}
	return nil
}

package util

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	rlp "github.com/defiweb/go-rlp"
	"golang.org/x/crypto/sha3"
)

const (
	LegacyTxType     = 0x00
	AccessListTxType = 0x01
	DynamicFeeTxType = 0x02
	BlobTxType       = 0x03
)

// -----------------------------------------------------------------------------
// Public Tx model – keep ONLY what the Hed‑Eth bridge needs.  Add more later.
// -----------------------------------------------------------------------------
type Tx struct {
	// Common fields --------------------------------------------------------
	Type     byte
	Nonce    uint64
	GasPrice *big.Int // (type 0/1 only)
	GasLimit uint64
	To       string // hex w/ 0x, "" == contract‑create
	Value    *big.Int
	Data     string // hex-encoded string instead of []byte

	// 1559 / 4844 extras ---------------------------------------------------
	GasTipCap  *big.Int // max priority fee (type 2)
	GasFeeCap  *big.Int // max fee per gas (type 2)
	BlobGas    uint64   // type 3
	BlobFeeCap *big.Int

	// Signature ------------------------------------------------------------
	ChainID *big.Int // 0 == unprotected
	V, R, S *big.Int

	// Convenience extras ---------------------------------------------------
	Hash        string // tx hash (computed on demand, not part of the sig)
	BlockHash   string // mirror node sets these
	BlockNumber uint64
	Index       uint64 // position in block
}

// -----------------------------------------------------------------------------
// Decode raw RLP‑encoded tx ----------------------------------------------------
// -----------------------------------------------------------------------------

// Decode parses raw transaction bytes. Only legacy/EIP‑155 supported right now.
func DecodeTx(raw []byte) (*Tx, error) {
	if len(raw) == 0 {
		return nil, errors.New("empty tx data")
	}

	// Typed envelope? First byte < 0x7f and next byte is RLP list tag.
	if raw[0] >= AccessListTxType && raw[0] <= BlobTxType {
		return nil, fmt.Errorf("typed tx %d not implemented yet", raw[0])
	}

	// Legacy → decode using DecodeLazy for dynamic structure
	dec, _, err := rlp.DecodeLazy(raw)
	if err != nil {
		return nil, fmt.Errorf("legacy rlp decode: %w", err)
	}

	// Check if the decoded data is a list
	list, err := dec.List()
	if err != nil {
		// If it's not a list, maybe it's a string containing the actual transaction
		if dec.IsString() {
			str, strErr := dec.String()
			if strErr != nil {
				return nil, fmt.Errorf("expected list or string, got error: %w", strErr)
			}
			// Try to decode the inner content
			innerDec, _, innerErr := rlp.DecodeLazy([]byte(str.Get()))
			if innerErr != nil {
				return nil, fmt.Errorf("failed to decode inner content: %w", innerErr)
			}
			list, err = innerDec.List()
			if err != nil {
				return nil, fmt.Errorf("inner content is not a list: %w", err)
			}
		} else {
			return nil, fmt.Errorf("expected list, got: %w", err)
		}
	}

	if len(list) != 9 {
		return nil, fmt.Errorf("legacy tx expects 9 fields, got %d", len(list))
	}

	// Convert each element to bytes
	fields := make([][]byte, 9)
	for i, item := range list {
		if item.IsString() {
			str, err := item.String()
			if err != nil {
				return nil, fmt.Errorf("failed to decode field %d as string: %w", i, err)
			}
			fields[i] = []byte(str.Get())
		} else {
			// Try to decode as uint for numeric values
			uintVal, err := item.Uint()
			if err != nil {
				return nil, fmt.Errorf("failed to decode field %d: %w", i, err)
			}
			// Convert uint to bytes
			// value already uint64, avoid redundant conversion
			fields[i] = new(big.Int).SetUint64(uintVal.Get()).Bytes()
		}
	}

	tx := &Tx{Type: LegacyTxType}
	var ok bool
	if tx.Nonce, ok = bytesToUint(fields[0]); !ok {
		return nil, errors.New("nonce overflow")
	}
	tx.GasPrice = new(big.Int).SetBytes(fields[1])
	if tx.GasLimit, ok = bytesToUint(fields[2]); !ok {
		return nil, errors.New("gasLimit overflow")
	}
	tx.To = bytesToHexAddr(fields[3])
	tx.Value = new(big.Int).SetBytes(fields[4])
	tx.Data = hex.EncodeToString(fields[5]) // Convert to hex string

	tx.V = new(big.Int).SetBytes(fields[6])
	tx.R = new(big.Int).SetBytes(fields[7])
	tx.S = new(big.Int).SetBytes(fields[8])
	tx.ChainID = deriveChainID(tx.V)

	return tx, nil
}

// -----------------------------------------------------------------------------
// Convenience helpers ---------------------------------------------------------
// -----------------------------------------------------------------------------

// Sender recovers the 0x…40 hex address from the signature.
func (tx *Tx) Sender() (string, error) {
	if tx.Type != LegacyTxType {
		return "", errors.New("Sender: unsupported tx type")
	}
	if tx.R.Sign() == 0 || tx.S.Sign() == 0 {
		return "", errors.New("Sender: missing sig values")
	}

	sighash, err := tx.signingHashLegacy()
	if err != nil {
		return "", err
	}

	// Manually construct the 65-byte signature for Ethereum compatibility
	var rBytes, sBytes [32]byte
	tx.R.FillBytes(rBytes[:])
	tx.S.FillBytes(sBytes[:])

	// Calculate recovery ID for EIP-155 transactions
	var recoveryID byte
	if tx.ChainID.Sign() != 0 {
		// EIP-155: recovery_id = V - 2*chain_id - 35
		v := new(big.Int).Set(tx.V)
		chainIDMul2 := new(big.Int).Mul(tx.ChainID, big.NewInt(2))
		v.Sub(v, chainIDMul2)
		v.Sub(v, big.NewInt(35))
		recoveryID = byte(v.Uint64())
	} else {
		// Unprotected transaction: recovery_id = V - 27
		recoveryID = byte(tx.V.Uint64() - 27)
	}

	// For compact signature format, the recovery code is:
	// 27 + recovery_id (+ 4 if compressed)
	// Try both compressed and uncompressed
	var lastErr error
	for _, compressed := range []bool{false, true} {
		recoveryCode := byte(27) + recoveryID
		if compressed {
			recoveryCode += 4
		}

		// Create 65-byte signature: recovery_code(1) + R(32) + S(32)
		// The compact signature format expects recovery code as the FIRST byte
		sig := make([]byte, 65)
		sig[0] = recoveryCode
		copy(sig[1:33], rBytes[:])
		copy(sig[33:65], sBytes[:])

		// Recover public key using the 65-byte signature
		pub, wasCompressed, err := ecdsa.RecoverCompact(sig, sighash)
		if err != nil {
			lastErr = err
			// Try the other compression format
			continue
		}
		if pub == nil {
			continue
		}

		// Use the recovered public key
		_ = wasCompressed

		h := sha3.NewLegacyKeccak256()
		uncompressed := pub.SerializeUncompressed()
		h.Write(uncompressed[1:]) // skip 0x04 prefix
		var out [32]byte
		h.Sum(out[:0])
		return "0x" + hex.EncodeToString(out[12:]), nil // last 20 bytes
	}

	return "", fmt.Errorf("failed to recover public key with either compression format, last error: %v", lastErr)
}

// signingHashLegacy returns Keccak256(RLP([nonce, gasPrice, gasLimit, to, value, data, chainID, 0, 0]))
func (tx *Tx) signingHashLegacy() ([]byte, error) {
	// For unprotected legacy tx (chainID==0) the extras are omitted.
	var payload rlp.List
	dataBytes, err := hex.DecodeString(tx.Data)
	if err != nil {
		return nil, fmt.Errorf("invalid hex data: %w", err)
	}

	if tx.ChainID.Sign() == 0 {
		payload = rlp.List{
			rlp.Uint(tx.Nonce),
			rlp.String(tx.GasPrice.Bytes()),
			rlp.Uint(tx.GasLimit),
			rlp.String(hexToBytes(tx.To)),
			rlp.String(tx.Value.Bytes()),
			rlp.String(dataBytes),
		}
	} else {
		payload = rlp.List{
			rlp.Uint(tx.Nonce),
			rlp.String(tx.GasPrice.Bytes()),
			rlp.Uint(tx.GasLimit),
			rlp.String(hexToBytes(tx.To)),
			rlp.String(tx.Value.Bytes()),
			rlp.String(dataBytes),
			rlp.String(tx.ChainID.Bytes()),
			rlp.Uint(0),
			rlp.Uint(0),
		}
	}
	enc, err := rlp.Encode(payload)
	if err != nil {
		return nil, err
	}
	h := sha3.NewLegacyKeccak256()
	h.Write(enc)
	return h.Sum(nil), nil
}

// -----------------------------------------------------------------------------
// Utility functions
// -----------------------------------------------------------------------------

func bytesToUint(b []byte) (uint64, bool) {
	if len(b) > 8 {
		return 0, false
	}
	var v uint64
	for _, by := range b {
		v = v<<8 | uint64(by)
	}
	return v, true
}

func bytesToHexAddr(b []byte) string {
	if len(b) == 0 {
		return "" // contract creation
	}
	return "0x" + hex.EncodeToString(b)
}

func hexToBytes(addr string) []byte {
	addr = strings.TrimPrefix(addr, "0x")
	if addr == "" {
		return []byte{}
	}
	out, _ := hex.DecodeString(addr)
	return out
}

func deriveChainID(v *big.Int) *big.Int {
	if v.Sign() == 0 {
		return big.NewInt(0)
	}
	vv := new(big.Int).Sub(v, big.NewInt(35))
	vv.Div(vv, big.NewInt(2))
	if vv.Sign() < 0 {
		return big.NewInt(0)
	}
	return vv
}

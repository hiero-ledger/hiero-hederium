package util

import (
	"encoding/hex"
	"errors"
	"strings"
)

func Decode(s string) ([]byte, error) {
	s = strings.TrimPrefix(s, "0x")
	if len(s)%2 == 1 {
		// hex.DecodeString requires even length
		s = "0" + s
	}
	out, err := hex.DecodeString(s)
	if err != nil {
		return nil, errors.New("util: invalid hex string")
	}
	return out, nil
}

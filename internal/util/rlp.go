package util

import drlp "github.com/defiweb/go-rlp"

type Decoder interface{ DecodeRLP([]byte) (int, error) }

func DecodeBytes(b []byte, out interface{}) error {
	dec, ok := out.(Decoder)
	if !ok {
		return drlp.ErrUnsupportedType
	}
	_, err := dec.DecodeRLP(b) // re-use the upstream implementation
	return err
}

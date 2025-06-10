package util

var prohibited = map[byte]bool{
	0xf2: true, // CALLCODE
	0xf4: true, // DELEGATECALL
	0xff: true, // SELFDESTRUCT
}

func HasProhibitedOpcodes(code []byte) bool {
	for i := 0; i < len(code); i++ {
		op := code[i]
		if prohibited[op] {
			return true
		}
		// Skip immediate data on PUSH1–PUSH32
		if op >= 0x60 && op <= 0x7f {
			i += int(op - 0x60 + 1)
		}
	}
	return false
}

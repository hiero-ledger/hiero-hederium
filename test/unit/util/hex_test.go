package util_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/LimeChain/Hederium/internal/util"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{name: "empty", input: "", want: []byte{}},
		{name: "prefix only", input: "0x", want: []byte{}},
		{name: "even length with prefix", input: "0xabcd", want: []byte{0xab, 0xcd}},
		{name: "odd length is left-padded", input: "0x1", want: []byte{0x01}},
		{name: "uppercase without prefix", input: "ABCD", want: []byte{0xab, 0xcd}},
		{name: "invalid characters", input: "0xzz", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.Decode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Decode(%q) expected an error, got %x", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Decode(%q) unexpected error: %v", tt.input, err)
			}
			if hex.EncodeToString(got) != hex.EncodeToString(tt.want) {
				t.Fatalf("Decode(%q) = %x, want %x", tt.input, got, tt.want)
			}
		})
	}
}

func FuzzDecode(f *testing.F) {
	seeds := []string{
		"",
		"0x",
		"0x0",
		"0x00",
		"0x1",
		"0xabcdef",
		"ABCDEF",
		"0x0123456789abcdefABCDEF",
		"0xzz",
		"0x0x12",
		"0X12",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		out, err := util.Decode(s)
		if err != nil {
			if out != nil {
				t.Fatalf("Decode(%q) returned bytes alongside an error", s)
			}
			return
		}

		// A successful decode must round-trip back to the normalised input:
		// "0x" prefix stripped, odd length left-padded with a zero, lowercase.
		want := strings.ToLower(strings.TrimPrefix(s, "0x"))
		if len(want)%2 == 1 {
			want = "0" + want
		}
		if got := hex.EncodeToString(out); got != want {
			t.Fatalf("Decode(%q) round-trip mismatch: got %q, want %q", s, got, want)
		}
	})
}

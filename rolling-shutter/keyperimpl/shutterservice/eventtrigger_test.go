package shutterservice

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/keyperimpl/shutterservice/help"
)

func TestOpValidate(t *testing.T) {
	tests := []struct {
		name    string
		op      Op
		wantErr bool
	}{
		{
			name:    "UintLt is valid",
			op:      UintLt,
			wantErr: false,
		},
		{
			name:    "UintLte is valid",
			op:      UintLte,
			wantErr: false,
		},
		{
			name:    "UintEq is valid",
			op:      UintEq,
			wantErr: false,
		},
		{
			name:    "UintGt is valid",
			op:      UintGt,
			wantErr: false,
		},
		{
			name:    "UintGte is valid",
			op:      UintGte,
			wantErr: false,
		},
		{
			name:    "BytesEq is valid",
			op:      BytesEq,
			wantErr: false,
		},
		{
			name:    "invalid operation value 6",
			op:      Op(6),
			wantErr: true,
		},
		{
			name:    "invalid operation value 100",
			op:      Op(100),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.op.Validate()
			if tt.wantErr {
				assert.ErrorContains(t, err, "invalid operation:")
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestLogValueRefValidate(t *testing.T) {
	tests := []struct {
		name    string
		ref     LogValueRef
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid topic reference - offset 0",
			ref: LogValueRef{
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "valid topic reference - offset 3",
			ref: LogValueRef{
				Offset: 3,
			},
			wantErr: false,
		},
		{
			name: "valid data reference - offset 4",
			ref: LogValueRef{
				Offset: 4,
			},
			wantErr: false,
		},
		{
			name: "valid data reference - offset 5",
			ref: LogValueRef{
				Offset: 5,
			},
			wantErr: false,
		},
		{
			name: "valid data reference - large offset",
			ref: LogValueRef{
				Offset: 100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ref.Validate()
			if tt.wantErr {
				assert.Error(t, err, tt.errMsg)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestValuePredicateValidate(t *testing.T) {
	tests := []struct {
		name      string
		predicate ValuePredicate
		numWords  uint64
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid UintLt predicate",
			predicate: ValuePredicate{
				Op:       UintLt,
				IntArgs:  []*big.Int{big.NewInt(100)},
				ByteArgs: [][]byte{},
			},
			wantErr: false,
		},
		{
			name: "valid UintEq predicate",
			predicate: ValuePredicate{
				Op:       UintEq,
				IntArgs:  []*big.Int{big.NewInt(42)},
				ByteArgs: [][]byte{},
			},
			wantErr: false,
		},
		{
			name: "valid BytesEq predicate",
			predicate: ValuePredicate{
				Op:       BytesEq,
				IntArgs:  []*big.Int{},
				ByteArgs: [][]byte{make([]byte, 32)}, // 32 bytes for 1 word
			},
			wantErr: false,
		},
		{
			name: "valid BytesEq predicate with 2 words",
			predicate: ValuePredicate{
				Op:       BytesEq,
				IntArgs:  []*big.Int{},
				ByteArgs: [][]byte{make([]byte, 64)}, // 64 bytes for 2 words
			},
			wantErr: false,
		},
		{
			name: "invalid operation",
			predicate: ValuePredicate{
				Op:       Op(999),
				IntArgs:  []*big.Int{},
				ByteArgs: [][]byte{},
			},
			wantErr: true,
			errMsg:  "invalid operation: 999",
		},
		{
			name: "UintLt with wrong number of int args - too few",
			predicate: ValuePredicate{
				Op:       UintLt,
				IntArgs:  []*big.Int{},
				ByteArgs: [][]byte{},
			},
			wantErr: true,
			errMsg:  "operation 0 requires exactly 1 integer argument(s), got 0",
		},
		{
			name: "UintLt with wrong number of int args - too many",
			predicate: ValuePredicate{
				Op:       UintLt,
				IntArgs:  []*big.Int{big.NewInt(1), big.NewInt(2)},
				ByteArgs: [][]byte{},
			},
			wantErr: true,
			errMsg:  "operation 0 requires exactly 1 integer argument(s), got 2",
		},
		{
			name: "BytesEq with wrong number of byte args - too few",
			predicate: ValuePredicate{
				Op:       BytesEq,
				IntArgs:  []*big.Int{},
				ByteArgs: [][]byte{},
			},
			wantErr: true,
			errMsg:  "operation 5 requires exactly 1 bytes argument(s), got 0",
		},
		{
			name: "BytesEq with wrong number of byte args - too many",
			predicate: ValuePredicate{
				Op:       BytesEq,
				IntArgs:  []*big.Int{},
				ByteArgs: [][]byte{make([]byte, 32), make([]byte, 32)},
			},
			wantErr: true,
			errMsg:  "operation 5 requires exactly 1 bytes argument(s), got 2",
		},
		{
			name: "UintLt with nil integer argument",
			predicate: ValuePredicate{
				Op:       UintLt,
				IntArgs:  []*big.Int{nil},
				ByteArgs: [][]byte{},
			},
			wantErr: true,
			errMsg:  "integer argument 0 cannot be nil for operation 0",
		},
		{
			name: "UintLt with negative integer argument",
			predicate: ValuePredicate{
				Op:       UintLt,
				IntArgs:  []*big.Int{big.NewInt(-1)},
				ByteArgs: [][]byte{},
			},
			wantErr: true,
			errMsg:  "integer argument 0 cannot be negative for operation 0",
		},
		{
			name: "UintLt with integer argument fitting exactly in 2 words",
			predicate: ValuePredicate{
				Op: UintLt,
				IntArgs: []*big.Int{func() *big.Int {
					// Create a number that requires exactly 512 bits (2 words = 64 bytes = 512 bits)
					val := big.NewInt(1)
					val.Lsh(val, 511) // 2^511
					return val
				}()},
				ByteArgs: [][]byte{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.predicate.Validate()
			if tt.wantErr {
				assert.ErrorContains(t, err, tt.errMsg)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestValuePredicateMatch(t *testing.T) {
	tests := []struct {
		name      string
		predicate ValuePredicate
		value     []byte
		want      bool
	}{
		// UintLt tests
		{
			name: "UintLt - value less than argument",
			predicate: ValuePredicate{
				Op:      UintLt,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(50).Bytes())
				return val
			}(),
			want: true,
		},
		{
			name: "UintLt - value equal to argument",
			predicate: ValuePredicate{
				Op:      UintLt,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(100).Bytes())
				return val
			}(),
			want: false,
		},
		{
			name: "UintLt - value greater than argument",
			predicate: ValuePredicate{
				Op:      UintLt,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(150).Bytes())
				return val
			}(),
			want: false,
		},
		{
			name: "UintLt - zero value",
			predicate: ValuePredicate{
				Op:      UintLt,
				IntArgs: []*big.Int{big.NewInt(1)},
			},
			value: make([]byte, 32), // 32 bytes of zeros
			want:  true,
		},
		// UintLte tests
		{
			name: "UintLte - value less than argument",
			predicate: ValuePredicate{
				Op:      UintLte,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(50).Bytes())
				return val
			}(),
			want: true,
		},
		{
			name: "UintLte - value equal to argument",
			predicate: ValuePredicate{
				Op:      UintLte,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(100).Bytes())
				return val
			}(),
			want: true,
		},
		{
			name: "UintLte - value greater than argument",
			predicate: ValuePredicate{
				Op:      UintLte,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(150).Bytes())
				return val
			}(),
			want: false,
		},
		// UintEq tests
		{
			name: "UintEq - values equal",
			predicate: ValuePredicate{
				Op:      UintEq,
				IntArgs: []*big.Int{big.NewInt(42)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(42).Bytes())
				return val
			}(),
			want: true,
		},
		{
			name: "UintEq - values not equal",
			predicate: ValuePredicate{
				Op:      UintEq,
				IntArgs: []*big.Int{big.NewInt(42)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(43).Bytes())
				return val
			}(),
			want: false,
		},
		{
			name: "UintEq - zero values",
			predicate: ValuePredicate{
				Op:      UintEq,
				IntArgs: []*big.Int{big.NewInt(0)},
			},
			value: make([]byte, 32), // 32 bytes of zeros
			want:  true,
		},
		// UintGt tests
		{
			name: "UintGt - value greater than argument",
			predicate: ValuePredicate{
				Op:      UintGt,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(150).Bytes())
				return val
			}(),
			want: true,
		},
		{
			name: "UintGt - value equal to argument",
			predicate: ValuePredicate{
				Op:      UintGt,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(100).Bytes())
				return val
			}(),
			want: false,
		},
		{
			name: "UintGt - value less than argument",
			predicate: ValuePredicate{
				Op:      UintGt,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(50).Bytes())
				return val
			}(),
			want: false,
		},
		// UintGte tests
		{
			name: "UintGte - value greater than argument",
			predicate: ValuePredicate{
				Op:      UintGte,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(150).Bytes())
				return val
			}(),
			want: true,
		},
		{
			name: "UintGte - value equal to argument",
			predicate: ValuePredicate{
				Op:      UintGte,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(100).Bytes())
				return val
			}(),
			want: true,
		},
		{
			name: "UintGte - value less than argument",
			predicate: ValuePredicate{
				Op:      UintGte,
				IntArgs: []*big.Int{big.NewInt(100)},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val[31:], big.NewInt(50).Bytes())
				return val
			}(),
			want: false,
		},
		// BytesEq tests
		{
			name: "BytesEq - equal bytes",
			predicate: ValuePredicate{
				Op: BytesEq,
				ByteArgs: [][]byte{func() []byte {
					val := make([]byte, 32)
					copy(val, "hello")
					return val
				}()},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val, "hello")
				return val
			}(),
			want: true,
		},
		{
			name: "BytesEq - different bytes",
			predicate: ValuePredicate{
				Op: BytesEq,
				ByteArgs: [][]byte{func() []byte {
					val := make([]byte, 32)
					copy(val, "hello")
					return val
				}()},
			},
			value: func() []byte {
				val := make([]byte, 32)
				copy(val, "world")
				return val
			}(),
			want: false,
		},
		{
			name: "BytesEq - empty bytes",
			predicate: ValuePredicate{
				Op:       BytesEq,
				ByteArgs: [][]byte{make([]byte, 32)}, // 32 bytes of zeros
			},
			value: make([]byte, 32), // 32 bytes of zeros
			want:  true,
		},
		{
			name: "BytesEq - 32-byte values (typical for Ethereum)",
			predicate: ValuePredicate{
				Op:       BytesEq,
				ByteArgs: [][]byte{make([]byte, 32)}, // all zeros
			},
			value: make([]byte, 32), // all zeros
			want:  true,
		},
		{
			name: "BytesEq - 64-byte values (2 words)",
			predicate: ValuePredicate{
				Op: BytesEq,
				ByteArgs: [][]byte{func() []byte {
					val := make([]byte, 64)
					copy(val[:5], "hello")
					copy(val[32:37], "world")
					return val
				}()},
			},
			value: func() []byte {
				val := make([]byte, 64)
				copy(val[:5], "hello")
				copy(val[32:37], "world")
				return val
			}(),
			want: true,
		},
		// Large number tests
		{
			name: "UintLt - large numbers",
			predicate: ValuePredicate{
				Op: UintLt,
				IntArgs: []*big.Int{func() *big.Int {
					val := big.NewInt(1)
					val.Lsh(val, 200) // 2^200
					return val
				}()},
			},
			value: func() []byte {
				val := big.NewInt(1)
				val.Lsh(val, 199) // 2^199
				bytes := val.Bytes()
				// Pad to 32 bytes
				result := make([]byte, 32)
				copy(result[32-len(bytes):], bytes)
				return result
			}(),
			want: true,
		},
		{
			name: "UintGt - large numbers",
			predicate: ValuePredicate{
				Op: UintGt,
				IntArgs: []*big.Int{func() *big.Int {
					val := big.NewInt(1)
					val.Lsh(val, 200) // 2^200
					return val
				}()},
			},
			value: func() []byte {
				val := big.NewInt(1)
				val.Lsh(val, 201) // 2^201
				bytes := val.Bytes()
				// Pad to 32 bytes
				result := make([]byte, 32)
				copy(result[32-len(bytes):], bytes)
				return result
			}(),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.predicate.Match(tt.value)
			assert.NilError(t, err, "Match should not return an error")
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValuePredicateEncodeDecode(t *testing.T) {
	tests := []struct {
		name      string
		predicate ValuePredicate
	}{
		{
			name: "UintLt predicate",
			predicate: ValuePredicate{
				Op:       UintLt,
				IntArgs:  []*big.Int{big.NewInt(100)},
				ByteArgs: [][]byte{},
			},
		},
		{
			name: "UintLte predicate",
			predicate: ValuePredicate{
				Op:       UintLte,
				IntArgs:  []*big.Int{big.NewInt(500)},
				ByteArgs: [][]byte{},
			},
		},
		{
			name: "UintEq predicate",
			predicate: ValuePredicate{
				Op:       UintEq,
				IntArgs:  []*big.Int{big.NewInt(42)},
				ByteArgs: [][]byte{},
			},
		},
		{
			name: "UintGt predicate",
			predicate: ValuePredicate{
				Op:       UintGt,
				IntArgs:  []*big.Int{big.NewInt(200)},
				ByteArgs: [][]byte{},
			},
		},
		{
			name: "UintGte predicate",
			predicate: ValuePredicate{
				Op:       UintGte,
				IntArgs:  []*big.Int{big.NewInt(1000)},
				ByteArgs: [][]byte{},
			},
		},
		{
			name: "BytesEq predicate with 32 bytes",
			predicate: ValuePredicate{
				Op:      BytesEq,
				IntArgs: []*big.Int{},
				ByteArgs: [][]byte{
					common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111").Bytes(),
				},
			},
		},
		{
			name: "BytesEq predicate with different 32 bytes",
			predicate: ValuePredicate{
				Op:      BytesEq,
				IntArgs: []*big.Int{},
				ByteArgs: [][]byte{
					common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef").Bytes(),
				},
			},
		},
		{
			name: "UintEq predicate with zero value",
			predicate: ValuePredicate{
				Op:       UintEq,
				IntArgs:  []*big.Int{big.NewInt(0)},
				ByteArgs: [][]byte{},
			},
		},
		{
			name: "UintLt predicate with large number",
			predicate: ValuePredicate{
				Op: UintLt,
				IntArgs: []*big.Int{func() *big.Int {
					val := big.NewInt(1)
					val.Lsh(val, 200) // 2^200
					return val
				}()},
				ByteArgs: [][]byte{},
			},
		},
		{
			name: "BytesEq predicate with all zeros",
			predicate: ValuePredicate{
				Op:       BytesEq,
				IntArgs:  []*big.Int{},
				ByteArgs: [][]byte{make([]byte, 32)}, // All zeros
			},
		},
		{
			name: "BytesEq predicate with all 0xFF",
			predicate: ValuePredicate{
				Op:      BytesEq,
				IntArgs: []*big.Int{},
				ByteArgs: [][]byte{func() []byte {
					data := make([]byte, 32)
					for i := range data {
						data[i] = 0xFF
					}
					return data
				}()},
			},
		},
		{
			name: "UintGte predicate with maximum uint64 value",
			predicate: ValuePredicate{
				Op:       UintGte,
				IntArgs:  []*big.Int{big.NewInt(int64(^uint64(0) >> 1))}, // Max int64 value
				ByteArgs: [][]byte{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encoding
			var buf bytes.Buffer
			err := tt.predicate.EncodeRLP(&buf)
			assert.NilError(t, err, "EncodeRLP should not fail")

			encodedData := buf.Bytes()
			assert.Assert(t, len(encodedData) > 0, "Encoded data should not be empty")

			// Test decoding
			var decoded ValuePredicate
			err = rlp.DecodeBytes(encodedData, &decoded)
			assert.NilError(t, err, "DecodeRLP should not fail")

			// Compare the original and decoded predicates
			assert.Equal(t, tt.predicate.Op, decoded.Op, "Operation should match")
			assert.Equal(t, len(tt.predicate.IntArgs), len(decoded.IntArgs), "Number of int args should match")
			assert.Equal(t, len(tt.predicate.ByteArgs), len(decoded.ByteArgs), "Number of byte args should match")

			// Compare integer arguments
			for i, originalIntArg := range tt.predicate.IntArgs {
				decodedIntArg := decoded.IntArgs[i]
				assert.Equal(t, originalIntArg.Cmp(decodedIntArg), 0, "Integer argument %d should match", i)
			}

			// Compare byte arguments
			for i, originalByteArg := range tt.predicate.ByteArgs {
				decodedByteArg := decoded.ByteArgs[i]
				assert.DeepEqual(t, originalByteArg, decodedByteArg)
			}

			// Test round-trip: encode the decoded predicate and compare with original encoding
			var buf2 bytes.Buffer
			err = decoded.EncodeRLP(&buf2)
			assert.NilError(t, err, "Second encoding should not fail")
			assert.DeepEqual(t, encodedData, buf2.Bytes())
		})
	}
}

func TestValuePredicateDecodeErrors(t *testing.T) {
	tests := []struct {
		name        string
		encodedData []byte
		expectedErr string
	}{
		{
			name:        "empty data",
			encodedData: []byte{},
			expectedErr: "failed to decode ValuePredicate",
		},
		{
			name:        "invalid RLP data",
			encodedData: []byte{0xFF, 0xFF, 0xFF},
			expectedErr: "failed to decode ValuePredicate",
		},
		{
			name: "invalid operation value",
			encodedData: func() []byte {
				// Manually encode an invalid operation
				var buf bytes.Buffer
				elements := []interface{}{uint64(999)} // Invalid operation
				err := rlp.Encode(&buf, elements)
				assert.NilError(t, err, "Encoding should not fail")
				return buf.Bytes()
			}(),
			expectedErr: "invalid operation",
		},
		{
			name: "missing integer argument for UintLt",
			encodedData: func() []byte {
				// Encode UintLt operation but without the required integer argument
				var buf bytes.Buffer
				elements := []interface{}{uint64(UintLt)} // Missing the integer argument
				err := rlp.Encode(&buf, elements)
				assert.NilError(t, err, "Encoding should not fail")
				return buf.Bytes()
			}(),
			expectedErr: "failed to read integer argument",
		},
		{
			name: "missing byte argument for BytesEq",
			encodedData: func() []byte {
				// Encode BytesEq operation but without the required byte argument
				var buf bytes.Buffer
				elements := []interface{}{uint64(BytesEq)} // Missing the byte argument
				err := rlp.Encode(&buf, elements)
				assert.NilError(t, err, "Encoding should not fail")
				return buf.Bytes()
			}(),
			expectedErr: "failed to read byte argument",
		},
		{
			name: "too many elements for UintEq operation",
			encodedData: func() []byte {
				// Encode UintEq with two integer arguments instead of one
				var buf bytes.Buffer
				elements := []interface{}{
					uint64(UintEq),  // Operation
					big.NewInt(42),  // First integer argument (valid)
					big.NewInt(100), // Second integer argument (invalid - UintEq only needs one)
				}
				err := rlp.Encode(&buf, elements)
				assert.NilError(t, err, "Encoding should not fail")
				return buf.Bytes()
			}(),
			expectedErr: "failed to decode ValuePredicate",
		},
		{
			name: "too many elements for BytesEq operation",
			encodedData: func() []byte {
				// Encode BytesEq with two byte arguments instead of one
				var buf bytes.Buffer
				elements := []interface{}{
					uint64(BytesEq),  // Operation
					make([]byte, 32), // First byte argument (valid)
					make([]byte, 32), // Second byte argument (invalid - BytesEq only needs one)
				}
				err := rlp.Encode(&buf, elements)
				assert.NilError(t, err, "Encoding should not fail")
				return buf.Bytes()
			}(),
			expectedErr: "failed to decode ValuePredicate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var decoded ValuePredicate
			err := rlp.DecodeBytes(tt.encodedData, &decoded)
			assert.Assert(t, err != nil, "DecodeRLP should fail for invalid data")
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}
}

func TestLogValueRefGetValue(t *testing.T) {
	tests := []struct {
		name string
		ref  LogValueRef
		log  *types.Log
		want []byte
	}{
		// Topic tests
		{
			name: "get topic 0",
			ref:  LogValueRef{Offset: 0},
			log: &types.Log{
				Topics: []common.Hash{
					common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"),
				},
				Data: make([]byte, 64),
			},
			want: common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef").Bytes(),
		},
		{
			name: "get topic 1",
			ref:  LogValueRef{Offset: 1},
			log: &types.Log{
				Topics: []common.Hash{
					common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
					common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"),
				},
				Data: make([]byte, 64),
			},
			want: common.HexToHash("0xfedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321").Bytes(),
		},
		{
			name: "get topic 3 (last valid topic)",
			ref:  LogValueRef{Offset: 3},
			log: &types.Log{
				Topics: []common.Hash{
					common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
					common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
					common.HexToHash("0x3333333333333333333333333333333333333333333333333333333333333333"),
					common.HexToHash("0x4444444444444444444444444444444444444444444444444444444444444444"),
				},
				Data: make([]byte, 64),
			},
			want: common.HexToHash("0x4444444444444444444444444444444444444444444444444444444444444444").Bytes(),
		},
		{
			name: "get topic that doesn't exist - returns nil",
			ref:  LogValueRef{Offset: 2},
			log: &types.Log{
				Topics: []common.Hash{
					common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
				},
				Data: make([]byte, 64),
			},
			want: nil,
		},
		// Data tests
		{
			name: "get first data word (offset 4, length 1)",
			ref: LogValueRef{
				Offset: 4,
			},
			log: &types.Log{
				Topics: []common.Hash{
					common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
				},
				Data: func() []byte {
					data := make([]byte, 64)
					// First word: 0x1111...1111
					for i := 0; i < 32; i++ {
						data[i] = 0x11
					}
					// Second word: 0x2222...2222
					for i := 32; i < 64; i++ {
						data[i] = 0x22
					}
					return data
				}(),
			},
			want: func() []byte {
				result := make([]byte, 32)
				for i := 0; i < 32; i++ {
					result[i] = 0x11
				}
				return result
			}(),
		},
		{
			name: "get second data word (offset 5, length 1)",
			ref:  LogValueRef{Offset: 5},
			log: &types.Log{
				Topics: []common.Hash{
					common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
				},
				Data: func() []byte {
					data := make([]byte, 64)
					// First word: 0x1111...1111
					for i := 0; i < 32; i++ {
						data[i] = 0x11
					}
					// Second word: 0x2222...2222
					for i := 32; i < 64; i++ {
						data[i] = 0x22
					}
					return data
				}(),
			},
			want: func() []byte {
				result := make([]byte, 32)
				for i := 0; i < 32; i++ {
					result[i] = 0x22
				}
				return result
			}(),
		},

		{
			name: "get data beyond log length - zero padded",
			// Third word, but log only has 2 words
			ref: LogValueRef{Offset: 6},
			log: &types.Log{
				Topics: []common.Hash{
					common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
				},
				Data: func() []byte {
					data := make([]byte, 64) // Only 2 words
					for i := 0; i < 64; i++ {
						data[i] = 0xff
					}
					return data
				}(),
			},
			want: make([]byte, 32), // Should return 32 zero bytes
		},
		{
			name: "get data from empty log data",
			ref:  LogValueRef{Offset: 4},
			log: &types.Log{
				Topics: []common.Hash{
					common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
				},
				Data: []byte{}, // Empty data
			},
			want: make([]byte, 32), // Should return 32 zero bytes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ref.GetValue(tt.log)
			assert.DeepEqual(t, tt.want, result)
		})
	}
}

func TestEventTriggerDefinitionValidate(t *testing.T) {
	tests := []struct {
		name       string
		definition EventTriggerDefinition
		wantErr    bool
	}{
		{
			name: "valid definition with no predicates",
			definition: EventTriggerDefinition{
				Contract:      common.HexToAddress("0x1234567890123456789012345678901234567890"),
				LogPredicates: []LogPredicate{},
			},
			wantErr: false,
		},
		{
			name: "valid definition with single predicate",
			definition: EventTriggerDefinition{
				Contract: common.HexToAddress("0x1234567890123456789012345678901234567890"),
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{make([]byte, 32)},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid definition with invalid log predicate",
			definition: EventTriggerDefinition{
				Contract: common.HexToAddress("0x1234567890123456789012345678901234567890"),
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid definition with duplicate BytesEq predicates for same topic",
			definition: EventTriggerDefinition{
				Contract: common.HexToAddress("0x1234567890123456789012345678901234567890"),
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{make([]byte, 32)},
						},
					},
					{
						// Same topic as previous predicate
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{make([]byte, 32)},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid definition with multiple non-BytesEq predicates for same topic",
			definition: EventTriggerDefinition{
				Contract: common.HexToAddress("0x1234567890123456789012345678901234567890"),
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       UintLt, // Not BytesEq, so multiple predicates allowed
							IntArgs:  []*big.Int{big.NewInt(100)},
							ByteArgs: [][]byte{},
						},
					},
					{
						// Same topic, but different operation
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       UintGt, // Not BytesEq, so allowed
							IntArgs:  []*big.Int{big.NewInt(50)},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.definition.Validate()
			if tt.wantErr {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestEventTriggerDefinitionToFilterQuery(t *testing.T) {
	contractAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	tests := []struct {
		name       string
		definition EventTriggerDefinition
		wantQuery  ethereum.FilterQuery
		wantErr    bool
	}{
		{
			name: "definition with no predicates",
			definition: EventTriggerDefinition{
				Contract:      contractAddr,
				LogPredicates: []LogPredicate{},
			},
			wantQuery: ethereum.FilterQuery{
				Addresses: []common.Address{contractAddr},
				Topics:    [][]common.Hash{},
			},
			wantErr: false,
		},
		{
			name: "definition with single topic predicate",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{make([]byte, 32)},
						},
					},
				},
			},
			wantQuery: ethereum.FilterQuery{
				Addresses: []common.Address{contractAddr},
				Topics:    [][]common.Hash{{common.BytesToHash(make([]byte, 32))}},
			},
			wantErr: false,
		},
		{
			name: "definition with multiple topic predicates",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111").Bytes()},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 2},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222").Bytes()},
						},
					},
				},
			},
			wantQuery: ethereum.FilterQuery{
				Addresses: []common.Address{contractAddr},
				Topics: [][]common.Hash{
					{common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")},
					{},
					{common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222")},
				},
			},
			wantErr: false,
		},
		{
			name: "definition with data predicate (ignored in filter)",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 4},
						ValuePredicate: ValuePredicate{
							Op:       UintLt,
							IntArgs:  []*big.Int{big.NewInt(100)},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
			wantQuery: ethereum.FilterQuery{
				Addresses: []common.Address{contractAddr},
				Topics:    [][]common.Hash{},
			},
			wantErr: false,
		},
		{
			name: "definition with non-BytesEq topic predicate (ignored in filter)",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 1},
						ValuePredicate: ValuePredicate{
							Op:       UintGt,
							IntArgs:  []*big.Int{big.NewInt(50)},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
			wantQuery: ethereum.FilterQuery{
				Addresses: []common.Address{contractAddr},
				Topics:    [][]common.Hash{},
			},
			wantErr: false,
		},
		{
			name: "definition with duplicate BytesEq predicates for same topic",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{make([]byte, 32)},
						},
					},
					{
						// Same topic
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{make([]byte, 32)},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "definition with invalid topic byte length",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{make([]byte, 16)}, // Invalid: not 32 bytes
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := tt.definition.ToFilterQuery()
			if tt.wantErr {
				assert.Assert(t, err != nil)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, tt.wantQuery.Addresses, query.Addresses)
				assert.DeepEqual(t, tt.wantQuery.Topics, query.Topics)
				// FromBlock, ToBlock, and BlockHash should be nil by default
				assert.Assert(t, query.FromBlock == nil)
				assert.Assert(t, query.ToBlock == nil)
				assert.Assert(t, query.BlockHash == nil)
			}
		})
	}
}

func TestEventTriggerDefinitionMatch(t *testing.T) {
	contractAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	tests := []struct {
		name       string
		definition EventTriggerDefinition
		log        *types.Log
		want       bool
	}{
		{
			name: "definition with no predicates matches any log",
			definition: EventTriggerDefinition{
				Contract:      contractAddr,
				LogPredicates: []LogPredicate{},
			},
			log: &types.Log{
				Address: contractAddr,
				Topics: []common.Hash{
					common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
				},
				Data: make([]byte, 64),
			},
			want: true,
		},
		{
			name: "log from different contract address should not match",
			definition: EventTriggerDefinition{
				Contract:      contractAddr,
				LogPredicates: []LogPredicate{},
			},
			log: &types.Log{
				Address: common.HexToAddress("0x9999999999999999999999999999999999999999"), // Different contract
				Topics: []common.Hash{
					common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
				},
				Data: make([]byte, 64),
			},
			want: false,
		},
		{
			name: "single matching topic predicate",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111").Bytes()},
						},
					},
				},
			},
			log: &types.Log{
				Address: contractAddr,
				Topics: []common.Hash{
					common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
				},
				Data: make([]byte, 64),
			},
			want: true,
		},
		{
			name: "single non-matching topic predicate",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111").Bytes()},
						},
					},
				},
			},
			log: &types.Log{
				Address: contractAddr,
				Topics: []common.Hash{
					common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
				},
				Data: make([]byte, 64),
			},
			want: false,
		},
		{
			name: "multiple matching predicates",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111").Bytes()},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 1},
						ValuePredicate: ValuePredicate{
							Op:       UintGt,
							IntArgs:  []*big.Int{big.NewInt(50)},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
			log: &types.Log{
				Address: contractAddr,
				Topics: []common.Hash{
					common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
					func() common.Hash {
						val := make([]byte, 32)
						copy(val[31:], big.NewInt(100).Bytes())
						return common.BytesToHash(val)
					}(),
				},
				Data: make([]byte, 64),
			},
			want: true,
		},
		{
			name: "multiple predicates with one not matching",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111").Bytes()},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 1},
						ValuePredicate: ValuePredicate{
							Op:       UintGt,
							IntArgs:  []*big.Int{big.NewInt(200)},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
			log: &types.Log{
				Address: contractAddr,
				Topics: []common.Hash{
					common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
					func() common.Hash {
						val := make([]byte, 32)
						copy(val[31:], big.NewInt(100).Bytes()) // 100 is not > 200
						return common.BytesToHash(val)
					}(),
				},
				Data: make([]byte, 64),
			},
			want: false,
		},
		{
			name: "data predicate matching",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{
							Offset: 4,
						},
						ValuePredicate: ValuePredicate{
							Op:       UintLt,
							IntArgs:  []*big.Int{big.NewInt(1000)},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
			log: &types.Log{
				Address: contractAddr,
				Topics: []common.Hash{
					common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
				},
				Data: func() []byte {
					data := make([]byte, 64)
					// First word: 500 (which is < 1000)
					copy(data[30:32], big.NewInt(500).Bytes())
					return data
				}(),
			},
			want: true,
		},
		{
			name: "data predicate not matching",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{
							Offset: 4,
						},
						ValuePredicate: ValuePredicate{
							Op:       UintLt,
							IntArgs:  []*big.Int{big.NewInt(100)},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
			log: &types.Log{
				Address: contractAddr,
				Topics: []common.Hash{
					common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
				},
				Data: func() []byte {
					data := make([]byte, 64)
					// First word: 500 (which is not < 100)
					copy(data[28:32], big.NewInt(500).Bytes())
					return data
				}(),
			},
			want: false,
		},
		{
			name: "topic reference that doesn't exist in log",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						// Topic index 2, but log only has 1 topic
						LogValueRef: LogValueRef{Offset: 2},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{make([]byte, 32)},
						},
					},
				},
			},
			log: &types.Log{
				Address: contractAddr,
				Topics: []common.Hash{
					common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
				},
				Data: make([]byte, 64),
			},
			want: false,
		},
		{
			name: "mixed topic and data predicates all matching",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111").Bytes()},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 4},
						ValuePredicate: ValuePredicate{
							Op:       UintEq,
							IntArgs:  []*big.Int{big.NewInt(42)},
							ByteArgs: [][]byte{},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 5},
						ValuePredicate: ValuePredicate{
							Op:       UintGte,
							IntArgs:  []*big.Int{big.NewInt(100)},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
			log: &types.Log{
				Address: contractAddr,
				Topics: []common.Hash{
					common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
				},
				Data: func() []byte {
					data := make([]byte, 96) // 3 words
					// First word: 42
					copy(data[31:32], big.NewInt(42).Bytes())
					// Second word: 150
					copy(data[63:64], big.NewInt(150).Bytes())
					return data
				}(),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.definition.Match(tt.log)
			assert.NilError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestEventTriggerDefinitionMarshalUnmarshal(t *testing.T) {
	contractAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	tests := []struct {
		name       string
		definition EventTriggerDefinition
	}{
		{
			name: "empty definition with no predicates",
			definition: EventTriggerDefinition{
				Contract:      contractAddr,
				LogPredicates: []LogPredicate{},
			},
		},
		{
			name: "definition with single BytesEq predicate",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111").Bytes()},
						},
					},
				},
			},
		},
		{
			name: "definition with single UintLt predicate",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 4},
						ValuePredicate: ValuePredicate{
							Op:       UintLt,
							IntArgs:  []*big.Int{big.NewInt(1000)},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
		},
		{
			name: "definition with multiple predicates of different types",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111").Bytes()},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 1},
						ValuePredicate: ValuePredicate{
							Op:       UintGt,
							IntArgs:  []*big.Int{big.NewInt(50)},
							ByteArgs: [][]byte{},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 4},
						ValuePredicate: ValuePredicate{
							Op:       UintEq,
							IntArgs:  []*big.Int{big.NewInt(42)},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
		},
		{
			name: "definition with large integer values",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 4},
						ValuePredicate: ValuePredicate{
							Op: UintLte,
							IntArgs: []*big.Int{func() *big.Int {
								val := big.NewInt(1)
								val.Lsh(val, 400) // Very large number
								return val
							}()},
							ByteArgs: [][]byte{},
						},
					},
				},
			},
		},
		{
			name: "definition with BytesEq predicate on multi-word data",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 4},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{make([]byte, 64)}, // 64 bytes for 2 words
						},
					},
				},
			},
		},
		{
			name: "definition with all operation types",
			definition: EventTriggerDefinition{
				Contract: contractAddr,
				LogPredicates: []LogPredicate{
					{
						LogValueRef: LogValueRef{Offset: 0},
						ValuePredicate: ValuePredicate{
							Op:       UintLt,
							IntArgs:  []*big.Int{big.NewInt(100)},
							ByteArgs: [][]byte{},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 1},
						ValuePredicate: ValuePredicate{
							Op:       UintLte,
							IntArgs:  []*big.Int{big.NewInt(200)},
							ByteArgs: [][]byte{},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 2},
						ValuePredicate: ValuePredicate{
							Op:       UintEq,
							IntArgs:  []*big.Int{big.NewInt(42)},
							ByteArgs: [][]byte{},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 3},
						ValuePredicate: ValuePredicate{
							Op:       UintGt,
							IntArgs:  []*big.Int{big.NewInt(300)},
							ByteArgs: [][]byte{},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 4},
						ValuePredicate: ValuePredicate{
							Op:       UintGte,
							IntArgs:  []*big.Int{big.NewInt(400)},
							ByteArgs: [][]byte{},
						},
					},
					{
						LogValueRef: LogValueRef{Offset: 5},
						ValuePredicate: ValuePredicate{
							Op:       BytesEq,
							IntArgs:  []*big.Int{},
							ByteArgs: [][]byte{common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef").Bytes()},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal the original definition
			marshaled := tt.definition.MarshalBytes()

			// Unmarshal to get a new definition
			var unmarshaled EventTriggerDefinition
			err := unmarshaled.UnmarshalBytes(marshaled)
			assert.NilError(t, err, "UnmarshalBytes should not fail")

			// Compare the original and unmarshaled definitions
			assert.DeepEqual(t, tt.definition.Contract, unmarshaled.Contract)
			assert.Equal(t, len(tt.definition.LogPredicates), len(unmarshaled.LogPredicates))

			for i, originalPredicate := range tt.definition.LogPredicates {
				unmarshaledPredicate := unmarshaled.LogPredicates[i]

				// Compare LogValueRef
				assert.Equal(t, originalPredicate.LogValueRef.Offset,
					unmarshaledPredicate.LogValueRef.Offset)

				// Compare ValuePredicate
				assert.Equal(t, originalPredicate.ValuePredicate.Op, unmarshaledPredicate.ValuePredicate.Op)

				// Compare IntArgs
				assert.Equal(t, len(originalPredicate.ValuePredicate.IntArgs),
					len(unmarshaledPredicate.ValuePredicate.IntArgs))
				for j, originalIntArg := range originalPredicate.ValuePredicate.IntArgs {
					unmarshaledIntArg := unmarshaledPredicate.ValuePredicate.IntArgs[j]
					assert.Equal(t, originalIntArg.Cmp(unmarshaledIntArg), 0)
				}

				// Compare ByteArgs
				assert.Equal(t, len(originalPredicate.ValuePredicate.ByteArgs),
					len(unmarshaledPredicate.ValuePredicate.ByteArgs))
				for j, originalByteArg := range originalPredicate.ValuePredicate.ByteArgs {
					unmarshaledByteArg := unmarshaledPredicate.ValuePredicate.ByteArgs[j]
					assert.DeepEqual(t, originalByteArg, unmarshaledByteArg)
				}
			}

			// Additional validation: marshal the unmarshaled definition and compare bytes
			remarshaled := unmarshaled.MarshalBytes()
			assert.NilError(t, err)
			assert.DeepEqual(t, marshaled, remarshaled)
		})
	}
}

func TestEventTriggerDefinitionUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "invalid version",
			data: []byte{0x99}, // Wrong version
		},
		{
			name: "old version",
			data: []byte{0x1}, // Old version
		},
		{
			name: "version only, no RLP data",
			data: []byte{Version},
		},
		{
			name: "invalid RLP data",
			data: []byte{Version, 0xFF, 0xFF, 0xFF}, // Invalid RLP
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var definition EventTriggerDefinition
			err := definition.UnmarshalBytes(tt.data)
			assert.Assert(t, err != nil, "UnmarshalBytes should fail for invalid data")
		})
	}
}

func TestLogValueRefEncodeRLP(t *testing.T) {
	tests := []struct {
		name     string
		ref      LogValueRef
		expected []byte
	}{
		{
			name:     "topic reference - offset 0",
			ref:      LogValueRef{Offset: 0},
			expected: []byte{0x80}, // RLP encoding of uint64(0)
		},
		{
			name:     "topic reference - offset 3",
			ref:      LogValueRef{Offset: 3},
			expected: []byte{0x03}, // RLP encoding of uint64(3)
		},
		{
			name:     "data reference - offset 4",
			ref:      LogValueRef{Offset: 4},
			expected: []byte{0x04}, // RLP encoding of uint64(4)
		},
		{
			name:     "data reference - offset 5",
			ref:      LogValueRef{Offset: 5},
			expected: []byte{0x05}, // RLP encoding of uint64(5)
		},
		{
			name:     "data reference - offset 10",
			ref:      LogValueRef{Offset: 10},
			expected: []byte{0x0a}, // RLP encoding of uint64(10)
		},
		{
			name:     "large offset",
			ref:      LogValueRef{Offset: 1000},
			expected: []byte{0x82, 0x03, 0xe8}, // RLP encoding of uint64(1000)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tt.ref.EncodeRLP(&buf)
			assert.NilError(t, err, "EncodeRLP should not fail")
			assert.DeepEqual(t, tt.expected, buf.Bytes())
		})
	}
}

func TestLogValueRefDecodeRLP(t *testing.T) {
	tests := []struct {
		name     string
		encoded  []byte
		expected LogValueRef
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "topic reference - offset 0",
			encoded:  []byte{0x80}, // RLP encoding of uint64(0)
			expected: LogValueRef{Offset: 0},
			wantErr:  false,
		},
		{
			name:     "topic reference - offset 3",
			encoded:  []byte{0x03}, // RLP encoding of uint64(3)
			expected: LogValueRef{Offset: 3},
			wantErr:  false,
		},
		{
			name:     "data reference - offset 4",
			encoded:  []byte{0x04}, // RLP encoding of uint64(4)
			expected: LogValueRef{Offset: 4},
			wantErr:  false,
		},
		{
			name:     "data reference - offset 5",
			encoded:  []byte{0x05}, // RLP encoding of uint64(5)
			expected: LogValueRef{Offset: 5},
			wantErr:  false,
		},
		{
			name:     "data reference - offset 10",
			encoded:  []byte{0x0a}, // RLP encoding of uint64(10)
			expected: LogValueRef{Offset: 10},
			wantErr:  false,
		},
		{
			name:     "large offset",
			encoded:  []byte{0x82, 0x03, 0xe8}, // RLP encoding of uint64(1000)
			expected: LogValueRef{Offset: 1000},
			wantErr:  false,
		},
		{
			name:    "invalid - empty RLP data",
			encoded: []byte{},
			wantErr: true,
			errMsg:  "failed to decode LogValueRef",
		},
		{
			name:    "invalid - malformed RLP",
			encoded: []byte{0xFF, 0xFF},
			wantErr: true,
			errMsg:  "failed to decode LogValueRef",
		},
		{
			name:    "invalid - incomplete list",
			encoded: []byte{0xc1, 0x04}, // List with only one element
			wantErr: true,
			errMsg:  "LogValueRef can't be a list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref LogValueRef
			err := rlp.DecodeBytes(tt.encoded, &ref)
			fmt.Println(ref)

			if tt.wantErr {
				assert.Assert(t, err != nil, "DecodeRLP should fail")
				assert.ErrorContains(t, err, tt.errMsg)
			} else {
				assert.NilError(t, err, "DecodeRLP should not fail")
				assert.Equal(t, tt.expected.Offset, ref.Offset)
			}
		})
	}
}

func TestLogValueRefRLPRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		ref  LogValueRef
	}{
		{
			name: "topic reference - offset 0",
			ref:  LogValueRef{Offset: 0},
		},
		{
			name: "topic reference - offset 3",
			ref:  LogValueRef{Offset: 3},
		},
		{
			name: "data reference - single word",
			ref:  LogValueRef{Offset: 4},
		},
		{
			name: "data reference - multiple words",
			ref:  LogValueRef{Offset: 5},
		},
		{
			name: "large values",
			ref:  LogValueRef{Offset: 65535},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			var buf bytes.Buffer
			err := tt.ref.EncodeRLP(&buf)
			assert.NilError(t, err, "EncodeRLP should not fail")

			encoded := buf.Bytes()
			assert.Assert(t, len(encoded) > 0, "Encoded data should not be empty")

			// Decode
			var decoded LogValueRef
			err = rlp.DecodeBytes(encoded, &decoded)
			assert.NilError(t, err, "DecodeRLP should not fail")

			// Verify round trip
			assert.Equal(t, tt.ref.Offset, decoded.Offset)

			// Encode again and verify consistency
			var buf2 bytes.Buffer
			err = decoded.EncodeRLP(&buf2)
			assert.NilError(t, err, "Second EncodeRLP should not fail")
			assert.DeepEqual(t, encoded, buf2.Bytes())
		})
	}
}

func TestWithEVM(t *testing.T) {
	setup := help.SetupBackend(t)
	one := big.NewInt(1)
	mOne := LogPredicate{
		LogValueRef:    LogValueRef{Offset: 1},
		ValuePredicate: ValuePredicate{Op: BytesEq, ByteArgs: [][]byte{Align(one.Bytes())}},
	}
	two := "two"
	mTwo := LogPredicate{
		LogValueRef: LogValueRef{Offset: 2},
		ValuePredicate: ValuePredicate{Op: BytesEq, ByteArgs: [][]byte{
			Align(crypto.Keccak256([]byte("two"))),
		}},
	}
	three := common.BytesToAddress(big.NewInt(84).Bytes())
	mThree := LogPredicate{
		LogValueRef:    LogValueRef{Offset: 3},
		ValuePredicate: ValuePredicate{Op: BytesEq, ByteArgs: [][]byte{Align(three[:])}},
	}
	four := []byte("first and slightly longer arg that should use more space and if i am right, then this will span multiple words")
	preFour := []byte("first and slightly longer arg that should use more space and if ")
	mFoure := LogPredicate{
		LogValueRef:    LogValueRef{Offset: 4},
		ValuePredicate: ValuePredicate{Op: BytesEq, ByteArgs: [][]byte{four}},
	}
	noMFour := LogPredicate{
		LogValueRef:    LogValueRef{Offset: 4},
		ValuePredicate: ValuePredicate{Op: BytesEq, ByteArgs: [][]byte{[]byte("no match")}},
	}
	preNotFour := LogPredicate{
		LogValueRef:    LogValueRef{Offset: 4},
		ValuePredicate: ValuePredicate{Op: BytesEq, ByteArgs: [][]byte{preFour}},
	}
	five := big.NewInt(42)
	six := []byte("second arg")
	mSix := LogPredicate{
		LogValueRef:    LogValueRef{Offset: 6},
		ValuePredicate: ValuePredicate{Op: BytesEq, ByteArgs: [][]byte{six}},
	}
	tx, err := setup.Contract.EmitSix(setup.Auth, one, two, three, four, five, six)
	assert.NilError(t, err, "error creating tx")
	vLog, err := help.CollectLog(t, setup, tx)
	assert.NilError(t, err, "error getting log")

	tests := []struct {
		predicates []LogPredicate
		match      bool
		name       string
	}{
		{predicates: []LogPredicate{mOne, mTwo, mThree, mSix}, match: true, name: "match one, two, three and six"},
		{predicates: []LogPredicate{mFoure}, match: true, name: "match four"},
		{predicates: []LogPredicate{mFoure, mSix}, match: true, name: "match four and six"},
		{
			predicates: []LogPredicate{mSix, mSix},
			match:      true,
			// Note: this is legal, although not practical
			name: "match duplicate six",
		},
		{predicates: []LogPredicate{preNotFour}, match: false, name: "prefix should not match whole"},
		{predicates: []LogPredicate{noMFour, mFoure}, match: false, name: "mismatch same offset"},
		{
			predicates: []LogPredicate{
				{
					LogValueRef: LogValueRef{Offset: 5},
					ValuePredicate: ValuePredicate{
						Op:      UintEq,
						IntArgs: []*big.Int{five},
					},
				},
			},
			match: true,
			name:  "match five GTE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etd := EventTriggerDefinition{
				Contract:      setup.ContractAddress,
				LogPredicates: tt.predicates,
			}
			err := etd.Validate()
			assert.NilError(t, err, "did not validate: %v", tt.name)

			encoded := etd.MarshalBytes()
			decoded := EventTriggerDefinition{}
			err = decoded.UnmarshalBytes(encoded)
			assert.NilError(t, err, "error when roundtrip decoding")
			doubleEncoded := decoded.MarshalBytes()

			equal := cmp.DeepEqual(encoded, doubleEncoded)
			assert.Check(t, equal, "did not survive roundtrip: %v", tt.name)
			match, err := etd.Match(vLog)
			assert.NilError(t, err, "error when matching: %v. err: %v", tt.name, err)
			assert.Equal(t, match, tt.match, "did not match expectation: %v\nlog data:\t%v\nmatch data:\t%v", tt.name, vLog, etd)
			match, err = decoded.Match(vLog)
			assert.NilError(t, err, "error when matching from decoded: %v. err: %v", tt.name, err)
			assert.Equal(t, match, tt.match, "did not match expectation after roundtrip: %v", tt.name)
		})
	}
}

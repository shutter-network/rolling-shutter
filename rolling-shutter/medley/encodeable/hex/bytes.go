package hex

import (
	"bytes"
	"encoding/hex"
	"io"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable"
)

func EncodeHex(src []byte) []byte {
	// copied from the encoding/hex package, without the string->byte conversion:
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)
	return dst
}

func DecodeHex(src []byte) ([]byte, error) {
	// copied from the encoding/hex package, without string -> byte conversion:

	// We can use the source slice itself as the destination
	// because the decode loop increments by one and then the 'seen' byte is not used anymore.
	n, err := hex.Decode(src, src)
	return src[:n], err
}

type Bytes struct {
	Value []byte
}

func ReadBytes(reader io.Reader, byteSize int) (Bytes, error) {
	value := make([]byte, byteSize)
	_, err := io.ReadFull(reader, value)
	if err != nil {
		return Bytes{}, err
	}
	return Bytes{value}, nil
}

func (a *Bytes) UnmarshalText(b []byte) error {
	v, err := DecodeHex(b)
	if err != nil {
		return err
	}
	a.Value = v
	return nil
}

func (a Bytes) MarshalText() ([]byte, error) {
	return EncodeHex(a.Value), nil
}

func (a *Bytes) Equal(b *Bytes) bool {
	return bytes.Equal(a.Value, b.Value)
}

func (a *Bytes) String() string {
	return encodeable.String(a)
}

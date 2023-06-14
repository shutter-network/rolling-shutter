package encodeable

import (
	"encoding"
)

type TextEncodeable interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

func String(a encoding.TextMarshaler) string {
	res, err := a.MarshalText()
	if err != nil {
		return ""
	}
	return string(res)
}

func FromString[T encoding.TextUnmarshaler](a T, s string) error {
	err := a.UnmarshalText([]byte(s))
	if err != nil {
		return err
	}
	return nil
}

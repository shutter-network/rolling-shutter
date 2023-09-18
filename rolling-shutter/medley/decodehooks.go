package medley

import (
	"encoding"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

func TextUnmarshalerHook(from reflect.Value, to reflect.Value) (interface{}, error) {
	data := from.Interface()
	if from.Kind() != reflect.String {
		return data, nil
	}
	// create new instance
	toNew := reflect.New(to.Type())
	resultPtr := toNew.Interface()
	umshl, ok := resultPtr.(encoding.TextUnmarshaler)
	if !ok {
		return data, nil
	}
	err := umshl.UnmarshalText([]byte(data.(string)))
	if err != nil {
		return nil, err
	}
	if to.Kind() == reflect.Pointer {
		return umshl, nil
	}
	// if to type is no ptr type,
	// return the element
	return toNew.Elem().Interface(), nil
}

func TextMarshalerHook(from reflect.Value, to reflect.Value) (interface{}, error) {
	if from.Kind() == reflect.Ptr {
		from = from.Elem()
	}

	data := from.Interface()
	if to.Kind() != reflect.String {
		return data, nil
	}
	fromType := from.Type()
	result := reflect.New(fromType).Interface()
	_, ok := result.(encoding.TextMarshaler)
	if !ok {
		return data, nil
	}

	marshaller, ok := data.(encoding.TextMarshaler)
	if !ok {
		return data, nil
	}
	mshl, err := marshaller.MarshalText()
	if err != nil {
		return nil, err
	}
	return string(mshl), nil
}

func MapstructureDecode(input, result any, hookFunc mapstructure.DecodeHookFunc) error {
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			Result:     result,
			DecodeHook: hookFunc,
		})
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

func MapstructureMarshal(input, result any) error {
	return MapstructureDecode(
		input,
		result,
		mapstructure.ComposeDecodeHookFunc(
			TextMarshalerHook,
		),
	)
}

func MapstructureUnmarshal(input, result any) error {
	return MapstructureDecode(
		input,
		result,
		mapstructure.ComposeDecodeHookFunc(
			TextUnmarshalerHook,
		),
	)
}

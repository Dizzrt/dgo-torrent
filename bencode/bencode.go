package bencode

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

var (
	ErrInvalidType = errors.New("invalid type for marshal")
)

func Unmarshal(r io.Reader) (any, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	res, err := parse(br)
	return res, err
}

func Marshal(v any) (string, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return marshal(val)
}

func marshal(v reflect.Value) (string, error) {
	ret := ""
	var err error = nil

	switch v.Kind() {
	case reflect.Int:
		fallthrough
	case reflect.Int64:
		ret, _ = encodeInt64(v.Int())
	case reflect.String:
		ret, _ = encodeStr(v.String())
	case reflect.Slice:
		ret, err = marshalSlice(v)
	case reflect.Map:
		ret, err = marshalMap(v)
	case reflect.Struct:
		ret, err = marshalStruct(v)
	default:
		fmt.Println("[debug]" + v.Kind().String())
		err = ErrInvalidType
	}

	return ret, err
}

func marshalSlice(v reflect.Value) (string, error) {
	ret := "l"
	for i := 0; i < v.Len(); i++ {
		value := v.Index(i).Interface()
		res, err := marshal(reflect.ValueOf(value))
		if err != nil {
			return "", err
		}

		ret += res
	}

	ret += "e"
	return ret, nil
}

func marshalMap(v reflect.Value) (string, error) {
	ret := "d"
	for _, key := range v.MapKeys() {
		value := v.MapIndex(key)
		marshaledValue, err := marshal(value)
		if err != nil {
			return "", nil
		}

		marshaledKey, err := marshal(key)
		if err != nil {
			return "", err
		}

		ret += marshaledKey + marshaledValue
	}

	ret += "e"
	return ret, nil
}

func marshalStruct(v reflect.Value) (string, error) {
	ret := "d"
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := v.Type().Field(i)

		key := ft.Tag.Get("bencode")
		if key == "" {
			key = strings.ToLower(ft.Name)
		}

		marshaledValue, err := marshal(fv)
		if err != nil {
			return "", err
		}

		marshaledKey, err := marshal(reflect.ValueOf(key))
		if err != nil {
			return "", err
		}

		ret += marshaledKey + marshaledValue
	}

	ret += "e"
	return ret, nil
}

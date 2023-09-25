package bencode

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
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

	keys := v.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	for _, key := range keys {
		value := v.MapIndex(key).Interface()
		marshaledValue, err := marshal(reflect.ValueOf(value))
		if err != nil {
			return "", err
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
	fields := make([]reflect.StructField, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		fields[i] = v.Type().Field(i)
	}

	sort.Slice(fields, func(i, j int) bool {
		tagI := fields[i].Tag.Get("bencode")
		tagJ := fields[j].Tag.Get("bencode")

		if tagI == "" {
			tagI = strings.ToLower(fields[i].Name)
		}

		if tagJ == "" {
			tagJ = strings.ToLower(fields[j].Name)
		}

		return tagI < tagJ
	})

	ret := "d"
	for _, field := range fields {
		key := field.Tag.Get("bencode")
		if key == "" {
			key = strings.ToLower(field.Name)
		}
		value := v.FieldByName(field.Name).Interface()

		marshaledKey, err := marshal(reflect.ValueOf(key))
		if err != nil {
			return "", err
		}

		marshaledValue, err := marshal(reflect.ValueOf(value))
		if err != nil {
			return "", err
		}

		ret += marshaledKey + marshaledValue
	}

	ret += "e"
	return ret, nil
}

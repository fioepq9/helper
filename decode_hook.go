package helper

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mitchellh/mapstructure"
)

func StringToSliceHookFunc(sep string) mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if to != reflect.SliceOf(from) {
			return data, nil
		}
		s := data.(string)
		return strings.Split(s, sep), nil
	}
}

func UnmarshalToStructHookFunc(unmarshal func(in []byte, out any) error) mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if to.Kind() != reflect.Struct {
			return data, nil
		}
		s := data.(string)
		out := reflect.New(to).Interface()
		err := unmarshal([]byte(s), out)
		return out, err
	}
}

func UnmarshalToMapHookFunc(unmarshal func(in []byte, out any) error) mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if to.Kind() != reflect.Map {
			return data, nil
		}
		s := data.(string)
		out := reflect.MakeMap(to).Interface()
		err := unmarshal([]byte(s), &out)
		return out, err
	}
}

func UnmarshalToSliceHookFunc(unmarshal func(in []byte, out any) error) mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if to.Kind() != reflect.Slice {
			return data, nil
		}
		s := data.(string)
		out := reflect.MakeSlice(to, 0, 0).Interface()
		err := unmarshal([]byte(s), &out)
		return out, err
	}
}

func StringToBoolHookFunc() mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if to != reflect.TypeOf(true) {
			return data, nil
		}
		s := data.(string)
		return s == "true", nil
	}
}

func StringToIntHookFunc() mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if to != reflect.TypeOf(1) {
			return data, nil
		}
		s := data.(string)
		return strconv.Atoi(s)
	}
}

func StringToFloat64HookFunc() mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if to != reflect.TypeOf(1.0) {
			return data, nil
		}
		s := data.(string)
		return strconv.ParseFloat(s, 64)
	}
}

func StringToBytesHookFunc() mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if to != reflect.TypeOf([]byte{}) {
			return data, nil
		}
		s := data.(string)
		return []byte(s), nil
	}
}

func StringToTimeHookFunc() mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		if to != reflect.TypeOf(time.Time{}) {
			return data, nil
		}
		s := data.(string)
		if !strings.HasPrefix(s, "now") {
			return data, errors.New("cannot find keyword now")
		}
		s = strings.TrimPrefix(s, "now")
		now := time.Now()
		if len(s) == 0 {
			return now, nil
		}
		verb, s := s[0], s[1:]
		d, err := time.ParseDuration(s)
		if err != nil {
			return data, errors.Wrapf(err, "cannot parse %s to time.Duration", s)
		}
		if verb == '+' {
			return now.Add(d), nil
		} else if verb == '-' {
			return now.Add(-d), nil
		} else {
			return data, errors.Newf("unsupported verb %c", verb)
		}
	}
}

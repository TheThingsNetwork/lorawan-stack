// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var errFormat = errors.DefineInvalidArgument("format", "invalid format `{input}`")

// stringToTimeHookFunc is a hook for mapstructure that decodes strings to time.Time.
func stringToTimeHookFunc(layout string) mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}
		return time.Parse(layout, data.(string))
	}
}

// stringSliceToStringMapHookFunc is a hook for mapstructure that decodes []string to map[string]string.
func stringSliceToStringMapHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Slice || f.Elem().Kind() != reflect.String ||
		t.Kind() != reflect.Map || t.Elem().Kind() != reflect.String {
		return data, nil
	}
	sl := data.([]string)
	m := make(map[string]string, len(sl))
	for _, s := range sl {
		p := strings.SplitN(s, "=", 2)
		if len(p) != 2 {
			return nil, errFormat.WithAttributes("input", s)
		}
		m[p[0]] = p[1]
	}

	return m, nil
}

// stringSliceToStringMapSliceHookFunc is a hook for mapstructure that decodes []string to map[string][]string.
// For example: [a=b a=c d=e] -> map[string][]string{a:[b c], d:[e]}
func stringSliceToStringMapStringSliceHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if (f.Kind() != reflect.String && (f.Kind() != reflect.Slice || f.Elem().Kind() != reflect.String)) || t.Kind() != reflect.Map || t.Elem().Kind() != reflect.Slice || t.Elem().Elem().Kind() != reflect.String {
		return data, nil
	}

	var slice []string
	switch v := data.(type) {
	case []string:
		slice = v
	case string:
		slice = strings.Fields(v)
	}

	m := make(map[string][]string, len(slice))

	for _, s := range slice {
		p := strings.SplitN(s, "=", 2)
		if len(p) != 2 {
			return nil, errFormat.WithAttributes("input", s)
		}
		v := m[p[0]]
		if v == nil {
			v = make([]string, 0, 1)
		}
		m[p[0]] = append(v, p[1])
	}

	return m, nil
}

// stringSliceToStringHookFunc is a hook for mapstructure that decodes []string to string by picking the first element.
func stringSliceToStringHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Slice || f.Elem().Kind() != reflect.String || t.Kind() != reflect.String {
		return data, nil
	}

	slice := data.([]string)

	if len(slice) >= 1 {
		return slice[0], nil
	}

	return "", nil
}

// stringToStringMapHookFunc is a hook for mapstructure that decodes string to map[string]string.
func stringToStringMapHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t.Kind() != reflect.Map || t.Elem().Kind() != reflect.String {
		return data, nil
	}

	str := data.(string)
	slice := strings.Fields(str)

	m := make(map[string]string, len(slice))
	for _, s := range slice {
		p := strings.SplitN(s, "=", 2)
		if len(p) != 2 {
			return nil, errFormat.WithAttributes("input", s)
		}
		m[p[0]] = p[1]
	}

	return m, nil
}

// stringToBufferMapHookFunc is a hook for mapstructure that decodes string or []string to map[string][]byte.
func stringToBufferMapHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if (f.Kind() != reflect.String && (f.Kind() != reflect.Slice || f.Elem().Kind() != reflect.String)) ||
		t.Kind() != reflect.Map || t.Elem().Kind() != reflect.Slice || t.Elem().Elem().Kind() != reflect.Uint8 {
		return data, nil
	}

	var slice []string
	switch v := data.(type) {
	case []string:
		slice = v
	case string:
		slice = strings.Fields(v)
	}

	m := make(map[string][]byte, len(slice))
	for _, s := range slice {
		p := strings.SplitN(s, "=", 2)
		if len(p) != 2 {
			return nil, errFormat.WithAttributes("input", s)
		}
		str := strings.TrimPrefix(p[1], "0x")
		buf, err := hex.DecodeString(str)
		if err != nil {
			return nil, errFormat.WithAttributes("input", s)
		}
		m[p[0]] = buf
	}

	return m, nil
}

func configurableInterfaceHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || !isConfigurableType(t) {
		return data, nil
	}

	str := data.(string)

	if t.Kind() == reflect.Ptr {
		rv := reflect.New(t.Elem())
		if err := rv.Interface().(Configurable).UnmarshalConfigString(str); err != nil {
			return nil, err
		}
		return rv.Interface(), nil
	}
	rv := reflect.New(t)
	if err := rv.Interface().(Configurable).UnmarshalConfigString(str); err != nil {
		return nil, err
	}
	return rv.Elem().Interface(), nil
}

func configurableInterfaceSliceHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Slice || f.Elem().Kind() != reflect.String || t.Kind() != reflect.Slice || !isConfigurableType(t.Elem()) {
		return data, nil
	}

	strs := data.([]string)
	res := reflect.MakeSlice(t, len(strs), len(strs))

	et := t.Elem()
	if et.Kind() == reflect.Ptr {
		for i, str := range strs {
			rv := reflect.New(et.Elem())
			if err := rv.Interface().(Configurable).UnmarshalConfigString(str); err != nil {
				return nil, err
			}
			res.Index(i).Set(rv)
		}
	} else {
		for i, str := range strs {
			rv := reflect.New(et)
			if err := rv.Interface().(Configurable).UnmarshalConfigString(str); err != nil {
				return nil, err
			}
			res.Index(i).Set(rv.Elem())
		}
	}
	return res.Interface(), nil
}

func stringToByteSliceHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t.Kind() != reflect.Slice || t.Elem().Kind() != reflect.Uint8 {
		return data, nil
	}

	str := strings.TrimPrefix(data.(string), "0x")
	slice, err := hex.DecodeString(str)
	if err != nil {
		return nil, fmt.Errorf("Could not decode hex: %s", err)
	}

	return slice, nil
}

func stringToByteArrayHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t.Kind() != reflect.Array || t.Elem().Kind() != reflect.Uint8 {
		return data, nil
	}

	str := strings.TrimPrefix(data.(string), "0x")
	slice, err := hex.DecodeString(str)
	if err != nil {
		return nil, fmt.Errorf("Could not decode hex: %s", err)
	}
	if len(slice) != t.Len() {
		return nil, fmt.Errorf("Invalid length: expected %d, got %d", t.Len(), len(slice))
	}

	rv := reflect.New(t).Elem()
	for i, v := range slice {
		rv.Index(i).SetUint(uint64(v))
	}
	return rv.Interface(), nil
}

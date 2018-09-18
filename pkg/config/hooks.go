// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
)

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
		if p := strings.SplitN(s, "=", 2); len(p) == 2 {
			m[p[0]] = p[1]
		}
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
		if p := strings.SplitN(s, "=", 2); len(p) == 2 {
			v := m[p[0]]
			if v == nil {
				v = make([]string, 0, 1)
			}

			m[p[0]] = append(v, p[1])
		}
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
		if p := strings.SplitN(s, "=", 2); len(p) == 2 {
			m[p[0]] = p[1]
		}
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
		if p := strings.SplitN(s, "=", 2); len(p) == 2 {
			str := strings.TrimPrefix(p[1], "0x")
			buf, err := hex.DecodeString(str)
			if err != nil {
				return nil, fmt.Errorf("Could not decode hex: %s", err)
			}
			m[p[0]] = buf
		}
	}

	return m, nil
}

func configurableInterfaceHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || !t.Implements(configurableI) {
		return data, nil
	}

	str := data.(string)

	u, ok := reflect.New(t).Interface().(Configurable)
	if !ok {
		return data, nil
	}

	return u.FromConfigString(str)
}

func configurableInterfaceSliceHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Slice || f.Elem().Kind() != reflect.String || t.Kind() != reflect.Slice || !t.Elem().Implements(configurableI) {
		return data, nil
	}

	strs := data.([]string)
	res := reflect.MakeSlice(t, len(strs), len(strs))

	for i, str := range strs {
		val, ok := reflect.New(t.Elem()).Interface().(Configurable)
		if !ok {
			return data, nil
		}

		v, err := val.FromConfigString(str)
		if err != nil {
			return nil, err
		}

		res.Index(i).Set(reflect.ValueOf(v))
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

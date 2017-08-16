// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package config

import (
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

// stringSliceToStringHookFunc is a hook function for mapstructure that converts []string to string by picking the first element.
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
	slice := strings.Split(str, " ")

	m := make(map[string]string, len(slice))
	for _, s := range slice {
		if p := strings.SplitN(s, "=", 2); len(p) == 2 {
			m[p[0]] = p[1]
		}
	}

	return m, nil
}

var iConfigurable = reflect.TypeOf((*Configurable)(nil)).Elem()

func configurableInterfaceHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || !t.Implements(iConfigurable) {
		return data, nil
	}

	str := data.(string)

	u, ok := reflect.New(t).Interface().(Configurable)
	if !ok {
		return data, nil
	}

	return u.FromConfigString(str)
}

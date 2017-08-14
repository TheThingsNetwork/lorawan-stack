package config

import (
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

// stringToTimeHookFunc is a hook for mapstructure that decodes strings to time.Time
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

// stringSliceToStringMapHookFunc is a hook for mapstructure that decodes []string to map[string]string
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

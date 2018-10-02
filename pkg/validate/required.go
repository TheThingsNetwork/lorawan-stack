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

package validate

import (
	stdliberr "errors"
	"reflect"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// IsZeroer is an interface, which reports whether it represents a zero value.
type IsZeroer interface {
	IsZero() bool
}

var isZeroerType = reflect.TypeOf((*IsZeroer)(nil)).Elem()

// isZeroValue is like isZero, but acts on values of reflect.Value type.
func isZeroValue(v reflect.Value) bool {
	iv := reflect.Indirect(v)
	if !iv.IsValid() {
		return true
	}

	if v.Type().Implements(isZeroerType) {
		return v.Interface().(IsZeroer).IsZero()
	}
	if iv.Type().Implements(isZeroerType) {
		return iv.Interface().(IsZeroer).IsZero()
	}

	v = iv

	switch v.Kind() {
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZero(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	case reflect.UnsafePointer:
		return v.Pointer() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

// isZero reports whether the value is the zero of its type.
func isZero(v interface{}) bool {
	if v == nil {
		return true
	}

	switch v := v.(type) {
	case nil:
		return true
	case bool:
		return !v
	case int:
		return v == 0
	case int64:
		return v == 0
	case int32:
		return v == 0
	case int16:
		return v == 0
	case int8:
		return v == 0
	case uint:
		return v == 0
	case uint64:
		return v == 0
	case uint32:
		return v == 0
	case uint16:
		return v == 0
	case uint8:
		return v == 0
	case float64:
		return v == 0
	case float32:
		return v == 0
	case string:
		return v == ""
	case []bool:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case []uint:
		return len(v) == 0
	case []uint64:
		return len(v) == 0
	case []uint32:
		return len(v) == 0
	case []uint16:
		return len(v) == 0
	case []uint8:
		return len(v) == 0
	case []int:
		return len(v) == 0
	case []int64:
		return len(v) == 0
	case []int32:
		return len(v) == 0
	case []int16:
		return len(v) == 0
	case []int8:
		return len(v) == 0
	case []float64:
		return len(v) == 0
	case []float32:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	case map[string]string:
		return len(v) == 0
	case *time.Time:
		return v == nil || v.IsZero()
	case time.Time:
		return v.IsZero()
	case *types.AES128Key:
		return v == nil || v.IsZero()
	case types.AES128Key:
		return v.IsZero()
	case *types.EUI64:
		return v == nil || v.IsZero()
	case types.EUI64:
		return v.IsZero()
	case *types.NetID:
		return v == nil || v.IsZero()
	case types.NetID:
		return v.IsZero()
	case *types.DevAddr:
		return v == nil || v.IsZero()
	case types.DevAddr:
		return v.IsZero()
	case *types.DevNonce:
		return v == nil || v.IsZero()
	case types.DevNonce:
		return v.IsZero()
	case *types.JoinNonce:
		return v == nil || v.IsZero()
	case types.JoinNonce:
		return v.IsZero()
	}
	return isZeroValue(reflect.ValueOf(v))
}

var errFieldSet = errors.DefineInvalidArgument("field_set", "field is set")

// Empty returns error if v is set.
// It is meant to be used as the first validator function passed as argument to Field.
// It uses IsZero, if v implements IsZeroer interface.
func Empty(v interface{}) error {
	if !isZero(v) {
		return errFieldSet
	}
	return nil
}

// errZeroValue is a flag value indicating a value is not required. If a value is not set,
// requirements on its value will be ignored.
var errZeroValue = stdliberr.New("Value not set")

// NotRequired returns an error, used internally in Field, if v is zero.
// It is meant to be used as the first validator function passed as argument to Field.
// It uses IsZero, if v implements IsZeroer interface.
func NotRequired(v interface{}) error {
	if isZero(v) {
		return errZeroValue
	}
	return nil
}

// Required returns error if v is empty.
// It is meant to be used as the first validator function passed as argument to Field.
// It uses IsZero, if v implements IsZeroer interface.
func Required(v interface{}) error {
	if isZero(v) {
		return errRequired
	}
	return nil
}

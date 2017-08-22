// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"fmt"
	"reflect"

	"github.com/gotnospirit/messageformat"
)

// Format formats the values into the provided string
func Format(format string, values Attributes) string {
	formatter, err := messageformat.New()
	if err != nil {
		return format
	}

	fm, err := formatter.Parse(format)
	if err != nil {
		return format
	}

	fixed := make(map[string]interface{}, len(values))
	for k, v := range values {
		fixed[k] = fix(v)
	}

	// todo format unsupported types
	res, err := fm.FormatMap(fixed)
	if err != nil {
		fmt.Println("err", err)
		return format
	}

	return res
}

// Fix coerces types that cannot be formatted by messageformat to string
func fix(v interface{}) interface{} {
	if v == nil {
		return "<nil>"
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Bool:
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Uintptr:
	case reflect.Float32:
	case reflect.Float64:
		return v
	case reflect.Ptr:
		// dereference and fix
		return fix(reflect.ValueOf(v).Elem())
	}
	return fmt.Sprintf("%v", v)
}

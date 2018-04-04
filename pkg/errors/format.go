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

package errors

import (
	"fmt"
	"reflect"

	"github.com/gotnospirit/messageformat"
)

// FormatErrorSignaler is used when an error occurs during the formatting of a MessageFormat.
var FormatErrorSignaler interface {
	Errorf(msg string, a ...interface{})
	WithError(err error, msg string)
}

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

	res, err := fm.FormatMap(fixed)
	if err != nil {
		FormatErrorSignaler.WithError(From(err), "Could not format the error descriptor")
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
	case reflect.Complex64:
	case reflect.Complex128:
	case reflect.String:
		return v
	case reflect.Ptr:
		// dereference and fix
		return fix(reflect.ValueOf(v).Elem())
	default:
		FormatErrorSignaler.Errorf("Non-primitive (e.g. %T) should not be used as attributes with ErrDescriptor objects", v)
	}
	return fmt.Sprintf("%v", v)
}

type emptyFmtErrorSignaler struct{}

func (e emptyFmtErrorSignaler) Errorf(_ string, _ ...interface{}) {}
func (e emptyFmtErrorSignaler) WithError(_ error, _ string)       {}

func init() {
	FormatErrorSignaler = emptyFmtErrorSignaler{}
}

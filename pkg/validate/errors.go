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

import "fmt"

// Errors is a slice of errors. It it is used to concatenate the different
// validation errors than can happen in a single field.
type Errors []error

// Error implements error.
func (e Errors) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		msg := e[0].Error()
		for i := 1; i < len(e); i++ {
			msg += "\n" + e[i].Error()
		}
		return msg
	}
}

// DescribeFieldName allows to prefix the errors with the name of the field.
func (e Errors) DescribeFieldName(fieldName string) error {
	if len(e) == 0 {
		return nil
	}
	for i := 0; i < len(e); i++ {
		e[i] = fmt.Errorf("%s: %s", fieldName, e[i].Error())
	}
	return e
}

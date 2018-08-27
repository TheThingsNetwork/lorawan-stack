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
	"fmt"
	"regexp"

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

var (
	passwordRegex = regexp.MustCompile("^.{8,}$")

	errPasswordLength = errors.DefineInvalidArgument("password_length", "password must be at least 8 characters long")
)

// Password checks whether the input value is a string and is at least 8 characters long.
func Password(v interface{}) error {
	password, ok := v.(string)
	if !ok {
		return errNotString.WithAttributes("type", fmt.Sprintf("%T", v))
	}

	if !passwordRegex.MatchString(password) {
		return errPasswordLength
	}

	return nil
}

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

package validate

import (
	"fmt"
	"regexp"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

var (
	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	errEmail = errors.DefineInvalidArgument("email", "`{email}` is not a valid email.")
)

// Email checks whether the input value is a valid email or not.
func Email(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return errNotString.WithAttributes("type", fmt.Sprintf("%T", v))
	}

	if !emailRegex.MatchString(str) {
		return errEmail.WithAttributes("email", str)
	}

	return nil
}

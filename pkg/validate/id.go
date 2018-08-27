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
	idRegex = regexp.MustCompile("^[a-z0-9](?:[_-]?[a-z0-9]){1,35}$")

	errID = errors.DefineInvalidArgument("id", "`{id}` must be at least 2 and at most 36 characters long and may consist of only letters, numbers, dashes and underscores. It may not start or end with a dash or an underscore.")
)

// ID checks whether the input value is a valid ID according:
//		- Length must be between 2 and 36
//		- It consists only of numbers, dashs, underscores and lowercase letters
//		- Must start by a number or lowercase letter
//		- It cannot match any of the blacklisted IDs
func ID(v interface{}) error {
	id, ok := v.(string)
	if !ok {
		return errNotString.WithAttributes("type", fmt.Sprintf("%T", v))
	}

	if !idRegex.MatchString(id) {
		return errID.WithAttributes("id", id)

	}

	return nil
}

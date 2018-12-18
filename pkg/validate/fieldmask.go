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
	"regexp"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var (
	fieldMaskPathRegex     = regexp.MustCompile("^[a-z0-9](?:[._]?[a-z0-9]){1,}$")
	fieldMaskPathMaxLength = 256

	errFieldMaskPath = errors.DefineInvalidArgument("fieldmaskpath", "`{fieldmaskpath}` may consist of only lowercase letters, numbers, underscores and dots. It may not start or end with a dot or underscore, or have two or more consecutive dots or underscores. Also, it must be between 2 and 256 characters in length.")
)

// FieldMaskPaths performs a basic sanity check on the allowed characters for valid fieldmask paths.
// Paths of a FieldMask may only contain lowercase letters, numbers, dots, underscores and must be between 2 and 256 characters in length.
func FieldMaskPaths(fm *types.FieldMask) error {
	for _, path := range fm.Paths {
		if len(path) > 256 || !fieldMaskPathRegex.MatchString(path) {
			return errFieldMaskPath.WithAttributes("fieldmaskpath", path)
		}
	}
	return nil
}

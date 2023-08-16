// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

// Package errors defines common error types for all upstreams.
package errors

import (
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// DeviceErrors contains errors during claiming/unclaiming a batch of devices.
type DeviceErrors struct {
	Errors map[types.EUI64]errors.ErrorDetails
}

// Error implements error.
func (e DeviceErrors) Error() string {
	var errs strings.Builder
	for devEUI, err := range e.Errors {
		_, err := errs.WriteString(fmt.Sprintf("%s: %s, ", devEUI, err.Error()))
		if err != nil {
			return err.Error()
		}
	}
	return fmt.Sprintf("Errors per Device EUI: %s", strings.TrimRight(errs.String(), ", "))
}

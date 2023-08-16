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

package errors

import (
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestErrors(t *testing.T) {
	a, _ := test.New(t)
	t.Parallel()

	eui1 := types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	eui2 := types.EUI64{0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	errs := DeviceErrors{
		Errors: map[types.EUI64]errors.ErrorDetails{
			eui1: errors.DefineNotFound("not_found", "not found"),
			eui2: errors.DefineAlreadyExists("already_exists", "already exists"),
		},
	}
	exp := `Errors per Device EUI: 42FFFFFFFFFFFFFF: error:pkg/deviceclaimingserver/enddevices/errors:not_found (not found), 43FFFFFFFFFFFFFF: error:pkg/deviceclaimingserver/enddevices/errors:already_exists (already exists)` // nolint:lll
	a.So(errs.Error(), should.Equal, exp)
}

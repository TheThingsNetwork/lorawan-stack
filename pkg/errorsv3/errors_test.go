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

package errors_test

import (
	"fmt"

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func Example() {
	var errApplicationNotFound = errors.DefineNotFound("application_not_found", "Application with ID `{id}` not found", "id")
	var errCouldNotCreateDevice = errors.Define("could_not_create_device", "Could not create Device")

	findApplication := func(id *ttnpb.ApplicationIdentifiers) (*ttnpb.Application, error) {
		// try really hard, but fail
		return nil, errApplicationNotFound.WithAttributes("id", id.ApplicationID)
	}

	createDevice := func(dev *ttnpb.EndDevice) error {
		app, err := findApplication(&dev.ApplicationIdentifiers)
		if err != nil {
			return err // you can just pass errors up
		}
		// create device
		_ = app
		return nil
	}

	if err := createDevice(&ttnpb.EndDevice{}); err != nil {
		fmt.Println(errCouldNotCreateDevice.WithCause(err))
	}

	// Output:
	// error:pkg/errorsv3_test:could_not_create_device (Could not create Device)
}

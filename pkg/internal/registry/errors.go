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

// Package registry contains commonly used device registry functionality.
package registry

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var errEndDeviceEUIsTaken = errors.DefineAlreadyExists(
	"end_device_euis_taken",
	"an end device with JoinEUI `{join_eui}` and DevEUI `{dev_eui}` is already registered as `{device_id}` in application `{application_id}`", // nolint:lll
)

// UniqueEUIViolationErr creates a unique EUI violation error with the given UID.
func UniqueEUIViolationErr(_ context.Context, joinEUI, devEUI types.EUI64, uid string) error {
	deviceIDs, err := unique.ToDeviceID(uid)
	if err != nil {
		return err
	}
	attributes := []any{
		"join_eui", joinEUI,
		"dev_eui", devEUI,
		"device_id", deviceIDs.GetDeviceId(),
		"application_id", deviceIDs.GetApplicationIds().GetApplicationId(),
	}
	return errEndDeviceEUIsTaken.WithAttributes(attributes...)
}

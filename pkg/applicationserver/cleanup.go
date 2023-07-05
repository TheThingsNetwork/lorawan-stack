// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package applicationserver

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/cleanup"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// RegistryCleaner is a service responsible for cleanup of the device registry.
type RegistryCleaner struct {
	DevRegistry DeviceRegistry
	LocalSet    map[string]struct{}
}

// RangeToLocalSet returns a set of devices that have data in the registry.
func (cleaner *RegistryCleaner) RangeToLocalSet(ctx context.Context) error {
	cleaner.LocalSet = make(map[string]struct{})
	err := cleaner.DevRegistry.Range(ctx, []string{"ids"}, func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, dev *ttnpb.EndDevice) bool {
		cleaner.LocalSet[unique.ID(ctx, ids)] = struct{}{}
		return true
	},
	)
	return err
}

// DeleteDeviceData deletes registry application data of all devices in the device id list.
func (cleaner *RegistryCleaner) DeleteDeviceData(ctx context.Context, devSet []string) error {
	for _, ids := range devSet {
		devIds, err := unique.ToDeviceID(ids)
		if err != nil {
			return err
		}
		ctx, err = unique.WithContext(ctx, ids)
		if err != nil {
			return err
		}
		_, err = cleaner.DevRegistry.Set(ctx, devIds, nil, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			return nil, nil, nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// CleanData cleans registry device data.
func (cleaner *RegistryCleaner) CleanData(ctx context.Context, isSet map[string]struct{}) error {
	complement := cleanup.ComputeSetComplement(isSet, cleaner.LocalSet)
	devIds := make([]string, len(complement))
	i := 0
	for id := range complement {
		devIds[i] = id
		i++
	}
	err := cleaner.DeleteDeviceData(ctx, devIds)
	if err != nil {
		return err
	}
	return nil
}

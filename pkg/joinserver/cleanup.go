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

package joinserver

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/cleanup"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// RegistryCleaner is a service responsible for cleanup of the device registry.
type RegistryCleaner struct {
	DevRegistry         DeviceRegistry
	AppAsRegistry       ApplicationActivationSettingRegistry
	LocalDeviceSet      map[string]struct{}
	LocalApplicationSet map[string]struct{}
}

// RangeToLocalSet returns a set of devices that have data in the registry.
func (cleaner *RegistryCleaner) RangeToLocalSet(ctx context.Context) error {
	cleaner.LocalDeviceSet = make(map[string]struct{})
	cleaner.LocalApplicationSet = make(map[string]struct{})
	err := cleaner.DevRegistry.RangeByID(ctx, []string{"ids"}, func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, dev *ttnpb.EndDevice) bool {
		cleaner.LocalDeviceSet[unique.ID(ctx, ids)] = struct{}{}
		return true
	})
	if err != nil {
		return err
	}
	return cleaner.AppAsRegistry.Range(ctx, []string{"application_server_id"}, func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, _ *ttnpb.ApplicationActivationSettings) bool {
		cleaner.LocalApplicationSet[unique.ID(ctx, ids)] = struct{}{}
		return true
	})
}

// DeleteApplicationAndDeviceData deletes registry application data of all devices in the device id list.
func (cleaner *RegistryCleaner) DeleteApplicationAndDeviceData(ctx context.Context, deviceList, applicationList []string) error {
	for _, ids := range deviceList {
		devIds, err := unique.ToDeviceID(ids)
		if err != nil {
			return err
		}
		ctx, err = unique.WithContext(ctx, ids)
		if err != nil {
			return err
		}
		err = DeleteDevice(ctx, cleaner.DevRegistry, devIds.ApplicationIds, devIds.DeviceId)
		if err != nil {
			return err
		}
	}
	for _, ids := range applicationList {
		appIds, err := unique.ToApplicationID(ids)
		if err != nil {
			return err
		}
		ctx, err = unique.WithContext(ctx, ids)
		if err != nil {
			return err
		}
		err = DeleteApplicationActivationSettings(ctx, cleaner.AppAsRegistry, appIds)
		if err != nil {
			return err
		}
	}
	return nil
}

// CleanData cleans registry device and application data.
func (cleaner *RegistryCleaner) CleanData(ctx context.Context, isDeviceSet, isApplicationSet map[string]struct{}) error {
	devComplement := cleanup.ComputeSetComplement(isDeviceSet, cleaner.LocalDeviceSet)
	devIds := make([]string, len(devComplement))
	i := 0
	for id := range devComplement {
		devIds[i] = id
		i++
	}
	appComplement := cleanup.ComputeSetComplement(isApplicationSet, cleaner.LocalApplicationSet)
	appIds := make([]string, len(appComplement))
	i = 0
	for id := range appComplement {
		appIds[i] = id
		i++
	}
	err := cleaner.DeleteApplicationAndDeviceData(ctx, devIds, appIds)
	if err != nil {
		return err
	}
	return nil
}

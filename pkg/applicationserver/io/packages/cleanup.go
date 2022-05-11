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

package packages

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/cleanup"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// RegistryCleaner is a service responsible for cleanup of the application packages registry.
type RegistryCleaner struct {
	ApplicationPackagesRegistry Registry
	LocalDeviceSet              map[string]struct{}
	LocalApplicationSet         map[string]struct{}
}

// RangeToLocalSet returns a set of applications and devices that have data in the registry.
func (cleaner *RegistryCleaner) RangeToLocalSet(ctx context.Context) error {
	cleaner.LocalDeviceSet = make(map[string]struct{})
	cleaner.LocalApplicationSet = make(map[string]struct{})
	err := cleaner.ApplicationPackagesRegistry.Range(ctx, []string{"ids"},
		func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, pb *ttnpb.ApplicationPackageAssociation) bool {
			cleaner.LocalDeviceSet[unique.ID(ctx, ids)] = struct{}{}
			return true
		},
		func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, pb *ttnpb.ApplicationPackageDefaultAssociation) bool {
			cleaner.LocalApplicationSet[unique.ID(ctx, ids)] = struct{}{}
			return true
		},
	)
	return err
}

// DeleteApplicationAndDeviceData deletes registry application and device data of all applications and devices in application and device lists.
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
		associations, err := cleaner.ApplicationPackagesRegistry.ListAssociations(ctx, devIds, []string{"ids"})
		if err != nil {
			return err
		}
		for _, association := range associations {
			_, err := cleaner.ApplicationPackagesRegistry.SetAssociation(ctx, association.GetIds(), nil,
				func(assoc *ttnpb.ApplicationPackageAssociation) (*ttnpb.ApplicationPackageAssociation, []string, error) {
					return nil, nil, nil
				},
			)
			if err != nil {
				return err
			}
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
		associations, err := cleaner.ApplicationPackagesRegistry.ListDefaultAssociations(ctx, appIds, []string{"ids"})
		if err != nil {
			return err
		}
		for _, association := range associations {
			_, err := cleaner.ApplicationPackagesRegistry.SetDefaultAssociation(ctx, association.GetIds(), nil,
				func(association *ttnpb.ApplicationPackageDefaultAssociation) (*ttnpb.ApplicationPackageDefaultAssociation, []string, error) {
					return nil, nil, nil
				},
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CleanData cleans registry application and device data.
func (cleaner *RegistryCleaner) CleanData(ctx context.Context, isDeviceSet map[string]struct{}, isApplicationSet map[string]struct{}) error {
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

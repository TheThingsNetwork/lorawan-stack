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

package pubsub

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/cleanup"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// RegistryCleaner is a service responsible for cleanup of the PubSub registry.
type RegistryCleaner struct {
	PubSubRegistry Registry
	LocalSet       map[string]struct{}
}

// RangeToLocalSet returns a set of applications that have data in the registry.
func (cleaner *RegistryCleaner) RangeToLocalSet(ctx context.Context) error {
	cleaner.LocalSet = make(map[string]struct{})
	err := cleaner.PubSubRegistry.Range(ctx, []string{"ids"},
		func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, pb *ttnpb.ApplicationPubSub) bool {
			cleaner.LocalSet[unique.ID(ctx, ids)] = struct{}{}
			return true
		},
	)
	return err
}

// DeleteApplicationData deletes registry application data of all applications in the application list.
func (cleaner *RegistryCleaner) DeleteApplicationData(ctx context.Context, applicationList []string) error {
	for _, ids := range applicationList {
		appIds, err := unique.ToApplicationID(ids)
		if err != nil {
			return err
		}
		ctx, err = unique.WithContext(ctx, ids)
		if err != nil {
			return err
		}
		pubsubs, err := cleaner.PubSubRegistry.List(ctx, appIds, []string{"ids"})
		if err != nil {
			return err
		}
		for _, pubsub := range pubsubs {
			_, err := cleaner.PubSubRegistry.Set(ctx, pubsub.GetIds(), nil,
				func(pubsub *ttnpb.ApplicationPubSub) (*ttnpb.ApplicationPubSub, []string, error) {
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

// CleanData cleans registry application data.
func (cleaner *RegistryCleaner) CleanData(ctx context.Context, isSet map[string]struct{}) error {
	complement := cleanup.ComputeSetComplement(isSet, cleaner.LocalSet)
	appIds := make([]string, len(complement))
	i := 0
	for id := range complement {
		appIds[i] = id
		i++
	}
	err := cleaner.DeleteApplicationData(ctx, appIds)
	if err != nil {
		return err
	}
	return nil
}

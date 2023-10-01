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

//go:build !tti
// +build !tti

package redis

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

func getBatchKey(uid string) (string, error) {
	devIDs, err := unique.ToDeviceID(uid)
	if err != nil {
		return "", errInvalidUID.WithCause(err)
	}

	return devIDs.ApplicationIds.GetApplicationId(), nil
}

func addToBatch(
	ctx context.Context, m map[string]*contextualUplinkBatch,
	confirmID string, uid string, up *ttnpb.ApplicationUp,
) error {
	key, err := getBatchKey(uid)
	if err != nil {
		return err
	}
	batch, ok := m[key]
	if !ok {
		batch = &contextualUplinkBatch{
			ctx:        ctx,
			confirmIDs: make([]string, 0),
			uplinks:    make([]*ttnpb.ApplicationUp, 0),
		}
		m[key] = batch
	}
	batch.uplinks = append(batch.uplinks, up)
	batch.confirmIDs = append(batch.confirmIDs, confirmID)
	return nil
}

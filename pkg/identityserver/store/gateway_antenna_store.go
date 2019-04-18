// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"context"
	"runtime/trace"

	"github.com/jinzhu/gorm"
)

func (s *store) replaceGatewayAntennas(ctx context.Context, gatewayUUID string, old []GatewayAntenna, new []GatewayAntenna) error {
	return replaceGatewayAntennas(ctx, s.DB, gatewayUUID, old, new)
}

func replaceGatewayAntennas(ctx context.Context, db *gorm.DB, gatewayUUID string, old []GatewayAntenna, new []GatewayAntenna) (err error) {
	defer trace.StartRegion(ctx, "update gateway antennas").End()
	db = db.Where(GatewayAntenna{GatewayID: gatewayUUID})
	if len(new) < len(old) {
		if err = db.Where("\"index\" >= ?", len(new)).Delete(&GatewayAntenna{}).Error; err != nil {
			return err
		}
	}
	for _, antenna := range new {
		antenna.GatewayID = gatewayUUID
		if err = db.Save(&antenna).Error; err != nil {
			return err
		}
	}
	return nil
}

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
	"fmt"
	"reflect"
	"runtime/trace"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetGatewayStore returns an GatewayStore on the given db (or transaction).
func GetGatewayStore(db *gorm.DB) GatewayStore {
	return &gatewayStore{db: db}
}

type gatewayStore struct {
	db *gorm.DB
}

// selectGatewayFields selects relevant fields (based on fieldMask) and preloads details if needed.
func selectGatewayFields(ctx context.Context, query *gorm.DB, fieldMask *types.FieldMask) *gorm.DB {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		return query.Preload("Attributes").Preload("Antennas")
	}
	var gatewayColumns []string
	var notFoundPaths []string
	for _, path := range ttnpb.TopLevelFields(fieldMask.Paths) {
		switch path {
		case "ids", "created_at", "updated_at":
			// always selected
		case attributesField:
			query = query.Preload("Attributes")
		case antennasField:
			query = query.Preload("Antennas")
		default:
			if columns, ok := gatewayColumnNames[path]; ok {
				gatewayColumns = append(gatewayColumns, columns...)
			} else {
				notFoundPaths = append(notFoundPaths, path)
			}
		}
	}
	if len(notFoundPaths) > 0 {
		warning.Add(ctx, fmt.Sprintf("unsupported field mask paths: %s", strings.Join(notFoundPaths, ", ")))
	}
	return query.Select(cleanFields(append(append(modelColumns, "gateway_id", "gateway_eui"), gatewayColumns...)...))
}

func (s *gatewayStore) CreateGateway(ctx context.Context, gtw *ttnpb.Gateway) (*ttnpb.Gateway, error) {
	defer trace.StartRegion(ctx, "create gateway").End()
	gtwModel := Gateway{
		GatewayID: gtw.GatewayID, // The ID is not mutated by fromPB.
	}
	gtwModel.fromPB(gtw, nil)
	gtwModel.SetContext(ctx)
	query := s.db.Create(&gtwModel)
	if query.Error != nil {
		return nil, query.Error
	}
	var gtwProto ttnpb.Gateway
	gtwModel.toPB(&gtwProto, nil)
	return &gtwProto, nil
}

func (s *gatewayStore) FindGateways(ctx context.Context, ids []*ttnpb.GatewayIdentifiers, fieldMask *types.FieldMask) ([]*ttnpb.Gateway, error) {
	defer trace.StartRegion(ctx, "find gateways").End()
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.GetGatewayID()
	}
	query := s.db.Scopes(withContext(ctx), withGatewayID(idStrings...))
	query = selectGatewayFields(ctx, query, fieldMask)
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(Gateway{}))
		query = query.Limit(limit).Offset(offset)
	}
	var gtwModels []Gateway
	query = query.Find(&gtwModels)
	setTotal(ctx, uint64(len(gtwModels)))
	if query.Error != nil {
		return nil, query.Error
	}
	gtwProtos := make([]*ttnpb.Gateway, len(gtwModels))
	for i, gtwModel := range gtwModels {
		gtwProto := &ttnpb.Gateway{}
		gtwModel.toPB(gtwProto, fieldMask)
		gtwProtos[i] = gtwProto
	}
	return gtwProtos, nil
}

func (s *gatewayStore) GetGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers, fieldMask *types.FieldMask) (*ttnpb.Gateway, error) {
	defer trace.StartRegion(ctx, "get gateway").End()
	query := s.db.Scopes(withContext(ctx), withGatewayID(id.GetGatewayID()))
	if id.EUI != nil {
		query = query.Scopes(withGatewayEUI(EUI64(*id.EUI)))
	}
	query = selectGatewayFields(ctx, query, fieldMask)
	var gtwModel Gateway
	if err := query.First(&gtwModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id.EntityIdentifiers())
		}
		return nil, err
	}
	gtwProto := &ttnpb.Gateway{}
	gtwModel.toPB(gtwProto, fieldMask)
	return gtwProto, nil
}

func (s *gatewayStore) UpdateGateway(ctx context.Context, gtw *ttnpb.Gateway, fieldMask *types.FieldMask) (updated *ttnpb.Gateway, err error) {
	defer trace.StartRegion(ctx, "update gateway").End()
	query := s.db.Scopes(withContext(ctx), withGatewayID(gtw.GetGatewayID()))
	query = selectGatewayFields(ctx, query, fieldMask)
	var gtwModel Gateway
	if err = query.First(&gtwModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(gtw.GatewayIdentifiers.EntityIdentifiers())
		}
		return nil, err
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	oldAttributes, oldAntennas := gtwModel.Attributes, gtwModel.Antennas
	columns := gtwModel.fromPB(gtw, fieldMask)
	if len(columns) > 0 {
		query = s.db.Select(append(columns, "updated_at")).Save(&gtwModel)
		if query.Error != nil {
			return nil, query.Error
		}
	}
	if !reflect.DeepEqual(oldAttributes, gtwModel.Attributes) {
		if err = replaceAttributes(ctx, s.db, "gateway", gtwModel.ID, oldAttributes, gtwModel.Attributes); err != nil {
			return nil, err
		}
	}
	if !reflect.DeepEqual(oldAntennas, gtwModel.Antennas) {
		if err = replaceGatewayAntennas(ctx, s.db, gtwModel.ID, oldAntennas, gtwModel.Antennas); err != nil {
			return nil, err
		}
	}
	updated = &ttnpb.Gateway{}
	gtwModel.toPB(updated, fieldMask)
	return updated, nil
}

func (s *gatewayStore) DeleteGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error {
	defer trace.StartRegion(ctx, "delete gateway").End()
	return deleteEntity(ctx, s.db, id.EntityIdentifiers())
}

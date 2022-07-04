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

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetGatewayStore returns an GatewayStore on the given db (or transaction).
func GetGatewayStore(db *gorm.DB) store.GatewayStore {
	return &gatewayStore{baseStore: newStore(db)}
}

type gatewayStore struct {
	*baseStore
}

// selectGatewayFields selects relevant fields (based on fieldMask) and preloads details if needed.
func selectGatewayFields(ctx context.Context, query *gorm.DB, fieldMask store.FieldMask) *gorm.DB {
	if len(fieldMask) == 0 {
		return query.Preload("Attributes").Preload("Antennas")
	}
	var gatewayColumns []string
	var notFoundPaths []string
	for _, path := range ttnpb.TopLevelFields(fieldMask) {
		switch path {
		case ids, createdAt, updatedAt, deletedAt:
			// always selected
		case attributesField:
			query = query.Preload("Attributes")
		case administrativeContactField:
			gatewayColumns = append(gatewayColumns, gatewayColumnNames[path]...)
			query = query.Preload("AdministrativeContact")
		case technicalContactField:
			gatewayColumns = append(gatewayColumns, gatewayColumnNames[path]...)
			query = query.Preload("TechnicalContact")
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
	return query.Select(cleanFields(
		mergeFields(modelColumns, gatewayColumns, []string{
			deletedAt, "gateway_id", "gateway_eui",
		})...,
	))
}

func (s *gatewayStore) CreateGateway(ctx context.Context, gtw *ttnpb.Gateway) (*ttnpb.Gateway, error) {
	defer trace.StartRegion(ctx, "create gateway").End()
	gtwModel := Gateway{
		GatewayID: gtw.GetIds().GetGatewayId(), // The ID is not mutated by fromPB.
	}
	gtwModel.fromPB(gtw, nil)
	var err error
	gtwModel.AdministrativeContactID, err = s.loadContact(ctx, gtwModel.AdministrativeContact)
	if err != nil {
		return nil, err
	}
	gtwModel.TechnicalContactID, err = s.loadContact(ctx, gtwModel.TechnicalContact)
	if err != nil {
		return nil, err
	}
	if err := s.createEntity(ctx, &gtwModel); err != nil {
		return nil, err
	}
	var gtwProto ttnpb.Gateway
	gtwModel.toPB(&gtwProto, nil)
	return &gtwProto, nil
}

func (s *gatewayStore) CountGateways(ctx context.Context) (uint64, error) {
	defer trace.StartRegion(ctx, "count gateways").End()

	var total uint64
	if err := s.query(ctx, Gateway{}, withGatewayID()).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (s *gatewayStore) FindGateways(
	ctx context.Context, ids []*ttnpb.GatewayIdentifiers, fieldMask store.FieldMask,
) ([]*ttnpb.Gateway, error) {
	defer trace.StartRegion(ctx, "find gateways").End()
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.GetGatewayId()
	}
	query := s.query(ctx, Gateway{}, withGatewayID(idStrings...))
	query = selectGatewayFields(ctx, query, fieldMask)
	query = query.Order(store.OrderFromContext(ctx, "gateways", "gateway_id", "ASC"))
	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 {
		var total uint64
		query.Count(&total)
		store.SetTotal(ctx, total)
		query = query.Limit(limit).Offset(offset)
	}
	var gtwModels []Gateway
	query = query.Find(&gtwModels)
	store.SetTotal(ctx, uint64(len(gtwModels)))
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

func (s *gatewayStore) GetGateway(
	ctx context.Context, id *ttnpb.GatewayIdentifiers, fieldMask store.FieldMask,
) (*ttnpb.Gateway, error) {
	defer trace.StartRegion(ctx, "get gateway").End()
	query := s.query(ctx, Gateway{}, withGatewayID(id.GetGatewayId()))
	if id.Eui != nil {
		query = query.Scopes(withGatewayEUI(EUI64(*id.Eui)))
	}
	query = selectGatewayFields(ctx, query, fieldMask)
	var gtwModel Gateway
	if err := query.First(&gtwModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id)
		}
		return nil, err
	}
	gtwProto := &ttnpb.Gateway{}
	gtwModel.toPB(gtwProto, fieldMask)
	return gtwProto, nil
}

func (s *gatewayStore) UpdateGateway(
	ctx context.Context, gtw *ttnpb.Gateway, fieldMask store.FieldMask,
) (updated *ttnpb.Gateway, err error) {
	defer trace.StartRegion(ctx, "update gateway").End()
	query := s.query(ctx, Gateway{}, withGatewayID(gtw.GetIds().GetGatewayId()))
	query = selectGatewayFields(ctx, query, fieldMask)
	var gtwModel Gateway
	if err = query.First(&gtwModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(gtw.GetIds())
		}
		return nil, err
	}
	if err = ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	oldAttributes, oldAntennas := gtwModel.Attributes, gtwModel.Antennas
	columns := gtwModel.fromPB(gtw, fieldMask)
	gtwModel.AdministrativeContactID, err = s.loadContact(ctx, gtwModel.AdministrativeContact)
	if err != nil {
		return nil, err
	}
	gtwModel.TechnicalContactID, err = s.loadContact(ctx, gtwModel.TechnicalContact)
	if err != nil {
		return nil, err
	}
	if err = s.updateEntity(ctx, &gtwModel, columns...); err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(oldAttributes, gtwModel.Attributes) {
		if err = s.replaceAttributes(ctx, gateway, gtwModel.ID, oldAttributes, gtwModel.Attributes); err != nil {
			return nil, err
		}
	}
	if !reflect.DeepEqual(oldAntennas, gtwModel.Antennas) {
		if err = s.replaceGatewayAntennas(ctx, gtwModel.ID, oldAntennas, gtwModel.Antennas); err != nil {
			return nil, err
		}
	}
	updated = &ttnpb.Gateway{}
	gtwModel.toPB(updated, fieldMask)
	return updated, nil
}

func (s *gatewayStore) DeleteGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error {
	defer trace.StartRegion(ctx, "delete gateway").End()
	return s.deleteEntity(ctx, id)
}

func (s *gatewayStore) RestoreGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error {
	defer trace.StartRegion(ctx, "restore gateway").End()
	return s.restoreEntity(ctx, id)
}

func (s *gatewayStore) PurgeGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error {
	defer trace.StartRegion(ctx, "purge gateway").End()
	query := s.query(ctx, Gateway{}, withSoftDeleted(), withGatewayID(id.GetGatewayId()))
	query = selectGatewayFields(ctx, query, nil)
	var gtwModel Gateway
	if err := query.First(&gtwModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errNotFoundForID(id)
		}
		return err
	}
	// delete gateway antennas before purging
	if len(gtwModel.Antennas) > 0 {
		if err := s.replaceGatewayAntennas(ctx, gtwModel.ID, gtwModel.Antennas, nil); err != nil {
			return err
		}
	}
	// delete gateway attributes before purging
	if len(gtwModel.Attributes) > 0 {
		if err := s.replaceAttributes(ctx, gateway, gtwModel.ID, gtwModel.Attributes, nil); err != nil {
			return err
		}
	}
	return s.purgeEntity(ctx, id)
}

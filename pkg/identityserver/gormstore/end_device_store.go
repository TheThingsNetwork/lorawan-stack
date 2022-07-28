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
	"sort"
	"strings"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetEndDeviceStore returns an EndDeviceStore on the given db (or transaction).
func GetEndDeviceStore(db *gorm.DB) store.EndDeviceStore {
	return &deviceStore{baseStore: newStore(db)}
}

type deviceStore struct {
	*baseStore
}

// selectEndDeviceFields selects relevant fields (based on fieldMask) and preloads details if needed.
func selectEndDeviceFields(ctx context.Context, query *gorm.DB, fieldMask store.FieldMask) *gorm.DB {
	if len(fieldMask) == 0 {
		return query.Preload("Attributes").Preload("Locations")
	}
	var deviceColumns []string
	var notFoundPaths []string
	for _, path := range ttnpb.TopLevelFields(fieldMask) {
		switch path {
		case ids, createdAt, updatedAt:
			// always selected
		case attributesField:
			query = query.Preload("Attributes")
		case locationsField:
			query = query.Preload("Locations")
		case pictureField:
			deviceColumns = append(deviceColumns, "picture_id")
			query = query.Preload("Picture")
		default:
			if columns, ok := deviceColumnNames[path]; ok {
				deviceColumns = append(deviceColumns, columns...)
			} else {
				notFoundPaths = append(notFoundPaths, path)
			}
		}
	}
	if len(notFoundPaths) > 0 {
		warning.Add(ctx, fmt.Sprintf("unsupported field mask paths: %s", strings.Join(notFoundPaths, ", ")))
	}
	return query.Select(cleanFields(
		mergeFields(modelColumns, deviceColumns, []string{
			"application_id", "device_id", "join_eui", "dev_eui",
		})...,
	))
}

func (s *deviceStore) CreateEndDevice(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, error) {
	defer trace.StartRegion(ctx, "create end device").End()
	devModel := EndDevice{
		ApplicationID: dev.GetIds().GetApplicationIds().GetApplicationId(), // The ApplicationID is not mutated by fromPB.
		DeviceID:      dev.GetIds().GetDeviceId(),                          // The DeviceID is not mutated by fromPB.
	}
	devModel.fromPB(dev, nil)
	if err := s.createEntity(ctx, &devModel); err != nil {
		return nil, err
	}
	var devProto ttnpb.EndDevice
	devModel.toPB(&devProto, nil)
	return &devProto, nil
}

func (*deviceStore) findEndDevices(
	ctx context.Context, query *gorm.DB, fieldMask store.FieldMask,
) ([]*ttnpb.EndDevice, error) {
	defer trace.StartRegion(ctx, "find end devices").End()
	query = selectEndDeviceFields(ctx, query, fieldMask)
	query = query.Order(store.OrderFromContext(ctx, "end_devices", "device_id", "ASC"))
	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 {
		var total uint64
		query.Count(&total)
		store.SetTotal(ctx, total)
		query = query.Limit(limit).Offset(offset)
	}
	var devModels []EndDevice
	query = query.Find(&devModels)
	store.SetTotal(ctx, uint64(len(devModels)))
	if query.Error != nil {
		return nil, query.Error
	}
	devProtos := make([]*ttnpb.EndDevice, len(devModels))
	for i, devModel := range devModels {
		devProto := &ttnpb.EndDevice{}
		devModel.toPB(devProto, fieldMask)
		devProtos[i] = devProto
	}
	return devProtos, nil
}

func (s *deviceStore) CountEndDevices(
	ctx context.Context, ids *ttnpb.ApplicationIdentifiers,
) (total uint64, err error) {
	defer trace.StartRegion(ctx, "count end devices").End()
	err = s.query(ctx, EndDevice{}, withApplicationID(ids.GetApplicationId())).Count(&total).Error
	return total, err
}

func (s *deviceStore) ListEndDevices(
	ctx context.Context, ids *ttnpb.ApplicationIdentifiers, fieldMask store.FieldMask,
) ([]*ttnpb.EndDevice, error) {
	// NOTE: tracing done in s.findEndDevices.
	query := s.query(ctx, EndDevice{})
	if ids != nil {
		query = query.Scopes(withApplicationID(ids.GetApplicationId()))
	}
	return s.findEndDevices(ctx, query, fieldMask)
}

var errMultipleApplicationIDs = errors.DefineInvalidArgument(
	"multiple_application_ids",
	"can not list devices for multiple application IDs",
)

func (s *deviceStore) FindEndDevices(
	ctx context.Context, ids []*ttnpb.EndDeviceIdentifiers, fieldMask store.FieldMask,
) ([]*ttnpb.EndDevice, error) {
	// NOTE: tracing done in s.findEndDevices.
	idStrings := make([]string, len(ids))
	var applicationID string
	for i, id := range ids {
		if applicationID != "" && applicationID != id.GetApplicationIds().GetApplicationId() {
			return nil, errMultipleApplicationIDs.New()
		}
		applicationID = id.GetApplicationIds().GetApplicationId()
		idStrings[i] = id.GetDeviceId()
	}
	query := s.query(ctx, EndDevice{}, withApplicationID(applicationID), withDeviceID(idStrings...))
	return s.findEndDevices(ctx, query, fieldMask)
}

func (s *deviceStore) GetEndDevice(
	ctx context.Context, id *ttnpb.EndDeviceIdentifiers, fieldMask store.FieldMask,
) (*ttnpb.EndDevice, error) {
	defer trace.StartRegion(ctx, "get end device").End()
	query := s.query(ctx,
		EndDevice{},
		withApplicationID(id.GetApplicationIds().GetApplicationId()),
		withDeviceID(id.GetDeviceId()),
	)
	if id.JoinEui != nil {
		query = query.Scopes(withJoinEUI(EUI64(*id.JoinEui)))
	}
	if id.DevEui != nil {
		query = query.Scopes(withDevEUI(EUI64(*id.DevEui)))
	}
	query = selectEndDeviceFields(ctx, query, fieldMask)
	var devModel EndDevice
	if err := query.First(&devModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id)
		}
		return nil, err
	}
	devProto := &ttnpb.EndDevice{}
	devModel.toPB(devProto, fieldMask)
	return devProto, nil
}

func (s *deviceStore) UpdateEndDevice(
	ctx context.Context, dev *ttnpb.EndDevice, fieldMask store.FieldMask,
) (updated *ttnpb.EndDevice, err error) {
	defer trace.StartRegion(ctx, "update end device").End()
	query := s.query(ctx,
		EndDevice{},
		withApplicationID(dev.GetIds().GetApplicationIds().GetApplicationId()),
		withDeviceID(dev.GetIds().GetDeviceId()),
	)
	query = selectEndDeviceFields(ctx, query, fieldMask)
	var devModel EndDevice
	if err = query.First(&devModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(dev.Ids)
		}
		return nil, err
	}
	if err = ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	oldAttributes, oldLocations, oldPicture := devModel.Attributes, devModel.Locations, devModel.Picture
	columns := devModel.fromPB(dev, fieldMask)
	newPicture := devModel.Picture
	if newPicture != oldPicture {
		if oldPicture != nil {
			if err = s.query(ctx, Picture{}).Delete(oldPicture).Error; err != nil {
				return nil, err
			}
		}
		if newPicture != nil {
			if err = s.createEntity(ctx, &newPicture); err != nil {
				return nil, err
			}
			devModel.PictureID, devModel.Picture = &newPicture.ID, nil
			columns = append(columns, "picture_id")
		}
	}
	if err = s.updateEntity(ctx, &devModel, columns...); err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(oldAttributes, devModel.Attributes) {
		if err = s.replaceAttributes(ctx, "device", devModel.ID, oldAttributes, devModel.Attributes); err != nil {
			return nil, err
		}
	}
	if !reflect.DeepEqual(oldLocations, devModel.Locations) {
		if err = s.replaceEndDeviceLocations(ctx, devModel.ID, oldLocations, devModel.Locations); err != nil {
			return nil, err
		}
	}
	devModel.Picture = newPicture
	updated = &ttnpb.EndDevice{}
	devModel.toPB(updated, fieldMask)
	return updated, nil
}

func (s *deviceStore) DeleteEndDevice(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) error {
	defer trace.StartRegion(ctx, "delete end device").End()
	return s.deleteEntity(ctx, id)
}

func batchUpdateLastSeenAtQuery(placeholders []string) string {
	return fmt.Sprintf(`
	WITH updates (application_id, device_id, last_seen_at) AS (VALUES %s)
	UPDATE end_devices AS end_device
	SET last_seen_at = updates.last_seen_at
	FROM updates
	WHERE end_device.application_id = updates.application_id
	AND end_device.device_id = updates.device_id
	AND (end_device.last_seen_at IS NULL OR end_device.last_seen_at < updates.last_seen_at)
	`, strings.Join(placeholders, ","))
}

func (s *deviceStore) BatchUpdateEndDeviceLastSeen(
	ctx context.Context, devsLastSeen []*ttnpb.BatchUpdateEndDeviceLastSeenRequest_EndDeviceLastSeenUpdate,
) error {
	defer trace.StartRegion(ctx, "batch update end devices").End()

	sort.Slice(devsLastSeen, func(i, j int) bool {
		return devsLastSeen[i].Ids.IDString() < devsLastSeen[j].Ids.IDString()
	})

	valueStrings := []string{}
	valueArgs := []interface{}{}

	for _, dev := range devsLastSeen {
		valueStrings = append(valueStrings, "(?::varchar, ?::varchar, ?::timestamptz)")
		valueArgs = append(valueArgs,
			dev.Ids.ApplicationIds.ApplicationId,
			dev.Ids.DeviceId,
			ttnpb.StdTime(dev.LastSeenAt),
		)
	}

	return s.query(ctx, EndDevice{}).Exec(batchUpdateLastSeenAtQuery(valueStrings), valueArgs...).Error
}

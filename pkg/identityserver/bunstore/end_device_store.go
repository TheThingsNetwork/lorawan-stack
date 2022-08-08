// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"sort"
	"time"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// EndDevice is the end_device model in the database.
type EndDevice struct {
	bun.BaseModel `bun:"table:end_devices,alias:dev"`

	Model

	ApplicationID string `bun:"application_id,notnull"`
	DeviceID      string `bun:"device_id,notnull"`

	Name        string `bun:"name,nullzero"`
	Description string `bun:"description,nullzero"`

	Attributes []*Attribute `bun:"rel:has-many,join:type=entity_type,join:id=entity_id,polymorphic:device"`

	JoinEUI *string `bun:"join_eui"`
	DevEUI  *string `bun:"dev_eui"`

	BrandID         string `bun:"brand_id,nullzero"`
	ModelID         string `bun:"model_id,nullzero"`
	HardwareVersion string `bun:"hardware_version,nullzero"`
	FirmwareVersion string `bun:"firmware_version,nullzero"`
	BandID          string `bun:"band_id,nullzero"`

	NetworkServerAddress     string `bun:"network_server_address,nullzero"`
	ApplicationServerAddress string `bun:"application_server_address,nullzero"`
	JoinServerAddress        string `bun:"join_server_address,nullzero"`

	ServiceProfileID string `bun:"service_profile_id,nullzero"`

	Locations []*EndDeviceLocation `bun:"rel:has-many,join:id=end_device_id"`

	PictureID *string  `bun:"picture_id"`
	Picture   *Picture `bun:"rel:belongs-to,join:picture_id=id"`

	ActivatedAt *time.Time `bun:"activated_at"`
	LastSeenAt  *time.Time `bun:"last_seen_at"`

	ClaimAuthenticationCodeSecret    []byte     `bun:"claim_authentication_code_secret"`
	ClaimAuthenticationCodeValidFrom *time.Time `bun:"claim_authentication_code_valid_from"`
	ClaimAuthenticationCodeValidTo   *time.Time `bun:"claim_authentication_code_valid_to"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *EndDevice) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func endDeviceToPB(m *EndDevice, fieldMask ...string) (*ttnpb.EndDevice, error) {
	var devEUI, joinEUI []byte
	if euiFromString := eui64FromString(m.DevEUI); euiFromString != nil {
		devEUI = euiFromString.Bytes()
	}
	if euiFromString := eui64FromString(m.JoinEUI); euiFromString != nil {
		joinEUI = euiFromString.Bytes()
	}

	pb := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: m.ApplicationID,
			},
			DeviceId: m.DeviceID,
			DevEui:   devEUI,
			JoinEui:  joinEUI,
		},

		CreatedAt: ttnpb.ProtoTimePtr(m.CreatedAt),
		UpdatedAt: ttnpb.ProtoTimePtr(m.UpdatedAt),

		Name:        m.Name,
		Description: m.Description,

		VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
			BrandId:         m.BrandID,
			ModelId:         m.ModelID,
			HardwareVersion: m.HardwareVersion,
			FirmwareVersion: m.FirmwareVersion,
			BandId:          m.BandID,
		},

		NetworkServerAddress:     m.NetworkServerAddress,
		ApplicationServerAddress: m.ApplicationServerAddress,
		JoinServerAddress:        m.JoinServerAddress,

		ClaimAuthenticationCode: func() *ttnpb.EndDeviceAuthenticationCode {
			return &ttnpb.EndDeviceAuthenticationCode{
				Value:     string(m.ClaimAuthenticationCodeSecret),
				ValidFrom: ttnpb.ProtoTime(m.ClaimAuthenticationCodeValidFrom),
				ValidTo:   ttnpb.ProtoTime(m.ClaimAuthenticationCodeValidTo),
			}
		}(),

		ServiceProfileId: m.ServiceProfileID,

		ActivatedAt: ttnpb.ProtoTime(m.ActivatedAt),
		LastSeenAt:  ttnpb.ProtoTime(m.LastSeenAt),
	}

	if len(m.Attributes) > 0 {
		pb.Attributes = make(map[string]string, len(m.Attributes))
		for _, a := range m.Attributes {
			pb.Attributes[a.Key] = a.Value
		}
	}

	if len(m.Locations) > 0 {
		pb.Locations = make(map[string]*ttnpb.Location, len(m.Locations))
		for _, location := range m.Locations {
			locationPB := locationToPB(location.Location)
			if locationPB == nil {
				// Unfortunately it's relatively common to have nil or zero locations in the database.
				// If that's the case, we still need to set the source on a zero location.
				locationPB = &ttnpb.Location{}
			}
			locationPB.Source = ttnpb.LocationSource(location.Source)
			pb.Locations[location.Service] = locationPB
		}
	}

	if m.Picture != nil {
		picture, err := pictureToPB(m.Picture)
		if err != nil {
			return nil, err
		}
		pb.Picture = picture
	}

	if len(fieldMask) == 0 {
		return pb, nil
	}

	res := &ttnpb.EndDevice{}
	if err := res.SetFields(pb, fieldMask...); err != nil {
		return nil, err
	}

	// Set fields that are always present.
	res.Ids = pb.Ids
	res.CreatedAt = pb.CreatedAt
	res.UpdatedAt = pb.UpdatedAt

	return res, nil
}

type endDeviceStore struct {
	*baseStore
}

func newEndDeviceStore(baseStore *baseStore) *endDeviceStore {
	return &endDeviceStore{
		baseStore: baseStore,
	}
}

func (s *endDeviceStore) CreateEndDevice(
	ctx context.Context, pb *ttnpb.EndDevice,
) (*ttnpb.EndDevice, error) {
	ctx, span := tracer.Start(ctx, "CreateEndDevice", trace.WithAttributes(
		attribute.String("application_id", pb.GetIds().GetApplicationIds().GetApplicationId()),
		attribute.String("device_id", pb.GetIds().GetDeviceId()),
	))
	defer span.End()

	model := &EndDevice{
		ApplicationID: pb.GetIds().GetApplicationIds().GetApplicationId(),
		DeviceID:      pb.GetIds().GetDeviceId(),
		DevEUI:        eui64ToString(types.MustEUI64(pb.GetIds().GetDevEui())),
		JoinEUI:       eui64ToString(types.MustEUI64(pb.GetIds().GetJoinEui())),
		Name:          pb.Name,
		Description:   pb.Description,

		BrandID:         pb.VersionIds.GetBrandId(),
		ModelID:         pb.VersionIds.GetModelId(),
		HardwareVersion: pb.VersionIds.GetHardwareVersion(),
		FirmwareVersion: pb.VersionIds.GetFirmwareVersion(),
		BandID:          pb.VersionIds.GetBandId(),

		NetworkServerAddress:     pb.NetworkServerAddress,
		ApplicationServerAddress: pb.ApplicationServerAddress,
		JoinServerAddress:        pb.JoinServerAddress,

		ClaimAuthenticationCodeSecret:    []byte(pb.ClaimAuthenticationCode.GetValue()),
		ClaimAuthenticationCodeValidFrom: cleanTimePtr(ttnpb.StdTime(pb.ClaimAuthenticationCode.GetValidFrom())),
		ClaimAuthenticationCodeValidTo:   cleanTimePtr(ttnpb.StdTime(pb.ClaimAuthenticationCode.GetValidTo())),

		ServiceProfileID: pb.ServiceProfileId,

		ActivatedAt: cleanTimePtr(ttnpb.StdTime(pb.ActivatedAt)),
		LastSeenAt:  cleanTimePtr(ttnpb.StdTime(pb.LastSeenAt)),
	}

	if pb.Picture != nil {
		picture, err := pictureFromPB(ctx, pb.Picture)
		if err != nil {
			return nil, err
		}
		model.Picture = picture

		_, err = s.DB.NewInsert().
			Model(model.Picture).
			Exec(ctx)
		if err != nil {
			return nil, wrapDriverError(err)
		}

		model.PictureID = &model.Picture.ID
	}

	_, err := s.DB.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	if len(pb.Attributes) > 0 {
		model.Attributes, err = s.replaceAttributes(
			ctx, nil, pb.Attributes, "device", model.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	if len(pb.Locations) > 0 {
		model.Locations, err = s.replaceEndDeviceLocations(
			ctx, nil, pb.Locations, model.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	pb, err = endDeviceToPB(model)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (*endDeviceStore) selectWithFields(q *bun.SelectQuery, fieldMask store.FieldMask) (*bun.SelectQuery, error) {
	if fieldMask == nil {
		q = q.ExcludeColumn()
	} else {
		columns := []string{
			"id",
			"created_at",
			"updated_at",
			"application_id",
			"device_id",
			"join_eui",
			"dev_eui",
		}
		for _, f := range fieldMask.TopLevel() {
			switch f {
			default:
				return nil, fmt.Errorf("unknown field %q", f)
			case "ids", "created_at", "updated_at":
				// Always selected.
			case "name", "description",
				"network_server_address", "application_server_address", "join_server_address",
				"service_profile_id",
				"activated_at", "last_seen_at":
				// Proto name equals model name.
				columns = append(columns, f)
			case "version_ids":
				columns = append(columns, "brand_id", "model_id", "hardware_version", "firmware_version", "band_id")
			case "attributes":
				q = q.Relation("Attributes")
			case "contact_info":
				q = q.Relation("ContactInfo")
			case "locations":
				q = q.Relation("Locations")
			case "picture":
				q = q.Relation("Picture")
			case "claim_authentication_code":
				columns = append(
					columns,
					"claim_authentication_code_secret",
					"claim_authentication_code_valid_from",
					"claim_authentication_code_valid_to",
				)
			}
		}
		q = q.Column(columns...)
	}
	return q, nil
}

func (s *endDeviceStore) listEndDevicesBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) ([]*ttnpb.EndDevice, error) {
	models := []*EndDevice{}
	selectQuery := s.DB.NewSelect().
		Model(&models).
		Apply(selectWithSoftDeletedFromContext(ctx)).
		Apply(by)

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering, paging and field masking.
	selectQuery = selectQuery.
		Apply(selectWithOrderFromContext(ctx, "device_id", map[string]string{
			"device_id":  "device_id",
			"name":       "name",
			"created_at": "created_at",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	selectQuery, err = s.selectWithFields(selectQuery, fieldMask)
	if err != nil {
		return nil, err
	}

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.EndDevice, len(models))
	for i, model := range models {
		pb, err := endDeviceToPB(model, fieldMask...)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (*endDeviceStore) selectWithID(
	ctx context.Context, applicationID string, deviceIDs ...string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Apply(selectWithContext(ctx))
		q = q.Where("?TableAlias.application_id = ?", applicationID)
		switch len(deviceIDs) {
		case 0:
			return q
		case 1:
			return q.Where("?TableAlias.device_id = ?", deviceIDs[0])
		default:
			return q.Where("?TableAlias.device_id IN (?)", bun.In(deviceIDs))
		}
	}
}

func (*endDeviceStore) selectWithJoinEUI(
	_ context.Context, joinEUI string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("?TableAlias.join_eui = ?", joinEUI)
	}
}

func (*endDeviceStore) selectWithDevEUI(
	_ context.Context, devEUI string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("?TableAlias.dev_eui = ?", devEUI)
	}
}

func (s *endDeviceStore) CountEndDevices(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (uint64, error) {
	ctx, span := tracer.Start(ctx, "CountEndDevices", trace.WithAttributes(
		attribute.String("application_id", ids.GetApplicationId()),
	))
	defer span.End()

	by := noopSelectModifier
	if ids != nil {
		by = s.selectWithID(ctx, ids.GetApplicationId())
	}

	models := []*EndDevice{}
	selectQuery := s.DB.NewSelect().
		Model(&models).
		Apply(selectWithSoftDeletedFromContext(ctx)).
		Apply(by)

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return 0, wrapDriverError(err)
	}

	return uint64(count), nil
}

func (s *endDeviceStore) ListEndDevices(
	ctx context.Context, ids *ttnpb.ApplicationIdentifiers, fieldMask store.FieldMask,
) ([]*ttnpb.EndDevice, error) {
	ctx, span := tracer.Start(ctx, "ListEndDevices", trace.WithAttributes(
		attribute.String("application_id", ids.GetApplicationId()),
	))
	defer span.End()

	by := noopSelectModifier
	if ids != nil {
		by = s.selectWithID(ctx, ids.GetApplicationId())
	}

	return s.listEndDevicesBy(ctx, by, fieldMask)
}

func (s *endDeviceStore) FindEndDevices(
	ctx context.Context, ids []*ttnpb.EndDeviceIdentifiers, fieldMask store.FieldMask,
) ([]*ttnpb.EndDevice, error) {
	ctx, span := tracer.Start(ctx, "FindEndDevices", trace.WithAttributes(
		attribute.StringSlice("end_device_ids", idStrings(ids...)),
	))
	defer span.End()

	by := noopSelectModifier
	if ids != nil {
		applicationID := ids[0].GetApplicationIds().GetApplicationId()
		span.SetAttributes(attribute.String("application_id", applicationID))

		deviceIDs := make([]string, len(ids))
		for i, id := range ids {
			if id.GetApplicationIds().GetApplicationId() != applicationID {
				return nil, fmt.Errorf(
					"inconsistent application ID %q, expected %q",
					id.GetApplicationIds().GetApplicationId(),
					applicationID,
				)
			}
			deviceIDs[i] = id.GetDeviceId()
		}

		by = s.selectWithID(ctx, applicationID, deviceIDs...)
	}

	return s.listEndDevicesBy(ctx, by, fieldMask)
}

func (s *endDeviceStore) getEndDeviceModelBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) (*EndDevice, error) {
	model := &EndDevice{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(selectWithSoftDeletedFromContext(ctx)).
		Apply(by)

	selectQuery, err := s.selectWithFields(selectQuery, fieldMask)
	if err != nil {
		return nil, err
	}

	if err := selectQuery.Scan(ctx); err != nil {
		return nil, wrapDriverError(err)
	}

	return model, nil
}

func (s *endDeviceStore) GetEndDevice(
	ctx context.Context, id *ttnpb.EndDeviceIdentifiers, fieldMask store.FieldMask,
) (*ttnpb.EndDevice, error) {
	ctx, span := tracer.Start(ctx, "GetEndDevice", trace.WithAttributes(
		attribute.String("application_id", id.GetApplicationIds().GetApplicationId()),
		attribute.String("device_id", id.GetDeviceId()),
	))
	defer span.End()

	var by []func(*bun.SelectQuery) *bun.SelectQuery
	if id.GetApplicationIds().GetApplicationId() != "" {
		if id.GetDeviceId() != "" {
			by = append(by, s.selectWithID(ctx, id.GetApplicationIds().GetApplicationId(), id.GetDeviceId()))
		} else {
			by = append(by, s.selectWithID(ctx, id.GetApplicationIds().GetApplicationId()))
		}
	}
	if euiString := eui64ToString(types.MustEUI64(id.GetDevEui())); euiString != nil {
		by = append(by, s.selectWithDevEUI(ctx, *euiString))
	}
	if euiString := eui64ToString(types.MustEUI64(id.GetJoinEui())); euiString != nil {
		by = append(by, s.selectWithJoinEUI(ctx, *euiString))
	}

	model, err := s.getEndDeviceModelBy(ctx, combineApply(by...), fieldMask)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrEndDeviceNotFound.WithAttributes(
				"application_id", id.GetApplicationIds().GetApplicationId(),
				"device_id", id.GetDeviceId(),
			)
		}
		return nil, err
	}
	pb, err := endDeviceToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func (s *endDeviceStore) updateEndDeviceModel( //nolint:gocyclo
	ctx context.Context, model *EndDevice, pb *ttnpb.EndDevice, fieldMask store.FieldMask,
) (err error) {
	columns := store.FieldMask{"updated_at"}

	for _, field := range fieldMask {
		switch field {
		case "ids.join_eui":
			model.JoinEUI = eui64ToString(types.MustEUI64(pb.GetIds().GetJoinEui()))
			columns = append(columns, "join_eui")

		case "ids.dev_eui":
			model.DevEUI = eui64ToString(types.MustEUI64(pb.GetIds().GetDevEui()))
			columns = append(columns, "dev_eui")

		case "name":
			model.Name = pb.Name
			columns = append(columns, "name")

		case "description":
			model.Description = pb.Description
			columns = append(columns, "description")

		case "attributes":
			model.Attributes, err = s.replaceAttributes(
				ctx, model.Attributes, pb.Attributes, "device", model.ID,
			)
			if err != nil {
				return err
			}

		case "version_ids":
			model.BrandID = pb.VersionIds.GetBrandId()
			model.ModelID = pb.VersionIds.GetModelId()
			model.HardwareVersion = pb.VersionIds.GetHardwareVersion()
			model.FirmwareVersion = pb.VersionIds.GetFirmwareVersion()
			model.BandID = pb.VersionIds.GetBandId()
			columns = append(columns, "brand_id", "model_id", "hardware_version", "firmware_version", "band_id")

		case "version_ids.brand_id":
			model.BrandID = pb.VersionIds.GetBrandId()
			columns = append(columns, "brand_id")

		case "version_ids.model_id":
			model.ModelID = pb.VersionIds.GetModelId()
			columns = append(columns, "model_id")

		case "version_ids.hardware_version":
			model.HardwareVersion = pb.VersionIds.GetHardwareVersion()
			columns = append(columns, "hardware_version")

		case "version_ids.firmware_version":
			model.FirmwareVersion = pb.VersionIds.GetFirmwareVersion()
			columns = append(columns, "firmware_version")

		case "version_ids.band_id":
			model.BandID = pb.VersionIds.GetBandId()
			columns = append(columns, "band_id")

		case "network_server_address":
			model.NetworkServerAddress = pb.NetworkServerAddress
			columns = append(columns, "network_server_address")

		case "application_server_address":
			model.ApplicationServerAddress = pb.ApplicationServerAddress
			columns = append(columns, "application_server_address")

		case "join_server_address":
			model.JoinServerAddress = pb.JoinServerAddress
			columns = append(columns, "join_server_address")

		case "service_profile_id":
			model.ServiceProfileID = pb.ServiceProfileId
			columns = append(columns, "service_profile_id")

		case "locations":
			model.Locations, err = s.replaceEndDeviceLocations(
				ctx, model.Locations, pb.Locations, model.ID,
			)
			if err != nil {
				return err
			}

		case "picture":
			if model.Picture != nil {
				_, err = s.DB.NewDelete().
					Model(model.Picture).
					WherePK().
					Exec(ctx)
				if err != nil {
					return wrapDriverError(err)
				}
			}
			if pb.Picture != nil {
				model.Picture, err = pictureFromPB(ctx, pb.Picture)
				if err != nil {
					return err
				}

				_, err = s.DB.NewInsert().
					Model(model.Picture).
					Exec(ctx)
				if err != nil {
					return wrapDriverError(err)
				}

				model.PictureID = &model.Picture.ID
				columns = append(columns, "picture_id")
			} else {
				model.Picture = nil
				model.PictureID = nil
				columns = append(columns, "picture_id")
			}

		case "activated_at":
			model.ActivatedAt = cleanTimePtr(ttnpb.StdTime(pb.ActivatedAt))
			columns = append(columns, "activated_at")

		case "last_seen_at":
			model.LastSeenAt = cleanTimePtr(ttnpb.StdTime(pb.LastSeenAt))
			columns = append(columns, "last_seen_at")

		case "claim_authentication_code":
			// NOTE: The old implementation didn't allow for updating the sub-fields, so we don't either.
			model.ClaimAuthenticationCodeSecret = []byte(pb.ClaimAuthenticationCode.GetValue())
			model.ClaimAuthenticationCodeValidFrom = cleanTimePtr(ttnpb.StdTime(pb.ClaimAuthenticationCode.GetValidFrom()))
			model.ClaimAuthenticationCodeValidTo = cleanTimePtr(ttnpb.StdTime(pb.ClaimAuthenticationCode.GetValidTo()))
			columns = append(
				columns,
				"claim_authentication_code_secret",
				"claim_authentication_code_valid_from",
				"claim_authentication_code_valid_to",
			)
		}
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Column(columns...).
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *endDeviceStore) UpdateEndDevice(
	ctx context.Context, pb *ttnpb.EndDevice, fieldMask store.FieldMask,
) (*ttnpb.EndDevice, error) {
	ctx, span := tracer.Start(ctx, "UpdateEndDevice", trace.WithAttributes(
		attribute.String("application_id", pb.GetIds().GetApplicationIds().GetApplicationId()),
		attribute.String("device_id", pb.GetIds().GetDeviceId()),
	))
	defer span.End()

	model, err := s.getEndDeviceModelBy(
		ctx, s.selectWithID(
			ctx,
			pb.GetIds().GetApplicationIds().GetApplicationId(),
			pb.GetIds().GetDeviceId(),
		), fieldMask,
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrEndDeviceNotFound.WithAttributes(
				"application_id", pb.GetIds().GetApplicationIds().GetApplicationId(),
				"device_id", pb.GetIds().GetDeviceId(),
			)
		}
		return nil, err
	}

	if err = s.updateEndDeviceModel(ctx, model, pb, fieldMask); err != nil {
		return nil, err
	}

	// Convert the result to protobuf.
	updatedPB, err := endDeviceToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}

	return updatedPB, nil
}

func (s *endDeviceStore) DeleteEndDevice(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) error {
	ctx, span := tracer.Start(ctx, "DeleteEndDevice", trace.WithAttributes(
		attribute.String("application_id", id.GetApplicationIds().GetApplicationId()),
		attribute.String("device_id", id.GetDeviceId()),
	))
	defer span.End()

	model, err := s.getEndDeviceModelBy(ctx, s.selectWithID(
		ctx,
		id.GetApplicationIds().GetApplicationId(),
		id.GetDeviceId(),
	), store.FieldMask{"ids", "attributes", "locations"})
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrEndDeviceNotFound.WithAttributes(
				"application_id", id.GetApplicationIds().GetApplicationId(),
				"device_id", id.GetDeviceId(),
			)
		}
		return err
	}

	if len(model.Attributes) > 0 {
		_, err = s.replaceAttributes(ctx, model.Attributes, nil, "device", model.ID)
		if err != nil {
			return err
		}
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	if _, err = s.replaceEndDeviceLocations(ctx, model.Locations, nil, model.ID); err != nil {
		return err
	}

	if model.PictureID != nil {
		_, err = s.DB.NewDelete().
			Model((*Picture)(nil)).
			Where("id = ?", *model.PictureID).
			Exec(ctx)
		if err != nil {
			return wrapDriverError(err)
		}
	}

	return nil
}

func (s *endDeviceStore) BatchUpdateEndDeviceLastSeen(
	ctx context.Context,
	devsLastSeen []*ttnpb.BatchUpdateEndDeviceLastSeenRequest_EndDeviceLastSeenUpdate,
) error {
	ctx, span := tracer.Start(ctx, "BatchUpdateEndDeviceLastSeen", trace.WithAttributes(
		attribute.Int("count", len(devsLastSeen)),
	))
	defer span.End()

	// Sort end devices by ID to avoid deadlocks.
	sort.Slice(devsLastSeen, func(i, j int) bool {
		return devsLastSeen[i].Ids.IDString() < devsLastSeen[j].Ids.IDString()
	})

	type lastSeenUpdate struct {
		ApplicationID string
		DeviceID      string
		LastSeenAt    *time.Time
	}
	lastSeenUpdates := make([]*lastSeenUpdate, len(devsLastSeen))
	for i, dev := range devsLastSeen {
		lastSeenUpdates[i] = &lastSeenUpdate{
			ApplicationID: dev.Ids.GetApplicationIds().GetApplicationId(),
			DeviceID:      dev.Ids.GetDeviceId(),
			LastSeenAt:    cleanTimePtr(ttnpb.StdTime(dev.GetLastSeenAt())),
		}
	}

	values := s.DB.NewValues(&lastSeenUpdates)

	updateQuery := s.DB.NewUpdate().
		With("_data", values).
		Model((*EndDevice)(nil)).
		TableExpr("_data").
		Set("last_seen_at = _data.last_seen_at")

	updateQuery = updateQuery.
		Where("?TableAlias.application_id = _data.application_id").
		Where("?TableAlias.device_id = _data.device_id").
		Where("?TableAlias.last_seen_at IS NULL OR ?TableAlias.last_seen_at < _data.last_seen_at")

	_, err := updateQuery.Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

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
	"strings"
	"time"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Gateway is the gateway model in the database.
type Gateway struct {
	bun.BaseModel `bun:"table:gateways,alias:gtw"`

	Model
	SoftDelete

	GatewayID  string  `bun:"gateway_id,notnull"`
	GatewayEUI *string `bun:"gateway_eui"`

	Name        string `bun:"name,nullzero"`
	Description string `bun:"description,nullzero"`

	Attributes []*Attribute `bun:"rel:has-many,join:type=entity_type,join:id=entity_id,polymorphic"`

	ContactInfo []*ContactInfo `bun:"rel:has-many,join:type=entity_type,join:id=entity_id,polymorphic"`

	AdministrativeContactID *string  `bun:"administrative_contact_id,type:uuid"`
	AdministrativeContact   *Account `bun:"rel:belongs-to,join:administrative_contact_id=id"`

	TechnicalContactID *string  `bun:"technical_contact_id,type:uuid"`
	TechnicalContact   *Account `bun:"rel:belongs-to,join:technical_contact_id=id"`

	BrandID         string `bun:"brand_id,nullzero"`
	ModelID         string `bun:"model_id,nullzero"`
	HardwareVersion string `bun:"hardware_version,nullzero"`
	FirmwareVersion string `bun:"firmware_version,nullzero"`

	GatewayServerAddress string `bun:"gateway_server_address,nullzero"`

	AutoUpdate    bool   `bun:"auto_update,notnull"`
	UpdateChannel string `bun:"update_channel,nullzero"`

	// Frequency Plan IDs separated by spaces.
	FrequencyPlanID string `bun:"frequency_plan_id,nullzero"`

	StatusPublic   bool `bun:"status_public,notnull"`
	LocationPublic bool `bun:"location_public,notnull"`

	ScheduleDownlinkLate   bool  `bun:"schedule_downlink_late,notnull"`
	EnforceDutyCycle       bool  `bun:"enforce_duty_cycle,notnull"`
	ScheduleAnytimeDelay   int64 `bun:"schedule_anytime_delay,notnull"`
	DownlinkPathConstraint int   `bun:"downlink_path_constraint,nullzero"`

	UpdateLocationFromStatus bool `bun:"update_location_from_status,notnull"`

	Antennas []*GatewayAntenna `bun:"rel:has-many,join:id=gateway_id"`

	LBSLNSSecret []byte `bun:"lbs_lns_secret,nullzero"`

	ClaimAuthenticationCodeSecret    []byte     `bun:"claim_authentication_code_secret"`
	ClaimAuthenticationCodeValidFrom *time.Time `bun:"claim_authentication_code_valid_from"`
	ClaimAuthenticationCodeValidTo   *time.Time `bun:"claim_authentication_code_valid_to"`

	TargetCUPSURI string `bun:"target_cups_uri,nullzero"`
	TargetCUPSKey []byte `bun:"target_cups_key"`

	RequireAuthenticatedConnection bool `bun:"require_authenticated_connection,nullzero"`

	SupportsLRFHSS bool `bun:"supports_lrfhss,notnull"`

	DisablePacketBrokerForwarding bool `bun:"disable_packet_broker_forwarding,notnull"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *Gateway) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func gatewayToPB(m *Gateway, fieldMask ...string) (*ttnpb.Gateway, error) {
	pb := &ttnpb.Gateway{
		Ids: &ttnpb.GatewayIdentifiers{
			GatewayId: m.GatewayID,
			Eui:       eui64FromString(m.GatewayEUI),
		},

		CreatedAt: ttnpb.ProtoTimePtr(m.CreatedAt),
		UpdatedAt: ttnpb.ProtoTimePtr(m.UpdatedAt),
		DeletedAt: ttnpb.ProtoTime(m.DeletedAt),

		Name:        m.Name,
		Description: m.Description,

		VersionIds: &ttnpb.GatewayVersionIdentifiers{
			BrandId:         m.BrandID,
			ModelId:         m.ModelID,
			HardwareVersion: m.HardwareVersion,
			FirmwareVersion: m.FirmwareVersion,
		},

		GatewayServerAddress: m.GatewayServerAddress,

		AutoUpdate:    m.AutoUpdate,
		UpdateChannel: m.UpdateChannel,

		// NOTE: FrequencyPlanId is deprecated. Compatibility is handled on the business layer.
		FrequencyPlanIds: strings.Split(m.FrequencyPlanID, " "),

		StatusPublic:   m.StatusPublic,
		LocationPublic: m.LocationPublic,

		ScheduleDownlinkLate:   m.ScheduleDownlinkLate,
		EnforceDutyCycle:       m.EnforceDutyCycle,
		ScheduleAnytimeDelay:   ttnpb.ProtoDurationPtr(time.Duration(m.ScheduleAnytimeDelay)),
		DownlinkPathConstraint: ttnpb.DownlinkPathConstraint(m.DownlinkPathConstraint),

		UpdateLocationFromStatus: m.UpdateLocationFromStatus,

		LbsLnsSecret: secretFromBytes(m.LBSLNSSecret),

		ClaimAuthenticationCode: func() *ttnpb.GatewayClaimAuthenticationCode {
			secret := secretFromBytes(m.ClaimAuthenticationCodeSecret)
			if secret == nil {
				return nil
			}
			return &ttnpb.GatewayClaimAuthenticationCode{
				Secret:    secret,
				ValidFrom: ttnpb.ProtoTime(m.ClaimAuthenticationCodeValidFrom),
				ValidTo:   ttnpb.ProtoTime(m.ClaimAuthenticationCodeValidTo),
			}
		}(),

		TargetCupsUri: m.TargetCUPSURI,
		TargetCupsKey: secretFromBytes(m.TargetCUPSKey),

		RequireAuthenticatedConnection: m.RequireAuthenticatedConnection,

		Lrfhss: func() *ttnpb.Gateway_LRFHSS {
			if m.SupportsLRFHSS {
				return &ttnpb.Gateway_LRFHSS{Supported: true}
			}
			return nil
		}(),

		DisablePacketBrokerForwarding: m.DisablePacketBrokerForwarding,
	}

	if len(m.Attributes) > 0 {
		pb.Attributes = make(map[string]string, len(m.Attributes))
		for _, a := range m.Attributes {
			pb.Attributes[a.Key] = a.Value
		}
	}

	if len(m.ContactInfo) > 0 {
		pb.ContactInfo = make([]*ttnpb.ContactInfo, len(m.ContactInfo))
		for i, contactInfo := range m.ContactInfo {
			pb.ContactInfo[i] = contactInfoToPB(contactInfo)
		}
		sort.Sort(contactInfoProtoSlice(pb.ContactInfo))
	}

	if m.AdministrativeContact != nil {
		pb.AdministrativeContact = m.AdministrativeContact.GetOrganizationOrUserIdentifiers()
	}
	if m.TechnicalContact != nil {
		pb.TechnicalContact = m.TechnicalContact.GetOrganizationOrUserIdentifiers()
	}

	if len(m.Antennas) > 0 {
		pb.Antennas = make([]*ttnpb.GatewayAntenna, len(m.Antennas))
		for i, antenna := range m.Antennas {
			pb.Antennas[i] = gatewayAntennaToPB(antenna)
		}
	}

	if len(fieldMask) == 0 {
		return pb, nil
	}

	res := &ttnpb.Gateway{}
	if err := res.SetFields(pb, fieldMask...); err != nil {
		return nil, err
	}

	// Set fields that are always present.
	res.Ids = pb.Ids
	res.CreatedAt = pb.CreatedAt
	res.UpdatedAt = pb.UpdatedAt
	res.DeletedAt = pb.DeletedAt

	return res, nil
}

type gatewayStore struct {
	*baseStore
}

func newGatewayStore(baseStore *baseStore) *gatewayStore {
	return &gatewayStore{
		baseStore: baseStore,
	}
}

func (s *gatewayStore) CreateGateway(
	ctx context.Context, pb *ttnpb.Gateway,
) (*ttnpb.Gateway, error) {
	ctx, span := tracer.Start(ctx, "CreateGateway", trace.WithAttributes(
		attribute.String("gateway_id", pb.GetIds().GetGatewayId()),
	))
	defer span.End()

	gatewayModel := &Gateway{
		GatewayID:   pb.GetIds().GetGatewayId(),
		GatewayEUI:  eui64ToString(pb.GetIds().GetEui()),
		Name:        pb.Name,
		Description: pb.Description,

		BrandID:         pb.VersionIds.GetBrandId(),
		ModelID:         pb.VersionIds.GetModelId(),
		HardwareVersion: pb.VersionIds.GetHardwareVersion(),
		FirmwareVersion: pb.VersionIds.GetFirmwareVersion(),

		GatewayServerAddress: pb.GatewayServerAddress,

		AutoUpdate:    pb.AutoUpdate,
		UpdateChannel: pb.UpdateChannel,

		// NOTE: pb.FrequencyPlanId is deprecated. Compatibility is handled on the business layer.
		FrequencyPlanID: strings.Join(pb.FrequencyPlanIds, " "),

		StatusPublic:   pb.StatusPublic,
		LocationPublic: pb.LocationPublic,

		ScheduleDownlinkLate:   pb.ScheduleDownlinkLate,
		EnforceDutyCycle:       pb.EnforceDutyCycle,
		ScheduleAnytimeDelay:   int64(ttnpb.StdDurationOrZero(pb.ScheduleAnytimeDelay)),
		DownlinkPathConstraint: int(pb.DownlinkPathConstraint),

		UpdateLocationFromStatus: pb.UpdateLocationFromStatus,

		LBSLNSSecret: secretToBytes(pb.LbsLnsSecret),

		ClaimAuthenticationCodeSecret:    secretToBytes(pb.ClaimAuthenticationCode.GetSecret()),
		ClaimAuthenticationCodeValidFrom: ttnpb.StdTime(pb.ClaimAuthenticationCode.GetValidFrom()),
		ClaimAuthenticationCodeValidTo:   ttnpb.StdTime(pb.ClaimAuthenticationCode.GetValidTo()),

		TargetCUPSURI: pb.TargetCupsUri,
		TargetCUPSKey: secretToBytes(pb.TargetCupsKey),

		RequireAuthenticatedConnection: pb.RequireAuthenticatedConnection,

		SupportsLRFHSS: pb.Lrfhss.GetSupported(),

		DisablePacketBrokerForwarding: pb.DisablePacketBrokerForwarding,
	}

	if contact := pb.AdministrativeContact; contact != nil {
		account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
		if err != nil {
			return nil, err
		}
		gatewayModel.AdministrativeContact = account
		gatewayModel.AdministrativeContactID = &account.ID
	}
	if contact := pb.TechnicalContact; contact != nil {
		account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
		if err != nil {
			return nil, err
		}
		gatewayModel.TechnicalContact = account
		gatewayModel.TechnicalContactID = &account.ID
	}

	_, err := s.DB.NewInsert().
		Model(gatewayModel).
		Exec(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	if len(pb.Attributes) > 0 {
		gatewayModel.Attributes, err = s.replaceAttributes(
			ctx, nil, pb.Attributes, "gateway", gatewayModel.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	if len(pb.ContactInfo) > 0 {
		gatewayModel.ContactInfo, err = s.replaceContactInfo(
			ctx, nil, pb.ContactInfo, "gateway", gatewayModel.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	if len(pb.Antennas) > 0 {
		gatewayModel.Antennas, err = s.replaceGatewayAntennas(
			ctx, nil, pb.Antennas, gatewayModel.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	pb, err = gatewayToPB(gatewayModel)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (*gatewayStore) selectWithFields(q *bun.SelectQuery, fieldMask store.FieldMask) (*bun.SelectQuery, error) {
	if fieldMask == nil {
		q = q.ExcludeColumn()
	} else {
		columns := []string{
			"id",
			"created_at",
			"updated_at",
			"deleted_at",
			"gateway_id",
			"gateway_eui",
		}
		for _, f := range fieldMask.TopLevel() {
			switch f {
			default:
				return nil, fmt.Errorf("unknown field %q", f)
			case "ids", "created_at", "updated_at", "deleted_at":
				// Always selected.
			case "name", "description",
				"gateway_server_address",
				"auto_update", "update_channel",
				"status_public", "location_public",
				"schedule_downlink_late",
				"enforce_duty_cycle",
				"downlink_path_constraint",
				"schedule_anytime_delay",
				"update_location_from_status",
				"lbs_lns_secret",
				"target_cups_uri", "target_cups_key",
				"require_authenticated_connection",
				"disable_packet_broker_forwarding":
				// Proto name equals model name.
				columns = append(columns, f)
			case "version_ids":
				columns = append(columns, "brand_id", "model_id", "hardware_version", "firmware_version")
			case "frequency_plan_id":
				// Ignore deprecated field.
			case "frequency_plan_ids":
				columns = append(columns, "frequency_plan_id")
			case "claim_authentication_code":
				columns = append(
					columns,
					"claim_authentication_code_secret",
					"claim_authentication_code_valid_from",
					"claim_authentication_code_valid_to",
				)
			case "lrfhss":
				columns = append(columns, "supports_lrfhss")
			case "attributes":
				q = q.Relation("Attributes")
			case "contact_info":
				q = q.Relation("ContactInfo")
			case "administrative_contact":
				q = q.Relation("AdministrativeContact", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Column("uid", "account_type")
				})
			case "technical_contact":
				q = q.Relation("TechnicalContact", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Column("uid", "account_type")
				})
			case "antennas":
				q = q.Relation("Antennas", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Order("index")
				})
			}
		}
		q = q.Column(columns...)
	}
	return q, nil
}

func (s *gatewayStore) listGatewaysBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) ([]*ttnpb.Gateway, error) {
	models := []*Gateway{}
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
		Apply(selectWithOrderFromContext(ctx, "gateway_id", map[string]string{
			"gateway_id": "gateway_id",
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
	pbs := make([]*ttnpb.Gateway, len(models))
	for i, model := range models {
		pb, err := gatewayToPB(model, fieldMask...)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (s *gatewayStore) selectWithID(ctx context.Context, ids ...string) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Apply(selectWithContext(ctx))
		switch len(ids) {
		case 0:
			return q
		case 1:
			return q.Where("?TableAlias.gateway_id = ?", ids[0])
		default:
			return q.Where("?TableAlias.gateway_id IN (?)", bun.In(ids))
		}
	}
}

func (s *gatewayStore) selectWithEUI(ctx context.Context, euis ...string) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Apply(selectWithContext(ctx))
		switch len(euis) {
		case 0:
			return q
		case 1:
			return q.Where("?TableAlias.gateway_eui = ?", euis[0])
		default:
			return q.Where("?TableAlias.gateway_eui IN (?)", bun.In(euis))
		}
	}
}

func (s *gatewayStore) FindGateways(
	ctx context.Context, ids []*ttnpb.GatewayIdentifiers, fieldMask store.FieldMask,
) ([]*ttnpb.Gateway, error) {
	ctx, span := tracer.Start(ctx, "FindGateways", trace.WithAttributes(
		attribute.StringSlice("gateway_ids", idStrings(ids...)),
	))
	defer span.End()

	return s.listGatewaysBy(ctx, s.selectWithID(ctx, idStrings(ids...)...), fieldMask)
}

func (s *gatewayStore) getGatewayModelBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) (*Gateway, error) {
	model := &Gateway{}
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

func (s *gatewayStore) GetGateway(
	ctx context.Context, id *ttnpb.GatewayIdentifiers, fieldMask store.FieldMask,
) (*ttnpb.Gateway, error) {
	ctx, span := tracer.Start(ctx, "GetGateway", trace.WithAttributes(
		attribute.String("gateway_id", id.GetGatewayId()),
	))
	defer span.End()

	by := s.selectWithID(ctx, id.GetGatewayId())
	if euiString := eui64ToString(id.GetEui()); euiString != nil {
		by = s.selectWithEUI(ctx, *euiString)
	}

	model, err := s.getGatewayModelBy(ctx, by, fieldMask)
	if err != nil {
		return nil, err
	}
	pb, err := gatewayToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func (s *gatewayStore) updateGatewayModel( //nolint:gocyclo
	ctx context.Context, model *Gateway, pb *ttnpb.Gateway, fieldMask store.FieldMask,
) (err error) {
	columns := store.FieldMask{"updated_at"}

	for _, field := range fieldMask {
		switch field {
		case "ids.eui":
			model.GatewayEUI = eui64ToString(pb.GetIds().GetEui())
			columns = append(columns, "gateway_eui")

		case "name":
			model.Name = pb.Name
			columns = append(columns, "name")

		case "description":
			model.Description = pb.Description
			columns = append(columns, "description")

		case "attributes":
			model.Attributes, err = s.replaceAttributes(
				ctx, model.Attributes, pb.Attributes, "gateway", model.ID,
			)
			if err != nil {
				return err
			}

		case "contact_info":
			model.ContactInfo, err = s.replaceContactInfo(
				ctx, model.ContactInfo, pb.ContactInfo, "gateway", model.ID,
			)
			if err != nil {
				return err
			}

		case "administrative_contact":
			if contact := pb.AdministrativeContact; contact != nil {
				account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
				if err != nil {
					return err
				}
				model.AdministrativeContact = account
				model.AdministrativeContactID = &account.ID
			} else {
				model.AdministrativeContact = nil
				model.AdministrativeContactID = nil
			}
			columns = append(columns, "administrative_contact_id")

		case "technical_contact":
			if contact := pb.TechnicalContact; contact != nil {
				account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
				if err != nil {
					return err
				}
				model.TechnicalContact = account
				model.TechnicalContactID = &account.ID
			} else {
				model.TechnicalContact = nil
				model.TechnicalContactID = nil
			}
			columns = append(columns, "technical_contact_id")

		case "version_ids":
			model.BrandID = pb.VersionIds.GetBrandId()
			model.ModelID = pb.VersionIds.GetModelId()
			model.HardwareVersion = pb.VersionIds.GetHardwareVersion()
			model.FirmwareVersion = pb.VersionIds.GetFirmwareVersion()
			columns = append(columns, "brand_id", "model_id", "hardware_version", "firmware_version")

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

		case "gateway_server_address":
			model.GatewayServerAddress = pb.GatewayServerAddress
			columns = append(columns, "gateway_server_address")

		case "auto_update":
			model.AutoUpdate = pb.AutoUpdate
			columns = append(columns, "auto_update")

		case "update_channel":
			model.UpdateChannel = pb.UpdateChannel
			columns = append(columns, "update_channel")

		case "frequency_plan_ids":
			model.FrequencyPlanID = strings.Join(pb.FrequencyPlanIds, " ")
			columns = append(columns, "frequency_plan_id")

		case "status_public":
			model.StatusPublic = pb.StatusPublic
			columns = append(columns, "status_public")

		case "location_public":
			model.LocationPublic = pb.LocationPublic
			columns = append(columns, "location_public")

		case "schedule_downlink_late":
			model.ScheduleDownlinkLate = pb.ScheduleDownlinkLate
			columns = append(columns, "schedule_downlink_late")

		case "enforce_duty_cycle":
			model.EnforceDutyCycle = pb.EnforceDutyCycle
			columns = append(columns, "enforce_duty_cycle")

		case "schedule_anytime_delay":
			model.ScheduleAnytimeDelay = int64(ttnpb.StdDurationOrZero(pb.ScheduleAnytimeDelay))
			columns = append(columns, "schedule_anytime_delay")

		case "downlink_path_constraint":
			model.DownlinkPathConstraint = int(pb.DownlinkPathConstraint)
			columns = append(columns, "downlink_path_constraint")

		case "update_location_from_status":
			model.UpdateLocationFromStatus = pb.UpdateLocationFromStatus
			columns = append(columns, "update_location_from_status")

		case "antennas":
			model.Antennas, err = s.replaceGatewayAntennas(
				ctx, model.Antennas, pb.Antennas, model.ID,
			)
			if err != nil {
				return err
			}

		case "lbs_lns_secret":
			model.LBSLNSSecret = secretToBytes(pb.LbsLnsSecret)
			columns = append(columns, "lbs_lns_secret")

		case "claim_authentication_code":
			// NOTE: The old implementation didn't allow for updating the sub-fields, so we don't either.
			model.ClaimAuthenticationCodeSecret = secretToBytes(pb.ClaimAuthenticationCode.GetSecret())
			model.ClaimAuthenticationCodeValidFrom = ttnpb.StdTime(pb.ClaimAuthenticationCode.GetValidFrom())
			model.ClaimAuthenticationCodeValidTo = ttnpb.StdTime(pb.ClaimAuthenticationCode.GetValidTo())
			columns = append(
				columns,
				"claim_authentication_code_secret",
				"claim_authentication_code_valid_from",
				"claim_authentication_code_valid_to",
			)

		case "target_cups_uri":
			model.TargetCUPSURI = pb.TargetCupsUri
			columns = append(columns, "target_cups_uri")

		case "target_cups_key":
			model.TargetCUPSKey = secretToBytes(pb.TargetCupsKey)
			columns = append(columns, "target_cups_key")

		case "require_authenticated_connection":
			model.RequireAuthenticatedConnection = pb.RequireAuthenticatedConnection
			columns = append(columns, "require_authenticated_connection")

		case "lrfhss", "lrfhss.supported":
			model.SupportsLRFHSS = pb.Lrfhss.GetSupported()
			columns = append(columns, "supports_lrfhss")

		case "disable_packet_broker_forwarding":
			model.DisablePacketBrokerForwarding = pb.DisablePacketBrokerForwarding
			columns = append(columns, "disable_packet_broker_forwarding")
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

func (s *gatewayStore) UpdateGateway(
	ctx context.Context, pb *ttnpb.Gateway, fieldMask store.FieldMask,
) (*ttnpb.Gateway, error) {
	ctx, span := tracer.Start(ctx, "UpdateGateway", trace.WithAttributes(
		attribute.String("gateway_id", pb.GetIds().GetGatewayId()),
	))
	defer span.End()

	model, err := s.getGatewayModelBy(
		ctx, s.selectWithID(ctx, pb.GetIds().GetGatewayId()), fieldMask,
	)
	if err != nil {
		return nil, err
	}

	if err = s.updateGatewayModel(ctx, model, pb, fieldMask); err != nil {
		return nil, err
	}

	// Convert the result to protobuf.
	updatedPB, err := gatewayToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}

	return updatedPB, nil
}

func (s *gatewayStore) DeleteGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error {
	ctx, span := tracer.Start(ctx, "DeleteGateway", trace.WithAttributes(
		attribute.String("gateway_id", id.GetGatewayId()),
	))
	defer span.End()

	model, err := s.getGatewayModelBy(ctx, s.selectWithID(ctx, id.GetGatewayId()), store.FieldMask{"ids"})
	if err != nil {
		return err
	}

	// TODO: Replace unique constraint to only check EUI for deleted_at = NULL.
	// _, err = s.DB.NewDelete().
	// 	Model(model).
	// 	WherePK().
	// 	Exec(ctx)
	// if err != nil {
	// 	return wrapDriverError(err)
	// }

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Set("deleted_at = ?", time.Now().UTC()).
		Set("gateway_eui = NULL").
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *gatewayStore) RestoreGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error {
	ctx, span := tracer.Start(ctx, "RestoreGateway", trace.WithAttributes(
		attribute.String("gateway_id", id.GetGatewayId()),
	))
	defer span.End()

	model, err := s.getGatewayModelBy(
		store.WithSoftDeleted(ctx, true),
		s.selectWithID(ctx, id.GetGatewayId()),
		store.FieldMask{"ids"},
	)
	if err != nil {
		return err
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		WhereAllWithDeleted().
		Set("deleted_at = NULL").
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *gatewayStore) PurgeGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) error {
	ctx, span := tracer.Start(ctx, "PurgeGateway", trace.WithAttributes(
		attribute.String("gateway_id", id.GetGatewayId()),
	))
	defer span.End()

	model, err := s.getGatewayModelBy(
		store.WithSoftDeleted(ctx, false),
		s.selectWithID(ctx, id.GetGatewayId()),
		store.FieldMask{"attributes", "contact_info", "antennas"},
	)
	if err != nil {
		return err
	}

	if len(model.Attributes) > 0 {
		_, err = s.replaceAttributes(ctx, model.Attributes, nil, "gateway", model.ID)
		if err != nil {
			return err
		}
	}

	if len(model.ContactInfo) > 0 {
		_, err = s.replaceContactInfo(ctx, model.ContactInfo, nil, "gateway", model.ID)
		if err != nil {
			return err
		}
	}

	if _, err = s.replaceGatewayAntennas(ctx, model.Antennas, nil, model.ID); err != nil {
		return err
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		ForceDelete().
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

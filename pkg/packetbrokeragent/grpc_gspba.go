// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package packetbrokeragent

import (
	"context"
	"fmt"
	"time"

	mappingpb "go.packetbroker.org/api/mapping/v2"
	packetbroker "go.packetbroker.org/api/v3"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// DefaultGatewayOnlineTTL is the default time-to-live of the online status reported to Packet Broker.
// Packet Broker Agent must bump the online status before the previous online status expires, to keep the gateway marked online.
// A low value results in more calls to Packet Broker Mapper to keep gateways online.
// A high value results in a longer time until gateways that go offline will be marked offline on the map.
const DefaultGatewayOnlineTTL = 10 * time.Minute

type messageEncrypter interface {
	encryptUplink(context.Context, *packetbroker.UplinkMessage) error
}

type frequencyPlansStore interface {
	GetByID(id string) (*frequencyplans.FrequencyPlan, error)
}

// GetFrequencyPlansStore defines a function that returns the frequencyPlansStore interface
type GetFrequencyPlansStore func(ctx context.Context) (frequencyPlansStore, error)

type gsPbaServer struct {
	ttnpb.UnimplementedGsPbaServer

	netID               types.NetID
	clusterID           string
	config              ForwarderConfig
	messageEncrypter    messageEncrypter
	contextDecoupler    contextDecoupler
	tenantIDExtractor   TenantIDExtractor
	frequencyPlansStore GetFrequencyPlansStore
	upstreamCh          chan *uplinkMessage
	mapperConn          *grpc.ClientConn
	entityRegistry      EntityRegistry
}

// PublishUplink is called by the Gateway Server when an uplink message arrives and needs to get forwarded to Packet Broker.
func (s *gsPbaServer) PublishUplink(ctx context.Context, up *ttnpb.GatewayUplinkMessage) (*emptypb.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ctx = events.ContextWithCorrelationID(ctx, append(
		up.Message.CorrelationIds,
		fmt.Sprintf("pba:uplink:%s", events.NewCorrelationID()),
	)...)
	up.Message.CorrelationIds = events.CorrelationIDsFromContext(ctx)

	msg, err := toPBUplink(ctx, up, s.config)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to convert outgoing uplink message")
		return nil, err
	}
	if err := s.messageEncrypter.encryptUplink(ctx, msg); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to encrypt outgoing uplink message")
		return nil, err
	}

	ctxMsg := &uplinkMessage{
		Context:       s.contextDecoupler.FromRequestContext(ctx),
		UplinkMessage: msg,
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.upstreamCh <- ctxMsg:
		return ttnpb.Empty, nil
	}
}

var (
	errPacketBrokerInternal = errors.DefineAborted("packet_broker_internal", "internal Packet Broker error")
	errNoGatewayID          = errors.DefineFailedPrecondition(
		"no_gateway_id", "no gateway identifier provided or included in configuration",
	)
)

// UpdateGateway is called by Gateway Server to update a gateway.
func (s *gsPbaServer) UpdateGateway( // nolint: gocyclo
	ctx context.Context, req *ttnpb.UpdatePacketBrokerGatewayRequest,
) (*ttnpb.UpdatePacketBrokerGatewayResponse, error) {
	logger := log.FromContext(ctx)
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	id := toPBGatewayIdentifier(&ttnpb.GatewayIdentifiers{
		GatewayId: req.Gateway.Ids.GatewayId,
		Eui:       req.Gateway.Ids.Eui,
	}, s.config)
	if id == nil {
		return nil, errNoGatewayID.New()
	}
	updateReq := &mappingpb.UpdateGatewayRequest{
		ForwarderNetId:     s.netID.MarshalNumber(),
		ForwarderTenantId:  s.tenantIDExtractor(ctx),
		ForwarderClusterId: s.clusterID,
		ForwarderGatewayId: id,
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "location_public") {
		updateReq.GatewayLocation = &packetbroker.GatewayLocationValue{}
		var val *packetbroker.GatewayLocation_Terrestrial
		if req.Gateway.LocationPublic && ttnpb.HasAnyField(req.FieldMask.GetPaths(), "antennas") && len(req.Gateway.Antennas) > 0 {
			val = &packetbroker.GatewayLocation_Terrestrial{
				AntennaCount: &wrapperspb.UInt32Value{
					Value: uint32(len(req.Gateway.Antennas)),
				},
				AntennaPlacement: toPBTerrestrialAntennaPlacement(req.Gateway.Antennas[0].Placement),
			}
			if loc := req.Gateway.Antennas[0].Location; loc != nil {
				val.Location = toPBLocation(loc)
			}
		} else {
			val = &packetbroker.GatewayLocation_Terrestrial{
				AntennaCount: &wrapperspb.UInt32Value{Value: 0},
			}
		}
		updateReq.GatewayLocation.Location = &packetbroker.GatewayLocation{
			Value: &packetbroker.GatewayLocation_Terrestrial_{
				Terrestrial: val,
			},
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "status_public") && ttnpb.HasAnyField(req.FieldMask.GetPaths(), "online") {
		updateReq.Online = &wrapperspb.BoolValue{}
		if req.Gateway.StatusPublic && req.Gateway.Online {
			updateReq.Online.Value = true
			updateReq.OnlineTtl = durationpb.New(s.config.GatewayOnlineTTL)
		}
	}

	var adminContact, techContact *packetbroker.ContactInfo

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "administrative_contact") {
		// Only replaces the administrative contact if the contact is validated and no error happens.
		adminUsr, err := s.fetchUserFromContact(
			ctx, req.Gateway.GetAdministrativeContact(),
			func(o *ttnpb.Organization) *ttnpb.OrganizationOrUserIdentifiers { return o.GetAdministrativeContact() },
		)
		if err == nil && adminUsr.PrimaryEmailAddressValidatedAt != nil {
			adminContact = &packetbroker.ContactInfo{Name: adminUsr.Ids.UserId, Email: adminUsr.PrimaryEmailAddress}
		} else if !errors.IsNotFound(err) {
			logger.
				WithField("gateway_id", req.Gateway.Ids).
				WithError(err).
				Warn("Failed to fetch administrative contact")
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "technical_contact") {
		// Only replaces the technical contact if the contact is validated and no error happens.
		techUsr, err := s.fetchUserFromContact(
			ctx, req.Gateway.GetTechnicalContact(),
			func(o *ttnpb.Organization) *ttnpb.OrganizationOrUserIdentifiers { return o.GetTechnicalContact() },
		)
		if err == nil && techUsr.PrimaryEmailAddressValidatedAt != nil {
			techContact = &packetbroker.ContactInfo{Name: techUsr.Ids.UserId, Email: techUsr.PrimaryEmailAddress}
		} else if !errors.IsNotFound(err) {
			logger.
				WithField("gateway_id", req.Gateway.Ids).
				WithError(err).
				Warn("Failed to fetch technical contact")
		}
	}

	if adminContact != nil {
		updateReq.AdministrativeContact = &packetbroker.ContactInfoValue{Value: adminContact}
	}
	if techContact != nil {
		updateReq.TechnicalContact = &packetbroker.ContactInfoValue{Value: techContact}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "frequency_plan_ids") {
		fps, err := s.frequencyPlansStore(ctx)
		if err != nil {
			return nil, err
		}

		fpSlice := make([]*frequencyplans.FrequencyPlan, 0, len(req.Gateway.FrequencyPlanIds))
		for _, fpID := range req.Gateway.FrequencyPlanIds {
			var fp *frequencyplans.FrequencyPlan
			fp, err = fps.GetByID(fpID)
			if err != nil {
				break
			}
			fpSlice = append(fpSlice, fp)
		}
		if err == nil {
			if fp, err := toPBFrequencyPlan(fpSlice...); err == nil {
				updateReq.FrequencyPlan = fp
			}
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "rx_rate") {
		updateReq.RxRate = req.Gateway.RxRate
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "tx_rate") {
		updateReq.TxRate = req.Gateway.TxRate
	}

	_, err := mappingpb.NewMapperClient(s.mapperConn).UpdateGateway(ctx, updateReq)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to update gateway")
		if errors.IsInternal(err) {
			return nil, errPacketBrokerInternal.WithCause(err)
		}
		return nil, err
	}

	res := &ttnpb.UpdatePacketBrokerGatewayResponse{}
	if updateReq.Online.GetValue() {
		res.OnlineTtl = durationpb.New(s.config.GatewayOnlineTTL)
	}
	return res, nil
}

func (s *gsPbaServer) fetchUserFromContact(
	ctx context.Context, contact *ttnpb.OrganizationOrUserIdentifiers,
	selectOrgContact func(*ttnpb.Organization) *ttnpb.OrganizationOrUserIdentifiers,
) (*ttnpb.User, error) {
	usrID := contact.GetUserIds()

	// If the contact is an organization, get its contact before fetching the user information.
	if orgID := contact.GetOrganizationIds(); orgID != nil {
		org, err := s.entityRegistry.GetOrganization(ctx, &ttnpb.GetOrganizationRequest{
			OrganizationIds: orgID,
			FieldMask:       ttnpb.FieldMask("administrative_contact", "technical_contact"),
		})
		if err != nil {
			return nil, err
		}
		usrID = selectOrgContact(org).GetUserIds()
	}

	return s.entityRegistry.GetUser(ctx, &ttnpb.GetUserRequest{
		UserIds:   usrID,
		FieldMask: ttnpb.FieldMask("ids", "primary_email_address", "primary_email_address_validated_at"),
	})
}

// EntityRegistry abstracts the Identity server user/organization functions.
type EntityRegistry interface {
	GetUser(ctx context.Context, req *ttnpb.GetUserRequest) (*ttnpb.User, error)
	GetOrganization(ctx context.Context, req *ttnpb.GetOrganizationRequest) (*ttnpb.Organization, error)
}

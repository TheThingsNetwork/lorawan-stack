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
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func TestGatewayStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Gateway{}, &GatewayAntenna{}, &Attribute{})
		store := GetGatewayStore(db)
		s := newStore(db)

		eui := &types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
		scheduleAnytimeDelay := time.Second
		targetCUPSURI := "https://thethings.example.com"
		otherTargetCUPSURI := "https://thenotthings.example.com"
		secret := &ttnpb.Secret{
			KeyId: "my-secret-key-id",
			Value: []byte("my very secret value"),
		}
		otherSecret := &ttnpb.Secret{
			KeyId: "my-secret-key-id",
			Value: []byte("my other very secret value"),
		}

		from := time.Now().UTC()
		to := from.Add(5 * time.Minute)
		gtwClaimAuthCode := ttnpb.GatewayClaimAuthenticationCode{
			ValidFrom: ttnpb.ProtoTimePtr(from),
			ValidTo:   ttnpb.ProtoTimePtr(to),
			Secret: &ttnpb.Secret{
				KeyId: "my-secret-key-id",
				Value: []byte("my very secret value"),
			},
		}
		otherGtwClaimAuthCode := ttnpb.GatewayClaimAuthenticationCode{
			ValidFrom: ttnpb.ProtoTimePtr(from),
			ValidTo:   ttnpb.ProtoTimePtr(to),
			Secret: &ttnpb.Secret{
				KeyId: "my-secret-key-id",
				Value: []byte("my other very secret value"),
			},
		}

		created, err := store.CreateGateway(ctx, &ttnpb.Gateway{
			Ids: &ttnpb.GatewayIdentifiers{
				GatewayId: "foo",
				Eui:       eui,
			},
			Name:        "Foo Gateway",
			Description: "The Amazing Foo Gateway",
			Attributes: map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "qux",
			},
			Antennas: []*ttnpb.GatewayAntenna{
				{
					Gain:      3,
					Location:  &ttnpb.Location{Latitude: 12.345, Longitude: 23.456, Altitude: 1090, Accuracy: 1},
					Placement: ttnpb.GatewayAntennaPlacement_OUTDOOR,
				},
			},
			ScheduleAnytimeDelay:          ttnpb.ProtoDurationPtr(scheduleAnytimeDelay),
			UpdateLocationFromStatus:      true,
			LbsLnsSecret:                  secret,
			ClaimAuthenticationCode:       &gtwClaimAuthCode,
			TargetCupsUri:                 targetCUPSURI,
			TargetCupsKey:                 secret,
			DisablePacketBrokerForwarding: true,
		})

		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.GetIds().GetGatewayId(), should.Equal, "foo")
			a.So(created.Name, should.Equal, "Foo Gateway")
			a.So(created.Description, should.Equal, "The Amazing Foo Gateway")
			a.So(created.Attributes, should.HaveLength, 3)
			if a.So(created.Antennas, should.HaveLength, 1) {
				a.So(created.Antennas[0].Gain, should.Equal, 3)
				a.So(created.Antennas[0].Placement, should.Equal, ttnpb.GatewayAntennaPlacement_OUTDOOR)
			}
			a.So(*ttnpb.StdTime(created.CreatedAt), should.HappenAfter, time.Now().Add(-1*time.Hour))
			a.So(*ttnpb.StdTime(created.UpdatedAt), should.HappenAfter, time.Now().Add(-1*time.Hour))
			a.So(*ttnpb.StdDuration(created.ScheduleAnytimeDelay), should.Equal, time.Second)
			a.So(created.UpdateLocationFromStatus, should.BeTrue)
			a.So(created.LbsLnsSecret, should.NotBeNil)
			a.So(created.LbsLnsSecret, should.Resemble, secret)
			a.So(created.ClaimAuthenticationCode, should.NotBeNil)
			a.So(created.ClaimAuthenticationCode.Secret, should.Resemble, gtwClaimAuthCode.Secret)
			a.So(created.TargetCupsUri, should.Equal, targetCUPSURI)
			a.So(created.TargetCupsKey, should.NotBeNil)
			a.So(created.TargetCupsKey, should.Resemble, secret)
			a.So(created.DisablePacketBrokerForwarding, should.BeTrue)
		}

		got, err := store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "foo"}, &pbtypes.FieldMask{Paths: []string{"name", "attributes", "lbs_lns_secret", "claim_authentication_code", "target_cups_uri", "target_cups_key", "disable_packet_broker_forwarding"}})

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.GetIds().GetGatewayId(), should.Equal, "foo")
			a.So(got.Name, should.Equal, "Foo Gateway")
			a.So(got.Description, should.BeEmpty)
			a.So(got.Attributes, should.HaveLength, 3)
			a.So(got.CreatedAt, should.Resemble, created.CreatedAt)
			a.So(got.UpdatedAt, should.Resemble, created.UpdatedAt)
			a.So(got.LbsLnsSecret, should.Resemble, created.LbsLnsSecret)
			a.So(got.ClaimAuthenticationCode.Secret, should.Resemble, created.ClaimAuthenticationCode.Secret)
			a.So(got.TargetCupsUri, should.Equal, created.TargetCupsUri)
			a.So(got.TargetCupsKey, should.Resemble, created.TargetCupsKey)
			a.So(got.DisablePacketBrokerForwarding, should.BeTrue)
		}

		byEUI, err := store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{Eui: &types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}}, &pbtypes.FieldMask{Paths: []string{"name"}})

		a.So(err, should.BeNil)
		if a.So(byEUI, should.NotBeNil) {
			a.So(byEUI.GetIds().GetGatewayId(), should.Equal, got.GetIds().GetGatewayId())
			a.So(byEUI.LbsLnsSecret, should.BeNil)
			a.So(byEUI.ClaimAuthenticationCode, should.BeNil)
			a.So(byEUI.TargetCupsKey, should.BeNil)
		}

		_, err = store.UpdateGateway(ctx, &ttnpb.Gateway{
			Ids: &ttnpb.GatewayIdentifiers{GatewayId: "bar"},
		}, nil)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		updated, err := store.UpdateGateway(ctx, &ttnpb.Gateway{
			Ids:         &ttnpb.GatewayIdentifiers{GatewayId: "foo"},
			Name:        "Foobar Gateway",
			Description: "The Amazing Foobar Gateway",
			Attributes: map[string]string{
				"foo": "bar",
				"baz": "baz",
				"qux": "foo",
			},
			Antennas: []*ttnpb.GatewayAntenna{
				{
					Gain:       6,
					Location:   &ttnpb.Location{Latitude: 12.345, Longitude: 23.456, Altitude: 1090, Accuracy: 1},
					Attributes: map[string]string{"direction": "west"},
					Placement:  ttnpb.GatewayAntennaPlacement_INDOOR,
				},
				{
					Gain:       6,
					Location:   &ttnpb.Location{Latitude: 12.345, Longitude: 23.456, Altitude: 1090, Accuracy: 1},
					Attributes: map[string]string{"direction": "east"},
					Placement:  ttnpb.GatewayAntennaPlacement_OUTDOOR,
				},
			},
			ScheduleAnytimeDelay:          nil,
			UpdateLocationFromStatus:      false,
			LbsLnsSecret:                  otherSecret,
			ClaimAuthenticationCode:       &otherGtwClaimAuthCode,
			TargetCupsUri:                 otherTargetCUPSURI,
			TargetCupsKey:                 otherSecret,
			DisablePacketBrokerForwarding: false,
		}, &pbtypes.FieldMask{Paths: []string{"description", "attributes", "antennas", "schedule_anytime_delay", "update_location_from_status", "lbs_lns_secret", "claim_authentication_code", "target_cups_uri", "target_cups_key", "disable_packet_broker_forwarding"}})

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Description, should.Equal, "The Amazing Foobar Gateway")
			a.So(updated.Attributes, should.HaveLength, 3)
			if a.So(updated.Antennas, should.HaveLength, 2) {
				a.So(updated.Antennas[0].Gain, should.Equal, 6)
				a.So(updated.Antennas[0].Attributes, should.HaveLength, 1)
				a.So(updated.Antennas[0].Placement, should.Equal, ttnpb.GatewayAntennaPlacement_INDOOR)
				a.So(updated.Antennas[1].Gain, should.Equal, 6)
				a.So(updated.Antennas[1].Attributes, should.HaveLength, 1)
				a.So(updated.Antennas[1].Placement, should.Equal, ttnpb.GatewayAntennaPlacement_OUTDOOR)
			}
			a.So(updated.CreatedAt, should.Resemble, created.CreatedAt)
			a.So(*ttnpb.StdTime(updated.UpdatedAt), should.HappenAfter, *ttnpb.StdTime(created.CreatedAt))
			a.So(*ttnpb.StdDuration(updated.ScheduleAnytimeDelay), should.Equal, time.Duration(0))
			a.So(updated.UpdateLocationFromStatus, should.BeFalse)
			a.So(updated.LbsLnsSecret, should.Resemble, otherSecret)
			a.So(updated.ClaimAuthenticationCode.Secret, should.Resemble, otherGtwClaimAuthCode.Secret)
			a.So(updated.TargetCupsKey, should.Resemble, otherSecret)
			a.So(updated.TargetCupsUri, should.Resemble, otherTargetCUPSURI)
			a.So(updated.DisablePacketBrokerForwarding, should.BeFalse)
		}

		got, err = store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "foo"}, nil)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.GetIds().GetGatewayId(), should.Equal, created.GetIds().GetGatewayId())
			a.So(got.Name, should.Equal, created.Name)
			a.So(got.Description, should.Equal, updated.Description)
			a.So(got.Attributes, should.Resemble, updated.Attributes)
			a.So(got.Antennas, should.HaveLength, len(updated.Antennas))
			a.So(got.CreatedAt, should.Resemble, created.CreatedAt)
			a.So(got.UpdatedAt, should.Resemble, updated.UpdatedAt)
			a.So(got.LbsLnsSecret, should.Resemble, otherSecret)
			a.So(got.ClaimAuthenticationCode.Secret, should.Resemble, otherGtwClaimAuthCode.Secret)
			a.So(got.TargetCupsKey, should.Resemble, otherSecret)
		}

		list, err := store.FindGateways(ctx, nil, &pbtypes.FieldMask{Paths: []string{"name"}})

		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 1) {
			a.So(list[0].Name, should.EndWith, got.Name)
		}

		updated, err = store.UpdateGateway(ctx, &ttnpb.Gateway{
			Ids:      &ttnpb.GatewayIdentifiers{GatewayId: "foo"},
			Antennas: []*ttnpb.GatewayAntenna{},
		}, &pbtypes.FieldMask{Paths: []string{"antennas"}})

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Antennas, should.HaveLength, 0)
		}

		_, _ = store.UpdateGateway(ctx, &ttnpb.Gateway{
			Ids: &ttnpb.GatewayIdentifiers{GatewayId: "foo"},
			Antennas: []*ttnpb.GatewayAntenna{
				{
					Gain:       6,
					Location:   &ttnpb.Location{Latitude: 12.345, Longitude: 23.456, Altitude: 1090, Accuracy: 1},
					Attributes: map[string]string{"direction": "west"},
					Placement:  ttnpb.GatewayAntennaPlacement_OUTDOOR,
				},
			},
		}, &pbtypes.FieldMask{Paths: []string{"antennas"}})

		err = store.DeleteGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "foo"})

		a.So(err, should.BeNil)

		err = store.RestoreGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "foo"})

		a.So(err, should.BeNil)

		got, err = store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "foo"}, nil)

		a.So(err, should.BeNil)

		err = store.DeleteGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "foo"})

		a.So(err, should.BeNil)

		got, err = store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "foo"}, nil)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		list, err = store.FindGateways(ctx, nil, nil)

		a.So(err, should.BeNil)
		a.So(list, should.BeEmpty)

		list, err = store.FindGateways(WithSoftDeleted(ctx, false), nil, nil)

		a.So(err, should.BeNil)
		a.So(list, should.NotBeEmpty)

		got, err = store.CreateGateway(ctx, &ttnpb.Gateway{
			Ids: &ttnpb.GatewayIdentifiers{
				GatewayId: "reuse-foo-eui",
				Eui:       eui,
			},
		})

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.GetIds().GetGatewayId(), should.Equal, "reuse-foo-eui")
			a.So(got.GetIds().GetEui(), should.Resemble, eui)
		}

		entity, _ := s.findDeletedEntity(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "foo"}, "id")

		err = store.PurgeGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "foo"})

		a.So(err, should.BeNil)

		var attribute []Attribute
		s.query(ctx, Attribute{}).Where(&Attribute{
			EntityID:   entity.PrimaryKey(),
			EntityType: "gateway",
		}).Find(&attribute)

		var antenna []GatewayAntenna
		s.query(ctx, GatewayAntenna{}).Where(&GatewayAntenna{
			GatewayID: entity.PrimaryKey(),
		}).Find(&antenna)

		a.So(attribute, should.HaveLength, 0)

		err = store.PurgeGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "reuse-foo-eui"})

		a.So(err, should.BeNil)

		// Check that gateway ids are released after purge
		got, err = store.CreateGateway(ctx, &ttnpb.Gateway{
			Ids: &ttnpb.GatewayIdentifiers{
				GatewayId: "foo",
				Eui:       eui,
			},
		})

		a.So(err, should.BeNil)
	})
}
